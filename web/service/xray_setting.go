package service

import (
	_ "embed"
	"encoding/json"

	"github.com/alireza0/x-ui/util/common"
	"github.com/alireza0/x-ui/xray"
)

type XraySettingService struct {
	SettingService
}

func (s *XraySettingService) SaveXraySetting(newXraySettings string) error {
	if err := s.CheckXrayConfig(newXraySettings); err != nil {
		return err
	}
	return s.SettingService.saveSetting("xrayTemplateConfig", newXraySettings)
}

func (s *XraySettingService) CheckXrayConfig(XrayTemplateConfig string) error {
	xrayConfig := &xray.Config{}
	err := json.Unmarshal([]byte(XrayTemplateConfig), xrayConfig)
	if err != nil {
		return common.NewError("xray template config invalid:", err)
	}
	return nil
}

func (s *XraySettingService) ApplyOutboundRouting(outboundTag string) error {
	// Get current xray config
	configStr, err := s.GetXrayConfigTemplate()
	if err != nil {
		return err
	}

	// Parse config
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return common.NewError("Failed to parse xray config:", err)
	}

	// Ensure routing exists
	if config["routing"] == nil {
		config["routing"] = make(map[string]interface{})
	}
	routing := config["routing"].(map[string]interface{})

	// Ensure rules array exists
	if routing["rules"] == nil {
		routing["rules"] = []interface{}{}
	}
	rules := routing["rules"].([]interface{})

	// Create or update the outbound routing rule
	outboundRuleExists := false
	for i, rule := range rules {
		ruleMap := rule.(map[string]interface{})
		// Check if this is an outbound rule (has outboundTag field)
		if _, hasOutbound := ruleMap["outboundTag"]; hasOutbound {
			// Update to use new outbound
			ruleMap["outboundTag"] = outboundTag
			rules[i] = ruleMap
			outboundRuleExists = true
			break
		}
	}

	// If no outbound rule exists, create a new one at the beginning
	if !outboundRuleExists {
		newRule := map[string]interface{}{
			"type":        "field",
			"outboundTag": outboundTag,
			"network":     "tcp,udp",
		}
		// Insert at the beginning
		rules = append([]interface{}{newRule}, rules...)
	}

	routing["rules"] = rules
	config["routing"] = routing

	// Save updated config
	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return common.NewError("Failed to serialize config:", err)
	}

	if err := s.SaveXraySetting(string(configBytes)); err != nil {
		return err
	}

	// Restart xray to apply changes
	xrayService := &XrayService{}
	return xrayService.RestartXray(true)
}
