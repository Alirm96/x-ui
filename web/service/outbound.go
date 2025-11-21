package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/alireza0/x-ui/database"
	"github.com/alireza0/x-ui/database/model"
	"github.com/alireza0/x-ui/logger"

	"gorm.io/gorm"
)

type OutboundService struct {
}

func (s *OutboundService) GetOutbounds(userId int) ([]*model.Outbound, error) {
	db := database.GetDB()
	var outbounds []*model.Outbound
	err := db.Model(model.Outbound{}).Where("user_id = ?", userId).Find(&outbounds).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return outbounds, nil
}

func (s *OutboundService) GetAllOutbounds() ([]*model.Outbound, error) {
	db := database.GetDB()
	var outbounds []*model.Outbound
	err := db.Model(model.Outbound{}).Find(&outbounds).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return outbounds, nil
}

func (s *OutboundService) GetOutbound(id int) (*model.Outbound, error) {
	db := database.GetDB()
	outbound := &model.Outbound{}
	err := db.Model(model.Outbound{}).First(outbound, id).Error
	if err != nil {
		return nil, err
	}
	return outbound, nil
}

func (s *OutboundService) GetOutboundByTag(tag string) (*model.Outbound, error) {
	db := database.GetDB()
	outbound := &model.Outbound{}
	err := db.Model(model.Outbound{}).Where("tag = ?", tag).First(outbound).Error
	if err != nil {
		return nil, err
	}
	return outbound, nil
}

func (s *OutboundService) checkTagExist(tag string, ignoreId int) (bool, error) {
	db := database.GetDB()
	db = db.Model(model.Outbound{}).Where("tag = ?", tag)
	if ignoreId > 0 {
		db = db.Where("id != ?", ignoreId)
	}
	var count int64
	err := db.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *OutboundService) AddOutbound(outbound *model.Outbound) (*model.Outbound, error) {
	fmt.Println("=== AddOutbound START ===")
	fmt.Println("Received outbound with GroupName:", outbound.GroupName)
	fmt.Println("Outbound Tag:", outbound.Tag)
	logger.Info("=== AddOutbound START ===")
	logger.Info("Received outbound with GroupName:", outbound.GroupName)
	logger.Info("Outbound Tag:", outbound.Tag)
	
	// Validate VLESS outbound configuration
	if outbound.Protocol == model.VLESS {
		if err := s.ValidateVLESSOutbound(outbound.Settings); err != nil {
			return nil, fmt.Errorf("VLESS validation failed: %w", err)
		}
	}
	
	db := database.GetDB()
	
	exist, err := s.checkTagExist(outbound.Tag, 0)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, fmt.Errorf("tag already exists")
	}

	now := time.Now().Unix()
	outbound.CreatedAt = now
	outbound.UpdatedAt = now
	outbound.LastTestResult = -1 // not tested yet
	outbound.TestFailCount = 0

	fmt.Println("Before Save - GroupName:", outbound.GroupName)
	logger.Info("Before Save - GroupName:", outbound.GroupName)
	err = db.Save(outbound).Error
	if err != nil {
		fmt.Println("Save failed:", err)
		logger.Error("Save failed:", err)
		return nil, err
	}
	fmt.Println("After Save - Outbound ID:", outbound.Id, "GroupName:", outbound.GroupName)
	logger.Info("After Save - Outbound ID:", outbound.Id, "GroupName:", outbound.GroupName)
	
	// Re-read from database to verify what was actually saved
	var savedOutbound model.Outbound
	err = db.First(&savedOutbound, outbound.Id).Error
	if err != nil {
		fmt.Println("Failed to re-read outbound:", err)
		logger.Error("Failed to re-read outbound:", err)
	} else {
		fmt.Println("Re-read from DB - ID:", savedOutbound.Id, "GroupName:", savedOutbound.GroupName)
		logger.Info("Re-read from DB - ID:", savedOutbound.Id, "GroupName:", savedOutbound.GroupName)
	}
	
	fmt.Println("=== AddOutbound END ===")
	logger.Info("=== AddOutbound END ===")
	return outbound, nil
}

