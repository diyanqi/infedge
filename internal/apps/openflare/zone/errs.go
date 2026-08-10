// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package zone

const (
	errZoneDomainRequired        = "域名不能为空"
	errZoneRootInvalid           = "zone 必须是有效的注册根域"
	errDomainInvalid             = "域名格式不合法"
	errDomainWildcardUnsupported = "不支持通配符域名"
	errDomainOutsideZone         = "域名不属于该 Zone"
	errZoneNotFound              = "Zone 不存在"
	errDomainNotFound            = "域名不存在"
	errDomainExists              = "域名已存在"
	errDomainInUse               = "该域名已被其他站点绑定，请先解除对应规则"
	errCertificateNotFound       = "所选证书不存在"
	errDomainBoundToRoute        = "域名已绑定反代路由，请先解除绑定"
	errZoneHasDomains            = "根域下仍有域名，请先删除全部域名"
	errStatsRangeInvalid         = "时间范围无效，请选择 24h、7d 或 30d"
	errSiteDomainRequired        = "请输入要接入的域名"
)

// IsConflict reports whether err is a domain conflict that maps to HTTP 409.
func IsConflict(err error) bool {
	return err != nil && (err.Error() == errDomainExists || err.Error() == errDomainInUse)
}
