package controller

import (
	"fmt"
	"strconv"

	"github.com/alireza0/x-ui/database/model"
	"github.com/alireza0/x-ui/util/common"
	"github.com/alireza0/x-ui/web/service"
	"github.com/alireza0/x-ui/web/session"

	"github.com/gin-gonic/gin"
)

type OutboundController struct {
	outboundService service.OutboundService
	settingService  service.SettingService
}

func NewOutboundController(g *gin.RouterGroup) *OutboundController {
	a := &OutboundController{}
	a.initRouter(g)
	return a
}

func (a *OutboundController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/outbound")

	g.POST("/list", a.getOutbounds)
	g.POST("/get/:id", a.getOutbound)
	g.POST("/add", a.addOutbound)
	g.POST("/del/:id", a.delOutbound)
	g.POST("/update/:id", a.updateOutbound)
	g.POST("/import", a.importOutbound)
	g.POST("/test/:id", a.testOutbound)
	g.POST("/testAll", a.testAllOutbounds)
	g.POST("/testHistory/:id", a.getTestHistory)
	g.POST("/getBest", a.getBestOutbound)
	g.POST("/migrateVLESS", a.migrateVLESS)
}

func (a *OutboundController) getOutbounds(c *gin.Context) {
	user := session.GetLoginUser(c)
	var userId int
	if user != nil {
		userId = user.Id
	} else {
		// For dev mode bypass, use a default user ID
		userId = 1
	}
	outbounds, err := a.outboundService.GetOutbounds(userId)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.outbounds.toasts.obtain"), err)
		return
	}
	jsonObj(c, outbounds, nil)
}

func (a *OutboundController) getOutbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "get"), err)
		return
	}
	outbound, err := a.outboundService.GetOutbound(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.outbounds.toasts.obtain"), err)
		return
	}
	jsonObj(c, outbound, nil)
}

func (a *OutboundController) addOutbound(c *gin.Context) {
	outbound := &model.Outbound{}
	err := c.ShouldBind(outbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.outbounds.create"), err)
		return
	}
	user := session.GetLoginUser(c)
	if user != nil {
		outbound.UserId = user.Id
	} else {
		// For dev mode bypass, use a default user ID
		outbound.UserId = 1
	}

	outbound, err = a.outboundService.AddOutbound(outbound)
	jsonMsgObj(c, I18nWeb(c, "pages.outbounds.create"), outbound, err)
}

func (a *OutboundController) delOutbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "delete"), err)
		return
	}
	err = a.outboundService.DelOutbound(id)
	jsonMsgObj(c, I18nWeb(c, "delete"), id, err)
}

func (a *OutboundController) updateOutbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.outbounds.update"), err)
		return
	}
	
	// Get existing outbound to preserve userId
	existing, err := a.outboundService.GetOutbound(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.outbounds.update"), err)
		return
	}
	
	outbound := &model.Outbound{
		Id: id,
		UserId: existing.UserId, // Preserve the original userId
	}
	err = c.ShouldBind(outbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.outbounds.update"), err)
		return
	}
	
	// Ensure userId is preserved (in case it wasn't in the request)
	if outbound.UserId == 0 {
		outbound.UserId = existing.UserId
	}
	
	outbound, err = a.outboundService.UpdateOutbound(outbound)
	jsonMsgObj(c, I18nWeb(c, "pages.outbounds.update"), outbound, err)
}