func (s *OutboundService) UpdateOutbound(outbound *model.Outbound) (*model.Outbound, error) {
	// Validate VLESS outbound configuration
	if outbound.Protocol == model.VLESS {
		if err := s.ValidateVLESSOutbound(outbound.Settings); err != nil {
			return nil, fmt.Errorf("VLESS validation failed: %w", err)
		}
	}
	
	db := database.GetDB()

	exist, err := s.checkTagExist(outbound.Tag, outbound.Id)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, fmt.Errorf("tag already exists")
	}

	outbound.UpdatedAt = time.Now().Unix()
	
	err = db.Save(outbound).Error
	if err != nil {
		return nil, err
	}
	return outbound, nil
}

func (s *OutboundService) DelOutbound(id int) error {
	db := database.GetDB()
	
	// Delete related test history
	err := db.Where("outbound_id = ?", id).Delete(&model.OutboundTestHistory{}).Error
	if err != nil {
		return err
	}
	
	return db.Delete(&model.Outbound{}, id).Error
}

func (s *OutboundService) ImportOutboundFromLink(link string, userId int, groupName string) (*model.Outbound, error) {
	fmt.Println("=== ImportOutboundFromLink START ===")
	fmt.Println("Received groupName parameter:", groupName)
	logger.Info("=== ImportOutboundFromLink START ===")
	logger.Info("Received groupName parameter:", groupName)
	
	link = strings.TrimSpace(link)
	
	var outbound *model.Outbound
	var err error
	
	if strings.HasPrefix(link, "vmess://") {
		outbound, err = s.parseVmessLink(link)
	} else if strings.HasPrefix(link, "vless://") {
		outbound, err = s.parseVlessLink(link)
	} else if strings.HasPrefix(link, "trojan://") {
		outbound, err = s.parseTrojanLink(link)
	} else if strings.HasPrefix(link, "ss://") {
		outbound, err = s.parseShadowsocksLink(link)
	} else {
		return nil, fmt.Errorf("unsupported link format")
	}
	
	if err != nil {
		logger.Error("Failed to parse link:", err)
		return nil, err
	}
	
	logger.Info("Link parsed successfully, protocol:", outbound.Protocol)
	
	outbound.UserId = userId
	outbound.Enable = true
	
	// Set group name
	if groupName == "" {
		groupName = "default"
		fmt.Println("GroupName was empty, setting to default")
		logger.Info("GroupName was empty, setting to default")
	}
	outbound.GroupName = groupName
	fmt.Println("Set outbound.GroupName to:", outbound.GroupName)
	logger.Info("Set outbound.GroupName to:", outbound.GroupName)
	
	// Ensure unique tag
	originalTag := outbound.Tag
	counter := 1
	for {
		exists, err := s.checkTagExist(outbound.Tag, 0)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
		outbound.Tag = fmt.Sprintf("%s-%d", originalTag, counter)
		counter++
	}
	
	fmt.Println("About to call AddOutbound with GroupName:", outbound.GroupName)
	logger.Info("About to call AddOutbound with GroupName:", outbound.GroupName)
	result, err := s.AddOutbound(outbound)
	if err != nil {
		fmt.Println("AddOutbound failed:", err)
		logger.Error("AddOutbound failed:", err)
		return nil, err
	}
	
	fmt.Println("AddOutbound returned outbound with ID:", result.Id, "GroupName:", result.GroupName)
	fmt.Println("=== ImportOutboundFromLink END ===")
	logger.Info("AddOutbound returned outbound with ID:", result.Id, "GroupName:", result.GroupName)
	logger.Info("=== ImportOutboundFromLink END ===")
	return result, nil
}

