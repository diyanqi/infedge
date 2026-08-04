// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

//nolint:revive // Exported domain API is consumed across the application layers.
package subscription

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Rain-kl/Wavelet/internal/apps/admin"
	"github.com/Rain-kl/Wavelet/internal/apps/oauth"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/Rain-kl/Wavelet/internal/repository"
	"github.com/Rain-kl/Wavelet/internal/shared/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(api *gin.RouterGroup, root *gin.Engine) {
	api.GET("/plans", ListPublicPlans)
	api.GET("/payment/channels", ListPublicChannels)
	user := api.Group("", oauth.LoginRequired())
	user.GET("/subscription", GetMySubscription)
	user.GET("/orders", ListMyOrders)
	user.POST("/orders", CreateMyOrder)
	user.POST("/redeem", Redeem)

	adminGroup := api.Group("/admin", oauth.LoginRequired(), admin.LoginAdminRequired())
	adminGroup.GET("/plans", ListPlans)
	adminGroup.POST("/plans", CreatePlan)
	adminGroup.PUT("/plans/:id", UpdatePlan)
	adminGroup.DELETE("/plans/:id", DeletePlan)
	adminGroup.GET("/redeem-codes", ListRedeemCodes)
	adminGroup.POST("/redeem-codes", CreateRedeemCode)
	adminGroup.GET("/payment/channels", ListChannels)
	adminGroup.POST("/payment/channels", CreateChannel)
	adminGroup.PUT("/payment/channels/:id", UpdateChannel)
	adminGroup.DELETE("/payment/channels/:id", DeleteChannel)
	if root != nil {
		root.POST("/payment/easy-pay/notify/:channel_id", Notify)
	}
}

// ListPublicPlans returns enabled subscription plans.
// @Summary 获取公开套餐列表
// @Tags custom-subscription
// @Produce json
// @Success 200 {object} response.Any{data=[]model.SubscriptionPlan}
// @Failure 500 {object} response.Any
// @Router /api/v1/custom/plans [get]
func ListPublicPlans(c *gin.Context) {
	rows, err := repository.ListSubscriptionPlans(c.Request.Context(), true)
	if err != nil {
		response.AbortInternal(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OK(rows))
}

// ListPlans returns all plans for administrators.
// @Summary 获取管理端套餐列表
// @Tags custom-subscription
// @Produce json
// @Security SessionCookie
// @Success 200 {object} response.Any{data=[]model.SubscriptionPlan}
// @Failure 401 {object} response.Any
// @Failure 500 {object} response.Any
// @Router /api/v1/custom/admin/plans [get]
func ListPlans(c *gin.Context) { listPlans(c, false) }

func listPlans(c *gin.Context, enabledOnly bool) {
	rows, err := repository.ListSubscriptionPlans(c.Request.Context(), enabledOnly)
	if err != nil {
		response.AbortInternal(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OK(rows))
}

// CreatePlan creates a subscription plan.
// @Summary 创建套餐
// @Tags custom-subscription
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param body body subscription.PlanInput true "套餐参数"
// @Success 200 {object} response.Any{data=model.SubscriptionPlan}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/admin/plans [post]
func CreatePlan(c *gin.Context) {
	var input PlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	row := planFromInput(input)
	if err := repository.CreateSubscriptionPlan(c.Request.Context(), row); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OK(row))
}

// UpdatePlan updates a subscription plan.
// @Summary 更新套餐
// @Tags custom-subscription
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param id path int true "套餐 ID"
// @Param body body subscription.PlanInput true "套餐参数"
// @Success 200 {object} response.Any{data=model.SubscriptionPlan}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/admin/plans/{id} [put]
func UpdatePlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.AbortBadRequest(c, "套餐 ID 无效")
		return
	}
	var input PlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	row, err := repository.GetSubscriptionPlanByID(c.Request.Context(), uint(id))
	if err == gorm.ErrRecordNotFound {
		response.AbortNotFound(c, errPlanNotFound)
		return
	}
	if err != nil {
		response.AbortInternal(c, err.Error())
		return
	}
	updated := planFromInput(input)
	updated.ID = row.ID
	if err := repository.SaveSubscriptionPlan(c.Request.Context(), updated); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OK(updated))
}

