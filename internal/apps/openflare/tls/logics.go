// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/Rain-kl/Wavelet/internal/repository"

	"github.com/Rain-kl/Wavelet/internal/infra/task"
	"github.com/Rain-kl/Wavelet/internal/model"
)

const maxUserDNSAccounts = 5

// CertificateInput TLS 证书创建/更新请求。
type CertificateInput struct {
	Name    string `json:"name"`
	CertPEM string `json:"cert_pem"`
	KeyPEM  string `json:"key_pem"`
	Remark  string `json:"remark"`
}

// CertificateContent TLS 证书 PEM 内容（仅 /content 端点返回）。
type CertificateContent struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	CertPEM       string `json:"cert_pem"`
	KeyPEM        string `json:"key_pem"`
	Remark        string `json:"remark"`
	Provider      string `json:"provider"`
	AcmeAccountID uint   `json:"acme_account_id"`
	DNSAccountID  uint   `json:"dns_account_id"`
	KeyAlgorithm  string `json:"key_algorithm"`
	AutoRenew     bool   `json:"auto_renew"`
	PrimaryDomain string `json:"primary_domain"`
	OtherDomains  string `json:"other_domains"`
	DisableCNAME  bool   `json:"disable_cname"`
	SkipDNS       bool   `json:"skip_dns"`
	DNS1          string `json:"dns1"`
	DNS2          string `json:"dns2"`
	ApplyStatus   string `json:"apply_status"`
	ApplyMessage  string `json:"apply_message"`
}

// ApplyInput ACME 证书申请/更新请求。
type ApplyInput struct {
	Name          string `json:"name"`
	Remark        string `json:"remark"`
	AcmeAccountID uint   `json:"acme_account_id"`
	DNSAccountID  uint   `json:"dns_account_id"`
	KeyAlgorithm  string `json:"key_algorithm"`
	AutoRenew     bool   `json:"auto_renew"`
	PrimaryDomain string `json:"primary_domain"`
	OtherDomains  string `json:"other_domains"`
	DisableCNAME  bool   `json:"disable_cname"`
	SkipDNS       bool   `json:"skip_dns"`
	DNS1          string `json:"dns1"`
	DNS2          string `json:"dns2"`
}

// DNSAccountInput DNS 账号创建/更新请求。
type DNSAccountInput struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Authorization string `json:"authorization"`
}

// ListCertificates 列出全部证书（不含 PEM）。
func ListCertificates(ctx context.Context) ([]model.TLSCertificate, error) {
	return repository.ListTLSCertificates(ctx)
}

// GetCertificate 获取证书详情（不含 PEM）。
func GetCertificate(ctx context.Context, id uint) (*model.TLSCertificate, error) {
	return repository.GetTLSCertificateByID(ctx, id)
}

// GetCertificateContent 获取证书 PEM 内容。
func GetCertificateContent(ctx context.Context, id uint) (*CertificateContent, error) {
	certificate, err := repository.GetTLSCertificateByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return certificateContent(certificate)
}

// GetCertificateContentOwned returns PEM content only for the certificate owner.
func GetCertificateContentOwned(ctx context.Context, id uint, ownerID uint64) (*CertificateContent, error) {
	certificate, err := repository.GetOwnedTLSCertificateByID(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}
	return certificateContent(certificate)
}

func certificateContent(certificate *model.TLSCertificate) (*CertificateContent, error) {
	keyPEM, err := openSensitive(certificate.KeyPEM)
	if err != nil {
		return nil, err
	}
	return &CertificateContent{
		ID:            certificate.ID,
		Name:          certificate.Name,
		CertPEM:       certificate.CertPEM,
		KeyPEM:        keyPEM,
		Remark:        certificate.Remark,
		Provider:      certificate.Provider,
		AcmeAccountID: certificate.AcmeAccountID,
		DNSAccountID:  certificate.DNSAccountID,
		KeyAlgorithm:  certificate.KeyAlgorithm,
		AutoRenew:     certificate.AutoRenew,
		PrimaryDomain: certificate.PrimaryDomain,
		OtherDomains:  certificate.OtherDomains,
		DisableCNAME:  certificate.DisableCNAME,
		SkipDNS:       certificate.SkipDNS,
		DNS1:          certificate.DNS1,
		DNS2:          certificate.DNS2,
		ApplyStatus:   certificate.ApplyStatus,
		ApplyMessage:  certificate.ApplyMessage,
	}, nil
}

