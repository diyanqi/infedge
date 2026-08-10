// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package useraccess exposes owner-scoped OpenFlare resources to ordinary users.
//
//nolint:revive // Exported HTTP handlers are registered by the router package.
package useraccess

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Rain-kl/Wavelet/internal/apps/oauth"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/config_version"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/origin"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/proxy_route"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/zone"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/Rain-kl/Wavelet/internal/repository"
	"github.com/Rain-kl/Wavelet/internal/shared/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	errNoPlan        = "请先选择有效套餐"
	errQuotaExceeded = "已达到当前套餐资源额度"
)

func RegisterRoutes(api *gin.RouterGroup) {
	group := api.Group("/resources", oauth.LoginRequired())
	group.GET("/zones", ListZones)
	group.POST("/zones", CreateZone)
	group.GET("/zones/:id", GetZone)
	group.POST("/zones/:id/update", UpdateZone)
	group.POST("/zones/:id/delete", DeleteZone)
	group.POST("/zones/:id/domains", CreateDomain)
	group.POST("/zones/:id/domains/:domain_id/update", UpdateDomain)
	group.POST("/zones/:id/verify", VerifyZone)
	group.POST("/zones/:id/domains/:domain_id/verify", VerifyDomain)
	group.POST("/sites", CreateSite)
	group.POST("/sites/:id/verify", VerifySite)
	group.GET("/origins", ListOrigins)
	group.POST("/origins", CreateOrigin)
	group.POST("/origins/:id/update", UpdateOrigin)
	group.POST("/origins/:id/delete", DeleteOrigin)
	group.GET("/proxy-routes", ListRoutes)
	group.GET("/proxy-routes/:id", GetRoute)
	group.POST("/proxy-routes", CreateRoute)
	group.POST("/proxy-routes/:id/update", UpdateRoute)
	group.POST("/proxy-routes/:id/delete", DeleteRoute)
	group.POST("/publish", Publish)
	registerPagesRoutes(group)
	registerWAFRoutes(group)
	registerTLSRoutes(group)
}

func userID(c *gin.Context) uint64 { return oauth.GetUserIDFromContext(c) }

func planForUser(ctx context.Context, id uint64) (*model.SubscriptionPlan, error) {
	current, err := repository.GetActiveSubscription(ctx, id, time.Now())
	if err == nil && current.Plan != nil && current.Plan.Enabled {
		return current.Plan, nil
	}
	return nil, errNoPlanError()
}

func errNoPlanError() error { return errors.New(errNoPlan) }

func checkQuota(ctx context.Context, userID uint64, resource string, quotaLimit int) error {
	if quotaLimit <= 0 {
		return nil
	}
	count, err := repository.CountOwnedResources(ctx, resource, userID)
	if err != nil {
		return err
	}
	if count >= int64(quotaLimit) {
		return errors.New(errQuotaExceeded)
	}
	return nil
}

// ListZones lists the current user's Zones.
// @Summary 获取我的 Zone 列表
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Success 200 {object} response.Any{data=[]zone.ListItem}
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/resources/zones [get]
func ListZones(c *gin.Context) {
	rows, err := zone.ListOwned(c.Request.Context(), userID(c))
	respond(c, rows, err)
}

// CreateZone creates a Zone owned by the current user.
// @Summary 创建我的 Zone
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param body body zone.Input true "Zone 参数"
// @Success 200 {object} response.Any{data=model.Zone}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 409 {object} response.Any
// @Router /api/v1/custom/resources/zones [post]
func CreateZone(c *gin.Context) {
	ctx, id := c.Request.Context(), userID(c)
	plan, err := planForUser(ctx, id)
	if err == nil {
		err = checkQuota(ctx, id, "zones", plan.MaxZones)
	}
	if err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	var input zone.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	row, err := zone.CreateOwned(ctx, id, input)
	respond(c, row, err)
}

