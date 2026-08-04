// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

//nolint:revive // Exported domain API is consumed across the application layers.
package model

import "time"

// OpenFlareNodeGroup groups edge nodes under a shared monthly quota.
type OpenFlareNodeGroup struct {
	ID                uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	Name              string          `json:"name" gorm:"size:128;uniqueIndex;not null"`
	MonthlyBytesLimit int64           `json:"monthly_bytes_limit" gorm:"not null;default:0"`
	CreatedAt         time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	Nodes             []OpenFlareNode `json:"nodes,omitempty" gorm:"many2many:of_node_group_nodes"`
}

func (OpenFlareNodeGroup) TableName() string { return "of_node_groups" }

// OpenFlareNodeGroupNode is the explicit join entity for node membership.
type OpenFlareNodeGroupNode struct {
	NodeGroupID uint      `json:"node_group_id" gorm:"primaryKey"`
	NodeID      uint      `json:"node_id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (OpenFlareNodeGroupNode) TableName() string { return "of_node_group_nodes" }