// CreateCertificate 从 PEM 创建证书。
func CreateCertificate(ctx context.Context, input CertificateInput) (*model.TLSCertificate, error) {
	return createCertificate(ctx, 0, input)
}

// CreateCertificateOwned creates a user-owned uploaded certificate.
func CreateCertificateOwned(ctx context.Context, ownerID uint64, input CertificateInput) (*model.TLSCertificate, error) {
	return createCertificate(ctx, ownerID, input)
}

// CreateCertificateFromFilesOwned creates an uploaded certificate for a user.
func CreateCertificateFromFilesOwned(ctx context.Context, ownerID uint64, name string, certFile *multipart.FileHeader, keyFile *multipart.FileHeader, remark string) (*model.TLSCertificate, error) {
	if certFile == nil || keyFile == nil {
		return nil, errors.New(errCertificateFilesRequired)
	}
	certContent, err := readMultipartFile(certFile)
	if err != nil {
		return nil, err
	}
	keyContent, err := readMultipartFile(keyFile)
	if err != nil {
		return nil, err
	}
	return CreateCertificateOwned(ctx, ownerID, CertificateInput{Name: name, CertPEM: certContent, KeyPEM: keyContent, Remark: remark})
}

func createCertificate(ctx context.Context, ownerID uint64, input CertificateInput) (*model.TLSCertificate, error) {
	certificate, err := buildCertificate(ctx, nil, input)
	if err != nil {
		return nil, err
	}
	certificate.OwnerID = ownerID
	if err = repository.CreateTLSCertificateRecord(ctx, certificate); err != nil {
		if isUniqueConstraintError(err) {
			return nil, errors.New(errCertificateNameExists)
		}
		return nil, err
	}
	return sanitizeCertificateForResponse(certificate), nil
}

// ListCertificatesOwned lists only certificates owned by the caller.
func ListCertificatesOwned(ctx context.Context, ownerID uint64) ([]model.TLSCertificate, error) {
	return repository.ListOwnedTLSCertificates(ctx, ownerID)
}

// CreateCertificateFromFiles 从上传文件创建证书。
func CreateCertificateFromFiles(ctx context.Context, name string, certFile *multipart.FileHeader, keyFile *multipart.FileHeader, remark string) (*model.TLSCertificate, error) {
	if certFile == nil || keyFile == nil {
		return nil, errors.New(errCertificateFilesRequired)
	}
	certContent, err := readMultipartFile(certFile)
	if err != nil {
		return nil, err
	}
	keyContent, err := readMultipartFile(keyFile)
	if err != nil {
		return nil, err
	}
	return CreateCertificate(ctx, CertificateInput{
		Name:    name,
		CertPEM: certContent,
		KeyPEM:  keyContent,
		Remark:  remark,
	})
}

// UpdateCertificate 更新上传证书。
func UpdateCertificate(ctx context.Context, id uint, input CertificateInput) (*model.TLSCertificate, error) {
	existing, err := repository.GetTLSCertificateByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return updateCertificate(ctx, existing, input, 0)
}

// UpdateCertificateOwned updates an uploaded certificate only for its owner.
func UpdateCertificateOwned(ctx context.Context, id uint, ownerID uint64, input CertificateInput) (*model.TLSCertificate, error) {
	existing, err := repository.GetOwnedTLSCertificateByID(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}
	return updateCertificate(ctx, existing, input, ownerID)
}

func updateCertificate(ctx context.Context, existing *model.TLSCertificate, input CertificateInput, ownerID uint64) (*model.TLSCertificate, error) {
	certificate, err := buildCertificate(ctx, existing, input)
	if err != nil {
		return nil, err
	}
	err = saveCertificate(ctx, certificate, ownerID)
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, errors.New(errCertificateNameExists)
		}
		return nil, err
	}
	return sanitizeCertificateForResponse(certificate), nil
}

