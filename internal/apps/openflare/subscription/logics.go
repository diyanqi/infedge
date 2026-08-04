// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

//nolint:revive // Exported domain API is consumed across the application layers.
package subscription

import (
	"context"
	"crypto/md5" //nolint:gosec // EasyPay protocol requires MD5 signatures.
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/Rain-kl/Wavelet/internal/repository"
	"gorm.io/gorm"
)

const (
	easyPayFenPerYuan  = 100
	orderRandomBytes   = 5
	redeemCodeLength   = 16
	redeemCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

type PlanInput struct {
	Name                string `json:"name" binding:"required,max=128"`
	Description         string `json:"description" binding:"max=2000"`
	PriceFen            int64  `json:"price_fen" binding:"gte=0"`
	BillingMonths       int    `json:"billing_months" binding:"gte=1,lte=24"`
	HighSpeedBytes      int64  `json:"high_speed_bytes" binding:"gte=0"`
	ThrottleBytesPerSec int64  `json:"throttle_bytes_per_sec" binding:"gte=0"`
	DailyPublishLimit   int    `json:"daily_publish_limit" binding:"gte=0"`
	MaxZones            int    `json:"max_zones" binding:"gte=0"`
	MaxOrigins          int    `json:"max_origins" binding:"gte=0"`
	MaxRoutes           int    `json:"max_routes" binding:"gte=0"`
	MaxPages            int    `json:"max_pages" binding:"gte=0"`
	Enabled             bool   `json:"enabled"`
}

type ChannelInput struct {
	Name      string `json:"name" binding:"required,max=128"`
	Gateway   string `json:"gateway" binding:"required,url,max=512"`
	PID       string `json:"pid" binding:"required,max=128"`
	SecretKey string `json:"secret_key" binding:"max=255"`
	Enabled   bool   `json:"enabled"`
	Sort      int    `json:"sort"`
}

type CreateRedeemCodeInput struct {
	PlanID uint `json:"plan_id" binding:"required"`
}

type RedeemRequest struct {
	Code string `json:"code" binding:"required,max=64"`
}

func planFromInput(input PlanInput) *model.SubscriptionPlan {
	return &model.SubscriptionPlan{
		Name: input.Name, Description: input.Description, PriceFen: input.PriceFen,
		BillingMonths: input.BillingMonths, HighSpeedBytes: input.HighSpeedBytes,
		ThrottleBytesPerSec: input.ThrottleBytesPerSec, DailyPublishLimit: input.DailyPublishLimit,
		MaxZones: input.MaxZones, MaxOrigins: input.MaxOrigins, MaxRoutes: input.MaxRoutes,
		MaxPages: input.MaxPages, Enabled: input.Enabled,
	}
}

func channelFromInput(input ChannelInput) *model.PaymentChannel {
	return &model.PaymentChannel{Name: input.Name, Gateway: strings.TrimRight(input.Gateway, "/"), PID: input.PID, SecretKey: input.SecretKey, Enabled: input.Enabled, Sort: input.Sort}
}

func GenerateRedeemCode(ctx context.Context, planID uint) (*model.RedeemCode, error) {
	plan, err := repository.GetSubscriptionPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(errPlanNotFound)
		}
		return nil, err
	}
	if !plan.Enabled {
		return nil, errors.New(errPlanDisabled)
	}
	code, err := newRedeemCode()
	if err != nil {
		return nil, err
	}
	row := &model.RedeemCode{
		Code: code, PlanID: plan.ID, Plan: plan, Status: model.RedeemCodeStatusUnused,
	}
	if err := repository.CreateRedeemCode(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func RedeemPlan(ctx context.Context, userID uint64, rawCode string, now time.Time) (*model.UserSubscription, error) {
	code := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(rawCode), "-", ""))
	if code == "" {
		return nil, errors.New(errRedeemCodeInvalid)
	}
	subscription, err := repository.RedeemSubscriptionWithCode(ctx, code, userID, now)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(errRedeemCodeInvalid)
	}
	return subscription, err
}

func newRedeemCode() (string, error) {
	code := make([]byte, redeemCodeLength)
	max := big.NewInt(int64(len(redeemCodeAlphabet)))
	for i := range code {
		value, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		code[i] = redeemCodeAlphabet[value.Int64()]
	}
	return string(code), nil
}