func (s *OutboundService) parseVmessLink(link string) (*model.Outbound, error) {
	// vmess://base64encoded
	encoded := strings.TrimPrefix(link, "vmess://")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode vmess link")
		}
	}
	
	var vmessData map[string]interface{}
	err = json.Unmarshal(decoded, &vmessData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vmess data")
	}
	
	outbound := &model.Outbound{
		Protocol: model.VMess,
		Address:  getString(vmessData, "add"),
		Port:     getInt(vmessData, "port"),
		Remark:   getString(vmessData, "ps"),
		Tag:      fmt.Sprintf("out-%s", getString(vmessData, "ps")),
	}
	
	if outbound.Tag == "out-" {
		outbound.Tag = fmt.Sprintf("out-vmess-%s-%d", outbound.Address, outbound.Port)
	}
	
	// Build settings
	settings := map[string]interface{}{
		"vnext": []map[string]interface{}{
			{
				"address": outbound.Address,
				"port":    outbound.Port,
				"users": []map[string]interface{}{
					{
						"id":       getString(vmessData, "id"),
						"alterId":  getInt(vmessData, "aid"),
						"security": getString(vmessData, "scy"),
					},
				},
			},
		},
	}
	
	settingsJSON, _ := json.Marshal(settings)
	outbound.Settings = string(settingsJSON)
	
	// Build stream settings
	streamSettings := map[string]interface{}{
		"network":  getString(vmessData, "net"),
		"security": getString(vmessData, "tls"),
	}
	
	if net := getString(vmessData, "net"); net == "ws" {
		streamSettings["wsSettings"] = map[string]interface{}{
			"path": getString(vmessData, "path"),
			"headers": map[string]interface{}{
				"Host": getString(vmessData, "host"),
			},
		}
	} else if net == "tcp" {
		if headerType := getString(vmessData, "type"); headerType == "http" {
			streamSettings["tcpSettings"] = map[string]interface{}{
				"header": map[string]interface{}{
					"type": "http",
					"request": map[string]interface{}{
						"path": []string{getString(vmessData, "path")},
						"headers": map[string]interface{}{
							"Host": []string{getString(vmessData, "host")},
						},
					},
				},
			}
		}
	}
	
	if getString(vmessData, "tls") == "tls" {
		streamSettings["tlsSettings"] = map[string]interface{}{
			"serverName": getString(vmessData, "sni"),
		}
	}
	
	streamJSON, _ := json.Marshal(streamSettings)
	outbound.StreamSettings = string(streamJSON)
	
	return outbound, nil
}

func (s *OutboundService) parseVlessLink(link string) (*model.Outbound, error) {
	// vless://uuid@address:port?params#remark
	link = strings.TrimPrefix(link, "vless://")
	
	parts := strings.Split(link, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid vless link format")
	}
	
	uuid := parts[0]
	remaining := parts[1]
	
	// Split address:port and query params
	var addressPort, queryString, remark string
	if idx := strings.Index(remaining, "?"); idx >= 0 {
		addressPort = remaining[:idx]
		remaining = remaining[idx+1:]
		if idx := strings.Index(remaining, "#"); idx >= 0 {
			queryString = remaining[:idx]
			remark, _ = url.QueryUnescape(remaining[idx+1:])
		} else {
			queryString = remaining
		}
	} else if idx := strings.Index(remaining, "#"); idx >= 0 {
		addressPort = remaining[:idx]
		remark, _ = url.QueryUnescape(remaining[idx+1:])
	} else {
		addressPort = remaining
	}
	
	// Parse address and port
	addrParts := strings.Split(addressPort, ":")
	if len(addrParts) != 2 {
		return nil, fmt.Errorf("invalid address:port format")
	}
	
	address := addrParts[0]
	port := 0
	fmt.Sscanf(addrParts[1], "%d", &port)
	
	// Parse query parameters
	params, _ := url.ParseQuery(queryString)
	
	outbound := &model.Outbound{
		Protocol: model.VLESS,
		Address:  address,
		Port:     port,
		Remark:   remark,
		Tag:      fmt.Sprintf("out-%s", remark),
	}
	
	if outbound.Tag == "out-" {
		outbound.Tag = fmt.Sprintf("out-vless-%s-%d", address, port)
	}
	
	// Build settings
	settings := map[string]interface{}{
		"vnext": []map[string]interface{}{
			{
				"address": address,
				"port":    port,
				"users": []map[string]interface{}{
					{
						"id":         uuid,
						"encryption": params.Get("encryption"),
						"flow":       params.Get("flow"),
					},
				},
			},
		},
	}
	
	settingsJSON, _ := json.Marshal(settings)
	outbound.Settings = string(settingsJSON)
	
	// Build stream settings
	streamSettings := map[string]interface{}{
		"network":  params.Get("type"),
		"security": params.Get("security"),
	}
	
	if params.Get("type") == "ws" {
		streamSettings["wsSettings"] = map[string]interface{}{
			"path": params.Get("path"),
			"headers": map[string]interface{}{
				"Host": params.Get("host"),
			},
		}
	} else if params.Get("type") == "grpc" {
		streamSettings["grpcSettings"] = map[string]interface{}{
			"serviceName": params.Get("serviceName"),
		}
	}
	
	if params.Get("security") == "tls" {
		streamSettings["tlsSettings"] = map[string]interface{}{
			"serverName": params.Get("sni"),
			"alpn":       strings.Split(params.Get("alpn"), ","),
		}
	} else if params.Get("security") == "reality" {
		streamSettings["realitySettings"] = map[string]interface{}{
			"serverName": params.Get("sni"),
			"publicKey":  params.Get("pbk"),
			"shortId":    params.Get("sid"),
			"fingerprint": params.Get("fp"),
		}
	}
	
	streamJSON, _ := json.Marshal(streamSettings)
	outbound.StreamSettings = string(streamJSON)
	
	return outbound, nil
}

