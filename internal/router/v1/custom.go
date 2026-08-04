// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package v1 contains router registrations for API V1
package v1

import (
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/subscription"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/useraccess"
	"github.com/gin-gonic/gin"
)

// RegisterCustomRoutes registers custom business routes to keep routing clean and stable.
func RegisterCustomRoutes(apiV1Router *gin.RouterGroup, roots ...*gin.Engine) {
	custom := apiV1Router.Group("/custom")
	var root *gin.Engine
	if len(roots) > 0 {
		root = roots[0]
	}
	subscription.RegisterRoutes(custom, root)
	useraccess.RegisterRoutes(custom)
}
