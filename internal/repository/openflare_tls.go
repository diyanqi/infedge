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

// HasTLSProxyRoutesTable 判断代理规则表是否已迁移。
func HasTLSProxyRoutesTable(ctx context.Context) bool {
	return db.DB(ctx).Migrator().HasTable(&model.TLSProxyRouteRef{})
}

// ListTLSCertificates 列出全部证书（不含 PEM 敏感字段的 JSON 暴露由 struct tag 控制）。
func ListTLSCertificates(ctx context.Context) ([]model.TLSCertificate, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	var certificates []model.TLSCertificate
	if err := conn.Order("id desc").Find(&certificates).Error; err != nil {
		return nil, err
	}
	return certificates, nil
}

// ListOwnedTLSCertificates lists certificates owned by a user.
func ListOwnedTLSCertificates(ctx context.Context, ownerID uint64) ([]model.TLSCertificate, error) {
	var certificates []model.TLSCertificate
	if err := db.DB(ctx).Where("owner_id = ?", ownerID).Order("id desc").Find(&certificates).Error; err != nil {
		return nil, err
	}
	return certificates, nil
}

// GetTLSCertificateByID 按 ID 查询证书。
func GetTLSCertificateByID(ctx context.Context, id uint) (*model.TLSCertificate, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	var certificate model.TLSCertificate
	if err := conn.First(&certificate, id).Error; err != nil {
		return nil, err
	}
	return &certificate, nil
}

// GetOwnedTLSCertificateByID loads a certificate only for its owner.
func GetOwnedTLSCertificateByID(ctx context.Context, id uint, ownerID uint64) (*model.TLSCertificate, error) {
	var certificate model.TLSCertificate
	if err := db.DB(ctx).Where("id = ? AND owner_id = ?", id, ownerID).First(&certificate).Error; err != nil {
		return nil, err
	}
	return &certificate, nil
}

// CreateTLSCertificateRecord 创建证书记录。
func CreateTLSCertificateRecord(ctx context.Context, certificate *model.TLSCertificate) error {
	conn := db.DB(ctx)
	if conn == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	return conn.Create(certificate).Error
}

// SaveTLSCertificate 保存证书记录。
func SaveTLSCertificate(ctx context.Context, certificate *model.TLSCertificate) error {
	conn := db.DB(ctx)
	if conn == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	return conn.Save(certificate).Error
}

// SaveOwnedTLSCertificate persists mutable certificate fields only when the
// certificate still belongs to ownerID.
func SaveOwnedTLSCertificate(ctx context.Context, certificate *model.TLSCertificate, ownerID uint64) error {
	conn := db.DB(ctx)
	if conn == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	result := conn.Model(&model.TLSCertificate{}).
		Where("id = ? AND owner_id = ?", certificate.ID, ownerID).
		Updates(map[string]any{
			"name":            certificate.Name,
			"cert_pem":        certificate.CertPEM,
			"key_pem":         certificate.KeyPEM,
			"not_before":      certificate.NotBefore,
			"not_after":       certificate.NotAfter,
			"remark":          certificate.Remark,
			"provider":        certificate.Provider,
			"acme_account_id": certificate.AcmeAccountID,
			"dns_account_id":  certificate.DNSAccountID,
			"key_algorithm":   certificate.KeyAlgorithm,
			"auto_renew":      certificate.AutoRenew,
			"primary_domain":  certificate.PrimaryDomain,
			"other_domains":   certificate.OtherDomains,
			"disable_cname":   certificate.DisableCNAME,
			"skip_dns":        certificate.SkipDNS,
			"dns1":            certificate.DNS1,
			"dns2":            certificate.DNS2,
			"apply_status":    certificate.ApplyStatus,
			"apply_message":   certificate.ApplyMessage,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteTLSCertificateRecord 删除证书记录。
func DeleteTLSCertificateRecord(ctx context.Context, id uint) error {
	conn := db.DB(ctx)
	if conn == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	return conn.Delete(&model.TLSCertificate{}, id).Error
}

// DeleteOwnedTLSCertificateRecord deletes a certificate only for its owner.
func DeleteOwnedTLSCertificateRecord(ctx context.Context, id uint, ownerID uint64) error {
	conn := db.DB(ctx)
	if conn == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	result := conn.Where("id = ? AND owner_id = ?", id, ownerID).Delete(&model.TLSCertificate{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CountTLSCertificatesByDNSAccountID 统计引用指定 DNS 账号的证书数量。
func CountTLSCertificatesByDNSAccountID(ctx context.Context, dnsAccountID uint) (int64, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return 0, errors.New(errDatabaseNotInitialized)
	}
	var count int64
	if err := conn.Model(&model.TLSCertificate{}).Where("dns_account_id = ?", dnsAccountID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ListTLSProxyRouteRefs 列出代理规则证书引用字段。
func ListTLSProxyRouteRefs(ctx context.Context) ([]model.TLSProxyRouteRef, error) {
	if !HasTLSProxyRoutesTable(ctx) {
		return nil, nil
	}
	var routes []model.TLSProxyRouteRef
	if err := db.DB(ctx).Order("id asc").Find(&routes).Error; err != nil {
		return nil, err
	}
	return routes, nil
}