// CreateSite onboards one exact root or subdomain in a single request.
// @Summary 直接创建我的 CDN 域名
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param body body zone.SiteInput true "域名参数"
// @Success 200 {object} response.Any{data=zone.Site}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 409 {object} response.Any
// @Router /api/v1/custom/resources/sites [post]
func CreateSite(c *gin.Context) {
	ctx, uid := c.Request.Context(), userID(c)
	var input zone.SiteInput
	if !bind(c, &input) {
		return
	}
	hasRoot, err := zone.HasOwnedRoot(ctx, uid, input.Domain)
	if err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	if !hasRoot {
		plan, planErr := planForUser(ctx, uid)
		if planErr == nil {
			planErr = checkQuota(ctx, uid, "zones", plan.MaxZones)
		}
		if planErr != nil {
			response.AbortBadRequest(c, planErr.Error())
			return
		}
	}
	row, err := zone.CreateOwnedSite(ctx, uid, input)
	respond(c, row, err)
}

// VerifySite verifies the exact domain created by CreateSite.
// @Summary 验证我的 CDN 域名
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Param id path int true "域名 ID"
// @Success 200 {object} response.Any{data=model.ZoneDomain}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/sites/{id}/verify [post]
func VerifySite(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	row, err := zone.VerifyOwnedSiteDomain(c.Request.Context(), id, userID(c))
	respond(c, row, err)
}

// GetZone returns an owned Zone and its domains.
// @Summary 获取我的 Zone 详情
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Param id path int true "Zone ID"
// @Success 200 {object} response.Any{data=zone.Overview}
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/zones/{id} [get]
func GetZone(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	row, err := zone.GetOwnedOverview(c.Request.Context(), id, userID(c))
	respond(c, row, err)
}

// UpdateZone updates an owned Zone.
// @Summary 更新我的 Zone
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param id path int true "Zone ID"
// @Param body body zone.Input true "Zone 参数"
// @Success 200 {object} response.Any{data=model.Zone}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Failure 409 {object} response.Any
// @Router /api/v1/custom/resources/zones/{id}/update [post]
func UpdateZone(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input zone.Input
	if !bind(c, &input) {
		return
	}
	row, err := zone.UpdateOwned(c.Request.Context(), id, userID(c), input)
	respond(c, row, err)
}

// DeleteZone deletes an owned Zone.
// @Summary 删除我的 Zone
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Param id path int true "Zone ID"
// @Success 200 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/zones/{id}/delete [post]
func DeleteZone(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	respondNil(c, zone.DeleteOwned(c.Request.Context(), id, userID(c)))
}

// CreateDomain adds a domain to an owned Zone.
// @Summary 添加我的域名
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param id path int true "Zone ID"
// @Param body body zone.DomainInput true "域名参数"
// @Success 200 {object} response.Any{data=model.ZoneDomain}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Failure 409 {object} response.Any
// @Router /api/v1/custom/resources/zones/{id}/domains [post]
func CreateDomain(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input zone.DomainInput
	if !bind(c, &input) {
		return
	}
	if input.CertID != nil {
		if _, err := repository.GetOwnedTLSCertificateByID(c.Request.Context(), *input.CertID, userID(c)); err != nil {
			response.AbortNotFound(c, "证书不存在")
			return
		}
	}
	row, err := zone.CreateOwnedDomain(c.Request.Context(), id, userID(c), input)
	respond(c, row, err)
}

// UpdateDomain updates the certificate assigned to an owned child domain.
// @Summary 更新我的子域配置
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param id path int true "Zone ID"
// @Param domain_id path int true "域名 ID"
// @Param body body zone.DomainInput true "域名参数"
// @Success 200 {object} response.Any{data=model.ZoneDomain}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Failure 409 {object} response.Any
// @Router /api/v1/custom/resources/zones/{id}/domains/{domain_id}/update [post]
func UpdateDomain(c *gin.Context) {
	zoneID, ok := parseID(c)
	if !ok {
		return
	}
	domainID, err := strconv.ParseUint(c.Param("domain_id"), 10, 32)
	if err != nil || domainID == 0 {
		response.AbortBadRequest(c, "ID 无效")
		return
	}
	var input zone.DomainInput
	if !bind(c, &input) {
		return
	}
	if input.CertID != nil {
		if _, err := repository.GetOwnedTLSCertificateByID(c.Request.Context(), *input.CertID, userID(c)); err != nil {
			response.AbortNotFound(c, "证书不存在")
			return
		}
	}
	row, err := zone.UpdateOwnedDomain(c.Request.Context(), zoneID, uint(domainID), userID(c), input)
	respond(c, row, err)
}