func (s *OutboundService) parseTrojanLink(link string) (*model.Outbound, error) {
	// trojan://password@address:port?params#remark
	link = strings.TrimPrefix(link, "trojan://")
	
	parts := strings.Split(link, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid trojan link format")
	}
	
	password := parts[0]
	remaining := parts[1]
	
	var addressPort, queryString, remark string
	if idx := strings.Index(remaining, "?"); idx >= 0 {
		addressPort = remaining[:idx]
		remaining = remaining[idx+1:]
		if idx := strings.Index(remaining, "#"); idx >= 0 {
			queryString = remaining[:idx]
			remark, _ = url.QueryUnescape(remaining[idx+1:])
		} else {
			queryString = remaining
		}
	} else if idx := strings.Index(remaining, "#"); idx >= 0 {
		addressPort = remaining[:idx]
		remark, _ = url.QueryUnescape(remaining[idx+1:])
	} else {
		addressPort = remaining
	}
	
	addrParts := strings.Split(addressPort, ":")
	if len(addrParts) != 2 {
		return nil, fmt.Errorf("invalid address:port format")
	}
	
	address := addrParts[0]
	port := 0
	fmt.Sscanf(addrParts[1], "%d", &port)
	
	params, _ := url.ParseQuery(queryString)
	
	outbound := &model.Outbound{
		Protocol: model.Trojan,
		Address:  address,
		Port:     port,
		Remark:   remark,
		Tag:      fmt.Sprintf("out-%s", remark),
	}
	
	if outbound.Tag == "out-" {
		outbound.Tag = fmt.Sprintf("out-trojan-%s-%d", address, port)
	}
	
	settings := map[string]interface{}{
		"servers": []map[string]interface{}{
			{
				"address":  address,
				"port":     port,
				"password": password,
			},
		},
	}
	
	settingsJSON, _ := json.Marshal(settings)
	outbound.Settings = string(settingsJSON)
	
	streamSettings := map[string]interface{}{
		"network":  params.Get("type"),
		"security": params.Get("security"),
	}
	
	if params.Get("type") == "ws" {
		streamSettings["wsSettings"] = map[string]interface{}{
			"path": params.Get("path"),
			"headers": map[string]interface{}{
				"Host": params.Get("host"),
			},
		}
	}
	
	if params.Get("security") == "tls" {
		streamSettings["tlsSettings"] = map[string]interface{}{
			"serverName": params.Get("sni"),
		}
	}
	
	streamJSON, _ := json.Marshal(streamSettings)
	outbound.StreamSettings = string(streamJSON)
	
	return outbound, nil
}