func CreateOrder(ctx context.Context, userID uint64, planID, channelID uint, notifyURL, returnURL string) (string, *model.PaymentOrder, error) {
	plan, err := repository.GetSubscriptionPlanByID(ctx, planID)
	if err != nil {
		return "", nil, errors.New(errPlanNotFound)
	}
	if !plan.Enabled {
		return "", nil, errors.New(errPlanDisabled)
	}
	if plan.PriceFen == 0 {
		if err := repository.ActivateSubscription(ctx, userID, plan.ID, time.Now()); err != nil {
			return "", nil, err
		}
		return "", nil, nil
	}
	channel, err := repository.GetPaymentChannelByID(ctx, channelID)
	if err != nil || !channel.Enabled {
		return "", nil, errors.New(errChannelNotFound)
	}
	orderNo, err := newOrderNo()
	if err != nil {
		return "", nil, err
	}
	order := &model.PaymentOrder{OrderNo: orderNo, UserID: userID, PlanID: plan.ID, ChannelID: channel.ID, AmountFen: plan.PriceFen, Status: model.PaymentOrderPending}
	if err := repository.CreatePaymentOrder(ctx, order); err != nil {
		return "", nil, err
	}
	params := map[string]string{
		"pid": channel.PID, "type": "alipay", "out_trade_no": orderNo,
		"notify_url": notifyURL, "return_url": returnURL,
		"name": plan.Name, "money": fmt.Sprintf("%.2f", float64(plan.PriceFen)/easyPayFenPerYuan),
	}
	params["sign"] = easyPaySign(params, channel.SecretKey)
	params["sign_type"] = "MD5"
	endpoint := channel.Gateway + "/submit.php"
	query := url.Values{}
	for key, value := range params {
		query.Set(key, value)
	}
	return endpoint + "?" + query.Encode(), order, nil
}

func HandleNotify(ctx context.Context, channelID uint, values url.Values) error {
	channel, err := repository.GetPaymentChannelByID(ctx, channelID)
	if err != nil {
		return errors.New(errChannelNotFound)
	}
	params := make(map[string]string, len(values))
	for key, value := range values {
		if len(value) > 0 {
			params[key] = value[0]
		}
	}
	sign := params["sign"]
	delete(params, "sign")
	delete(params, "sign_type")
	if !strings.EqualFold(sign, easyPaySign(params, channel.SecretKey)) {
		return errors.New(errInvalidPaymentSign)
	}
	if params["pid"] != channel.PID {
		return errors.New(errInvalidPaymentChannel)
	}
	if params["trade_status"] != "TRADE_SUCCESS" && params["trade_status"] != "TRADE_FINISHED" {
		return nil
	}
	order, err := repository.GetPaymentOrderByOrderNo(ctx, params["out_trade_no"])
	if err != nil {
		return errors.New(errOrderNotFound)
	}
	if order.ChannelID != channelID {
		return errors.New(errInvalidPaymentChannel)
	}
	amount, err := parseMoneyFen(params["money"])
	if err != nil || amount != order.AmountFen {
		return errors.New(errPaymentAmountMismatch)
	}
	return repository.MarkPaymentOrderPaid(ctx, params["out_trade_no"], params["trade_no"], time.Now())
}

func parseMoneyFen(raw string) (int64, error) {
	amount, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || amount < 0 || math.IsInf(amount, 0) || math.IsNaN(amount) {
		return 0, errors.New("金额无效")
	}
	return int64(math.Round(amount * easyPayFenPerYuan)), nil
}

func easyPaySign(values map[string]string, secret string) string {
	keys := make([]string, 0, len(values))
	for key, value := range values {
		if key == "sign" || key == "sign_type" || strings.TrimSpace(value) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values[key])
	}
	digest := md5.Sum([]byte(strings.Join(parts, "&") + secret)) //nolint:gosec // EasyPay requires MD5.
	return hex.EncodeToString(digest[:])
}

func newOrderNo() (string, error) {
	buf := make([]byte, orderRandomBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return time.Now().Format("20060102150405") + strconv.FormatInt(time.Now().UnixNano()%1000000, 10) + hex.EncodeToString(buf), nil
}