func saveCertificate(ctx context.Context, certificate *model.TLSCertificate, ownerID uint64) error {
	if ownerID > 0 {
		return repository.SaveOwnedTLSCertificate(ctx, certificate, ownerID)
	}
	return repository.SaveTLSCertificate(ctx, certificate)
}

// DeleteCertificate 删除证书。
func DeleteCertificate(ctx context.Context, id uint) error {
	if err := ensureCertificateNotReferenced(ctx, id); err != nil {
		return err
	}
	if _, err := repository.GetTLSCertificateByID(ctx, id); err != nil {
		return err
	}
	return repository.DeleteTLSCertificateRecord(ctx, id)
}

// DeleteCertificateOwned deletes a certificate only for its owner.
func DeleteCertificateOwned(ctx context.Context, id uint, ownerID uint64) error {
	if err := ensureCertificateNotReferenced(ctx, id); err != nil {
		return err
	}
	if _, err := repository.GetOwnedTLSCertificateByID(ctx, id, ownerID); err != nil {
		return err
	}
	return repository.DeleteOwnedTLSCertificateRecord(ctx, id, ownerID)
}

// ApplyCertificate 申请 ACME 证书。
func ApplyCertificate(ctx context.Context, input ApplyInput) (*model.TLSCertificate, error) {
	return applyCertificate(ctx, 0, input)
}

// ApplyCertificateOwned starts an ACME request for one user-owned certificate.
func ApplyCertificateOwned(ctx context.Context, ownerID uint64, input ApplyInput) (*model.TLSCertificate, error) {
	return applyCertificate(ctx, ownerID, input)
}

func applyCertificate(ctx context.Context, ownerID uint64, input ApplyInput) (*model.TLSCertificate, error) {
	cert := &model.TLSCertificate{
		OwnerID:  ownerID,
		Provider: tlsProviderACME,
		CertPEM:  " ",
		KeyPEM:   " ",
	}
	fillAcmeCertificateFields(cert, input)
	if cert.Name == "" {
		return nil, errors.New(errCertificateNameRequired)
	}
	if err := ensureDNSAccountAccessible(ctx, ownerID, input.DNSAccountID); err != nil {
		return nil, err
	}
	if err := repository.CreateTLSCertificateRecord(ctx, cert); err != nil {
		if isUniqueConstraintError(err) {
			return nil, errors.New(errCertificateNameExists)
		}
		return nil, err
	}

	go func(c *model.TLSCertificate) {
		asyncCtx := context.WithoutCancel(ctx)
		_ = obtainTLSCertificate(asyncCtx, c)
	}(cert)

	return sanitizeCertificateForResponse(cert), nil
}

// UpdateACMECertificate 更新 ACME 证书配置。
func UpdateACMECertificate(ctx context.Context, id uint, input ApplyInput) (*model.TLSCertificate, error) {
	cert, err := repository.GetTLSCertificateByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return updateACMECertificate(ctx, cert, input, 0)
}

// UpdateACMECertificateOwned updates an ACME certificate only for its owner.
func UpdateACMECertificateOwned(ctx context.Context, id uint, ownerID uint64, input ApplyInput) (*model.TLSCertificate, error) {
	cert, err := repository.GetOwnedTLSCertificateByID(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}
	return updateACMECertificate(ctx, cert, input, ownerID)
}

func updateACMECertificate(ctx context.Context, cert *model.TLSCertificate, input ApplyInput, ownerID uint64) (*model.TLSCertificate, error) {
	if cert.Provider != tlsProviderACME {
		return nil, errors.New(errCertificateOnlyACME)
	}
	if err := ensureDNSAccountAccessible(ctx, ownerID, input.DNSAccountID); err != nil {
		return nil, err
	}
	fillAcmeCertificateFields(cert, input)
	if cert.Name == "" {
		return nil, errors.New(errCertificateNameRequired)
	}
	saveErr := saveCertificate(ctx, cert, ownerID)
	if saveErr != nil {
		if isUniqueConstraintError(saveErr) {
			return nil, errors.New(errCertificateNameExists)
		}
		return nil, saveErr
	}

	go func(c *model.TLSCertificate) {
		asyncCtx := context.WithoutCancel(ctx)
		_ = obtainTLSCertificate(asyncCtx, c)
	}(cert)

	return sanitizeCertificateForResponse(cert), nil
}