func (s *OutboundService) parseShadowsocksLink(link string) (*model.Outbound, error) {
	// ss://base64encoded#remark or ss://method:password@address:port#remark
	link = strings.TrimPrefix(link, "ss://")
	
	var remark string
	if idx := strings.Index(link, "#"); idx >= 0 {
		remark, _ = url.QueryUnescape(link[idx+1:])
		link = link[:idx]
	}
	
	var method, password, address string
	var port int
	
	// Try base64 decode first
	decoded, err := base64.URLEncoding.DecodeString(link)
	if err != nil {
		decoded, err = base64.StdEncoding.DecodeString(link)
	}
	
	if err == nil {
		// Successfully decoded
		parts := strings.Split(string(decoded), "@")
		if len(parts) == 2 {
			methodPass := strings.Split(parts[0], ":")
			if len(methodPass) == 2 {
				method = methodPass[0]
				password = methodPass[1]
			}
			
			addrParts := strings.Split(parts[1], ":")
			if len(addrParts) == 2 {
				address = addrParts[0]
				fmt.Sscanf(addrParts[1], "%d", &port)
			}
		}
	} else {
		// Not base64, try direct parsing
		parts := strings.Split(link, "@")
		if len(parts) == 2 {
			methodPass := strings.Split(parts[0], ":")
			if len(methodPass) == 2 {
				method = methodPass[0]
				password = methodPass[1]
			}
			
			addrParts := strings.Split(parts[1], ":")
			if len(addrParts) == 2 {
				address = addrParts[0]
				fmt.Sscanf(addrParts[1], "%d", &port)
			}
		}
	}
	
	outbound := &model.Outbound{
		Protocol: model.Shadowsocks,
		Address:  address,
		Port:     port,
		Remark:   remark,
		Tag:      fmt.Sprintf("out-%s", remark),
	}
	
	if outbound.Tag == "out-" {
		outbound.Tag = fmt.Sprintf("out-ss-%s-%d", address, port)
	}
	
	settings := map[string]interface{}{
		"servers": []map[string]interface{}{
			{
				"address":  address,
				"port":     port,
				"method":   method,
				"password": password,
			},
		},
	}
	
	settingsJSON, _ := json.Marshal(settings)
	outbound.Settings = string(settingsJSON)
	outbound.StreamSettings = "{}"
	
	return outbound, nil
}

func (s *OutboundService) TestOutbound(id int, testURL string, timeout int) (int, error) {
	outbound, err := s.GetOutbound(id)
	if err != nil {
		return -1, err
	}
	
	// For now, we'll do a simple HTTP test
	// In production, you would configure xray to use this outbound and test through it
	latency, err := s.measureLatency(outbound, testURL, timeout)
	
	// Record test result
	db := database.GetDB()
	history := &model.OutboundTestHistory{
		OutboundId:   id,
		TestTime:     time.Now().Unix(),
		Success:      err == nil,
		LatencyMs:    latency,
		ErrorMessage: "",
	}
	
	if err != nil {
		history.ErrorMessage = err.Error()
		outbound.TestFailCount++
	} else {
		outbound.TestFailCount = 0
	}
	
	outbound.LastTestTime = history.TestTime
	outbound.LastTestResult = latency
	
	db.Save(history)
	db.Save(outbound)
	
	return latency, err
}

