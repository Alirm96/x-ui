package controller

import (
	"strconv"

	"github.com/alireza0/x-ui/database/model"
	"github.com/alireza0/x-ui/web/service"
	"github.com/gin-gonic/gin"
)

type SubscriptionController struct {
	BaseController
	subscriptionService service.SubscriptionService
}

func NewSubscriptionController(g *gin.RouterGroup) *SubscriptionController {
	a := &SubscriptionController{}
	a.initRouter(g)
	return a
}

func (a *SubscriptionController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/subscription")

	g.POST("/list", a.getSubscriptions)
	g.POST("/add", a.addSubscription)
	g.POST("/update/:id", a.updateSubscription)
	g.POST("/delete/:id", a.deleteSubscription)
	g.POST("/refresh/:id", a.refreshSubscription)
}

// getSubscriptions returns all subscriptions
func (a *SubscriptionController) getSubscriptions(c *gin.Context) {
	subscriptions, err := a.subscriptionService.GetSubscriptions()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.getSubscriptionsFailed"), err)
		return
	}
	jsonObj(c, subscriptions, nil)
}

// addSubscription creates a new subscription
func (a *SubscriptionController) addSubscription(c *gin.Context) {
	subscription := &model.Subscription{}
	err := c.ShouldBind(subscription)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.invalidData"), err)
		return
	}

	// Validate required fields
	if subscription.Remark == "" {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.remarkRequired"), nil)
		return
	}
	if subscription.URL == "" {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.urlRequired"), nil)
		return
	}
	if subscription.GroupName == "" {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.groupRequired"), nil)
		return
	}

	err = a.subscriptionService.AddSubscription(subscription)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.addSubscriptionFailed"), err)
		return
	}

	jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.addSubscriptionSuccess"), nil)
}

// updateSubscription updates an existing subscription
func (a *SubscriptionController) updateSubscription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.invalidId"), err)
		return
	}

	subscription := &model.Subscription{}
	err = c.ShouldBind(subscription)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.invalidData"), err)
		return
	}

	subscription.Id = id

	// Validate required fields
	if subscription.Remark == "" {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.remarkRequired"), nil)
		return
	}
	if subscription.URL == "" {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.urlRequired"), nil)
		return
	}
	if subscription.GroupName == "" {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.groupRequired"), nil)
		return
	}

	err = a.subscriptionService.UpdateSubscription(subscription)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.updateSubscriptionFailed"), err)
		return
	}

	jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.updateSubscriptionSuccess"), nil)
}

// deleteSubscription deletes a subscription
func (a *SubscriptionController) deleteSubscription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.invalidId"), err)
		return
	}

	err = a.subscriptionService.DeleteSubscription(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.deleteSubscriptionFailed"), err)
		return
	}

	jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.deleteSubscriptionSuccess"), nil)
}

// refreshSubscription manually updates a subscription
func (a *SubscriptionController) refreshSubscription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.invalidId"), err)
		return
	}

	addedCount, err := a.subscriptionService.UpdateSubscriptionContent(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.subscriptions.toasts.refreshSubscriptionFailed"), err)
		return
	}

	jsonMsgObj(c, I18nWeb(c, "pages.subscriptions.toasts.refreshSubscriptionSuccess"), addedCount, nil)
}
