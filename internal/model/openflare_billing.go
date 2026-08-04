// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

//nolint:revive // Exported domain API is consumed across the application layers.
package model

import "time"

const (
	SubscriptionStatusActive  = "active"
	SubscriptionStatusExpired = "expired"
	SubscriptionStatusPending = "pending"
	PaymentOrderPending       = "pending"
	PaymentOrderPaid          = "paid"
	PaymentOrderClosed        = "closed"
)

// SubscriptionPlan defines user quotas and limits.
type SubscriptionPlan struct {
	ID                  uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name                string    `json:"name" gorm:"size:128;not null"`
	Description         string    `json:"description" gorm:"type:text;not null;default:''"`
	PriceFen            int64     `json:"price_fen" gorm:"not null;default:0"`
	BillingMonths       int       `json:"billing_months" gorm:"not null;default:1"`
	HighSpeedBytes      int64     `json:"high_speed_bytes" gorm:"not null;default:0"`
	ThrottleBytesPerSec int64     `json:"throttle_bytes_per_sec" gorm:"not null;default:0"`
	DailyPublishLimit   int       `json:"daily_publish_limit" gorm:"not null;default:1"`
	MaxZones            int       `json:"max_zones" gorm:"not null;default:1"`
	MaxOrigins          int       `json:"max_origins" gorm:"not null;default:1"`
	MaxRoutes           int       `json:"max_routes" gorm:"not null;default:1"`
	MaxPages            int       `json:"max_pages" gorm:"not null;default:0"`
	Enabled             bool      `json:"enabled" gorm:"not null;default:true;index"`
	CreatedAt           time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (SubscriptionPlan) TableName() string { return "of_subscription_plans" }

// UserSubscription is the current or historical plan assignment for a user.
type UserSubscription struct {
	ID        uint              `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint64            `json:"user_id" gorm:"not null;index"`
	PlanID    uint              `json:"plan_id" gorm:"not null;index"`
	Status    string            `json:"status" gorm:"size:32;not null;index"`
	StartsAt  time.Time         `json:"starts_at"`
	ExpiresAt time.Time         `json:"expires_at"`
	Plan      *SubscriptionPlan `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
	CreatedAt time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
}

func (UserSubscription) TableName() string { return "of_user_subscriptions" }

// PaymentChannel stores one EasyPay-compatible merchant channel.
type PaymentChannel struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"size:128;not null"`
	Gateway   string    `json:"gateway" gorm:"size:512;not null"`
	PID       string    `json:"pid" gorm:"column:pid;size:128;not null"`
	SecretKey string    `json:"-" gorm:"size:255;not null"`
	Enabled   bool      `json:"enabled" gorm:"not null;default:true;index"`
	Sort      int       `json:"sort" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (PaymentChannel) TableName() string { return "of_payment_channels" }

// PaymentOrder is an idempotent EasyPay order.
type PaymentOrder struct {
	ID        uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderNo   string     `json:"order_no" gorm:"size:64;uniqueIndex;not null"`
	UserID    uint64     `json:"user_id" gorm:"not null;index"`
	PlanID    uint       `json:"plan_id" gorm:"not null;index"`
	ChannelID uint       `json:"channel_id" gorm:"not null;index"`
	AmountFen int64      `json:"amount_fen" gorm:"not null;default:0"`
	Status    string     `json:"status" gorm:"size:32;not null;index"`
	TradeNo   string     `json:"trade_no" gorm:"size:128;not null;default:''"`
	PaidAt    *time.Time `json:"paid_at"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (PaymentOrder) TableName() string { return "of_payment_orders" }
