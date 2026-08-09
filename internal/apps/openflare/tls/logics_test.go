// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Rain-kl/Wavelet/internal/repository"

	"github.com/Rain-kl/Wavelet/internal/infra/config"
	db "github.com/Rain-kl/Wavelet/internal/infra/persistence"
	"github.com/Rain-kl/Wavelet/internal/infra/task"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var tlsTestDBMu sync.Mutex

func setupTLSTestDB(t *testing.T) func() {
	t.Helper()
	tlsTestDBMu.Lock()

	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)
	require.NoError(t, sqliteDB.AutoMigrate(
		&model.TLSCertificate{},
		&model.Zone{},
		&model.ZoneDomain{},
		&model.DNSAccount{},
		&model.AcmeAccount{},
		&model.TaskExecution{}, // 异步任务执行记录也需要 migrate
	))

	db.SetDB(sqliteDB)
	oldSecret := config.Config.App.SessionSecret
	config.Config.App.SessionSecret = "test_session_secret_for_tls_encryption"

	mr, err := miniredis.Run()
	require.NoError(t, err)

	task.AsynqClient = asynq.NewClient(asynq.RedisClientOpt{
		Addr: mr.Addr(),
	})

	task.RegisterTaskMeta(SSLSingleRenewMeta)

	return func() {
		if task.AsynqClient != nil {
			_ = task.AsynqClient.Close()
			task.AsynqClient = nil
		}
		mr.Close()
		db.SetDB(nil)
		config.Config.App.SessionSecret = oldSecret
		tlsTestDBMu.Unlock()
	}
}

func TestDeleteCertificateRejectsZoneDomainReference(t *testing.T) {
	cleanup := setupTLSTestDB(t)
	defer cleanup()
	ctx := context.Background()

	certPEM, keyPEM := generateTestCertificatePair(t, []string{"api.example.com"})
	certificate, err := CreateCertificate(ctx, CertificateInput{Name: "api-cert", CertPEM: certPEM, KeyPEM: keyPEM})
	require.NoError(t, err)
	zone := &model.Zone{Domain: "example.com"}
	require.NoError(t, db.DB(ctx).Create(zone).Error)
	require.NoError(t, db.DB(ctx).Create(&model.ZoneDomain{ZoneID: zone.ID, Domain: "api.example.com", CertID: &certificate.ID}).Error)

	err = DeleteCertificate(ctx, certificate.ID)
	require.EqualError(t, err, errCertificateDeleteReferenced)
}

func generateTestCertificatePair(t *testing.T, dnsNames []string) (string, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: dnsNames[0],
		},
		DNSNames:    dnsNames,
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	return string(certPEM), string(keyPEM)
}

func TestCreateCertificateEncryptsPrivateKey(t *testing.T) {
	cleanup := setupTLSTestDB(t)
	defer cleanup()
	ctx := context.Background()

	certPEM, keyPEM := generateTestCertificatePair(t, []string{"secure.example.com"})
	certificate, err := CreateCertificate(ctx, CertificateInput{
		Name:    "secure-cert",
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
	})
	require.NoError(t, err)

	stored, err := repository.GetTLSCertificateByID(ctx, certificate.ID)
	require.NoError(t, err)
	assert.NotEqual(t, keyPEM, stored.KeyPEM)
	assert.Contains(t, stored.KeyPEM, sensitiveValuePrefix)

	content, err := GetCertificateContent(ctx, certificate.ID)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(keyPEM), strings.TrimSpace(content.KeyPEM))
}

func TestCreateOwnedDNSAccountEnforcesFiveLimit(t *testing.T) {
	cleanup := setupTLSTestDB(t)
	defer cleanup()
	ctx := context.Background()

	for i := 0; i < maxUserDNSAccounts; i++ {
		account, err := CreateOwnedDNSAccount(ctx, 100, DNSAccountInput{
			Name:          "cloudflare-" + string(rune('a'+i)),
			Type:          "cloudflare",
			Authorization: `{"api_token":"token-"}`,
		})
		require.NoError(t, err)
		require.Equal(t, uint64(100), account.OwnerID)
	}

	_, err := CreateOwnedDNSAccount(ctx, 100, DNSAccountInput{
		Name:          "overflow",
		Type:          "cloudflare",
		Authorization: `{"api_token":"token"}`,
	})
	require.EqualError(t, err, errDNSAccountLimit)

	// 其他用户不受影响，仍可创建自己的账号。
	account, err := CreateOwnedDNSAccount(ctx, 200, DNSAccountInput{
		Name:          "another-user",
		Type:          "aliyun",
		Authorization: `{"access_key_id":"id","access_key_secret":"secret"}`,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(200), account.OwnerID)
}

func TestOwnedDNSAccountScopedToOwner(t *testing.T) {
	cleanup := setupTLSTestDB(t)
	defer cleanup()
	ctx := context.Background()

	account, err := CreateOwnedDNSAccount(ctx, 100, DNSAccountInput{
		Name:          "my-account",
		Type:          "cloudflare",
		Authorization: `{"api_token":"token"}`,
	})
	require.NoError(t, err)

	_, err = UpdateOwnedDNSAccount(ctx, account.ID, 200, DNSAccountInput{
		Name:          "hijack",
		Type:          "cloudflare",
		Authorization: `{"api_token":"other"}`,
	})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	err = DeleteOwnedDNSAccount(ctx, account.ID, 200)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	rows, err := ListOwnedDNSAccounts(ctx, 100)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, account.ID, rows[0].ID)
}
