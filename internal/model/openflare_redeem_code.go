// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

//nolint:revive // Exported domain API is consumed across the application layers.
package model

import "time"

const (
	RedeemCodeStatusUnused = "unused"
	RedeemCodeStatusUsed   = "used"
)

// RedeemCode grants one month of a subscription plan to one user.
type RedeemCode struct {
	ID        uint              `json:"id" gorm:"primaryKey;autoIncrement"`
	Code      string            `json:"code" gorm:"size:64;uniqueIndex;not null"`
	PlanID    uint              `json:"plan_id" gorm:"not null;index"`
	Status    string            `json:"status" gorm:"size:16;not null;index"`
	UsedBy    *uint64           `json:"used_by,omitempty" gorm:"index"`
	UsedAt    *time.Time        `json:"used_at,omitempty"`
	Plan      *SubscriptionPlan `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
	CreatedAt time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
}

func (RedeemCode) TableName() string { return "of_redeem_codes" }
