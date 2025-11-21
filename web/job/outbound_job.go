package job

import (
	"github.com/alireza0/x-ui/logger"
	"github.com/alireza0/x-ui/web/service"
)

type OutboundTestJob struct {
	outboundService service.OutboundService
	settingService  service.SettingService
}

func NewOutboundTestJob() *OutboundTestJob {
	return new(OutboundTestJob)
}

// Run is an interface method of Job interface
func (j *OutboundTestJob) Run() {
	logger.Info("Running outbound connection tests...")

	// Get all enabled outbounds
	outbounds, err := j.outboundService.GetAllOutbounds()
	if err != nil {
		logger.Warning("Failed to get outbounds for testing:", err)
		return
	}

	// Get test configuration from settings
	testURL, err := j.settingService.GetOutboundTestURL()
	if err != nil {
		testURL = "https://www.google.com"
	}
	
	timeout, err := j.settingService.GetOutboundTestTimeout()
	if err != nil {
		timeout = 10
	}

	successCount := 0
	failCount := 0

	for _, outbound := range outbounds {
		if !outbound.Enable {
			continue
		}

		logger.Infof("Testing outbound: %s (%s)", outbound.Remark, outbound.Tag)
		
		latency, err := j.outboundService.TestOutbound(outbound.Id, testURL, timeout)
		if err != nil {
			logger.Warning("Outbound test failed for", outbound.Tag, ":", err)
			failCount++
		} else {
			logger.Infof("Outbound %s latency: %dms", outbound.Tag, latency)
			successCount++
		}
	}

	logger.Infof("Outbound tests completed: %d succeeded, %d failed", successCount, failCount)
}

type OutboundCleanupJob struct {
	outboundService service.OutboundService
	settingService  service.SettingService
}

func NewOutboundCleanupJob() *OutboundCleanupJob {
	return new(OutboundCleanupJob)
}

// Run is an interface method of Job interface
func (j *OutboundCleanupJob) Run() {
	logger.Info("Running outbound cleanup job...")

	// Get cleanup threshold from settings
	daysThreshold, err := j.settingService.GetOutboundCleanupDays()
	if err != nil {
		daysThreshold = 7
	}

	err = j.outboundService.CleanupFailedOutbounds(daysThreshold)
	if err != nil {
		logger.Warning("Failed to cleanup outbounds:", err)
		return
	}

	logger.Info("Outbound cleanup completed")
}

// OutboundAutoRouteJob updates xray routing to use best outbound
type OutboundAutoRouteJob struct {
	outboundService    service.OutboundService
	inboundService     service.InboundService
	settingService     service.SettingService
	xraySettingService service.XraySettingService
}

func NewOutboundAutoRouteJob() *OutboundAutoRouteJob {
	return new(OutboundAutoRouteJob)
}

// Run is an interface method of Job interface
func (j *OutboundAutoRouteJob) Run() {
	// Check if auto-routing is enabled
	enabled, err := j.settingService.GetOutboundAutoRoute()
	if err != nil || !enabled {
		return
	}

	logger.Info("Running outbound auto-routing job...")

	// Get best outbound
	bestOutbound, err := j.outboundService.GetBestOutbound()
	if err != nil {
		logger.Warning("Failed to get best outbound:", err)
		return
	}

	if bestOutbound == nil {
		logger.Warning("No suitable outbound found for auto-routing")
		return
	}

	logger.Infof("Best outbound selected: %s (%s) with %dms latency", 
		bestOutbound.Remark, bestOutbound.Tag, bestOutbound.LastTestResult)

	// Update xray routing configuration to use this outbound
	err = j.xraySettingService.ApplyOutboundRouting(bestOutbound.Tag)
	if err != nil {
		logger.Warning("Failed to apply outbound routing:", err)
		return
	}
	
	// Update routing rules for default inbounds to use the best outbound
	err = j.inboundService.CreateDefaultRoutingRules(bestOutbound.Tag)
	if err != nil {
		logger.Warning("Failed to update default inbound routing rules:", err)
		return
	}
	
	logger.Info("Outbound auto-routing completed successfully")
}