// VerifyZone verifies a root domain through DNS TXT.
func VerifyZone(c *gin.Context) {
	zoneID, ok := parseID(c)
	if !ok {
		return
	}
	row, err := zone.VerifyOwnedZone(c.Request.Context(), zoneID, userID(c))
	respond(c, row, err)
}

// VerifyDomain verifies a child domain through DNS TXT.
func VerifyDomain(c *gin.Context) {
	zoneID, ok := parseID(c)
	if !ok {
		return
	}
	domainID, err := strconv.ParseUint(c.Param("domain_id"), 10, 32)
	if err != nil {
		response.AbortBadRequest(c, "ID 无效")
		return
	}
	row, verifyErr := zone.VerifyOwnedDomain(c.Request.Context(), zoneID, uint(domainID), userID(c))
	respond(c, row, verifyErr)
}

// ListOrigins lists the current user's origins.
// @Summary 获取我的源站列表
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Success 200 {object} response.Any{data=[]model.Origin}
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/resources/origins [get]
func ListOrigins(c *gin.Context) {
	rows, err := origin.ListOwnedOrigins(c.Request.Context(), userID(c))
	respond(c, rows, err)
}

// CreateOrigin creates an origin owned by the current user.
// @Summary 创建我的源站
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param body body origin.Input true "源站参数"
// @Success 200 {object} response.Any{data=model.Origin}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/resources/origins [post]
func CreateOrigin(c *gin.Context) {
	ctx, id := c.Request.Context(), userID(c)
	plan, err := planForUser(ctx, id)
	if err == nil {
		err = checkQuota(ctx, id, "origins", plan.MaxOrigins)
	}
	if err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	var input origin.Input
	if !bind(c, &input) {
		return
	}
	row, err := origin.CreateOwnedOrigin(ctx, id, input)
	respond(c, row, err)
}

// UpdateOrigin updates an owned origin.
// @Summary 更新我的源站
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param id path int true "源站 ID"
// @Param body body origin.Input true "源站参数"
// @Success 200 {object} response.Any{data=model.Origin}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/origins/{id}/update [post]
func UpdateOrigin(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input origin.Input
	if !bind(c, &input) {
		return
	}
	row, err := origin.UpdateOwnedOrigin(c.Request.Context(), id, userID(c), input)
	respond(c, row, err)
}

// DeleteOrigin deletes an owned origin.
// @Summary 删除我的源站
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Param id path int true "源站 ID"
// @Success 200 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/origins/{id}/delete [post]
func DeleteOrigin(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	respondNil(c, origin.DeleteOwnedOrigin(c.Request.Context(), id, userID(c)))
}

// ListRoutes lists the current user's proxy routes.
// @Summary 获取我的 CDN 规则
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Success 200 {object} response.Any{data=[]proxy_route.View}
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/resources/proxy-routes [get]
func ListRoutes(c *gin.Context) {
	rows, err := proxy_route.ListOwnedProxyRoutes(c.Request.Context(), userID(c))
	respond(c, rows, err)
}

// GetRoute returns an owned proxy route.
// @Summary 获取我的 CDN 规则详情
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Param id path int true "规则 ID"
// @Success 200 {object} response.Any{data=proxy_route.View}
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/proxy-routes/{id} [get]
func GetRoute(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	row, err := proxy_route.GetOwnedProxyRoute(c.Request.Context(), id, userID(c))
	respond(c, row, err)
}

