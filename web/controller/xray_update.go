package controller

import (
	"fmt"
	"sync"

	"github.com/alireza0/x-ui/web/service"
	"github.com/gin-gonic/gin"
)

type XrayUpdateController struct {
	BaseController
	
	updateService service.XrayUpdateService
	
	// Track download progress
	downloadProgress struct {
		sync.RWMutex
		Downloaded int64
		Total      int64
		InProgress bool
		Error      string
	}
}

func NewXrayUpdateController(g *gin.RouterGroup) *XrayUpdateController {
	a := &XrayUpdateController{}
	a.initRouter(g)
	return a
}

func (a *XrayUpdateController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/xui/update")
	g.Use(a.checkLogin)
	
	g.GET("/current-version", a.getCurrentVersion)
	g.POST("/check", a.checkForUpdates)
	g.POST("/download", a.downloadAndInstall)
	g.GET("/progress", a.getDownloadProgress)
}

// getCurrentVersion returns the currently installed Xray version
func (a *XrayUpdateController) getCurrentVersion(c *gin.Context) {
	version, err := a.updateService.GetCurrentVersion()
	if err != nil {
		jsonMsg(c, "Failed to get current version", err)
		return
	}
	
	jsonObj(c, gin.H{
		"version": version,
	}, nil)
}

// checkForUpdates checks if there's a new version available
func (a *XrayUpdateController) checkForUpdates(c *gin.Context) {
	var req struct {
		IncludePrerelease bool `json:"includePrerelease" form:"includePrerelease"`
	}
	
	if err := c.ShouldBind(&req); err != nil {
		jsonMsg(c, "Invalid request", err)
		return
	}
	
	versionInfo, err := a.updateService.CheckForUpdates(req.IncludePrerelease)
	if err != nil {
		jsonMsg(c, "Failed to check for updates", err)
		return
	}
	
	jsonObj(c, versionInfo, nil)
}

// downloadAndInstall downloads and installs the new version
func (a *XrayUpdateController) downloadAndInstall(c *gin.Context) {
	var req struct {
		DownloadURL string `json:"downloadUrl" form:"downloadUrl" binding:"required"`
	}
	
	if err := c.ShouldBind(&req); err != nil {
		jsonMsg(c, "Invalid request", err)
		return
	}
	
	// Check if a download is already in progress
	a.downloadProgress.RLock()
	if a.downloadProgress.InProgress {
		a.downloadProgress.RUnlock()
		jsonMsg(c, "Download already in progress", nil)
		return
	}
	a.downloadProgress.RUnlock()
	
	// Start download in background
	a.downloadProgress.Lock()
	a.downloadProgress.InProgress = true
	a.downloadProgress.Downloaded = 0
	a.downloadProgress.Total = 0
	a.downloadProgress.Error = ""
	a.downloadProgress.Unlock()
	
	go func() {
		progressCallback := func(downloaded, total int64) {
			a.downloadProgress.Lock()
			a.downloadProgress.Downloaded = downloaded
			a.downloadProgress.Total = total
			a.downloadProgress.Unlock()
		}
		
		err := a.updateService.DownloadAndInstall(req.DownloadURL, progressCallback)
		
		a.downloadProgress.Lock()
		a.downloadProgress.InProgress = false
		if err != nil {
			a.downloadProgress.Error = err.Error()
		}
		a.downloadProgress.Unlock()
		
		if err == nil {
			// Restart Xray after successful installation
			a.updateService.RestartXray()
		}
	}()
	
	jsonMsg(c, "Download started", nil)
}

// getDownloadProgress returns the current download progress
func (a *XrayUpdateController) getDownloadProgress(c *gin.Context) {
	a.downloadProgress.RLock()
	defer a.downloadProgress.RUnlock()
	
	percentage := float64(0)
	if a.downloadProgress.Total > 0 {
		percentage = float64(a.downloadProgress.Downloaded) / float64(a.downloadProgress.Total) * 100
	}
	
	jsonObj(c, gin.H{
		"inProgress": a.downloadProgress.InProgress,
		"downloaded": a.downloadProgress.Downloaded,
		"total":      a.downloadProgress.Total,
		"percentage": fmt.Sprintf("%.2f", percentage),
		"error":      a.downloadProgress.Error,
	}, nil)
}
