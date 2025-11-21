package controller

import (
	"errors"
	"time"

	"github.com/alireza0/x-ui/logger"
	"github.com/alireza0/x-ui/web/entity"
	"github.com/alireza0/x-ui/web/service"
	"github.com/alireza0/x-ui/web/session"

	"github.com/gin-gonic/gin"
)

type updateUserForm struct {
	OldUsername string `json:"oldUsername" form:"oldUsername"`
	OldPassword string `json:"oldPassword" form:"oldPassword"`
	NewUsername string `json:"newUsername" form:"newUsername"`
	NewPassword string `json:"newPassword" form:"newPassword"`
}

type SettingController struct {
	settingService service.SettingService
	userService    service.UserService
	panelService   service.PanelService
}

func NewSettingController(g *gin.RouterGroup) *SettingController {
	a := &SettingController{}
	a.initRouter(g)
	return a
}

func (a *SettingController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/setting")

	g.POST("/all", a.getAllSetting)
	g.POST("/defaultSettings", a.getDefaultSettings)
	g.POST("/update", a.updateSetting)
	g.POST("/updateUser", a.updateUser)
	g.POST("/restartPanel", a.restartPanel)
	g.GET("/getDefaultJsonConfig", a.getDefaultXrayConfig)
}

func (a *SettingController) getAllSetting(c *gin.Context) {
	allSetting, err := a.settingService.GetAllSetting()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.getSettings"), err)
		return
	}
	jsonObj(c, allSetting, nil)
}

func (a *SettingController) getDefaultSettings(c *gin.Context) {
	result, err := a.settingService.GetDefaultSettings(c.Request.Host)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.getSettings"), err)
		return
	}
	jsonObj(c, result, nil)
}

func (a *SettingController) updateSetting(c *gin.Context) {
	// Load current settings first
	currentSettings, err := a.settingService.GetAllSetting()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifySettings"), err)
		return
	}
	
	logger.Info("=== DEBUG: GetAllSetting returned WebPort:", currentSettings.WebPort)
	logger.Info("=== DEBUG: Full current settings:", currentSettings)
	
	// Parse form data into a temporary struct to avoid overwriting existing fields
	incomingSettings := &entity.AllSetting{}
	err = c.ShouldBind(incomingSettings)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifySettings"), err)
		return
	}
	
	logger.Info("=== DEBUG: After ShouldBind, incoming WebPort:", incomingSettings.WebPort)
	
	// Only update the outbound-related fields that were actually sent
	if c.PostForm("outboundTestInterval") != "" {
		currentSettings.OutboundTestInterval = incomingSettings.OutboundTestInterval
	}
	if c.PostForm("outboundTestURL") != "" {
		currentSettings.OutboundTestURL = incomingSettings.OutboundTestURL
	}
	if c.PostForm("outboundTestTimeout") != "" {
		currentSettings.OutboundTestTimeout = incomingSettings.OutboundTestTimeout
	}
	if c.PostForm("outboundCleanupDays") != "" {
		currentSettings.OutboundCleanupDays = incomingSettings.OutboundCleanupDays
	}
	if c.PostForm("outboundAutoRoute") != "" {
		currentSettings.OutboundAutoRoute = incomingSettings.OutboundAutoRoute
	}
	if c.PostForm("outboundRouteInterval") != "" {
		currentSettings.OutboundRouteInterval = incomingSettings.OutboundRouteInterval
	}
	
	logger.Info("=== DEBUG: Before UpdateAllSetting, final WebPort:", currentSettings.WebPort)
	
	err = a.settingService.UpdateAllSetting(currentSettings)
	jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifySettings"), err)
}

func (a *SettingController) updateUser(c *gin.Context) {
	form := &updateUserForm{}
	err := c.ShouldBind(form)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifySettings"), err)
		return
	}
	user := session.GetLoginUser(c)
	if user.Username != form.OldUsername || user.Password != form.OldPassword {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifyUser"), errors.New(I18nWeb(c, "pages.settings.toasts.originalUserPassIncorrect")))
		return
	}
	if form.NewUsername == "" || form.NewPassword == "" {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifyUser"), errors.New(I18nWeb(c, "pages.settings.toasts.userPassMustBeNotEmpty")))
		return
	}
	err = a.userService.UpdateUser(user.Id, form.NewUsername, form.NewPassword)
	if err == nil {
		user.Username = form.NewUsername
		user.Password = form.NewPassword
		session.SetLoginUser(c, user)
	}
	jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifyUser"), err)
}

func (a *SettingController) restartPanel(c *gin.Context) {
	err := a.panelService.RestartPanel(time.Second * 3)
	jsonMsg(c, I18nWeb(c, "pages.settings.restartPanel"), err)
}

func (a *SettingController) getDefaultXrayConfig(c *gin.Context) {
	defaultJsonConfig, err := a.settingService.GetDefaultXrayConfig()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.getSettings"), err)
		return
	}
	jsonObj(c, defaultJsonConfig, nil)
}
