// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeDNSProviderType(t *testing.T) {
	assert.Equal(t, ProviderCloudflare, normalizeDNSProviderType("cloudflare"))
	assert.Equal(t, ProviderAliyun, normalizeDNSProviderType("aliyun"))
	assert.Equal(t, ProviderAliyun, normalizeDNSProviderType("alicloud"))
	assert.Equal(t, ProviderAliyun, normalizeDNSProviderType("alidns"))
	assert.Equal(t, ProviderTencent, normalizeDNSProviderType("tencent"))
	assert.Equal(t, ProviderTencent, normalizeDNSProviderType("tencentcloud"))
	assert.Equal(t, ProviderTencent, normalizeDNSProviderType("dnspod"))
	assert.Equal(t, ProviderTencent, normalizeDNSProviderType("qcloud"))
	assert.Equal(t, ProviderHuawei, normalizeDNSProviderType("huawei"))
	assert.Equal(t, ProviderHuawei, normalizeDNSProviderType("huaweicloud"))
	assert.Equal(t, ProviderCloudflare, normalizeDNSProviderType(" Cloudflare "))
	assert.Equal(t, "", normalizeDNSProviderType("aws"))
}

func TestNormalizeDNSAuthorization(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		raw       string
		want      string
		wantError string
	}{
		{
			name:     "cloudflare plain token",
			provider: ProviderCloudflare,
			raw:      "  cf-token  ",
			want:     `{"api_token":"cf-token"}`,
		},
		{
			name:     "cloudflare json token alias",
			provider: ProviderCloudflare,
			raw:      `{"token":"cf-json-token"}`,
			want:     `{"api_token":"cf-json-token"}`,
		},
		{
			name:      "cloudflare missing token",
			provider:  ProviderCloudflare,
			raw:       `{}`,
			wantError: errDNSAccountCloudflareMissing,
		},
		{
			name:     "aliyun canonical",
			provider: ProviderAliyun,
			raw:      `{"access_key":"ak","access_key_secret":"sk"}`,
			want:     `{"access_key":"ak","access_key_secret":"sk"}`,
		},
		{
			name:     "aliyun aliases and optional fields",
			provider: ProviderAliyun,
			raw:      `{"api_key":"ak-id","secret_key":"ak-secret","security_token":"token","region_id":"cn-beijing"}`,
			want:     `{"access_key":"ak-id","access_key_secret":"ak-secret","region_id":"cn-beijing","security_token":"token"}`,
		},
		{
			name:      "aliyun missing secret",
			provider:  ProviderAliyun,
			raw:       `{"access_key":"ak"}`,
			wantError: errDNSAccountAliyunMissing,
		},
		{
			name:     "tencent canonical",
			provider: ProviderTencent,
			raw:      `{"secret_id":"id","secret_key":"key"}`,
			want:     `{"secret_id":"id","secret_key":"key"}`,
		},
		{
			name:     "tencent aliases and optional fields",
			provider: ProviderTencent,
			raw:      `{"access_key_id":"q-id","access_key_secret":"q-key","session_token":"st","region":"ap-guangzhou"}`,
			want:     `{"region":"ap-guangzhou","secret_id":"q-id","secret_key":"q-key","session_token":"st"}`,
		},
		{
			name:      "tencent missing id",
			provider:  ProviderTencent,
			raw:       `{"secret_key":"key"}`,
			wantError: errDNSAccountTencentMissing,
		},
		{
			name:     "huawei canonical",
			provider: ProviderHuawei,
			raw:      `{"access_key_id":"ak","secret_access_key":"sk","region":"cn-north-4"}`,
			want:     `{"access_key_id":"ak","region":"cn-north-4","secret_access_key":"sk"}`,
		},
		{
			name:     "huawei aliases",
			provider: ProviderHuawei,
			raw:      `{"ak":"ak-id","sk":"ak-secret","region":"cn-east-3"}`,
			want:     `{"access_key_id":"ak-id","region":"cn-east-3","secret_access_key":"ak-secret"}`,
		},
		{
			name:      "huawei missing region",
			provider:  ProviderHuawei,
			raw:       `{"access_key_id":"ak","secret_access_key":"sk"}`,
			wantError: errDNSAccountHuaweiMissing,
		},
		{
			name:      "aliyun requires json",
			provider:  ProviderAliyun,
			raw:       "ak:sk",
			wantError: errDNSAccountAuthJSONInvalid,
		},
		{
			name:      "empty credentials",
			provider:  ProviderHuawei,
			raw:       "  ",
			wantError: errDNSAccountCredentialsRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeDNSAuthorization(tt.provider, tt.raw)
			if tt.wantError != "" {
				require.EqualError(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, got)
		})
	}
}

func TestBuildDNSAccountNormalizesCredentials(t *testing.T) {
	account, err := buildDNSAccount(0, DNSAccountInput{
		Name:          "华为云",
		Type:          "huaweicloud",
		Authorization: `{"ak":"ak-id","sk":"ak-secret","region":"cn-north-4"}`,
	})
	require.NoError(t, err)
	assert.Equal(t, ProviderHuawei, account.Type)
	assert.JSONEq(t, `{"access_key_id":"ak-id","secret_access_key":"ak-secret","region":"cn-north-4"}`, account.Authorization)

	_, err = buildDNSAccount(0, DNSAccountInput{
		Name:          "bad",
		Type:          "huawei",
		Authorization: `{"access_key_id":"ak"}`,
	})
	require.EqualError(t, err, errDNSAccountHuaweiMissing)

	_, err = buildDNSAccount(0, DNSAccountInput{
		Name:          "unsupported",
		Type:          "aws",
		Authorization: `{"key":"value"}`,
	})
	require.EqualError(t, err, errDNSAccountTypeUnsupported)
}