// ConvertCertificateToACME 将上传证书转为 ACME 管理。
func ConvertCertificateToACME(ctx context.Context, id uint, input ApplyInput) (*model.TLSCertificate, error) {
	cert, err := repository.GetTLSCertificateByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertCertificateToACME(ctx, cert, input, 0)
}

// ConvertCertificateToACMEOwned converts an uploaded certificate only for its owner.
func ConvertCertificateToACMEOwned(ctx context.Context, id uint, ownerID uint64, input ApplyInput) (*model.TLSCertificate, error) {
	cert, err := repository.GetOwnedTLSCertificateByID(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}
	return convertCertificateToACME(ctx, cert, input, ownerID)
}

func convertCertificateToACME(ctx context.Context, cert *model.TLSCertificate, input ApplyInput, ownerID uint64) (*model.TLSCertificate, error) {
	if cert.Provider != "upload" {
		return nil, errors.New(errCertificateOnlyUploadConvert)
	}
	if cert.ApplyStatus == tlsApplyStatusApplying {
		return nil, errors.New(errCertificateAlreadyApplying)
	}
	if err := ensureDNSAccountAccessible(ctx, ownerID, input.DNSAccountID); err != nil {
		return nil, err
	}
	fillAcmeCertificateFields(cert, input)
	if cert.Name == "" {
		return nil, errors.New(errCertificateNameRequired)
	}
	cert.ApplyMessage = ""
	if err := saveCertificate(ctx, cert, ownerID); err != nil {
		if isUniqueConstraintError(err) {
			return nil, errors.New(errCertificateNameExists)
		}
		return nil, err
	}

	go func(c *model.TLSCertificate) {
		asyncCtx := context.WithoutCancel(ctx)
		if err := obtainTLSCertificate(asyncCtx, c); err != nil {
			return
		}
		latest, err := repository.GetTLSCertificateByID(asyncCtx, c.ID)
		if err != nil {
			return
		}
		latest.Provider = tlsProviderACME
		latest.ApplyStatus = tlsApplyStatusReady
		latest.ApplyMessage = ""
		_ = saveCertificate(asyncCtx, latest, ownerID)
	}(cert)

	return sanitizeCertificateForResponse(cert), nil
}

// RenewCertificate 续期 ACME 证书。
func RenewCertificate(ctx context.Context, id uint) (*model.TLSCertificate, error) {
	cert, err := repository.GetTLSCertificateByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return renewCertificate(ctx, cert, 0)
}

// RenewCertificateOwned renews an ACME certificate only for its owner.
func RenewCertificateOwned(ctx context.Context, id uint, ownerID uint64) (*model.TLSCertificate, error) {
	cert, err := repository.GetOwnedTLSCertificateByID(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}
	return renewCertificate(ctx, cert, ownerID)
}

func renewCertificate(ctx context.Context, cert *model.TLSCertificate, ownerID uint64) (*model.TLSCertificate, error) {
	if cert.Provider != tlsProviderACME {
		return nil, errors.New(errCertificateOnlyACMERenew)
	}

	payload, err := json.Marshal(SSLSingleRenewPayload{ID: cert.ID})
	if err != nil {
		return nil, err
	}

	_, err = task.DispatchTask(ctx, TaskTypeSSLSingleRenew, payload, "manual")
	if err != nil {
		return nil, err
	}

	cert.ApplyStatus = tlsApplyStatusApplying
	cert.ApplyMessage = ""
	if err := saveCertificate(ctx, cert, ownerID); err != nil {
		return nil, err
	}
	return sanitizeCertificateForResponse(cert), nil
}

