package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alireza0/x-ui/database"
	"github.com/alireza0/x-ui/database/model"
	"github.com/alireza0/x-ui/logger"
)

type SubscriptionService struct {
	isRunning bool
	stopChan  chan bool
}

// StartAutoUpdate starts the auto-update scheduler
func (s *SubscriptionService) StartAutoUpdate() {
	if s.isRunning {
		return
	}

	s.isRunning = true
	s.stopChan = make(chan bool)

	go s.autoUpdateLoop()
	logger.Info("Subscription auto-update scheduler started")
}

// StopAutoUpdate stops the auto-update scheduler
func (s *SubscriptionService) StopAutoUpdate() {
	if !s.isRunning {
		return
	}

	s.isRunning = false
	s.stopChan <- true
	logger.Info("Subscription auto-update scheduler stopped")
}

// autoUpdateLoop runs the auto-update check every hour
func (s *SubscriptionService) autoUpdateLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Run once on startup
	s.checkAndUpdateSubscriptions()

	for {
		select {
		case <-ticker.C:
			s.checkAndUpdateSubscriptions()
		case <-s.stopChan:
			return
		}
	}
}

// checkAndUpdateSubscriptions checks and updates subscriptions that need updating
func (s *SubscriptionService) checkAndUpdateSubscriptions() {
	subscriptions, err := s.GetSubscriptions()
	if err != nil {
		logger.Warning("Failed to get subscriptions for auto-update: %v", err)
		return
	}

	now := time.Now().Unix()
	
	for _, subscription := range subscriptions {
		// Skip disabled subscriptions
		if !subscription.Enabled {
			continue
		}

		// Skip subscriptions with manual-only update (interval = 0)
		if subscription.UpdateInterval <= 0 {
			continue
		}

		// Check if it's time to update
		intervalSeconds := int64(subscription.UpdateInterval) * 3600 // hours to seconds
		
		// If never updated or interval has passed
		if subscription.LastUpdate == 0 || (now-subscription.LastUpdate) >= intervalSeconds {
			logger.Info("Auto-updating subscription: %s", subscription.Remark)
			
			count, err := s.UpdateSubscriptionContent(subscription.Id)
			if err != nil {
				logger.Warning("Failed to auto-update subscription %s: %v", subscription.Remark, err)
			} else {
				logger.Info("Successfully auto-updated subscription %s, added %d outbound(s)", subscription.Remark, count)
			}
		}
	}
}

// GetSubscriptions returns all subscriptions
func (s *SubscriptionService) GetSubscriptions() ([]*model.Subscription, error) {
	db := database.GetDB()
	var subscriptions []*model.Subscription
	err := db.Model(&model.Subscription{}).Order("created_at desc").Find(&subscriptions).Error
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}

