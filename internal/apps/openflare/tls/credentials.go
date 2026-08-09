// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/Rain-kl/Wavelet/pkg/httppool"
	"github.com/Rain-kl/Wavelet/pkg/logger"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	alidns "github.com/go-acme/alidns-20150109/v4/client"
	dnspod "github.com/go-acme/tencentclouddnspod/v20210323"
	hwauthbasic "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	hwconfig "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	hwdns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	hwmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	hwregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const (
	// ProviderCloudflare is the canonical Cloudflare provider name.
	ProviderCloudflare = "cloudflare"
	// ProviderAliyun is the canonical Alibaba Cloud DNS provider name.
	ProviderAliyun = "aliyun"
	// ProviderTencent is the canonical Tencent Cloud DNSPod provider name.
	ProviderTencent = "tencent"
	// ProviderHuawei is the canonical Huawei Cloud DNS provider name.
	ProviderHuawei = "huawei"

	dnsCredentialTestTimeout = 15 * time.Second
	aliyunDefaultRegionID    = "cn-hangzhou"
)

// TestDNSAccount validates credentials against a read-only DNS list API.
func TestDNSAccount(ctx context.Context, input DNSAccountInput) error {
	dnsType, authorization, err := normalizeDNSAccountCredentials(input)
	if err != nil {
		return err
	}
	creds, err := unmarshalCredentials(authorization)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, dnsCredentialTestTimeout)
	defer cancel()

	switch dnsType {
	case ProviderCloudflare:
		err = verifyCloudflareCredentials(ctx, creds)
	case ProviderAliyun:
		err = verifyAliyunCredentials(ctx, creds)
	case ProviderTencent:
		err = verifyTencentCredentials(ctx, creds)
	case ProviderHuawei:
		err = verifyHuaweiCredentials(creds)
	default:
		return errors.New(errDNSAccountTypeUnsupported)
	}
	if err != nil {
		logger.ErrorF(ctx, "dns account credential test failed: %v", err)
		return fmt.Errorf("%s：%v", errDNSAccountTestFailed, err)
	}
	return nil
}

func buildDNSAccount(ownerID uint64, input DNSAccountInput) (*model.DNSAccount, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("DNS 账号参数不完整")
	}
	dnsType, authorization, err := normalizeDNSAccountCredentials(input)
	if err != nil {
		return nil, err
	}
	sealed, err := sealSensitive(authorization)
	if err != nil {
		return nil, err
	}
	return &model.DNSAccount{
		OwnerID:       ownerID,
		Name:          name,
		Type:          dnsType,
		Authorization: sealed,
	}, nil
}

func normalizeDNSAccountCredentials(input DNSAccountInput) (string, string, error) {
	dnsType := normalizeDNSProviderType(input.Type)
	if dnsType == "" {
		return "", "", errors.New(errDNSAccountTypeUnsupported)
	}
	authorization, err := normalizeDNSAuthorization(dnsType, input.Authorization)
	if err != nil {
		return "", "", err
	}
	return dnsType, authorization, nil
}

func normalizeDNSProviderType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case ProviderCloudflare:
		return ProviderCloudflare
	case ProviderAliyun, "alicloud", "alidns":
		return ProviderAliyun
	case ProviderTencent, "tencentcloud", "dnspod", "qcloud":
		return ProviderTencent
	case ProviderHuawei, "huaweicloud":
		return ProviderHuawei
	default:
		return ""
	}
}

func normalizeDNSAuthorization(dnsType, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New(errDNSAccountCredentialsRequired)
	}
	if dnsType == ProviderCloudflare {
		return normalizeCloudflareAuthorization(raw)
	}

	creds, err := unmarshalCredentials(raw)
	if err != nil {
		return "", err
	}
	switch dnsType {
	case ProviderAliyun:
		return normalizeAliyunCredentials(creds)
	case ProviderTencent:
		return normalizeTencentCredentials(creds)
	case ProviderHuawei:
		return normalizeHuaweiCredentials(creds)
	default:
		return "", errors.New(errDNSAccountTypeUnsupported)
	}
}

func normalizeCloudflareAuthorization(raw string) (string, error) {
	token := strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "{") {
		creds, err := unmarshalCredentials(raw)
		if err != nil {
			return "", err
		}
		token = strings.TrimSpace(firstNonEmpty(creds["api_token"], creds["token"]))
	}
	if token == "" {
		return "", errors.New(errDNSAccountCloudflareMissing)
	}
	return marshalCredentials(map[string]string{"api_token": token})
}

func normalizeAliyunCredentials(creds map[string]string) (string, error) {
	accessKey := strings.TrimSpace(firstNonEmpty(creds["access_key"], creds["access_key_id"], creds["api_key"]))
	accessKeySecret := strings.TrimSpace(firstNonEmpty(creds["access_key_secret"], creds["secret_key"]))
	if accessKey == "" || accessKeySecret == "" {
		return "", errors.New(errDNSAccountAliyunMissing)
	}
	normalized := map[string]string{
		"access_key":        accessKey,
		"access_key_secret": accessKeySecret,
	}
	if token := strings.TrimSpace(creds["security_token"]); token != "" {
		normalized["security_token"] = token
	}
	if regionID := strings.TrimSpace(creds["region_id"]); regionID != "" {
		normalized["region_id"] = regionID
	}
	return marshalCredentials(normalized)
}

