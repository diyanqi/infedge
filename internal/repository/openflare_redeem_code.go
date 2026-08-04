// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

//nolint:revive // Exported domain API is consumed across the application layers.
package repository

import (
	"context"
	"time"

	db "github.com/Rain-kl/Wavelet/internal/infra/persistence"
	"github.com/Rain-kl/Wavelet/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateRedeemCode(ctx context.Context, row *model.RedeemCode) error {
	return db.DB(ctx).Create(row).Error
}

func ListRedeemCodes(ctx context.Context) ([]model.RedeemCode, error) {
	var rows []model.RedeemCode
	err := db.DB(ctx).Preload("Plan").Order("id desc").Find(&rows).Error
	return rows, err
}

// RedeemSubscriptionWithCode atomically consumes one code and grants one month.
func RedeemSubscriptionWithCode(ctx context.Context, code string, userID uint64, now time.Time) (*model.UserSubscription, error) {
	var subscription *model.UserSubscription
	err := db.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var redeemCode model.RedeemCode
		if err := tx.Clauses(clause.Locking{Strength: pagesRowLockStrength}).
			Where("code = ? AND status = ?", code, model.RedeemCodeStatusUnused).
			First(&redeemCode).Error; err != nil {
			return err
		}

		usedBy := userID
		result := tx.Model(&model.RedeemCode{}).
			Where("id = ? AND status = ?", redeemCode.ID, model.RedeemCodeStatusUnused).
			Updates(map[string]any{
				columnStatus: model.RedeemCodeStatusUsed,
				"used_by":    usedBy,
				"used_at":    now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}

		var err error
		subscription, err = activateSubscriptionTx(tx, userID, redeemCode.PlanID, 1, now)
		return err
	})
	return subscription, err
}
