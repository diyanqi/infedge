// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"strconv"
	"strings"
	"time"

	db "github.com/Rain-kl/Wavelet/internal/infra/persistence"
	"github.com/Rain-kl/Wavelet/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const repositoryUpdatedAtColumn = "updated_at"

// ListUserTrafficUsage returns monthly usage keyed by user ID.
func ListUserTrafficUsage(ctx context.Context, monthStart time.Time) (map[uint64]int64, error) {
	var rows []model.UserTrafficMonthly
	if err := db.DB(ctx).Where("month_start = ?", monthStart.UTC()).Find(&rows).Error; err != nil {
		return nil, err
	}
	usage := make(map[uint64]int64, len(rows))
	for _, row := range rows {
		usage[row.UserID] = row.BytesSent
	}
	return usage, nil
}

// AddUserTrafficUsage atomically folds newly delivered access-log bytes into calendar months.
func AddUserTrafficUsage(ctx context.Context, records []*model.OpenFlareAccessLog) error {
	type aggregate struct {
		userID     uint64
		monthStart time.Time
		bytes      int64
	}
	aggregates := make(map[string]aggregate)
	for _, record := range records {
		if record == nil || record.OwnerID == 0 || record.BytesSent <= 0 {
			continue
		}
		loggedAt := record.LoggedAt.UTC()
		monthStart := time.Date(loggedAt.Year(), loggedAt.Month(), 1, 0, 0, 0, 0, time.UTC)
		key := strings.Join([]string{formatUint(record.OwnerID), monthStart.Format("2006-01-02")}, "|")
		item := aggregates[key]
		item.userID = record.OwnerID
		item.monthStart = monthStart
		item.bytes += record.BytesSent
		aggregates[key] = item
	}
	if len(aggregates) == 0 {
		return nil
	}
	return db.DB(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range aggregates {
			row := &model.UserTrafficMonthly{UserID: item.userID, MonthStart: item.monthStart, BytesSent: item.bytes}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "user_id"}, {Name: "month_start"}},
				DoUpdates: clause.Assignments(map[string]any{
					"bytes_sent":              gorm.Expr("bytes_sent + EXCLUDED.bytes_sent"),
					repositoryUpdatedAtColumn: time.Now().UTC(),
				}),
			}).Create(row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func formatUint(value uint64) string {
	return strconv.FormatUint(value, 10)
}
