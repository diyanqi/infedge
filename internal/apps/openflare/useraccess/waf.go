package useraccess

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Rain-kl/Wavelet/internal/apps/openflare/waf"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/Rain-kl/Wavelet/internal/repository"
	"github.com/Rain-kl/Wavelet/internal/shared/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type userPolicyView struct {
	CNAME             string         `json:"cname"`
	GlobalRules       []waf.RuleView `json:"global_rules"`
	DefaultLimitRate  string         `json:"default_limit_rate"`
	DefaultLimitReqIP string         `json:"default_limit_req_per_ip"`
}

func registerWAFRoutes(group *gin.RouterGroup) {
	group.GET("/waf/rule-groups", listUserRules)
	group.POST("/waf/rule-groups", createUserRule)
	group.GET("/waf/rule-groups/:id", getUserRule)
	group.POST("/waf/rule-groups/:id/update", updateUserRule)
	group.POST("/waf/rule-groups/:id/graph", updateUserRuleGraph)
	group.POST("/waf/rule-groups/:id/delete", deleteUserRule)
	group.GET("/waf/routes/:id/rule-groups", getUserRouteRules)
	group.POST("/waf/routes/:id/rule-groups", replaceUserRouteRules)
	group.GET("/policies", getUserPolicies)
}

func parseWAFID(c *gin.Context, key string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 32)
	if err != nil || id == 0 {
		response.AbortBadRequest(c, "ID 无效")
		return 0, false
	}
	return uint(id), true
}

func wafResult(c *gin.Context, value any, err error) {
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.AbortNotFound(c, "WAF 规则不存在")
		} else {
			response.AbortBadRequest(c, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, response.OK(value))
}

func listUserRules(c *gin.Context) {
	rows, err := waf.ListRulesForOwner(c.Request.Context(), userID(c))
	wafResult(c, rows, err)
}

func createUserRule(c *gin.Context) {
	var input waf.CreateRuleInput
	if !bind(c, &input) {
		return
	}
	if strings.TrimSpace(input.Host) == "" {
		response.AbortBadRequest(c, "用户 WAF 规则必须绑定一个子域 Host")
		return
	}
	row, err := waf.CreateRuleForOwner(c.Request.Context(), userID(c), input)
	wafResult(c, row, err)
}

func getUserRule(c *gin.Context) {
	id, ok := parseWAFID(c, "id")
	if !ok {
		return
	}
	row, err := waf.GetRuleForOwner(c.Request.Context(), id, userID(c))
	wafResult(c, row, err)
}

func updateUserRule(c *gin.Context) {
	id, ok := parseWAFID(c, "id")
	if !ok {
		return
	}
	var input waf.UpdateRuleMetaInput
	if !bind(c, &input) {
		return
	}
	row, err := waf.UpdateRuleMetaForOwner(c.Request.Context(), id, userID(c), input)
	wafResult(c, row, err)
}

func updateUserRuleGraph(c *gin.Context) {
	id, ok := parseWAFID(c, "id")
	if !ok {
		return
	}
	var input waf.SaveRuleGraphInput
	if !bind(c, &input) {
		return
	}
	row, err := waf.SaveRuleGraphForOwner(c.Request.Context(), id, userID(c), input)
	wafResult(c, row, err)
}

func deleteUserRule(c *gin.Context) {
	id, ok := parseWAFID(c, "id")
	if !ok {
		return
	}
	group, err := repository.GetOpenFlareWAFRuleGroupForOwner(c.Request.Context(), id, userID(c))
	if err != nil {
		wafResult(c, nil, err)
		return
	}
	if group.IsGlobal {
		response.AbortBadRequest(c, "全局 WAF 规则只读")
		return
	}
	wafResult(c, nil, repository.DeleteOpenFlareWAFRuleGroupWithBindings(c.Request.Context(), id))
}

func ownedSingleHostRoute(c *gin.Context) (uint, string, bool) {
	routeID, ok := parseWAFID(c, "id")
	if !ok {
		return 0, "", false
	}
	if _, err := repository.GetOwnedProxyRouteByID(c.Request.Context(), routeID, userID(c)); err != nil {
		response.AbortNotFound(c, "CDN 规则不存在")
		return 0, "", false
	}
	domains, err := repository.ListZoneDomainsByRouteID(c.Request.Context(), routeID)
	if err != nil || len(domains) != 1 {
		response.AbortBadRequest(c, "普通用户的防火墙规则必须只绑定一个子域")
		return 0, "", false
	}
	return routeID, domains[0].Domain, true
}

func getUserRouteRules(c *gin.Context) {
	routeID, _, ok := ownedSingleHostRoute(c)
	if !ok {
		return
	}
	row, err := waf.GetSiteRuleGroupsForOwner(c.Request.Context(), routeID, userID(c))
	wafResult(c, row, err)
}

func replaceUserRouteRules(c *gin.Context) {
	routeID, host, ok := ownedSingleHostRoute(c)
	if !ok {
		return
	}
	var input waf.IDsRequest
	if !bind(c, &input) {
		return
	}
	for _, id := range input.IDs {
		group, err := repository.GetOpenFlareWAFRuleGroupForOwner(c.Request.Context(), id, userID(c))
		if err != nil || group.IsGlobal {
			response.AbortBadRequest(c, "只能绑定自己的 WAF 规则；全局规则会自动优先执行")
			return
		}
		if group.Host != "" && !strings.EqualFold(group.Host, host) {
			response.AbortBadRequest(c, "WAF 规则 Host 与子域不匹配")
			return
		}
	}
	row, err := waf.ReplaceSiteRuleGroups(c.Request.Context(), routeID, input.IDs)
	wafResult(c, row, err)
}

func getUserPolicies(c *gin.Context) {
	rules, err := waf.ListRulesForOwner(c.Request.Context(), userID(c))
	if err != nil {
		wafResult(c, nil, err)
		return
	}
	global := make([]waf.RuleView, 0, len(rules))
	for _, rule := range rules {
		if rule.IsGlobal {
			global = append(global, rule)
		}
	}
	rate, _ := repository.GetSystemConfigByKey(c.Request.Context(), model.ConfigKeyOpenRestyDefaultLimitRate)
	requestRate, _ := repository.GetSystemConfigByKey(c.Request.Context(), model.ConfigKeyOpenRestyDefaultLimitReqPerIP)
	c.JSON(http.StatusOK, response.OK(userPolicyView{CNAME: "cname.edge.infvar.com", GlobalRules: global, DefaultLimitRate: rate.Value, DefaultLimitReqIP: requestRate.Value}))
}