// GetSubscriptionById returns subscription by ID
func (s *SubscriptionService) GetSubscriptionById(id int) (*model.Subscription, error) {
	db := database.GetDB()
	subscription := &model.Subscription{}
	err := db.Model(&model.Subscription{}).Where("id = ?", id).First(subscription).Error
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

// GetSubscriptionsByGroup returns subscriptions in a specific group
func (s *SubscriptionService) GetSubscriptionsByGroup(groupName string) ([]*model.Subscription, error) {
	db := database.GetDB()
	var subscriptions []*model.Subscription
	err := db.Model(&model.Subscription{}).Where("group_name = ?", groupName).Find(&subscriptions).Error
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}

// AddSubscription creates a new subscription
func (s *SubscriptionService) AddSubscription(subscription *model.Subscription) error {
	db := database.GetDB()
	subscription.CreatedAt = time.Now().Unix()
	subscription.UpdatedAt = time.Now().Unix()
	return db.Create(subscription).Error
}

// UpdateSubscription updates an existing subscription
func (s *SubscriptionService) UpdateSubscription(subscription *model.Subscription) error {
	db := database.GetDB()
	subscription.UpdatedAt = time.Now().Unix()
	return db.Save(subscription).Error
}

// DeleteSubscription deletes a subscription by ID
func (s *SubscriptionService) DeleteSubscription(id int) error {
	db := database.GetDB()
	return db.Delete(&model.Subscription{}, id).Error
}

// UpdateSubscriptionContent downloads and processes subscription content
func (s *SubscriptionService) UpdateSubscriptionContent(id int) (int, error) {
	subscription, err := s.GetSubscriptionById(id)
	if err != nil {
		return 0, err
	}

	if !subscription.Enabled {
		return 0, fmt.Errorf("subscription is disabled")
	}

	if subscription.URL == "" {
		return 0, fmt.Errorf("subscription URL is empty")
	}

	// Download subscription content
	content, err := s.downloadSubscription(subscription.URL, subscription.UserAgent)
	if err != nil {
		return 0, fmt.Errorf("failed to download subscription: %w", err)
	}

	// Try to decode if it's base64
	decoded, err := s.decodeSubscription(content)
	if err == nil && decoded != "" {
		content = decoded
	}

	// Parse share links from content
	links := s.parseShareLinks(content)
	if len(links) == 0 {
		return 0, fmt.Errorf("no valid share links found in subscription")
	}

	logger.Info("Found %d share links in subscription %s", len(links), subscription.Remark)

	// Delete existing outbounds from this subscription
	err = s.deleteSubscriptionOutbounds(id)
	if err != nil {
		logger.Warning("Failed to delete old subscription outbounds: %v", err)
	}

	// Add new outbounds
	outboundService := &OutboundService{}
	addedCount := 0
	
	for _, link := range links {
		outbound, err := s.parseShareLink(link, subscription.GroupName, id)
		if err != nil {
			logger.Warning("Failed to parse share link: %v", err)
			continue
		}
		
		_, err = outboundService.AddOutbound(outbound)
		if err != nil {
			logger.Warning("Failed to add outbound %s: %v", outbound.Remark, err)
			continue
		}
		addedCount++
	}

	// Update last update time
	subscription.LastUpdate = time.Now().Unix()
	err = s.UpdateSubscription(subscription)
	if err != nil {
		logger.Warning("Failed to update subscription timestamp: %v", err)
	}

	return addedCount, nil
}

// downloadSubscription downloads subscription content from URL
func (s *SubscriptionService) downloadSubscription(subURL string, userAgent string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", subURL, nil)
	if err != nil {
		return "", err
	}

	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// decodeSubscription attempts to decode base64 subscription content
func (s *SubscriptionService) decodeSubscription(content string) (string, error) {
	content = strings.TrimSpace(content)
	
	// Try standard base64 decoding
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err == nil && len(decoded) > 0 {
		return string(decoded), nil
	}

	// Try URL-safe base64 decoding
	decoded, err = base64.URLEncoding.DecodeString(content)
	if err == nil && len(decoded) > 0 {
		return string(decoded), nil
	}

	// Try raw base64 (no padding)
	decoded, err = base64.RawStdEncoding.DecodeString(content)
	if err == nil && len(decoded) > 0 {
		return string(decoded), nil
	}

	return "", fmt.Errorf("not valid base64")
}

// parseShareLinks extracts share links from subscription content
func (s *SubscriptionService) parseShareLinks(content string) []string {
	var links []string
	
	// Split by newlines
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line is a valid share link
		if strings.HasPrefix(line, "vmess://") ||
			strings.HasPrefix(line, "vless://") ||
			strings.HasPrefix(line, "trojan://") ||
			strings.HasPrefix(line, "ss://") ||
			strings.HasPrefix(line, "shadowsocks://") {
			links = append(links, line)
		}
	}

	return links
}

// parseShareLink parses a share link and creates an Outbound
func (s *SubscriptionService) parseShareLink(link string, groupName string, subID int) (*model.Outbound, error) {
	if strings.HasPrefix(link, "vmess://") {
		return s.parseVmessLink(link, groupName, subID)
	} else if strings.HasPrefix(link, "vless://") {
		return s.parseVlessLink(link, groupName, subID)
	} else if strings.HasPrefix(link, "trojan://") {
		return s.parseTrojanLink(link, groupName, subID)
	} else if strings.HasPrefix(link, "ss://") || strings.HasPrefix(link, "shadowsocks://") {
		return s.parseShadowsocksLink(link, groupName, subID)
	}

	return nil, fmt.Errorf("unsupported protocol")
}

// parseVmessLink parses vmess:// share link
func (s *SubscriptionService) parseVmessLink(link string, groupName string, subID int) (*model.Outbound, error) {
	// Remove "vmess://" prefix
	encoded := strings.TrimPrefix(link, "vmess://")
	
	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode vmess link: %w", err)
		}
	}

	// Parse JSON
	var vmessData map[string]interface{}
	err = json.Unmarshal(decoded, &vmessData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vmess JSON: %w", err)
	}

	// Extract fields
	address, _ := vmessData["add"].(string)
	port, _ := strconv.Atoi(fmt.Sprintf("%v", vmessData["port"]))
	uuid, _ := vmessData["id"].(string)
	alterId, _ := strconv.Atoi(fmt.Sprintf("%v", vmessData["aid"]))
	remark, _ := vmessData["ps"].(string)
	network, _ := vmessData["net"].(string)
	tls, _ := vmessData["tls"].(string)

	if address == "" || port == 0 || uuid == "" {
		return nil, fmt.Errorf("invalid vmess link: missing required fields")
	}

	// Create settings
	settings := map[string]interface{}{
		"vnext": []map[string]interface{}{
			{
				"address": address,
				"port":    port,
				"users": []map[string]interface{}{
					{
						"id":       uuid,
						"alterId":  alterId,
						"security": "auto",
					},
				},
			},
		},
	}

	settingsJSON, _ := json.Marshal(settings)

	// Create stream settings
	streamSettings := map[string]interface{}{
		"network": network,
		"security": tls,
	}

	if network == "ws" {
		path, _ := vmessData["path"].(string)
		host, _ := vmessData["host"].(string)
		streamSettings["wsSettings"] = map[string]interface{}{
			"path": path,
			"headers": map[string]interface{}{
				"Host": host,
			},
		}
	}

	streamSettingsJSON, _ := json.Marshal(streamSettings)

	// Generate unique tag
	tag := fmt.Sprintf("sub%d-%s-%d", subID, remark, time.Now().UnixNano())
	tag = strings.ReplaceAll(tag, " ", "-")

	outbound := &model.Outbound{
		Protocol:       model.VMess,
		Address:        address,
		Port:           port,
		Settings:       string(settingsJSON),
		StreamSettings: string(streamSettingsJSON),
		Tag:            tag,
		Remark:         remark,
		Enable:         true,
		GroupName:      groupName,
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
	}

	return outbound, nil
}