// ListDNSAccounts 列出 DNS 账号。
func ListDNSAccounts(ctx context.Context) ([]model.DNSAccount, error) {
	return repository.ListDNSAccounts(ctx)
}

// CreateDNSAccount 创建 DNS 账号。
func CreateDNSAccount(ctx context.Context, input DNSAccountInput) (*model.DNSAccount, error) {
	account, err := buildDNSAccount(0, input)
	if err != nil {
		return nil, err
	}
	if err := repository.CreateDNSAccountRecord(ctx, account); err != nil {
		return nil, err
	}
	return sanitizeDNSAccountForResponse(account), nil
}

// UpdateDNSAccount 更新平台级 DNS 账号。
func UpdateDNSAccount(ctx context.Context, id uint, input DNSAccountInput) (*model.DNSAccount, error) {
	account, err := repository.GetPlatformDNSAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	built, err := buildDNSAccount(0, input)
	if err != nil {
		return nil, err
	}
	account.Name = built.Name
	account.Type = built.Type
	account.Authorization = built.Authorization
	if err := repository.SaveDNSAccount(ctx, account); err != nil {
		return nil, err
	}
	return sanitizeDNSAccountForResponse(account), nil
}

// DeleteDNSAccount 删除平台级 DNS 账号。
func DeleteDNSAccount(ctx context.Context, id uint) error {
	if _, err := repository.GetPlatformDNSAccountByID(ctx, id); err != nil {
		return err
	}
	count, err := repository.CountTLSCertificatesByDNSAccountID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New(errDNSAccountInUse)
	}
	return repository.DeleteDNSAccountRecord(ctx, id)
}

// ListOwnedDNSAccounts 列出某个普通用户自己的 DNS 账号。
func ListOwnedDNSAccounts(ctx context.Context, ownerID uint64) ([]model.DNSAccount, error) {
	return repository.ListOwnedDNSAccounts(ctx, ownerID)
}

// ListDNSAccountsForOwner 返回平台账号与当前用户自己的账号，供 ACME 申请选择。
func ListDNSAccountsForOwner(ctx context.Context, ownerID uint64) ([]model.DNSAccount, error) {
	return repository.ListDNSAccountsForOwner(ctx, ownerID)
}

// CreateOwnedDNSAccount 创建普通用户自己的 DNS 账号，最多 5 个。
func CreateOwnedDNSAccount(ctx context.Context, ownerID uint64, input DNSAccountInput) (*model.DNSAccount, error) {
	count, err := repository.CountOwnedDNSAccounts(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if count >= maxUserDNSAccounts {
		return nil, errors.New(errDNSAccountLimit)
	}
	account, err := buildDNSAccount(ownerID, input)
	if err != nil {
		return nil, err
	}
	if err := repository.CreateDNSAccountRecord(ctx, account); err != nil {
		return nil, err
	}
	return sanitizeDNSAccountForResponse(account), nil
}

// UpdateOwnedDNSAccount 更新普通用户自己的 DNS 账号。
func UpdateOwnedDNSAccount(ctx context.Context, id uint, ownerID uint64, input DNSAccountInput) (*model.DNSAccount, error) {
	account, err := repository.GetOwnedDNSAccountByID(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}
	built, err := buildDNSAccount(ownerID, input)
	if err != nil {
		return nil, err
	}
	account.Name = built.Name
	account.Type = built.Type
	account.Authorization = built.Authorization
	if err := repository.SaveOwnedDNSAccount(ctx, account, ownerID); err != nil {
		return nil, err
	}
	return sanitizeDNSAccountForResponse(account), nil
}

// DeleteOwnedDNSAccount 删除普通用户自己的 DNS 账号。
func DeleteOwnedDNSAccount(ctx context.Context, id uint, ownerID uint64) error {
	if _, err := repository.GetOwnedDNSAccountByID(ctx, id, ownerID); err != nil {
		return err
	}
	count, err := repository.CountTLSCertificatesByDNSAccountID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New(errDNSAccountInUse)
	}
	return repository.DeleteOwnedDNSAccountRecord(ctx, id, ownerID)
}

