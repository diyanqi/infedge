// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package root registers custom business routes and frontend serving.
package root

import (
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/subscription"
	"github.com/gin-gonic/gin"
)

// RegisterCustomRootRoutes registers custom business routes that belong to the root path.
func RegisterCustomRootRoutes(r *gin.Engine) {
	// Payment callbacks are registered by the v1 custom route setup. This root
	// hook intentionally remains available for future webhook integrations.
	_ = subscription.HandleNotify
	_ = r
}