// parseVlessLink parses vless:// share link
func (s *SubscriptionService) parseVlessLink(link string, groupName string, subID int) (*model.Outbound, error) {
	// vless://uuid@address:port?params#remark
	link = strings.TrimPrefix(link, "vless://")
	
	// Split by @ to get uuid and rest
	parts := strings.SplitN(link, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid vless link format")
	}

	uuid := parts[0]
	rest := parts[1]

	// Split by ? to get address:port and params
	addressParts := strings.SplitN(rest, "?", 2)
	addressPort := addressParts[0]

	// Parse address and port
	hostPort := strings.SplitN(addressPort, ":", 2)
	if len(hostPort) != 2 {
		return nil, fmt.Errorf("invalid address:port format")
	}

	address := hostPort[0]
	port, err := strconv.Atoi(hostPort[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	// Parse query parameters and fragment
	var params url.Values
	var remark string
	
	if len(addressParts) > 1 {
		// Split by # to get params and remark
		paramParts := strings.SplitN(addressParts[1], "#", 2)
		params, _ = url.ParseQuery(paramParts[0])
		if len(paramParts) > 1 {
			remark, _ = url.QueryUnescape(paramParts[1])
		}
	}

	// Create settings
	settings := map[string]interface{}{
		"vnext": []map[string]interface{}{
			{
				"address": address,
				"port":    port,
				"users": []map[string]interface{}{
					{
						"id":         uuid,
						"encryption": "none",
					},
				},
			},
		},
	}

	settingsJSON, _ := json.Marshal(settings)

	// Create stream settings
	network := params.Get("type")
	if network == "" {
		network = "tcp"
	}

	security := params.Get("security")
	streamSettings := map[string]interface{}{
		"network":  network,
		"security": security,
	}

	if network == "ws" {
		streamSettings["wsSettings"] = map[string]interface{}{
			"path": params.Get("path"),
			"headers": map[string]interface{}{
				"Host": params.Get("host"),
			},
		}
	}

	if security == "tls" {
		streamSettings["tlsSettings"] = map[string]interface{}{
			"serverName": params.Get("sni"),
		}
	}

	streamSettingsJSON, _ := json.Marshal(streamSettings)

	// Generate unique tag
	tag := fmt.Sprintf("sub%d-%s-%d", subID, remark, time.Now().UnixNano())
	tag = strings.ReplaceAll(tag, " ", "-")

	outbound := &model.Outbound{
		Protocol:       model.VLESS,
		Address:        address,
		Port:           port,
		Settings:       string(settingsJSON),
		StreamSettings: string(streamSettingsJSON),
		Tag:            tag,
		Remark:         remark,
		Enable:         true,
		GroupName:      groupName,
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
	}

	return outbound, nil
}

// parseTrojanLink parses trojan:// share link
func (s *SubscriptionService) parseTrojanLink(link string, groupName string, subID int) (*model.Outbound, error) {
	// trojan://password@address:port?params#remark
	link = strings.TrimPrefix(link, "trojan://")
	
	// Split by @ to get password and rest
	parts := strings.SplitN(link, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid trojan link format")
	}

	password := parts[0]
	rest := parts[1]

	// Split by ? to get address:port and params
	addressParts := strings.SplitN(rest, "?", 2)
	addressPort := addressParts[0]

	// Parse address and port
	hostPort := strings.SplitN(addressPort, ":", 2)
	if len(hostPort) != 2 {
		return nil, fmt.Errorf("invalid address:port format")
	}

	address := hostPort[0]
	port, err := strconv.Atoi(hostPort[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	// Parse query parameters and fragment
	var params url.Values
	var remark string
	
	if len(addressParts) > 1 {
		// Split by # to get params and remark
		paramParts := strings.SplitN(addressParts[1], "#", 2)
		params, _ = url.ParseQuery(paramParts[0])
		if len(paramParts) > 1 {
			remark, _ = url.QueryUnescape(paramParts[1])
		}
	}

	// Create settings
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

	// Create stream settings
	network := params.Get("type")
	if network == "" {
		network = "tcp"
	}

	security := params.Get("security")
	if security == "" {
		security = "tls"
	}

	streamSettings := map[string]interface{}{
		"network":  network,
		"security": security,
	}

	if network == "ws" {
		streamSettings["wsSettings"] = map[string]interface{}{
			"path": params.Get("path"),
			"headers": map[string]interface{}{
				"Host": params.Get("host"),
			},
		}
	}

	if security == "tls" {
		streamSettings["tlsSettings"] = map[string]interface{}{
			"serverName": params.Get("sni"),
		}
	}

	streamSettingsJSON, _ := json.Marshal(streamSettings)

	// Generate unique tag
	tag := fmt.Sprintf("sub%d-%s-%d", subID, remark, time.Now().UnixNano())
	tag = strings.ReplaceAll(tag, " ", "-")

	outbound := &model.Outbound{
		Protocol:       model.Trojan,
		Address:        address,
		Port:           port,
		Settings:       string(settingsJSON),
		StreamSettings: string(streamSettingsJSON),
		Tag:            tag,
		Remark:         remark,
		Enable:         true,
		GroupName:      groupName,
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
	}

	return outbound, nil
}

// parseShadowsocksLink parses ss:// share link
func (s *SubscriptionService) parseShadowsocksLink(link string, groupName string, subID int) (*model.Outbound, error) {
	// ss://base64(method:password)@address:port#remark
	link = strings.TrimPrefix(link, "ss://")
	link = strings.TrimPrefix(link, "shadowsocks://")

	// Split by # to get main part and remark
	parts := strings.SplitN(link, "#", 2)
	mainPart := parts[0]
	var remark string
	if len(parts) > 1 {
		remark, _ = url.QueryUnescape(parts[1])
	}

	// Split by @ to get userinfo and server
	userInfoParts := strings.SplitN(mainPart, "@", 2)
	if len(userInfoParts) != 2 {
		return nil, fmt.Errorf("invalid shadowsocks link format")
	}

	// Decode userinfo
	userInfo, err := base64.StdEncoding.DecodeString(userInfoParts[0])
	if err != nil {
		userInfo, err = base64.RawStdEncoding.DecodeString(userInfoParts[0])
		if err != nil {
			return nil, fmt.Errorf("failed to decode userinfo: %w", err)
		}
	}

	// Parse method:password
	methodPassword := strings.SplitN(string(userInfo), ":", 2)
	if len(methodPassword) != 2 {
		return nil, fmt.Errorf("invalid method:password format")
	}

	method := methodPassword[0]
	password := methodPassword[1]

	// Parse address:port
	hostPort := strings.SplitN(userInfoParts[1], ":", 2)
	if len(hostPort) != 2 {
		return nil, fmt.Errorf("invalid address:port format")
	}

	address := hostPort[0]
	port, err := strconv.Atoi(hostPort[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	// Create settings
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

	// Create stream settings (shadowsocks usually uses TCP)
	streamSettings := map[string]interface{}{
		"network":  "tcp",
		"security": "none",
	}

	streamSettingsJSON, _ := json.Marshal(streamSettings)

	// Generate unique tag
	tag := fmt.Sprintf("sub%d-%s-%d", subID, remark, time.Now().UnixNano())
	tag = strings.ReplaceAll(tag, " ", "-")

	outbound := &model.Outbound{
		Protocol:       model.Shadowsocks,
		Address:        address,
		Port:           port,
		Settings:       string(settingsJSON),
		StreamSettings: string(streamSettingsJSON),
		Tag:            tag,
		Remark:         remark,
		Enable:         true,
		GroupName:      groupName,
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
	}

	return outbound, nil
}

// deleteSubscriptionOutbounds deletes all outbounds created by a subscription
func (s *SubscriptionService) deleteSubscriptionOutbounds(subID int) error {
	db := database.GetDB()
	// Delete outbounds with tags starting with "sub{id}-"
	prefix := fmt.Sprintf("sub%d-", subID)
	return db.Where("tag LIKE ?", prefix+"%").Delete(&model.Outbound{}).Error
}