// DeletePlan deletes a subscription plan.
// @Summary 删除套餐
// @Tags custom-subscription
// @Produce json
// @Security SessionCookie
// @Param id path int true "套餐 ID"
// @Success 200 {object} response.Any
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/admin/plans/{id} [delete]
func DeletePlan(c *gin.Context) { deleteByID(c, true) }

// ListRedeemCodes returns all subscription redemption codes for administrators.
// @Summary 获取套餐兑换码
// @Tags custom-subscription
// @Produce json
// @Security SessionCookie
// @Success 200 {object} response.Any{data=[]model.RedeemCode}
// @Failure 401 {object} response.Any
// @Failure 500 {object} response.Any
// @Router /api/v1/custom/admin/redeem-codes [get]
func ListRedeemCodes(c *gin.Context) {
	rows, err := repository.ListRedeemCodes(c.Request.Context())
	if err != nil {
		response.AbortInternal(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OK(rows))
}

// CreateRedeemCode creates a one-time code for one month of the selected plan.
// @Summary 创建套餐兑换码
// @Tags custom-subscription
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param body body subscription.CreateRedeemCodeInput true "兑换码参数"
// @Success 200 {object} response.Any{data=model.RedeemCode}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 500 {object} response.Any
// @Router /api/v1/custom/admin/redeem-codes [post]
func CreateRedeemCode(c *gin.Context) {
	var input CreateRedeemCodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	row, err := GenerateRedeemCode(c.Request.Context(), input.PlanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == errPlanNotFound || err.Error() == errPlanDisabled {
			response.AbortBadRequest(c, err.Error())
			return
		}
		response.AbortInternal(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OK(row))
}

// ListPublicChannels returns enabled payment channels without secrets.
// @Summary 获取公开支付渠道
// @Tags custom-subscription
// @Produce json
// @Success 200 {object} response.Any{data=[]model.PaymentChannel}
// @Router /api/v1/custom/payment/channels [get]
func ListPublicChannels(c *gin.Context) { listChannels(c, true) }

// ListChannels returns all payment channels for administrators.
// @Summary 获取管理端支付渠道
// @Tags custom-subscription
// @Produce json
// @Security SessionCookie
// @Success 200 {object} response.Any{data=[]model.PaymentChannel}
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/admin/payment/channels [get]
func ListChannels(c *gin.Context) { listChannels(c, false) }

func listChannels(c *gin.Context, enabledOnly bool) {
	rows, err := repository.ListPaymentChannels(c.Request.Context(), enabledOnly)
	if err != nil {
		response.AbortInternal(c, err.Error())
		return
	}
	for i := range rows {
		rows[i].SecretKey = ""
	}
	c.JSON(http.StatusOK, response.OK(rows))
}

// CreateChannel creates an EasyPay channel.
// @Summary 创建易支付渠道
// @Tags custom-subscription
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param body body subscription.ChannelInput true "渠道参数"
// @Success 200 {object} response.Any{data=model.PaymentChannel}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/admin/payment/channels [post]
func CreateChannel(c *gin.Context) {
	var input ChannelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	row := channelFromInput(input)
	if err := repository.CreatePaymentChannel(c.Request.Context(), row); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	row.SecretKey = ""
	c.JSON(http.StatusOK, response.OK(row))
}

// UpdateChannel updates an EasyPay channel.
// @Summary 更新易支付渠道
// @Tags custom-subscription
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param id path int true "渠道 ID"
// @Param body body subscription.ChannelInput true "渠道参数"
// @Success 200 {object} response.Any{data=model.PaymentChannel}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/admin/payment/channels/{id} [put]
func UpdateChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.AbortBadRequest(c, "支付渠道 ID 无效")
		return
	}
	var input ChannelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	row, err := repository.GetPaymentChannelByID(c.Request.Context(), uint(id))
	if err == gorm.ErrRecordNotFound {
		response.AbortNotFound(c, errChannelNotFound)
		return
	}
	if err != nil {
		response.AbortInternal(c, err.Error())
		return
	}
	updated := channelFromInput(input)
	updated.ID = row.ID
	if updated.SecretKey == "" {
		updated.SecretKey = row.SecretKey
	}
	if err := repository.SavePaymentChannel(c.Request.Context(), updated); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	updated.SecretKey = ""
	c.JSON(http.StatusOK, response.OK(updated))
}

