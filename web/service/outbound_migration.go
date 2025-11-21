package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alireza0/x-ui/database"
	"github.com/alireza0/x-ui/database/model"
	"github.com/alireza0/x-ui/logger"
)

// MigrateVLESSOutbounds migrates old VLESS outbounds with multiple users
// to comply with Xray-core v25.9.11+ requirement of one user per outbound
func (s *OutboundService) MigrateVLESSOutbounds() error {
	logger.Info("Starting VLESS outbound migration...")
	
	db := database.GetDB()
	var outbounds []*model.Outbound
	
	// Get all VLESS outbounds
	err := db.Model(model.Outbound{}).Where("protocol = ?", model.VLESS).Find(&outbounds).Error
	if err != nil {
		return fmt.Errorf("failed to fetch VLESS outbounds: %w", err)
	}
	
	if len(outbounds) == 0 {
		logger.Info("No VLESS outbounds found")
		return nil
	}
	
	logger.Info(fmt.Sprintf("Found %d VLESS outbounds to check", len(outbounds)))
	
	var migratedCount int
	var newOutbounds []*model.Outbound
	var outboundsToDelete []int
	
	for _, outbound := range outbounds {
		// Parse settings
		var settings map[string]interface{}
		if err := json.Unmarshal([]byte(outbound.Settings), &settings); err != nil {
			logger.Warning(fmt.Sprintf("Failed to parse settings for outbound %d: %v", outbound.Id, err))
			continue
		}
		
		// Check if vnext exists
		vnextRaw, ok := settings["vnext"]
		if !ok {
			continue
		}
		
		vnextArray, ok := vnextRaw.([]interface{})
		if !ok || len(vnextArray) == 0 {
			continue
		}
		
		// Get the first vnext entry
		vnextMap, ok := vnextArray[0].(map[string]interface{})
		if !ok {
			continue
		}
		
		// Check users array
		usersRaw, ok := vnextMap["users"]
		if !ok {
			continue
		}
		
		usersArray, ok := usersRaw.([]interface{})
		if !ok {
			continue
		}
		
		// If only one user, no migration needed
		if len(usersArray) <= 1 {
			continue
		}
		
		logger.Info(fmt.Sprintf("Migrating outbound '%s' (ID: %d) with %d users", 
			outbound.Remark, outbound.Id, len(usersArray)))
		
		// Create separate outbounds for each user (skip first, update it instead)
		for i, userRaw := range usersArray {
			userMap, ok := userRaw.(map[string]interface{})
			if !ok {
				continue
			}
			
			if i == 0 {
				// Update the original outbound with only the first user
				newSettings := map[string]interface{}{
					"vnext": []map[string]interface{}{
						{
							"address": vnextMap["address"],
							"port":    vnextMap["port"],
							"users": []interface{}{userMap},
						},
					},
				}
				
				newSettingsJSON, _ := json.Marshal(newSettings)
				outbound.Settings = string(newSettingsJSON)
				
				// Update in database
				if err := db.Save(outbound).Error; err != nil {
					logger.Error(fmt.Sprintf("Failed to update outbound %d: %v", outbound.Id, err))
					continue
				}
				
				logger.Info(fmt.Sprintf("  - Updated original outbound (user 1/%d)", len(usersArray)))
			} else {
				// Create new outbound for additional users
				newOutbound := &model.Outbound{
					UserId:         outbound.UserId,
					Remark:         fmt.Sprintf("%s-user%d", outbound.Remark, i+1),
					Enable:         outbound.Enable,
					Port:           outbound.Port,
					Protocol:       outbound.Protocol,
					Settings:       "", // Will set below
					StreamSettings: outbound.StreamSettings,
					Tag:            fmt.Sprintf("%s-user%d", outbound.Tag, i+1),
					GroupName:      outbound.GroupName,
					Address:        outbound.Address,
					IsSystem:       false,
					LastTestTime:   0,
					LastTestResult: 0,
					TestFailCount:  0,
					CreatedAt:      time.Now().Unix(),
					UpdatedAt:      time.Now().Unix(),
				}
				
				// Create settings with single user
				newSettings := map[string]interface{}{
					"vnext": []map[string]interface{}{
						{
							"address": vnextMap["address"],
							"port":    vnextMap["port"],
							"users": []interface{}{userMap},
						},
					},
				}
				
				newSettingsJSON, _ := json.Marshal(newSettings)
				newOutbound.Settings = string(newSettingsJSON)
				
				newOutbounds = append(newOutbounds, newOutbound)
				logger.Info(fmt.Sprintf("  - Created new outbound for user %d/%d", i+1, len(usersArray)))
			}
		}
		
		migratedCount++
	}
	
	// Insert new outbounds
	if len(newOutbounds) > 0 {
		if err := db.Create(&newOutbounds).Error; err != nil {
			return fmt.Errorf("failed to create new outbounds: %w", err)
		}
		logger.Info(fmt.Sprintf("Created %d new outbounds", len(newOutbounds)))
	}
	
	// Delete old outbounds if needed (optional - we're updating them instead)
	if len(outboundsToDelete) > 0 {
		if err := db.Where("id IN ?", outboundsToDelete).Delete(&model.Outbound{}).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to delete old outbounds: %v", err))
		}
	}
	
	if migratedCount > 0 {
		logger.Info(fmt.Sprintf("Migration complete: %d outbounds migrated, %d new outbounds created", 
			migratedCount, len(newOutbounds)))
	} else {
		logger.Info("No migration needed - all outbounds are compliant")
	}
	
	return nil
}

// ValidateVLESSOutbound checks if a VLESS outbound complies with Xray-core v25.9.11+ requirements
func (s *OutboundService) ValidateVLESSOutbound(settings string) error {
	var settingsMap map[string]interface{}
	if err := json.Unmarshal([]byte(settings), &settingsMap); err != nil {
		return fmt.Errorf("invalid JSON settings: %w", err)
	}
	
	vnextRaw, ok := settingsMap["vnext"]
	if !ok {
		return fmt.Errorf("missing 'vnext' in settings")
	}
	
	vnextArray, ok := vnextRaw.([]interface{})
	if !ok || len(vnextArray) == 0 {
		return fmt.Errorf("'vnext' must be a non-empty array")
	}
	
	for i, vnextRaw := range vnextArray {
		vnextMap, ok := vnextRaw.(map[string]interface{})
		if !ok {
			return fmt.Errorf("vnext[%d] must be an object", i)
		}
		
		usersRaw, ok := vnextMap["users"]
		if !ok {
			return fmt.Errorf("vnext[%d] missing 'users'", i)
		}
		
		usersArray, ok := usersRaw.([]interface{})
		if !ok {
			return fmt.Errorf("vnext[%d].users must be an array", i)
		}
		
		if len(usersArray) != 1 {
			return fmt.Errorf("vnext[%d].users must have exactly one member (has %d). Multiple users should use multiple VLESS outbounds and routing balancer instead", i, len(usersArray))
		}
	}
	
	return nil
}
