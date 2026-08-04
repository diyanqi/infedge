// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"time"

	db "github.com/Rain-kl/Wavelet/internal/infra/persistence"
	"github.com/Rain-kl/Wavelet/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ListOwnedProxyRoutes returns only routes owned by userID.
func ListOwnedProxyRoutes(ctx context.Context, userID uint64) ([]*model.ProxyRoute, error) {
	var rows []*model.ProxyRoute
	err := db.DB(ctx).Where("owner_id = ?", userID).Order("id desc").Find(&rows).Error
	return rows, err
}

// GetOwnedProxyRouteByID returns a route only when it belongs to userID.
func GetOwnedProxyRouteByID(ctx context.Context, id uint, userID uint64) (*model.ProxyRoute, error) {
	var row model.ProxyRoute
	err := db.DB(ctx).Where("id = ? AND owner_id = ?", id, userID).First(&row).Error
	return &row, err
}

// SetProxyRouteOwner assigns a route to a user after creation.
func SetProxyRouteOwner(ctx context.Context, id uint, userID uint64) error {
	return db.DB(ctx).Model(&model.ProxyRoute{}).Where("id = ?", id).Update("owner_id", userID).Error
}

// ListOwnedOrigins returns only origins owned by userID.
func ListOwnedOrigins(ctx context.Context, userID uint64) ([]model.Origin, error) {
	var rows []model.Origin
	err := db.DB(ctx).Where("owner_id = ?", userID).Order("id desc").Find(&rows).Error
	return rows, err
}

// GetOwnedOriginByID returns an origin only when it belongs to userID.
func GetOwnedOriginByID(ctx context.Context, id uint, userID uint64) (*model.Origin, error) {
	var row model.Origin
	err := db.DB(ctx).Where("id = ? AND owner_id = ?", id, userID).First(&row).Error
	return &row, err
}

// SetOriginOwner assigns an origin to a user after creation.
func SetOriginOwner(ctx context.Context, id uint, userID uint64) error {
	return db.DB(ctx).Model(&model.Origin{}).Where("id = ?", id).Update("owner_id", userID).Error
}

// ListOwnedZones returns only zones owned by userID.
func ListOwnedZones(ctx context.Context, userID uint64) ([]model.Zone, error) {
	var rows []model.Zone
	err := db.DB(ctx).Where("owner_id = ?", userID).Order("domain asc").Find(&rows).Error
	return rows, err
}

// ListOwnedZoneDomains returns all domains under zones owned by userID.
func ListOwnedZoneDomains(ctx context.Context, userID uint64) ([]model.ZoneDomain, error) {
	var domains []model.ZoneDomain
	if err := db.DB(ctx).Joins("JOIN of_zones ON of_zones.id = of_zone_domains.zone_id").
		Where("of_zones.owner_id = ?", userID).Order("of_zone_domains.domain asc").Find(&domains).Error; err != nil {
		return nil, err
	}
	return domains, nil
}

// GetOwnedZoneByID returns a zone only when it belongs to userID.
func GetOwnedZoneByID(ctx context.Context, id uint, userID uint64) (*model.Zone, error) {
	var row model.Zone
	err := db.DB(ctx).Where("id = ? AND owner_id = ?", id, userID).First(&row).Error
	return &row, err
}

// SetZoneOwner assigns a zone to a user after creation.
func SetZoneOwner(ctx context.Context, id uint, userID uint64) error {
	return db.DB(ctx).Model(&model.Zone{}).Where("id = ?", id).Update("owner_id", userID).Error
}

// AreZoneDomainsOwned verifies all requested domains belong to the same user-owned zones.
func AreZoneDomainsOwned(ctx context.Context, domainIDs []uint, userID uint64) (bool, error) {
	if len(domainIDs) == 0 {
		return true, nil
	}
	var count int64
	err := db.DB(ctx).Model(&model.ZoneDomain{}).
		Joins("JOIN of_zones ON of_zones.id = of_zone_domains.zone_id").
		Where("of_zone_domains.id IN ? AND of_zones.owner_id = ?", domainIDs, userID).
		Count(&count).Error
	return count == int64(len(uniqueZoneDomainIDs(domainIDs))), err
}

// SetConfigVersionCreator records the numeric owner of a published version.
func SetConfigVersionCreator(ctx context.Context, version string, userID uint64) error {
	return db.DB(ctx).Model(&model.ConfigVersion{}).Where("version = ?", version).Update("created_by_user_id", userID).Error
}

// CountOwnedResources counts user-owned resources for quota checks.
func CountOwnedResources(ctx context.Context, table string, userID uint64) (int64, error) {
	allowed := map[string]any{
		"zones":          &model.Zone{},
		"origins":        &model.Origin{},
		"proxy_routes":   &model.ProxyRoute{},
		"pages_projects": &model.PagesProject{},
	}
	entity, ok := allowed[table]
	if !ok {
		return 0, gorm.ErrInvalidData
	}
	var count int64
	err := db.DB(ctx).Model(entity).Where("owner_id = ?", userID).Count(&count).Error
	return count, err
}

// CountUserPublishesSince counts versions published by a user after since.
func CountUserPublishesSince(ctx context.Context, userID uint64, since time.Time) (int64, error) {
	var count int64
	err := db.DB(ctx).Model(&model.ConfigVersion{}).
		Where("created_by_user_id = ? AND created_at >= ?", userID, since).
		Count(&count).Error
	return count, err
}

// ReserveUserPublish atomically consumes one daily publish entitlement.
func ReserveUserPublish(ctx context.Context, userID uint64, dayStart time.Time, limit int) (bool, error) {
	if limit <= 0 {
		return true, nil
	}
	var reserved bool
	err := db.DB(ctx).Transaction(func(tx *gorm.DB) error {
		counter := &model.UserPublishDailyCounter{UserID: userID, DayStart: dayStart, Used: 0}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(counter).Error; err != nil {
			return err
		}
		var current model.UserPublishDailyCounter
		if err := tx.Clauses(clause.Locking{Strength: pagesRowLockStrength}).Where(
			"user_id = ? AND day_start = ?", userID, dayStart,
		).First(&current).Error; err != nil {
			return err
		}
		if current.Used >= limit {
			return nil
		}
		if err := tx.Model(&current).Updates(map[string]any{
			"used": current.Used + 1, repositoryUpdatedAtColumn: time.Now().UTC(),
		}).Error; err != nil {
			return err
		}
		reserved = true
		return nil
	})
	return reserved, err
}

// ReleaseUserPublish returns a reservation when publishing fails before a version is created.
func ReleaseUserPublish(ctx context.Context, userID uint64, dayStart time.Time) error {
	return db.DB(ctx).Model(&model.UserPublishDailyCounter{}).Where(
		"user_id = ? AND day_start = ? AND used > 0", userID, dayStart,
	).UpdateColumn("used", gorm.Expr("used - 1")).Error
}
