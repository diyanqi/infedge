// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"
)

const (
	tableOfZones       = "of_zones"
	tableOfZoneDomains = "of_zone_domains"
)

// Zone OpenFlare 注册根域实体。
type Zone struct {
	ID                 uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	OwnerID            uint64     `json:"owner_id" gorm:"not null;default:0;index"`
	Domain             string     `json:"domain" gorm:"uniqueIndex:idx_of_zones_domain;size:255;not null"`
	ClaimsOwnership    bool       `json:"claims_ownership" gorm:"not null;default:false"`
	VerificationStatus string     `json:"verification_status" gorm:"size:32;not null;default:'pending'"`
	VerificationToken  string     `json:"verification_token,omitempty" gorm:"size:128;not null;default:''"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 表名。
func (Zone) TableName() string {
	return tableOfZones
}

// ZoneDomain OpenFlare Zone 下的明确域名实体。
type ZoneDomain struct {
	ID                 uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	ZoneID             uint       `json:"zone_id" gorm:"not null;index:idx_of_zone_domains_zone_id"`
	ProxyRouteID       *uint      `json:"proxy_route_id" gorm:"index:idx_of_zone_domains_proxy_route_id"`
	Domain             string     `json:"domain" gorm:"uniqueIndex:idx_of_zone_domains_domain;size:255;not null"`
	CertID             *uint      `json:"cert_id" gorm:"index:idx_of_zone_domains_cert_id"`
	VerificationStatus string     `json:"verification_status" gorm:"size:32;not null;default:'pending'"`
	VerificationToken  string     `json:"verification_token,omitempty" gorm:"size:128;not null;default:''"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 表名。
func (ZoneDomain) TableName() string {
	return tableOfZoneDomains
}

// ZoneDomainCount is the per-zone explicit domain count for list queries.
type ZoneDomainCount struct {
	ZoneID uint  `json:"zone_id" gorm:"column:zone_id"`
	Count  int64 `json:"count" gorm:"column:count"`
}
