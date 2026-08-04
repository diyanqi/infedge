// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package model

import "time"

// UserTrafficMonthly stores the authoritative monthly delivered bytes for a user.
type UserTrafficMonthly struct {
	ID         uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     uint64    `json:"user_id" gorm:"uniqueIndex:idx_of_user_traffic_monthly_user_month"`
	MonthStart time.Time `json:"month_start" gorm:"uniqueIndex:idx_of_user_traffic_monthly_user_month"`
	BytesSent  int64     `json:"bytes_sent" gorm:"not null;default:0"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the monthly traffic table name.
func (UserTrafficMonthly) TableName() string { return "of_user_traffic_monthly" }
