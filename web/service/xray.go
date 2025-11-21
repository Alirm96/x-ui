package service

import (
	"encoding/json"
	"errors"
	"runtime"
	"sync"

	"github.com/alireza0/x-ui/database/model"
	"github.com/alireza0/x-ui/logger"
	"github.com/alireza0/x-ui/util/json_util"
	"github.com/alireza0/x-ui/xray"

	"go.uber.org/atomic"
)

var (
	p                 *xray.Process
	lock              sync.Mutex
	isNeedXrayRestart atomic.Bool
	result            string
)

type XrayService struct {
	inboundService InboundService
	settingService SettingService
	xrayAPI        xray.XrayAPI
}

func (s *XrayService) IsXrayRunning() bool {
	return p != nil && p.IsRunning()
}

func (s *XrayService) GetXrayErr() error {
	if p == nil {
		return nil
	}

	err := p.GetErr()

	if runtime.GOOS == "windows" && err.Error() == "exit status 1" {
		// exit status 1 on Windows means that Xray process was killed
		// as we kill process to stop in on Windows, this is not an error
		return nil
	}

	return err
}

func (s *XrayService) GetXrayResult() string {
	if result != "" {
		return result
	}
	if s.IsXrayRunning() {
		return ""
	}
	if p == nil {
		return ""
	}

	result = p.GetResult()

	if runtime.GOOS == "windows" && result == "exit status 1" {
		// exit status 1 on Windows means that Xray process was killed
		// as we kill process to stop in on Windows, this is not an error
		return ""
	}

	return result
}

func (s *XrayService) GetXrayVersion() string {
	if p == nil {
		return "Unknown"
	}
	return p.GetVersion()
}

func RemoveIndex(s []interface{}, index int) []interface{} {
	return append(s[:index], s[index+1:]...)
}

