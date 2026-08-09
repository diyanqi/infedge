// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"errors"

	db "github.com/Rain-kl/Wavelet/internal/infra/persistence"
	"github.com/Rain-kl/Wavelet/internal/model"
	"gorm.io/gorm"
)

// ListDNSAccounts 列出平台级 DNS 账号（授权信息不通过 JSON 暴露）。
func ListDNSAccounts(ctx context.Context) ([]model.DNSAccount, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	var accounts []model.DNSAccount
	if err := conn.Where("owner_id = 0").Order("id desc").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// ListOwnedDNSAccounts 列出某个普通用户自己的 DNS 账号。
func ListOwnedDNSAccounts(ctx context.Context, ownerID uint64) ([]model.DNSAccount, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	var accounts []model.DNSAccount
	if err := conn.Where("owner_id = ?", ownerID).Order("id desc").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// ListDNSAccountsForOwner 返回平台账号与当前用户自己的账号，供普通用户 ACME 申请选择。
func ListDNSAccountsForOwner(ctx context.Context, ownerID uint64) ([]model.DNSAccount, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	var accounts []model.DNSAccount
	if err := conn.Where("owner_id = 0 OR owner_id = ?", ownerID).Order("id desc").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// GetDNSAccountByID 按 ID 查询 DNS 账号。
func GetDNSAccountByID(ctx context.Context, id uint) (*model.DNSAccount, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	var account model.DNSAccount
	if err := conn.First(&account, id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// GetPlatformDNSAccountByID 按 ID 查询平台级 DNS 账号。
func GetPlatformDNSAccountByID(ctx context.Context, id uint) (*model.DNSAccount, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	var account model.DNSAccount
	if err := conn.Where("id = ? AND owner_id = 0", id).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// GetOwnedDNSAccountByID 按 ID 和 owner 查询普通用户的 DNS 账号。
func GetOwnedDNSAccountByID(ctx context.Context, id uint, ownerID uint64) (*model.DNSAccount, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	var account model.DNSAccount
	if err := conn.Where("id = ? AND owner_id = ?", id, ownerID).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// CountOwnedDNSAccounts 统计某个普通用户自己的 DNS 账号数量。
func CountOwnedDNSAccounts(ctx context.Context, ownerID uint64) (int64, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return 0, errors.New(errDatabaseNotInitialized)
	}
	var count int64
	if err := conn.Model(&model.DNSAccount{}).Where("owner_id = ?", ownerID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CreateDNSAccountRecord 创建 DNS 账号。
func CreateDNSAccountRecord(ctx context.Context, account *model.DNSAccount) error {
	conn := db.DB(ctx)
	if conn == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	return conn.Create(account).Error
}

// SaveDNSAccount 保存 DNS 账号。
func SaveDNSAccount(ctx context.Context, account *model.DNSAccount) error {
	conn := db.DB(ctx)
	if conn == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	return conn.Save(account).Error
}

// SaveOwnedDNSAccount 仅当账号仍属于 ownerID 时更新其可变字段。
func SaveOwnedDNSAccount(ctx context.Context, account *model.DNSAccount, ownerID uint64) error {
	conn := db.DB(ctx)
	if conn == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	result := conn.Model(&model.DNSAccount{}).
		Where("id = ? AND owner_id = ?", account.ID, ownerID).
		Updates(map[string]any{
			colName:         account.Name,
			colType:         account.Type,
			"authorization": account.Authorization,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteDNSAccountRecord 删除 DNS 账号。
func DeleteDNSAccountRecord(ctx context.Context, id uint) error {
	conn := db.DB(ctx)
	if conn == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	return conn.Delete(&model.DNSAccount{}, id).Error
}

// DeleteOwnedDNSAccountRecord 仅当账号属于 ownerID 时删除。
func DeleteOwnedDNSAccountRecord(ctx context.Context, id uint, ownerID uint64) error {
	conn := db.DB(ctx)
	if conn == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	result := conn.Where("id = ? AND owner_id = ?", id, ownerID).Delete(&model.DNSAccount{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