// DeleteChannel deletes an EasyPay channel.
// @Summary 删除易支付渠道
// @Tags custom-subscription
// @Produce json
// @Security SessionCookie
// @Param id path int true "渠道 ID"
// @Success 200 {object} response.Any
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/admin/payment/channels/{id} [delete]
func DeleteChannel(c *gin.Context) { deleteByID(c, false) }

func deleteByID(c *gin.Context, plan bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.AbortBadRequest(c, "ID 无效")
		return
	}
	if plan {
		err = repository.DeleteSubscriptionPlan(c.Request.Context(), uint(id))
	} else {
		err = repository.DeletePaymentChannel(c.Request.Context(), uint(id))
	}
	if err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OKNil())
}

type CreateOrderRequest struct {
	PlanID    uint `json:"plan_id" binding:"required"`
	ChannelID uint `json:"channel_id"`
}
type CreateOrderResponse struct {
	Order      *model.PaymentOrder `json:"order"`
	PaymentURL string              `json:"payment_url"`
}

// CreateMyOrder creates a payment order or activates a free plan.
// @Summary 创建套餐订单
// @Tags custom-subscription
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param body body subscription.CreateOrderRequest true "订单参数"
// @Success 200 {object} response.Any{data=subscription.CreateOrderResponse}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/orders [post]
func CreateMyOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	base := requestBaseURL(c)
	paymentURL, order, err := CreateOrder(c.Request.Context(), oauth.GetUserIDFromContext(c), req.PlanID, req.ChannelID, base+"/payment/easy-pay/notify/"+strconv.Itoa(int(req.ChannelID)), base+"/payment/result")
	if err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OK(CreateOrderResponse{Order: order, PaymentURL: paymentURL}))
}

// GetMySubscription returns the current user's active subscription.
// @Summary 获取我的套餐
// @Tags custom-subscription
// @Produce json
// @Security SessionCookie
// @Success 200 {object} response.Any{data=model.UserSubscription}
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/subscription [get]
func GetMySubscription(c *gin.Context) {
	row, err := repository.GetActiveSubscription(c.Request.Context(), oauth.GetUserIDFromContext(c), time.Now())
	if err == gorm.ErrRecordNotFound {
		response.AbortNotFound(c, errNoSubscription)
		return
	}
	if err != nil {
		response.AbortInternal(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OK(row))
}

// ListMyOrders returns the current user's payment orders.
// @Summary 获取我的支付订单
// @Tags custom-subscription
// @Produce json
// @Security SessionCookie
// @Success 200 {object} response.Any{data=[]model.PaymentOrder}
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/orders [get]
func ListMyOrders(c *gin.Context) {
	rows, err := repository.ListPaymentOrdersByUser(c.Request.Context(), oauth.GetUserIDFromContext(c))
	if err != nil {
		response.AbortInternal(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OK(rows))
}

// Redeem exchanges a one-time code for one month of its subscription plan.
// @Summary 兑换套餐码
// @Tags custom-subscription
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param body body subscription.RedeemRequest true "兑换码"
// @Success 200 {object} response.Any{data=model.UserSubscription}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/redeem [post]
func Redeem(c *gin.Context) {
	var req RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	subscription, err := RedeemPlan(c.Request.Context(), oauth.GetUserIDFromContext(c), req.Code, time.Now())
	if err != nil {
		if err.Error() == errRedeemCodeInvalid {
			response.AbortBadRequest(c, err.Error())
			return
		}
		response.AbortInternal(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OK(subscription))
}

// Notify receives an EasyPay asynchronous payment callback.
// @Summary 接收易支付回调
// @Tags custom-subscription
// @Accept application/x-www-form-urlencoded
// @Produce plain
// @Param channel_id path int true "渠道 ID"
// @Success 200 {string} string "success"
// @Failure 400 {string} string "fail"
// @Router /payment/easy-pay/notify/{channel_id} [post]
func Notify(c *gin.Context) {
	channelID, err := strconv.ParseUint(c.Param("channel_id"), 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}
	if err := c.Request.ParseForm(); err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}
	if err := HandleNotify(c.Request.Context(), uint(channelID), c.Request.Form); err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}
	c.String(http.StatusOK, "success")
}

func requestBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}
