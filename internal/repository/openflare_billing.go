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
)

func ListSubscriptionPlans(ctx context.Context, enabledOnly bool) ([]model.SubscriptionPlan, error) {
	query := db.DB(ctx).Order("price_fen asc, id asc")
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	var rows []model.SubscriptionPlan
	return rows, query.Find(&rows).Error
}

func GetSubscriptionPlanByID(ctx context.Context, id uint) (*model.SubscriptionPlan, error) {
	var row model.SubscriptionPlan
	err := db.DB(ctx).First(&row, id).Error
	return &row, err
}

func SaveSubscriptionPlan(ctx context.Context, row *model.SubscriptionPlan) error {
	return db.DB(ctx).Save(row).Error
}

func CreateSubscriptionPlan(ctx context.Context, row *model.SubscriptionPlan) error {
	return db.DB(ctx).Create(row).Error
}

func DeleteSubscriptionPlan(ctx context.Context, id uint) error {
	return db.DB(ctx).Delete(&model.SubscriptionPlan{}, id).Error
}

func ListPaymentChannels(ctx context.Context, enabledOnly bool) ([]model.PaymentChannel, error) {
	query := db.DB(ctx).Order("sort asc, id asc")
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	var rows []model.PaymentChannel
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func GetPaymentChannelByID(ctx context.Context, id uint) (*model.PaymentChannel, error) {
	var row model.PaymentChannel
	err := db.DB(ctx).First(&row, id).Error
	return &row, err
}

func SavePaymentChannel(ctx context.Context, row *model.PaymentChannel) error {
	return db.DB(ctx).Save(row).Error
}

func CreatePaymentChannel(ctx context.Context, row *model.PaymentChannel) error {
	return db.DB(ctx).Create(row).Error
}

func DeletePaymentChannel(ctx context.Context, id uint) error {
	return db.DB(ctx).Delete(&model.PaymentChannel{}, id).Error
}

func GetActiveSubscription(ctx context.Context, userID uint64, now time.Time) (*model.UserSubscription, error) {
	var row model.UserSubscription
	err := db.DB(ctx).Preload("Plan").Where(
		"user_id = ? AND status = ? AND starts_at <= ? AND expires_at > ?",
		userID, model.SubscriptionStatusActive, now, now,
	).Order("expires_at desc").First(&row).Error
	return &row, err
}

// ListActiveSubscriptionPlans returns the current enabled plan for each subscribed user.
func ListActiveSubscriptionPlans(ctx context.Context, now time.Time) ([]model.UserSubscription, error) {
	var rows []model.UserSubscription
	err := db.DB(ctx).Preload("Plan").Where(
		"status = ? AND starts_at <= ? AND expires_at > ?",
		model.SubscriptionStatusActive, now, now,
	).Order("user_id asc, expires_at desc").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]model.UserSubscription, 0, len(rows))
	seen := make(map[uint64]struct{}, len(rows))
	for _, row := range rows {
		if row.Plan == nil || !row.Plan.Enabled {
			continue
		}
		if _, ok := seen[row.UserID]; ok {
			continue
		}
		seen[row.UserID] = struct{}{}
		result = append(result, row)
	}
	return result, nil
}

func CreatePaymentOrder(ctx context.Context, row *model.PaymentOrder) error {
	return db.DB(ctx).Create(row).Error
}

// ActivateSubscription grants a plan without a payment transaction.
func ActivateSubscription(ctx context.Context, userID uint64, planID uint, now time.Time) error {
	return db.DB(ctx).Transaction(func(tx *gorm.DB) error {
		_, err := activateSubscriptionTx(tx, userID, planID, 0, now)
		return err
	})
}

// activateSubscriptionTx activates a plan inside an existing transaction.
// months=0 uses the plan billing period; a positive value overrides it.
func activateSubscriptionTx(tx *gorm.DB, userID uint64, planID uint, months int, now time.Time) (*model.UserSubscription, error) {
	var plan model.SubscriptionPlan
	if err := tx.First(&plan, planID).Error; err != nil {
		return nil, err
	}
	var current model.UserSubscription
	if err := tx.Where("user_id = ? AND status = ?", userID, model.SubscriptionStatusActive).
		Order("expires_at desc").First(&current).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	start := now
	if current.ID != 0 && current.ExpiresAt.After(start) {
		start = current.ExpiresAt
		current.Status = model.SubscriptionStatusExpired
		if err := tx.Save(&current).Error; err != nil {
			return nil, err
		}
	}
	if months <= 0 {
		months = plan.BillingMonths
	}
	row := &model.UserSubscription{
		UserID: userID, PlanID: plan.ID, Status: model.SubscriptionStatusActive,
		StartsAt: start, ExpiresAt: start.AddDate(0, months, 0),
	}
	if err := tx.Create(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func GetPaymentOrderByOrderNo(ctx context.Context, orderNo string) (*model.PaymentOrder, error) {
	var row model.PaymentOrder
	err := db.DB(ctx).Where("order_no = ?", orderNo).First(&row).Error
	return &row, err
}

func ListPaymentOrdersByUser(ctx context.Context, userID uint64) ([]model.PaymentOrder, error) {
	var rows []model.PaymentOrder
	err := db.DB(ctx).Where("user_id = ?", userID).Order("id desc").Find(&rows).Error
	return rows, err
}

// MarkPaymentOrderPaid activates a plan exactly once.
func MarkPaymentOrderPaid(ctx context.Context, orderNo, tradeNo string, now time.Time) error {
	return db.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var order model.PaymentOrder
		if err := tx.Clauses().Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}
		if order.Status == model.PaymentOrderPaid {
			return nil
		}
		if err := tx.Model(&order).Updates(map[string]any{
			columnStatus: model.PaymentOrderPaid, "trade_no": tradeNo, "paid_at": now,
		}).Error; err != nil {
			return err
		}
		var current model.UserSubscription
		if err := tx.Where("user_id = ? AND status = ?", order.UserID, model.SubscriptionStatusActive).
			Order("expires_at desc").First(&current).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		start := now
		if current.ID != 0 && current.ExpiresAt.After(start) {
			start = current.ExpiresAt
			current.Status = model.SubscriptionStatusExpired
			if err := tx.Save(&current).Error; err != nil {
				return err
			}
		}
		var plan model.SubscriptionPlan
		if err := tx.First(&plan, order.PlanID).Error; err != nil {
			return err
		}
		return tx.Create(&model.UserSubscription{
			UserID: order.UserID, PlanID: order.PlanID, Status: model.SubscriptionStatusActive,
			StartsAt: start, ExpiresAt: start.AddDate(0, plan.BillingMonths, 0),
		}).Error
	})
}
