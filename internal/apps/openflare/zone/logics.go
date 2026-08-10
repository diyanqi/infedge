// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package zone manages registered roots and their explicit hostnames.
package zone

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/Rain-kl/Wavelet/internal/repository"
	"golang.org/x/net/publicsuffix"
	"gorm.io/gorm"
)

const (
	zoneVerificationStatusVerified = "verified"
	zoneVerificationStatusPending  = "pending"
	verificationTokenBytes         = 24
)

// Input is the mutable Zone payload.
type Input struct {
	Domain          string `json:"domain"`
	ClaimsOwnership bool   `json:"claims_ownership"`
}

// DomainInput is the mutable Zone-domain payload.
type DomainInput struct {
	Domain string `json:"domain"`
	CertID *uint  `json:"cert_id"`
}

// SiteInput is the one-step domain onboarding payload.
type SiteInput struct {
	Domain string `json:"domain"`
}

// Site is the result of onboarding one explicit hostname.
type Site struct {
	Zone   model.Zone       `json:"zone"`
	Domain model.ZoneDomain `json:"domain"`
}

// Overview joins a Zone with its explicit domains.
type Overview struct {
	Zone    model.Zone         `json:"zone"`
	Domains []model.ZoneDomain `json:"domains"`
}

// ListItem is a Zone list row with denormalized domain count for the UI.
type ListItem struct {
	ID          uint      `json:"id"`
	Domain      string    `json:"domain"`
	DomainCount int64     `json:"domain_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func zoneRoot(domain string) (string, error) {
	return publicsuffix.EffectiveTLDPlusOne(strings.ToLower(strings.TrimSpace(domain)))
}

func normalizeDomain(raw string) (string, error) {
	domain := strings.ToLower(strings.TrimSpace(raw))
	if domain == "" {
		return "", errors.New(errZoneDomainRequired)
	}
	if strings.Contains(domain, "*") {
		return "", errors.New(errDomainWildcardUnsupported)
	}
	if strings.Contains(domain, "://") || strings.Contains(domain, "/") || strings.Contains(domain, "?") || strings.Contains(domain, "#") || strings.Contains(domain, "@") {
		return "", errors.New(errDomainInvalid)
	}
	if _, err := zoneRoot(domain); err != nil {
		return "", errors.New(errDomainInvalid)
	}
	return domain, nil
}

// Create persists a validated registered root.
func Create(ctx context.Context, input Input) (*model.Zone, error) {
	return create(ctx, 0, input)
}

func create(ctx context.Context, ownerID uint64, input Input) (*model.Zone, error) {
	domain, err := normalizeDomain(input.Domain)
	if err != nil {
		return nil, err
	}
	root, err := zoneRoot(domain)
	if err != nil || root != domain {
		return nil, errors.New(errZoneRootInvalid)
	}
	status, verifiedAt := verificationState(ownerID)
	zone := &model.Zone{
		Domain:             domain,
		OwnerID:            ownerID,
		ClaimsOwnership:    input.ClaimsOwnership,
		VerificationStatus: status,
		VerificationToken:  newVerificationToken(),
		VerifiedAt:         verifiedAt,
	}
	if err := repository.CreateZone(ctx, zone); err != nil {
		if isUnique(err) {
			return nil, errors.New(errDomainExists)
		}
		return nil, err
	}
	return zone, nil
}

// Update replaces a Zone's mutable fields.
func Update(ctx context.Context, id uint, input Input) (*model.Zone, error) {
	zone, err := repository.GetZoneByID(ctx, id)
	if err != nil {
		return nil, err
	}
	domain, err := normalizeDomain(input.Domain)
	if err != nil {
		return nil, err
	}
	root, err := zoneRoot(domain)
	if err != nil || root != domain {
		return nil, errors.New(errZoneRootInvalid)
	}
	zone.Domain = domain
	if err := repository.SaveZone(ctx, zone); err != nil {
		if isUnique(err) {
			return nil, errors.New(errDomainExists)
		}
		return nil, err
	}
	return zone, nil
}

// List returns all Zones in stable domain order, with domain counts for list cards.
func List(ctx context.Context) ([]ListItem, error) {
	zones, err := repository.ListZones(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := repository.ListZoneDomainCounts(ctx)
	if err != nil {
		return nil, err
	}
	counts := make(map[uint]int64, len(rows))
	for _, row := range rows {
		counts[row.ZoneID] = row.Count
	}

	items := make([]ListItem, 0, len(zones))
	for _, zone := range zones {
		items = append(items, ListItem{
			ID:          zone.ID,
			Domain:      zone.Domain,
			DomainCount: counts[zone.ID],
			CreatedAt:   zone.CreatedAt,
			UpdatedAt:   zone.UpdatedAt,
		})
	}
	return items, nil
}

// ListOwned returns zones visible to one ordinary user.
func ListOwned(ctx context.Context, userID uint64) ([]ListItem, error) {
	zones, err := repository.ListOwnedZones(ctx, userID)
	if err != nil {
		return nil, err
	}
	counts, err := repository.ListZoneDomainCounts(ctx)
	if err != nil {
		return nil, err
	}
	countMap := make(map[uint]int64, len(counts))
	for _, row := range counts {
		countMap[row.ZoneID] = row.Count
	}
	items := make([]ListItem, 0, len(zones))
	for _, row := range zones {
		items = append(items, ListItem{ID: row.ID, Domain: row.Domain, DomainCount: countMap[row.ID], CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return items, nil
}

// GetOwnedOverview returns a zone only when owned by the user.
func GetOwnedOverview(ctx context.Context, id uint, userID uint64) (*Overview, error) {
	if _, err := repository.GetOwnedZoneByID(ctx, id, userID); err != nil {
		return nil, err
	}
	return GetOverview(ctx, id)
}

// CreateOwned creates a zone and assigns its owner.
func CreateOwned(ctx context.Context, userID uint64, input Input) (*model.Zone, error) {
	return create(ctx, userID, input)
}

// HasOwnedRoot reports whether the domain's derived Zone already exists.
func HasOwnedRoot(ctx context.Context, userID uint64, rawDomain string) (bool, error) {
	domain, err := normalizeDomain(rawDomain)
	if err != nil {
		return false, err
	}
	root, err := zoneRoot(domain)
	if err != nil {
		return false, err
	}
	_, err = repository.GetZoneByDomainAndOwner(ctx, root, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

func createSite(ctx context.Context, ownerID uint64, input SiteInput) (*Site, error) {
	domain, err := normalizeDomain(input.Domain)
	if err != nil {
		if strings.TrimSpace(input.Domain) == "" {
			return nil, errors.New(errSiteDomainRequired)
		}
		return nil, err
	}
	root, err := zoneRoot(domain)
	if err != nil {
		return nil, errors.New(errDomainInvalid)
	}
	if err := ensureDomainNotBound(ctx, domain, nil); err != nil {
		return nil, err
	}
	zone, err := repository.GetZoneByDomainAndOwner(ctx, root, ownerID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		status, verifiedAt := verificationState(ownerID)
		zone = &model.Zone{
			OwnerID:            ownerID,
			Domain:             root,
			ClaimsOwnership:    false,
			VerificationStatus: status,
			VerificationToken:  newVerificationToken(),
			VerifiedAt:         verifiedAt,
		}
		if err := repository.CreateZone(ctx, zone); err != nil {
			if isUnique(err) {
				return nil, errors.New(errDomainExists)
			}
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	domainItem, err := createDomain(ctx, zone.ID, DomainInput{Domain: domain}, false, ownerID == 0)
	if err != nil {
		return nil, err
	}
	return &Site{Zone: *zone, Domain: *domainItem}, nil
}

// CreateSite onboards one explicit hostname for an administrator.
func CreateSite(ctx context.Context, input SiteInput) (*Site, error) {
	return createSite(ctx, 0, input)
}

// CreateOwnedSite onboards one explicit hostname for an ordinary user.
func CreateOwnedSite(ctx context.Context, userID uint64, input SiteInput) (*Site, error) {
	return createSite(ctx, userID, input)
}

// UpdateOwned updates an owned zone.
func UpdateOwned(ctx context.Context, id uint, userID uint64, input Input) (*model.Zone, error) {
	if _, err := repository.GetOwnedZoneByID(ctx, id, userID); err != nil {
		return nil, err
	}
	zone, err := Update(ctx, id, input)
	if zone != nil {
		zone.OwnerID = userID
	}
	return zone, err
}

// DeleteOwned deletes an owned zone together with its domains and the sites
// (proxy routes) bound to those domains. Ordinary users manage "sites" by
// domain, so a root domain group can be removed in one action.
func DeleteOwned(ctx context.Context, id uint, userID uint64) error {
	return repository.DeleteOwnedZoneCascade(ctx, id, userID)
}

// CreateOwnedDomain creates a domain under an owned zone.
func CreateOwnedDomain(ctx context.Context, zoneID uint, userID uint64, input DomainInput) (*model.ZoneDomain, error) {
	if _, err := repository.GetOwnedZoneByID(ctx, zoneID, userID); err != nil {
		return nil, err
	}
	return createDomain(ctx, zoneID, input, true, false)
}

// UpdateOwnedDomain updates a domain and keeps all ownership checks in the
// owner-scoped path used by the ordinary-user console.
func UpdateOwnedDomain(ctx context.Context, zoneID, domainID uint, userID uint64, input DomainInput) (*model.ZoneDomain, error) {
	if _, err := repository.GetOwnedZoneByID(ctx, zoneID, userID); err != nil {
		return nil, err
	}
	if _, err := repository.GetZoneDomainByZoneAndID(ctx, zoneID, domainID); err != nil {
		return nil, err
	}
	return UpdateDomain(ctx, zoneID, domainID, input)
}

// GetOverview returns a Zone and its domains.
func GetOverview(ctx context.Context, id uint) (*Overview, error) {
	zone, err := repository.GetZoneByID(ctx, id)
	if err != nil {
		return nil, err
	}
	domains, err := repository.ListZoneDomainsByZoneID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &Overview{Zone: *zone, Domains: domains}, nil
}

// CreateDomain adds a validated exact hostname to a Zone.
func CreateDomain(ctx context.Context, zoneID uint, input DomainInput) (*model.ZoneDomain, error) {
	return createDomain(ctx, zoneID, input, true, true)
}

func createDomain(ctx context.Context, zoneID uint, input DomainInput, inheritZoneOwnership, autoVerified bool) (*model.ZoneDomain, error) {
	zone, err := repository.GetZoneByID(ctx, zoneID)
	if err != nil {
		return nil, err
	}
	domain, err := normalizeDomain(input.Domain)
	if err != nil {
		return nil, err
	}
	root, err := zoneRoot(domain)
	if err != nil || root != zone.Domain {
		return nil, errors.New(errDomainOutsideZone)
	}
	if input.CertID != nil {
		if _, err := repository.GetTLSCertificateByID(ctx, *input.CertID); err != nil {
			return nil, errors.New(errCertificateNotFound)
		}
	}
	if err := ensureDomainNotBound(ctx, domain, nil); err != nil {
		return nil, err
	}
	status := zoneVerificationStatusPending
	var verifiedAt *time.Time
	if autoVerified || (inheritZoneOwnership && zone.ClaimsOwnership && zone.VerificationStatus == zoneVerificationStatusVerified) {
		status = zoneVerificationStatusVerified
		now := time.Now().UTC()
		verifiedAt = &now
	}
	item := &model.ZoneDomain{
		ZoneID:             zoneID,
		Domain:             domain,
		CertID:             input.CertID,
		VerificationStatus: status,
		VerificationToken:  newVerificationToken(),
		VerifiedAt:         verifiedAt,
	}
	if err := repository.CreateZoneDomain(ctx, item); err != nil {
		if isUnique(err) {
			return nil, errors.New(errDomainExists)
		}
		return nil, err
	}
	return item, nil
}

// VerifyOwnedZone checks the TXT record proving control of a root domain.
func VerifyOwnedZone(ctx context.Context, id uint, ownerID uint64) (*model.Zone, error) {
	zone, err := repository.GetOwnedZoneByID(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}
	if !zone.ClaimsOwnership {
		return nil, errors.New("该根域未声明拥有全部权利，请逐个验证子域")
	}
	if err := verifyTXT(ctx, "_openflare-verification."+zone.Domain, zone.VerificationToken); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	zone.VerificationStatus, zone.VerifiedAt = zoneVerificationStatusVerified, &now
	if err := repository.SaveZone(ctx, zone); err != nil {
		return nil, err
	}
	return zone, nil
}

// VerifyOwnedDomain checks the TXT record proving control of one hostname.
func VerifyOwnedDomain(ctx context.Context, zoneID, domainID uint, ownerID uint64) (*model.ZoneDomain, error) {
	zone, err := repository.GetOwnedZoneByID(ctx, zoneID, ownerID)
	if err != nil {
		return nil, err
	}
	domain, err := repository.GetZoneDomainByZoneAndID(ctx, zoneID, domainID)
	if err != nil {
		return nil, err
	}
	if zone.ClaimsOwnership && zone.VerificationStatus == zoneVerificationStatusVerified {
		now := time.Now().UTC()
		domain.VerificationStatus, domain.VerifiedAt = zoneVerificationStatusVerified, &now
	} else {
		if err := verifyTXT(ctx, "_openflare-verification."+domain.Domain, domain.VerificationToken); err != nil {
			return nil, err
		}
		now := time.Now().UTC()
		domain.VerificationStatus, domain.VerifiedAt = zoneVerificationStatusVerified, &now
	}
	if err := repository.SaveZoneDomain(ctx, domain); err != nil {
		return nil, err
	}
	return domain, nil
}

// VerifyOwnedSiteDomain verifies a domain returned by the direct onboarding API.
func VerifyOwnedSiteDomain(ctx context.Context, domainID uint, ownerID uint64) (*model.ZoneDomain, error) {
	domain, err := repository.GetOwnedZoneDomainByIDAnyZone(ctx, domainID, ownerID)
	if err != nil {
		return nil, err
	}
	if ownerID != 0 {
		if err := verifyTXT(ctx, "_openflare-verification."+domain.Domain, domain.VerificationToken); err != nil {
			return nil, err
		}
	}
	now := time.Now().UTC()
	domain.VerificationStatus, domain.VerifiedAt = zoneVerificationStatusVerified, &now
	if err := repository.SaveZoneDomain(ctx, domain); err != nil {
		return nil, err
	}
	return domain, nil
}

// EnsureOwnedDomainsReady checks verification and the fixed CNAME before deploy.
func EnsureOwnedDomainsReady(ctx context.Context, ownerID uint64) error {
	domains, err := repository.ListOwnedZoneDomains(ctx, ownerID)
	if err != nil {
		return err
	}
	for _, domain := range domains {
		if domain.ProxyRouteID == nil {
			continue
		}
		if domain.VerificationStatus != zoneVerificationStatusVerified {
			return errors.New("域名 " + domain.Domain + " 尚未完成 DNS TXT 所有权验证")
		}
		if err := checkCNAME(ctx, domain.Domain); err != nil {
			return err
		}
	}
	return nil
}

func newVerificationToken() string {
	buf := make([]byte, verificationTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(buf)
}

func verificationState(ownerID uint64) (string, *time.Time) {
	if ownerID != 0 {
		return zoneVerificationStatusPending, nil
	}
	now := time.Now().UTC()
	return zoneVerificationStatusVerified, &now
}

func verifyTXT(ctx context.Context, name, expected string) error {
	records, err := net.DefaultResolver.LookupTXT(ctx, name)
	if err != nil {
		return errors.New("未找到 DNS TXT 验证记录，请稍后重试")
	}
	for _, record := range records {
		if strings.TrimSpace(record) == expected {
			return nil
		}
	}
	return errors.New("DNS TXT 验证值不匹配")
}

func checkCNAME(ctx context.Context, domain string) error {
	target, err := net.DefaultResolver.LookupCNAME(ctx, domain)
	if err != nil || strings.TrimSuffix(strings.ToLower(strings.TrimSpace(target)), ".") != "cname.edge.infvar.com" {
		return errors.New("域名必须将 CNAME 指向 cname.edge.infvar.com；禁止自行优选 IP，该地址已每小时进行全国拨测优选")
	}
	return nil
}

// UpdateDomain replaces a Zone-domain's mutable fields.
func UpdateDomain(ctx context.Context, zoneID, id uint, input DomainInput) (*model.ZoneDomain, error) {
	item, err := repository.GetZoneDomainByZoneAndID(ctx, zoneID, id)
	if err != nil {
		return nil, err
	}
	domain, err := normalizeDomain(input.Domain)
	if err != nil {
		return nil, err
	}
	zone, err := repository.GetZoneByID(ctx, zoneID)
	if err != nil {
		return nil, err
	}
	root, err := zoneRoot(domain)
	if err != nil || root != zone.Domain {
		return nil, errors.New(errDomainOutsideZone)
	}
	if input.CertID != nil {
		if _, err = repository.GetTLSCertificateByID(ctx, *input.CertID); err != nil {
			return nil, errors.New(errCertificateNotFound)
		}
	}
	if err := ensureDomainNotBound(ctx, domain, []uint{item.ID}); err != nil {
		return nil, err
	}
	item.Domain, item.CertID = domain, input.CertID
	if err = repository.SaveZoneDomain(ctx, item); err != nil {
		if isUnique(err) {
			return nil, errors.New(errDomainExists)
		}
		return nil, err
	}
	return item, nil
}

func ensureDomainNotBound(ctx context.Context, domain string, excludeIDs []uint) error {
	_, err := repository.GetBoundZoneDomainByDomain(ctx, domain, excludeIDs)
	if err == nil {
		return errors.New(errDomainInUse)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

// DeleteDomain removes a Zone domain that is not bound to a proxy route.
func DeleteDomain(ctx context.Context, zoneID, id uint) error {
	item, err := repository.GetZoneDomainByZoneAndID(ctx, zoneID, id)
	if err != nil {
		return err
	}
	if item.ProxyRouteID != nil {
		return errors.New(errDomainBoundToRoute)
	}
	return repository.DeleteZoneDomain(ctx, item)
}

// Delete removes a Zone that has no remaining domains.
func Delete(ctx context.Context, id uint) error {
	if _, err := repository.GetZoneByID(ctx, id); err != nil {
		return err
	}
	count, err := repository.CountZoneDomainsByZoneID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New(errZoneHasDomains)
	}
	return repository.DeleteZone(ctx, id)
}

func isUnique(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}
