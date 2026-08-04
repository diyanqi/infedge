// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	db "github.com/Rain-kl/Wavelet/internal/infra/persistence"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupRedeemCodeTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)
	sqlDB, err := sqliteDB.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, sqliteDB.AutoMigrate(
		&model.SubscriptionPlan{},
		&model.UserSubscription{},
		&model.RedeemCode{},
	))
	db.SetDB(sqliteDB)
	t.Cleanup(func() { db.SetDB(nil) })
	return sqliteDB
}

func TestRedeemSubscriptionWithCodeGrantsOneMonthAndConsumesCode(t *testing.T) {
	conn := setupRedeemCodeTestDB(t)
	ctx := context.Background()
	now := time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC)
	plan := model.SubscriptionPlan{Name: "Monthly", BillingMonths: 12, Enabled: true}
	require.NoError(t, conn.Create(&plan).Error)
	code := model.RedeemCode{Code: "TESTCODE", PlanID: plan.ID, Status: model.RedeemCodeStatusUnused}
	require.NoError(t, conn.Create(&code).Error)

	subscription, err := RedeemSubscriptionWithCode(ctx, code.Code, 1001, now)
	require.NoError(t, err)
	require.Equal(t, plan.ID, subscription.PlanID)
	require.Equal(t, now, subscription.StartsAt)
	require.Equal(t, now.AddDate(0, 1, 0), subscription.ExpiresAt)

	var storedCode model.RedeemCode
	require.NoError(t, conn.First(&storedCode, code.ID).Error)
	require.Equal(t, model.RedeemCodeStatusUsed, storedCode.Status)
	require.Equal(t, uint64(1001), *storedCode.UsedBy)
	require.Equal(t, now, *storedCode.UsedAt)

	_, err = RedeemSubscriptionWithCode(ctx, code.Code, 1002, now)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	var subscriptionCount int64
	require.NoError(t, conn.Model(&model.UserSubscription{}).Count(&subscriptionCount).Error)
	require.Equal(t, int64(1), subscriptionCount)
}

func TestRedeemSubscriptionWithCodeExtendsCurrentSubscription(t *testing.T) {
	conn := setupRedeemCodeTestDB(t)
	ctx := context.Background()
	now := time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC)
	plan := model.SubscriptionPlan{Name: "Monthly", BillingMonths: 1, Enabled: true}
	require.NoError(t, conn.Create(&plan).Error)
	current := model.UserSubscription{
		UserID: 1001, PlanID: plan.ID, Status: model.SubscriptionStatusActive,
		StartsAt: now.AddDate(0, -1, 0), ExpiresAt: now.AddDate(0, 0, 10),
	}
	require.NoError(t, conn.Create(&current).Error)
	code := model.RedeemCode{Code: "EXTEND01", PlanID: plan.ID, Status: model.RedeemCodeStatusUnused}
	require.NoError(t, conn.Create(&code).Error)

	subscription, err := RedeemSubscriptionWithCode(ctx, code.Code, current.UserID, now)
	require.NoError(t, err)
	require.Equal(t, current.ExpiresAt, subscription.StartsAt)
	require.Equal(t, current.ExpiresAt.AddDate(0, 1, 0), subscription.ExpiresAt)

	var storedCurrent model.UserSubscription
	require.NoError(t, conn.First(&storedCurrent, current.ID).Error)
	require.Equal(t, model.SubscriptionStatusExpired, storedCurrent.Status)
}

func TestRedeemSubscriptionWithCodeOnlySucceedsOnceConcurrently(t *testing.T) {
	conn := setupRedeemCodeTestDB(t)
	ctx := context.Background()
	now := time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC)
	plan := model.SubscriptionPlan{Name: "Monthly", BillingMonths: 1, Enabled: true}
	require.NoError(t, conn.Create(&plan).Error)
	code := model.RedeemCode{Code: "ONCEONLY", PlanID: plan.ID, Status: model.RedeemCodeStatusUnused}
	require.NoError(t, conn.Create(&code).Error)

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for userID := uint64(2001); userID <= 2002; userID++ {
		wg.Add(1)
		go func(userID uint64) {
			defer wg.Done()
			_, err := RedeemSubscriptionWithCode(ctx, code.Code, userID, now)
			errs <- err
		}(userID)
	}
	wg.Wait()
	close(errs)

	var successCount int
	for err := range errs {
		if err == nil {
			successCount++
		} else {
			require.ErrorIs(t, err, gorm.ErrRecordNotFound)
		}
	}
	require.Equal(t, 1, successCount)
	var subscriptionCount int64
	require.NoError(t, conn.Model(&model.UserSubscription{}).Count(&subscriptionCount).Error)
	require.Equal(t, int64(1), subscriptionCount)
}
