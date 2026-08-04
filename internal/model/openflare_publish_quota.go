// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package model

import "time"

// UserPublishDailyCounter reserves a user's daily publish entitlements atomically.
type UserPublishDailyCounter struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint64    `json:"user_id" gorm:"uniqueIndex:idx_of_user_publish_daily_user_day"`
	DayStart  time.Time `json:"day_start" gorm:"uniqueIndex:idx_of_user_publish_daily_user_day"`
	Used      int       `json:"used" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the daily publish counter table name.
func (UserPublishDailyCounter) TableName() string { return "of_user_publish_daily_counters" }