// ensureDNSAccountAccessible 校验普通用户只能使用平台账号或自己的 DNS 账号。
func ensureDNSAccountAccessible(ctx context.Context, ownerID uint64, dnsAccountID uint) error {
	if ownerID == 0 || dnsAccountID == 0 {
		return nil
	}
	account, err := repository.GetDNSAccountByID(ctx, dnsAccountID)
	if err != nil {
		return err
	}
	if account.OwnerID != 0 && account.OwnerID != ownerID {
		return errors.New("DNS 账号不属于当前用户")
	}
	return nil
}

// GetDefaultAcmeAccount 获取默认 ACME 账号。
func GetDefaultAcmeAccount(ctx context.Context) (*model.AcmeAccount, error) {
	account, err := repository.GetDefaultAcmeAccount(ctx)
	if err != nil {
		return nil, err
	}
	return sanitizeAcmeAccountForResponse(account), nil
}

func buildCertificate(_ context.Context, existing *model.TLSCertificate, input CertificateInput) (*model.TLSCertificate, error) {
	name := strings.TrimSpace(input.Name)
	certPEM := strings.TrimSpace(input.CertPEM)
	keyPEM := strings.TrimSpace(input.KeyPEM)
	remark := strings.TrimSpace(input.Remark)
	if name == "" {
		return nil, errors.New(errCertificateNameRequired)
	}
	if certPEM == "" || keyPEM == "" {
		return nil, errors.New(errCertificateContentRequired)
	}
	parsed, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errCertificateContentInvalid, err)
	}
	if len(parsed.Certificate) == 0 {
		return nil, errors.New(errCertificateContentInvalid)
	}
	leaf, err := parseLeafCertificate(certPEM)
	if err != nil {
		return nil, err
	}
	sealedKey, err := sealSensitive(keyPEM)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		existing = &model.TLSCertificate{
			Provider:    "upload",
			ApplyStatus: tlsApplyStatusReady,
		}
	}
	existing.Name = name
	existing.CertPEM = certPEM
	existing.KeyPEM = sealedKey
	existing.NotBefore = leaf.NotBefore
	existing.NotAfter = leaf.NotAfter
	existing.Remark = remark
	return existing, nil
}

func fillAcmeCertificateFields(cert *model.TLSCertificate, input ApplyInput) {
	cert.Name = strings.TrimSpace(input.Name)
	cert.Remark = strings.TrimSpace(input.Remark)
	cert.AcmeAccountID = input.AcmeAccountID
	cert.DNSAccountID = input.DNSAccountID
	cert.KeyAlgorithm = input.KeyAlgorithm
	cert.AutoRenew = input.AutoRenew
	cert.PrimaryDomain = strings.TrimSpace(input.PrimaryDomain)
	cert.OtherDomains = strings.TrimSpace(input.OtherDomains)
	cert.DisableCNAME = input.DisableCNAME
	cert.SkipDNS = input.SkipDNS
	cert.DNS1 = strings.TrimSpace(input.DNS1)
	cert.DNS2 = strings.TrimSpace(input.DNS2)
	cert.ApplyStatus = tlsApplyStatusApplying
}

func ensureCertificateNotReferenced(ctx context.Context, id uint) error {
	count, err := repository.CountZoneDomainsByCertificateID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New(errCertificateDeleteReferenced)
	}
	return nil
}

func sanitizeCertificateForResponse(certificate *model.TLSCertificate) *model.TLSCertificate {
	if certificate == nil {
		return nil
	}
	certCopy := *certificate
	certCopy.CertPEM = ""
	certCopy.KeyPEM = ""
	return &certCopy
}

func sanitizeDNSAccountForResponse(account *model.DNSAccount) *model.DNSAccount {
	if account == nil {
		return nil
	}
	certCopy := *account
	certCopy.Authorization = ""
	return &certCopy
}

func sanitizeAcmeAccountForResponse(account *model.AcmeAccount) *model.AcmeAccount {
	if account == nil {
		return nil
	}
	certCopy := *account
	certCopy.PrivateKey = ""
	return &certCopy
}
