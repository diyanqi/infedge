// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package node_group manages shared node traffic quota groups.
//
//nolint:revive // Exported domain API is consumed across the application layers.
package node_group

import (
	"net/http"
	"strconv"

	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/Rain-kl/Wavelet/internal/repository"
	"github.com/Rain-kl/Wavelet/internal/shared/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Input struct {
	Name              string `json:"name" binding:"required,max=128"`
	MonthlyBytesLimit int64  `json:"monthly_bytes_limit" binding:"gte=0"`
	NodeIDs           []uint `json:"node_ids"`
}

func RegisterRoutes(api *gin.RouterGroup) {
	api.GET("", List)
	api.POST("", Create)
	api.PUT("/:id", Update)
	api.DELETE("/:id", Delete)
}

// List returns all node groups and their members.
// @Summary 获取节点组列表
// @Tags openflare-node-group
// @Produce json
// @Security SessionCookie
// @Success 200 {object} response.Any{data=[]model.OpenFlareNodeGroup}
// @Failure 401 {object} response.Any
// @Router /api/v1/d/node-groups [get]
func List(c *gin.Context) {
	rows, err := repository.ListNodeGroups(c.Request.Context())
	if err != nil {
		response.AbortInternal(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OK(rows))
}

// Create creates a node group with a shared monthly traffic limit.
// @Summary 创建节点组
// @Tags openflare-node-group
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param body body node_group.Input true "节点组参数"
// @Success 200 {object} response.Any{data=model.OpenFlareNodeGroup}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/d/node-groups [post]
func Create(c *gin.Context) {
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	row := &model.OpenFlareNodeGroup{Name: input.Name, MonthlyBytesLimit: input.MonthlyBytesLimit}
	nodes := make([]model.OpenFlareNode, 0, len(input.NodeIDs))
	for _, id := range input.NodeIDs {
		nodes = append(nodes, model.OpenFlareNode{ID: id})
	}
	row.Nodes = nodes
	if err := repository.CreateNodeGroup(c.Request.Context(), row, input.NodeIDs); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OK(row))
}

// Update updates a node group and its members.
// @Summary 更新节点组
// @Tags openflare-node-group
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param id path int true "节点组 ID"
// @Param body body node_group.Input true "节点组参数"
// @Success 200 {object} response.Any{data=model.OpenFlareNodeGroup}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/d/node-groups/{id} [put]
func Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.AbortBadRequest(c, "节点组 ID 无效")
		return
	}
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	row, err := repository.GetNodeGroupByID(c.Request.Context(), id)
	if err == gorm.ErrRecordNotFound {
		response.AbortNotFound(c, "节点组不存在")
		return
	}
	if err != nil {
		response.AbortInternal(c, err.Error())
		return
	}
	row.Name, row.MonthlyBytesLimit = input.Name, input.MonthlyBytesLimit
	row.Nodes = make([]model.OpenFlareNode, 0, len(input.NodeIDs))
	for _, nodeID := range input.NodeIDs {
		row.Nodes = append(row.Nodes, model.OpenFlareNode{ID: nodeID})
	}
	if err := repository.SaveNodeGroup(c.Request.Context(), row); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OK(row))
}

// Delete removes a node group and clears its node memberships.
// @Summary 删除节点组
// @Tags openflare-node-group
// @Produce json
// @Security SessionCookie
// @Param id path int true "节点组 ID"
// @Success 200 {object} response.Any
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/d/node-groups/{id} [delete]
func Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.AbortBadRequest(c, "节点组 ID 无效")
		return
	}
	if err := repository.DeleteNodeGroup(c.Request.Context(), id); err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, response.OKNil())
}

func parseID(c *gin.Context) (uint, error) {
	raw, err := strconv.ParseUint(c.Param("id"), 10, 32)
	return uint(raw), err
}