// CreateRoute creates a proxy route using only owned references.
// @Summary 创建我的 CDN 规则
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param body body proxy_route.Input true "规则参数"
// @Success 200 {object} response.Any{data=proxy_route.View}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/resources/proxy-routes [post]
func CreateRoute(c *gin.Context) {
	ctx, uid := c.Request.Context(), userID(c)
	plan, err := planForUser(ctx, uid)
	if err == nil {
		err = checkQuota(ctx, uid, "proxy_routes", plan.MaxRoutes)
	}
	var input proxy_route.Input
	if err == nil && !bind(c, &input) {
		return
	}
	if err == nil {
		err = validateReferences(ctx, uid, input)
	}
	if err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	row, err := proxy_route.CreateOwnedProxyRoute(ctx, uid, input)
	respond(c, row, err)
}

// UpdateRoute updates an owned proxy route.
// @Summary 更新我的 CDN 规则
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param id path int true "规则 ID"
// @Param body body proxy_route.Input true "规则参数"
// @Success 200 {object} response.Any{data=proxy_route.View}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/proxy-routes/{id}/update [post]
func UpdateRoute(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input proxy_route.Input
	if !bind(c, &input) {
		return
	}
	if err := validateReferences(c.Request.Context(), userID(c), input); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	row, err := proxy_route.UpdateOwnedProxyRoute(c.Request.Context(), id, userID(c), input)
	respond(c, row, err)
}

// DeleteRoute deletes an owned proxy route.
// @Summary 删除我的 CDN 规则
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Param id path int true "规则 ID"
// @Success 200 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/proxy-routes/{id}/delete [post]
func DeleteRoute(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	respondNil(c, proxy_route.DeleteOwnedProxyRoute(c.Request.Context(), id, userID(c)))
}

// Publish publishes the global configuration on behalf of the current user.
// @Summary 发布全站配置
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Success 200 {object} response.Any{data=model.ConfigVersion}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/resources/publish [post]
func Publish(c *gin.Context) {
	uid := userID(c)
	if err := zone.EnsureOwnedDomainsReady(c.Request.Context(), uid); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	row, err := config_version.PublishForUser(c.Request.Context(), uid, strconv.FormatUint(uid, 10), false)
	respond(c, row, err)
}

func validateReferences(ctx context.Context, uid uint64, input proxy_route.Input) error {
	owned, err := repository.AreZoneDomainsOwned(ctx, input.ZoneDomainIDs, uid)
	if err != nil {
		return err
	}
	if !owned {
		return errors.New("域名不属于当前用户")
	}
	domains, err := repository.ListZoneDomainsByIDs(ctx, input.ZoneDomainIDs)
	if err != nil {
		return err
	}
	for _, domain := range domains {
		if domain.VerificationStatus != "verified" {
			return errors.New("域名 " + domain.Domain + " 尚未完成 DNS TXT 所有权验证")
		}
		if !input.EnableHTTPS || domain.CertID == nil || *domain.CertID == 0 {
			continue
		}
		if _, err := repository.GetOwnedTLSCertificateByID(ctx, *domain.CertID, uid); err != nil {
			return errors.New("HTTPS 证书不属于当前用户")
		}
	}
	if input.OriginID != nil {
		if _, err := repository.GetOwnedOriginByID(ctx, *input.OriginID, uid); err != nil {
			return errors.New("源站不属于当前用户")
		}
	}
	if input.PagesProjectID != nil && *input.PagesProjectID != 0 {
		if _, err := repository.GetPagesProjectByIDAndOwner(ctx, *input.PagesProjectID, uid); err != nil {
			return errors.New("pages 项目不属于当前用户")
		}
	}
	return nil
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.AbortBadRequest(c, "ID 无效")
		return 0, false
	}
	return uint(id), true
}
func bind(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		response.AbortBadRequest(c, err.Error())
		return false
	}
	return true
}
func respond(c *gin.Context, data any, err error) {
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			response.AbortNotFound(c, "资源不存在")
		case zone.IsConflict(err):
			response.AbortConflict(c, err.Error())
		default:
			response.AbortBadRequest(c, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, response.OK(data))
}
func respondNil(c *gin.Context, err error) {
	if err != nil {
		respond(c, nil, err)
		return
	}
	c.JSON(http.StatusOK, response.OKNil())
}
