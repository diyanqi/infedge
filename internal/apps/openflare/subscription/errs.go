// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

//nolint:revive // Exported domain API is consumed across the application layers.
package subscription

const (
	errPlanNotFound          = "套餐不存在"
	errPlanDisabled          = "套餐已停用"
	errChannelNotFound       = "支付渠道不存在"
	errOrderNotFound         = "订单不存在"
	errInvalidPaymentSign    = "支付签名无效"
	errInvalidPaymentChannel = "支付渠道无效"
	errPaymentAmountMismatch = "支付金额不匹配"
	errNoPaymentChannel      = "暂无可用支付渠道"
	errNoSubscription        = "当前没有有效套餐"
	errRedeemCodeInvalid     = "兑换码无效或已使用"
)