func normalizeTencentCredentials(creds map[string]string) (string, error) {
	secretID := strings.TrimSpace(firstNonEmpty(creds["secret_id"], creds["secretId"], creds["access_key_id"]))
	secretKey := strings.TrimSpace(firstNonEmpty(creds["secret_key"], creds["secretKey"], creds["access_key_secret"]))
	if secretID == "" || secretKey == "" {
		return "", errors.New(errDNSAccountTencentMissing)
	}
	normalized := map[string]string{
		"secret_id":  secretID,
		"secret_key": secretKey,
	}
	if token := strings.TrimSpace(creds["session_token"]); token != "" {
		normalized["session_token"] = token
	}
	if region := strings.TrimSpace(creds["region"]); region != "" {
		normalized["region"] = region
	}
	return marshalCredentials(normalized)
}

func normalizeHuaweiCredentials(creds map[string]string) (string, error) {
	accessKeyID := strings.TrimSpace(firstNonEmpty(creds["access_key_id"], creds["access_key"], creds["ak"]))
	secretAccessKey := strings.TrimSpace(firstNonEmpty(creds["secret_access_key"], creds["secret_key"], creds["sk"]))
	region := strings.TrimSpace(creds["region"])
	if accessKeyID == "" || secretAccessKey == "" || region == "" {
		return "", errors.New(errDNSAccountHuaweiMissing)
	}
	return marshalCredentials(map[string]string{
		"access_key_id":     accessKeyID,
		"secret_access_key": secretAccessKey,
		"region":            region,
	})
}

func unmarshalCredentials(raw string) (map[string]string, error) {
	var creds map[string]string
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return nil, errors.New(errDNSAccountAuthJSONInvalid)
	}
	return creds, nil
}

func marshalCredentials(creds map[string]string) (string, error) {
	data, err := json.Marshal(creds)
	if err != nil {
		return "", errors.New(errDNSAccountCredentialsInvalid)
	}
	return string(data), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func verifyCloudflareCredentials(ctx context.Context, creds map[string]string) error {
	token := strings.TrimSpace(creds["api_token"])
	if token == "" {
		return errors.New(errDNSAccountCloudflareMissing)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.cloudflare.com/client/v4/user/tokens/verify", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := httppool.NewClient(dnsCredentialTestTimeout)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cloudflare API 返回 %d", resp.StatusCode)
	}
	var body struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	if !body.Success {
		return errors.New("cloudflare Token 校验失败")
	}
	return nil
}

func verifyAliyunCredentials(ctx context.Context, creds map[string]string) error {
	regionID := strings.TrimSpace(creds["region_id"])
	if regionID == "" {
		regionID = aliyunDefaultRegionID
	}
	config := new(openapi.Config).
		SetRegionId(regionID).
		SetAccessKeyId(creds["access_key"]).
		SetAccessKeySecret(creds["access_key_secret"])
	if token := strings.TrimSpace(creds["security_token"]); token != "" {
		config.SetSecurityToken(token)
	}
	client, err := alidns.NewClient(config)
	if err != nil {
		return err
	}
	_, err = alidns.DescribeDomainsWithContext(ctx, client, new(alidns.DescribeDomainsRequest), &dara.RuntimeOptions{})
	return err
}

func verifyTencentCredentials(ctx context.Context, creds map[string]string) error {
	var credential common.CredentialIface
	if token := strings.TrimSpace(creds["session_token"]); token != "" {
		credential = common.NewTokenCredential(creds["secret_id"], creds["secret_key"], token)
	} else {
		credential = common.NewCredential(creds["secret_id"], creds["secret_key"])
	}

	clientProfile := profile.NewClientProfile()
	clientProfile.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	client, err := dnspod.NewClient(credential, creds["region"], clientProfile)
	if err != nil {
		return err
	}
	_, err = dnspod.DescribeDomainListWithContext(ctx, client, dnspod.NewDescribeDomainListRequest())
	return err
}

func verifyHuaweiCredentials(creds map[string]string) error {
	auth, err := hwauthbasic.NewCredentialsBuilder().
		WithAk(creds["access_key_id"]).
		WithSk(creds["secret_access_key"]).
		SafeBuild()
	if err != nil {
		return err
	}
	region, err := hwregion.SafeValueOf(creds["region"])
	if err != nil {
		return fmt.Errorf("华为云区域无效：%w", err)
	}
	hcClient, err := hwdns.DnsClientBuilder().
		WithHttpConfig(hwconfig.DefaultHttpConfig().WithTimeout(dnsCredentialTestTimeout)).
		WithRegion(region).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return err
	}
	_, err = hwdns.NewDnsClient(hcClient).ListPublicZones(&hwmodel.ListPublicZonesRequest{})
	return err
}
