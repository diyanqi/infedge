// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

//nolint:revive // Exported domain API is consumed across the application layers.
package repository

import (
	"context"
	"strconv"
	"time"

	db "github.com/Rain-kl/Wavelet/internal/infra/persistence"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/Rain-kl/Wavelet/pkg/protocol"
	"gorm.io/gorm"
)

func ListNodeGroups(ctx context.Context) ([]model.OpenFlareNodeGroup, error) {
	var rows []model.OpenFlareNodeGroup
	err := db.DB(ctx).Preload("Nodes").Order("id asc").Find(&rows).Error
	return rows, err
}

func GetNodeGroupByID(ctx context.Context, id uint) (*model.OpenFlareNodeGroup, error) {
	var row model.OpenFlareNodeGroup
	err := db.DB(ctx).Preload("Nodes").First(&row, id).Error
	return &row, err
}

func SaveNodeGroup(ctx context.Context, row *model.OpenFlareNodeGroup) error {
	return db.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(row).Error; err != nil {
			return err
		}
		if err := tx.Where("node_group_id = ?", row.ID).Delete(&model.OpenFlareNodeGroupNode{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.OpenFlareNode{}).Where("node_group_id = ?", row.ID).Update("node_group_id", nil).Error; err != nil {
			return err
		}
		for _, node := range row.Nodes {
			if err := tx.Create(&model.OpenFlareNodeGroupNode{NodeGroupID: row.ID, NodeID: node.ID}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.OpenFlareNode{}).Where("id = ?", node.ID).Update("node_group_id", row.ID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func CreateNodeGroup(ctx context.Context, row *model.OpenFlareNodeGroup, nodeIDs []uint) error {
	return db.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		for _, nodeID := range nodeIDs {
			if err := tx.Create(&model.OpenFlareNodeGroupNode{NodeGroupID: row.ID, NodeID: nodeID}).Error; err != nil {
				return err
			}
		}
		if len(nodeIDs) > 0 {
			return tx.Model(&model.OpenFlareNode{}).Where("id IN ?", nodeIDs).Update("node_group_id", row.ID).Error
		}
		return nil
	})
}

func DeleteNodeGroup(ctx context.Context, id uint) error {
	return db.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.OpenFlareNode{}).Where("node_group_id = ?", id).Update("node_group_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.OpenFlareNodeGroupNode{}, "node_group_id = ?", id).Error; err != nil {
			return err
		}
		return tx.Delete(&model.OpenFlareNodeGroup{}, id).Error
	})
}

// BuildNodeTrafficQuota returns the server-side monthly policy for a node.
func BuildNodeTrafficQuota(ctx context.Context, node *model.OpenFlareNode, now time.Time) (*protocol.TrafficQuota, error) {
	if node == nil {
		return nil, nil
	}
	utcNow := now.UTC()
	monthStart := time.Date(utcNow.Year(), utcNow.Month(), 1, 0, 0, 0, 0, time.UTC)
	quota := &protocol.TrafficQuota{}
	quota.HighSpeedLimitBytes = node.MonthlyBytesLimit
	used, usedErr := sumNodeTraffic(ctx, []string{node.NodeID}, monthStart, now)
	if usedErr != nil {
		return nil, usedErr
	}
	quota.RemainingBytes = remainingBytes(quota.HighSpeedLimitBytes, used)
	if node.NodeGroupID != nil && *node.NodeGroupID != 0 {
		group, groupErr := GetNodeGroupByID(ctx, *node.NodeGroupID)
		if groupErr != nil {
			return nil, groupErr
		}
		quota.GroupID = group.ID
		quota.GroupLimitBytes = group.MonthlyBytesLimit
		var nodeIDs []string
		for _, groupNode := range group.Nodes {
			nodeIDs = append(nodeIDs, groupNode.NodeID)
		}
		used, usedErr := sumNodeTraffic(ctx, nodeIDs, monthStart, now)
		if usedErr != nil {
			return nil, usedErr
		}
		quota.GroupRemainingBytes = remainingBytes(group.MonthlyBytesLimit, used)
	}
	users, userErr := buildUserTrafficQuotas(ctx, monthStart, now)
	if userErr != nil {
		return nil, userErr
	}
	quota.Users = users
	return quota, nil
}

func remainingBytes(limit, used int64) int64 {
	if limit <= 0 {
		return 0
	}
	if used >= limit {
		return 0
	}
	return limit - used
}

func sumNodeTraffic(ctx context.Context, nodeIDs []string, since, until time.Time) (int64, error) {
	var total int64
	for _, nodeID := range nodeIDs {
		if nodeID == "" {
			continue
		}
		_, _, bytesSent, err := CountOpenFlareAccessLogs(ctx, model.OpenFlareAccessLogQuery{
			NodeID: nodeID, Since: since, Until: until, Page: 1, PageSize: 1,
		})
		if err != nil {
			return 0, err
		}
		total += bytesSent
	}
	return total, nil
}

func buildUserTrafficQuotas(ctx context.Context, monthStart, now time.Time) (map[string]protocol.UserTrafficQuota, error) {
	subscriptions, err := ListActiveSubscriptionPlans(ctx, now)
	if err != nil {
		return nil, err
	}
	usage, err := ListUserTrafficUsage(ctx, monthStart)
	if err != nil {
		return nil, err
	}
	nodes, err := ListOpenFlareNodes(ctx)
	if err != nil {
		return nil, err
	}
	onlineNodes := 0
	for _, item := range nodes {
		if item.Status == openFlareNodeStatusOnline {
			onlineNodes++
		}
	}
	if onlineNodes == 0 {
		onlineNodes = 1
	}
	quotas := make(map[string]protocol.UserTrafficQuota, len(subscriptions))
	for _, subscription := range subscriptions {
		if subscription.Plan == nil {
			continue
		}
		plan := subscription.Plan
		used := usage[subscription.UserID]
		quotas[strconv.FormatUint(subscription.UserID, 10)] = protocol.UserTrafficQuota{
			UserID:                   subscription.UserID,
			HighSpeedLimitBytes:      plan.HighSpeedBytes,
			RemainingBytes:           remainingBytes(plan.HighSpeedBytes, used),
			ThrottleBytesPerSec:      plan.ThrottleBytesPerSec,
			AllocatedRateBytesPerSec: divideRate(plan.ThrottleBytesPerSec, onlineNodes),
		}
	}
	return quotas, nil
}

func divideRate(rate int64, nodes int) int64 {
	if rate <= 0 || nodes <= 1 {
		return rate
	}
	return max(rate/int64(nodes), 1)
}