func (a *OutboundController) importOutbound(c *gin.Context) {
	fmt.Println("========================================")
	fmt.Println("IMPORT OUTBOUND FUNCTION CALLED")
	fmt.Println("Content-Type:", c.GetHeader("Content-Type"))
	fmt.Println("========================================")
	
	// Use ShouldBind which handles both JSON and form data
	var data struct {
		Link      string `json:"link" form:"link"`
		GroupName string `json:"groupName" form:"groupName"`
	}
	
	if err := c.ShouldBind(&data); err != nil {
		fmt.Println("ShouldBind error:", err)
		jsonMsg(c, "Import outbound", common.NewError("invalid request data"))
		return
	}
	
	fmt.Println("=== DATA RECEIVED ===")
	fmt.Println("Link:", data.Link)
	fmt.Println("GroupName:", data.GroupName)
	
	if data.Link == "" {
		fmt.Println("ERROR: link is empty")
		jsonMsg(c, "Import outbound", common.NewError("link is required"))
		return
	}
	
	groupName := data.GroupName
	if groupName == "" {
		groupName = "default"
		fmt.Println("GroupName was empty, defaulting to:", groupName)
	}
	
	user := session.GetLoginUser(c)
	var userId int
	if user != nil {
		userId = user.Id
	} else {
		userId = 1
	}
	
	fmt.Println("Calling ImportOutboundFromLink with groupName:", groupName)
	outbound, err := a.outboundService.ImportOutboundFromLink(data.Link, userId, groupName)
	if err != nil {
		fmt.Println("ImportOutboundFromLink failed:", err)
		jsonMsg(c, "Import outbound", err)
		return
	}
	
	fmt.Println("Import successful! Outbound ID:", outbound.Id, "GroupName:", outbound.GroupName)
	jsonMsgObj(c, "Outbound imported successfully", outbound, nil)
}

func (a *OutboundController) testOutbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Test outbound", err)
		return
	}

	// Get test settings from configuration
	testURL := "https://www.google.com"
	timeout := 10

	// You can get these from settings if available
	// testURL, _ = a.settingService.GetOutboundTestURL()
	// timeout, _ = a.settingService.GetOutboundTestTimeout()

	latency, err := a.outboundService.TestOutbound(id, testURL, timeout)
	if err != nil {
		jsonMsgObj(c, "Test failed", map[string]interface{}{
			"id":      id,
			"latency": -1,
			"error":   err.Error(),
		}, err)
		return
	}

	jsonMsgObj(c, "Test successful", map[string]interface{}{
		"id":      id,
		"latency": latency,
	}, nil)
}

func (a *OutboundController) testAllOutbounds(c *gin.Context) {
	user := session.GetLoginUser(c)
	var userId int
	if user != nil {
		userId = user.Id
	} else {
		// For dev mode bypass, use a default user ID
		userId = 1
	}
	outbounds, err := a.outboundService.GetOutbounds(userId)
	if err != nil {
		jsonMsg(c, "Test all outbounds", err)
		return
	}

	testURL := "https://www.google.com"
	timeout := 10

	results := make([]map[string]interface{}, 0)
	for _, outbound := range outbounds {
		if !outbound.Enable {
			continue
		}

		latency, err := a.outboundService.TestOutbound(outbound.Id, testURL, timeout)
		result := map[string]interface{}{
			"id":      outbound.Id,
			"tag":     outbound.Tag,
			"remark":  outbound.Remark,
			"latency": latency,
		}
		if err != nil {
			result["error"] = err.Error()
		}
		results = append(results, result)
	}

	jsonMsgObj(c, "Tests completed", results, nil)
}

func (a *OutboundController) getTestHistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Get test history", err)
		return
	}

	limit := 50 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	history, err := a.outboundService.GetTestHistory(id, limit)
	if err != nil {
		jsonMsg(c, "Get test history", err)
		return
	}

	jsonObj(c, history, nil)
}

func (a *OutboundController) getBestOutbound(c *gin.Context) {
	outbound, err := a.outboundService.GetBestOutbound()
	if err != nil {
		jsonMsg(c, "Get best outbound", err)
		return
	}
	jsonObj(c, outbound, nil)
}

func (a *OutboundController) migrateVLESS(c *gin.Context) {
	err := a.outboundService.MigrateVLESSOutbounds()
	if err != nil {
		jsonMsg(c, "VLESS migration failed", err)
		return
	}
	jsonMsg(c, "VLESS migration completed successfully. Please restart Xray for changes to take effect.", nil)
}