func (s *OutboundService) measureLatency(outbound *model.Outbound, testURL string, timeout int) (int, error) {
	// Test the outbound by creating a temporary xray instance with SOCKS5 proxy
	// and routing traffic through the outbound
	
	// Generate a temporary port for SOCKS5
	socksPort := 10808 + outbound.Id // Use different port per outbound to avoid conflicts
	
	// Create temporary xray config
	xrayConfig := fmt.Sprintf(`{
		"log": {
			"loglevel": "error"
		},
		"inbounds": [{
			"port": %d,
			"protocol": "socks",
			"settings": {
				"auth": "noauth",
				"udp": false
			},
			"tag": "socks-in"
		}],
		"outbounds": [{
			"protocol": "%s",
			"settings": %s,
			"streamSettings": %s,
			"tag": "test-out"
		}],
		"routing": {
			"rules": [{
				"type": "field",
				"inboundTag": ["socks-in"],
				"outboundTag": "test-out"
			}]
		}
	}`, socksPort, outbound.Protocol, outbound.Settings, outbound.StreamSettings)
	
	// Write config to temporary file
	tmpDir := "/tmp"
	configFile := fmt.Sprintf("%s/xray-test-%d.json", tmpDir, outbound.Id)
	err := os.WriteFile(configFile, []byte(xrayConfig), 0600)
	if err != nil {
		return -1, fmt.Errorf("failed to write config: %v", err)
	}
	defer os.Remove(configFile)
	
	// Start temporary xray process
	xrayBin := "/app/bin/xray-linux-amd64"
	cmd := exec.Command(xrayBin, "run", "-c", configFile)
	err = cmd.Start()
	if err != nil {
		return -1, fmt.Errorf("failed to start xray: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()
	
	// Wait a bit for xray to start
	time.Sleep(500 * time.Millisecond)
	
	// Test through the SOCKS5 proxy
	var bestLatency int = -1
	var lastErr error
	
	proxyURL, err := url.Parse(fmt.Sprintf("socks5://127.0.0.1:%d", socksPort))
	if err != nil {
		return -1, fmt.Errorf("invalid proxy URL: %v", err)
	}
	
	for i := 0; i < 2; i++ {
		start := time.Now()
		
		client := &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			Transport: &http.Transport{
				Proxy:             http.ProxyURL(proxyURL),
				DisableKeepAlives: true,
			},
		}
		
		resp, err := client.Get(testURL)
		if err != nil {
			lastErr = err
			time.Sleep(100 * time.Millisecond)
			continue
		}
		defer resp.Body.Close()
		
		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("HTTP error: %d", resp.StatusCode)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		
		latency := int(time.Since(start).Milliseconds())
		if bestLatency == -1 || latency < bestLatency {
			bestLatency = latency
		}
		
		time.Sleep(100 * time.Millisecond)
	}
	
	if bestLatency == -1 {
		if lastErr != nil {
			return -1, lastErr
		}
		return -1, fmt.Errorf("all test attempts failed")
	}
	
	return bestLatency, nil
}

func (s *OutboundService) GetTestHistory(outboundId int, limit int) ([]*model.OutboundTestHistory, error) {
	db := database.GetDB()
	var history []*model.OutboundTestHistory
	
	query := db.Model(model.OutboundTestHistory{}).Where("outbound_id = ?", outboundId).Order("test_time DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&history).Error
	if err != nil {
		return nil, err
	}
	
	return history, nil
}

func (s *OutboundService) CleanupFailedOutbounds(daysThreshold int) error {
	db := database.GetDB()
	
	thresholdTime := time.Now().AddDate(0, 0, -daysThreshold).Unix()
	
	// Find outbounds that have been failing for more than threshold days
	var failedOutbounds []*model.Outbound
	err := db.Model(model.Outbound{}).
		Where("last_test_time > 0").
		Where("last_test_time < ?", thresholdTime).
		Where("last_test_result = -1").
		Find(&failedOutbounds).Error
	
	if err != nil {
		return err
	}
	
	logger.Infof("Found %d outbounds to cleanup", len(failedOutbounds))
	
	for _, outbound := range failedOutbounds {
		logger.Infof("Deleting failed outbound: %s (last test: %d)", outbound.Remark, outbound.LastTestTime)
		err = s.DelOutbound(outbound.Id)
		if err != nil {
			logger.Warning("Failed to delete outbound %d: %v", outbound.Id, err)
		}
	}
	
	return nil
}

func (s *OutboundService) GetBestOutbound() (*model.Outbound, error) {
	db := database.GetDB()
	var outbound model.Outbound
	
	// Get enabled outbound with best (lowest) latency
	// Exclude system outbounds (direct/freedom and block/blackhole protocols)
	err := db.Model(model.Outbound{}).
		Where("enable = ?", true).
		Where("last_test_result > 0").
		Where("protocol NOT IN ?", []string{"freedom", "blackhole"}).
		Order("last_test_result ASC").
		First(&outbound).Error
	
	if err != nil {
		return nil, err
	}
	
	return &outbound, nil
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			var i int
			fmt.Sscanf(v, "%d", &i)
			return i
		}
	}
	return 0
}