func (s *XrayService) GetXrayConfig() (*xray.Config, error) {
	templateConfig, err := s.settingService.GetXrayConfigTemplate()
	if err != nil {
		return nil, err
	}

	xrayConfig := &xray.Config{}
	err = json.Unmarshal([]byte(templateConfig), xrayConfig)
	if err != nil {
		return nil, err
	}

	s.inboundService.AddTraffic(nil, nil)

	inbounds, err := s.inboundService.GetAllInbounds()
	if err != nil {
		return nil, err
	}
	for _, inbound := range inbounds {
		if !inbound.Enable {
			continue
		}
		// get settings clients
		settings := map[string]interface{}{}
		json.Unmarshal([]byte(inbound.Settings), &settings)
		clients, ok := settings["clients"].([]interface{})
		if ok {
			// check users active or not
			clientStats := inbound.ClientStats
			for _, clientTraffic := range clientStats {
				indexDecrease := 0
				for index, client := range clients {
					c := client.(map[string]interface{})
					if c["email"] == clientTraffic.Email {
						if !clientTraffic.Enable {
							clients = RemoveIndex(clients, index-indexDecrease)
							indexDecrease++
						}
					}
				}
			}

			// clear client config for additional parameters
			var final_clients []interface{}
			for _, client := range clients {

				c := client.(map[string]interface{})

				if c["enable"] != nil {
					if enable, ok := c["enable"].(bool); ok && !enable {
						continue
					}
				}
				for key := range c {
					if key != "email" && key != "id" && key != "password" && key != "flow" && key != "method" {
						delete(c, key)
					}
					if c["flow"] == "xtls-rprx-vision-udp443" {
						c["flow"] = "xtls-rprx-vision"
					}
				}
				final_clients = append(final_clients, interface{}(c))
			}

			settings["clients"] = final_clients
			modifiedSettings, err := json.MarshalIndent(settings, "", "  ")
			if err != nil {
				return nil, err
			}

			inbound.Settings = string(modifiedSettings)
		}

		if len(inbound.StreamSettings) > 0 {
			// Unmarshal stream JSON
			var stream map[string]interface{}
			json.Unmarshal([]byte(inbound.StreamSettings), &stream)

			// Remove the "settings" field under "tlsSettings" and "realitySettings"
			tlsSettings, ok1 := stream["tlsSettings"].(map[string]interface{})
			realitySettings, ok2 := stream["realitySettings"].(map[string]interface{})
			if ok1 || ok2 {
				if ok1 {
					delete(tlsSettings, "settings")
				} else if ok2 {
					delete(realitySettings, "settings")
				}
			}

			delete(stream, "externalProxy")

			newStream, err := json.MarshalIndent(stream, "", "  ")
			if err != nil {
				return nil, err
			}
			inbound.StreamSettings = string(newStream)
		}

		inboundConfig := inbound.GenXrayInboundConfig()
		xrayConfig.InboundConfigs = append(xrayConfig.InboundConfigs, *inboundConfig)
	}

	// Load outbounds from database (replaces template outbounds)
	outboundService := OutboundService{}
	dbOutbounds, err := outboundService.GetAllOutbounds()
	if err != nil {
		return nil, err
	}

	// Convert database outbounds to xray config format
	var xrayOutbounds []map[string]interface{}
	for _, outbound := range dbOutbounds {
		if !outbound.Enable {
			continue
		}

		xrayOutbound := map[string]interface{}{
			"protocol": string(outbound.Protocol),
			"tag":      outbound.Tag,
		}

		// Parse settings JSON
		if outbound.Settings != "" && outbound.Settings != "{}" {
			var settings map[string]interface{}
			if err := json.Unmarshal([]byte(outbound.Settings), &settings); err == nil {
				xrayOutbound["settings"] = settings
			}
		}

		// Parse stream settings JSON
		if outbound.StreamSettings != "" && outbound.StreamSettings != "{}" {
			var streamSettings map[string]interface{}
			if err := json.Unmarshal([]byte(outbound.StreamSettings), &streamSettings); err == nil {
				xrayOutbound["streamSettings"] = streamSettings
			}
		}

		// Add server info for proxy protocols
		if outbound.Address != "" && outbound.Port > 0 {
			if outbound.Protocol == model.VMess || outbound.Protocol == model.VLESS {
				// VMess/VLESS use vnext
				if xrayOutbound["settings"] == nil {
					xrayOutbound["settings"] = map[string]interface{}{}
				}
				settings := xrayOutbound["settings"].(map[string]interface{})
				
				// Check if vnext already exists in settings
				existingVnext, hasVnext := settings["vnext"].([]interface{})
				if hasVnext && len(existingVnext) > 0 {
					// Vnext exists, update address/port if needed but preserve users
					for i := range existingVnext {
						if vnextMap, ok := existingVnext[i].(map[string]interface{}); ok {
							vnextMap["address"] = outbound.Address
							vnextMap["port"] = outbound.Port
						}
					}
				} else {
					// No vnext exists, create basic structure
					settings["vnext"] = []map[string]interface{}{
						{
							"address": outbound.Address,
							"port":    outbound.Port,
						},
					}
				}
			} else if outbound.Protocol == model.Trojan || outbound.Protocol == model.Shadowsocks {
				// Trojan/Shadowsocks use servers
				if xrayOutbound["settings"] == nil {
					xrayOutbound["settings"] = map[string]interface{}{}
				}
				settings := xrayOutbound["settings"].(map[string]interface{})
				settings["servers"] = []map[string]interface{}{
					{
						"address": outbound.Address,
						"port":    outbound.Port,
					},
				}
			}
		}

		xrayOutbounds = append(xrayOutbounds, xrayOutbound)
	}

	// Marshal outbounds to JSON
	outboundsJSON, err := json.Marshal(xrayOutbounds)
	if err != nil {
		return nil, err
	}
	xrayConfig.OutboundConfigs = json_util.RawMessage(outboundsJSON)

	return xrayConfig, nil
}

func (s *XrayService) GetXrayTraffic() ([]*xray.Traffic, []*xray.ClientTraffic, error) {
	if !s.IsXrayRunning() {
		return nil, nil, errors.New("xray is not running")
	}
	s.xrayAPI.Init(p.GetAPIPort())
	defer s.xrayAPI.Close()
	return s.xrayAPI.GetTraffic(true)
}

func (s *XrayService) RestartXray(isForce bool) error {
	lock.Lock()
	defer lock.Unlock()
	logger.Debug("restart xray, force:", isForce)

	xrayConfig, err := s.GetXrayConfig()
	if err != nil {
		return err
	}

	if p != nil && p.IsRunning() {
		if !isForce && p.GetConfig().Equals(xrayConfig) {
			logger.Debug("It does not need to restart xray")
			return nil
		}
		p.Stop()
	}

	p = xray.NewProcess(xrayConfig)
	result = ""
	err = p.Start()
	if err != nil {
		return err
	}
	return nil
}

func (s *XrayService) StopXray() error {
	lock.Lock()
	defer lock.Unlock()
	logger.Debug("stop xray")
	if s.IsXrayRunning() {
		return p.Stop()
	}
	return errors.New("xray is not running")
}

func (s *XrayService) SetToNeedRestart() {
	isNeedXrayRestart.Store(true)
}

func (s *XrayService) IsNeedRestartAndSetFalse() bool {
	return isNeedXrayRestart.CompareAndSwap(true, false)
}
