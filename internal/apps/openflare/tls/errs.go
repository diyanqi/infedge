// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package tls defines shared error messages for certificate management.
package tls

const (
	errCertificateNameRequired      = "certificate name cannot be empty"
	errCertificateNameExists        = "certificate name already exists"
	errCertificateContentRequired   = "certificate content and key content cannot be empty"
	errCertificateContentInvalid    = "certificate or key format is invalid"
	errCertificateDeleteReferenced  = "certificate is still referenced by proxy routes"
	errCertificateOnlyACME          = "only acme certificates can be updated via this endpoint"
	errCertificateOnlyUploadConvert = "only uploaded certificates can be converted to acme"
	errCertificateAlreadyApplying   = "certificate is already applying"
	errCertificateOnlyACMERenew     = "only acme certificates can be renewed"
	errCertificateFilesRequired     = "certificate file and key file cannot be empty"
	errCertificatePEMInvalid        = "证书 PEM 内容不合法"

	errDNSAccountInUse = "该 DNS 账号已被证书使用，无法删除"
	errDNSAccountLimit = "每个用户最多创建 5 个 DNS 账号"

	// #nosec G101 -- user-facing error messages, not credential material.
	errDNSAccountCredentialsRequired = "DNS 账号凭据不能为空"
	errDNSAccountTypeUnsupported     = "不支持的 DNS 服务商"
	errDNSAccountAuthJSONInvalid     = "凭据必须为 JSON 对象"
	// #nosec G101 -- user-facing error messages, not credential material.
	errDNSAccountCredentialsInvalid = "DNS 账号凭据格式不正确"
	errDNSAccountCloudflareMissing  = "cloudflare 凭据缺少 API Token"
	errDNSAccountAliyunMissing      = "阿里云凭据缺少 AccessKey 或 AccessKeySecret"
	errDNSAccountTencentMissing     = "腾讯云凭据缺少 SecretId 或 SecretKey"
	errDNSAccountHuaweiMissing      = "华为云凭据缺少 AccessKeyId、SecretAccessKey 或 Region"
	errDNSAccountTestFailed         = "测试连接失败"
)
