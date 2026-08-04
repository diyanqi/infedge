package useraccess

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Rain-kl/Wavelet/internal/apps/oauth"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/pages"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/Rain-kl/Wavelet/internal/repository"
	"github.com/Rain-kl/Wavelet/internal/shared/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerPagesRoutes(group *gin.RouterGroup) {
	group.GET("/pages", listPages)
	group.GET("/pages/:id", getPage)
	group.POST("/pages", createPage)
	group.POST("/pages/:id/update", updatePage)
	group.POST("/pages/:id/delete", deletePage)
	group.GET("/pages/:id/source", getPageSource)
	group.POST("/pages/:id/source/update", updatePageSource)
	group.POST("/pages/:id/source/delete", deletePageSource)
	group.POST("/pages/:id/source/check", checkPageSource)
	group.POST("/pages/:id/source/sync", syncPageSource)
	group.GET("/pages/:id/deployments", listPageDeployments)
	group.POST("/pages/:id/deployments/upload", uploadPageDeployment)
	group.POST("/pages/:id/deployments/upload-from-url", uploadPageDeploymentFromURL)
	group.POST("/pages/:id/deployments/:deployment_id/activate", activatePageDeployment)
	group.POST("/pages/:id/deployments/:deployment_id/delete", deletePageDeployment)
	group.GET("/pages/deployments/:deployment_id/files", listPageDeploymentFiles)
}

func pageID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.AbortBadRequest(c, "ID 无效")
		return 0, false
	}
	return uint(id), true
}

func deploymentID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("deployment_id"), 10, 32)
	if err != nil || id == 0 {
		response.AbortBadRequest(c, "部署 ID 无效")
		return 0, false
	}
	return uint(id), true
}

func pageActor(c *gin.Context) string { return fmt.Sprintf("user:%d", userID(c)) }

func pageAdmin(c *gin.Context) bool {
	user, ok := oauth.GetFromContext[*model.User](c, oauth.UserObjKey)
	return ok && user != nil && user.IsAdmin
}

func ensurePage(c *gin.Context, id uint) bool {
	if pageAdmin(c) {
		return true
	}
	if err := pages.EnsureProjectOwned(c.Request.Context(), id, userID(c)); err != nil {
		response.AbortNotFound(c, "Pages 项目不存在")
		return false
	}
	return true
}

func handlePageResult(c *gin.Context, value any, err error) {
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.AbortNotFound(c, "Pages 项目不存在")
		} else {
			response.AbortBadRequest(c, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, response.OK(value))
}

func listPages(c *gin.Context) {
	var rows []pages.View
	var err error
	if pageAdmin(c) {
		rows, err = pages.ListProjects(c.Request.Context())
	} else {
		rows, err = pages.ListProjectsOwned(c.Request.Context(), userID(c))
	}
	handlePageResult(c, rows, err)
}

func getPage(c *gin.Context) {
	id, ok := pageID(c)
	if !ok || !ensurePage(c, id) {
		return
	}
	var row *pages.View
	var err error
	if pageAdmin(c) {
		row, err = pages.GetProject(c.Request.Context(), id)
	} else {
		row, err = pages.GetProjectOwned(c.Request.Context(), id, userID(c))
	}
	handlePageResult(c, row, err)
}

func createPage(c *gin.Context) {
	ctx, uid := c.Request.Context(), userID(c)
	var input pages.Input
	if !bind(c, &input) {
		return
	}
	var row *pages.View
	var err error
	if pageAdmin(c) {
		row, err = pages.CreateProject(ctx, input)
	} else {
		plan, planErr := planForUser(ctx, uid)
		if planErr == nil {
			planErr = checkQuota(ctx, uid, "pages_projects", plan.MaxPages)
		}
		if planErr != nil {
			response.AbortBadRequest(c, planErr.Error())
			return
		}
		row, err = pages.CreateProjectOwned(ctx, uid, input)
	}
	handlePageResult(c, row, err)
}

func updatePage(c *gin.Context) {
	id, ok := pageID(c)
	if !ok || !ensurePage(c, id) {
		return
	}
	var input pages.Input
	if !bind(c, &input) {
		return
	}
	row, err := pages.UpdateProject(c.Request.Context(), id, input)
	handlePageResult(c, row, err)
}

func deletePage(c *gin.Context) {
	id, ok := pageID(c)
	if !ok || !ensurePage(c, id) {
		return
	}
	handlePageResult(c, nil, pages.DeleteProject(c.Request.Context(), id))
}

func getPageSource(c *gin.Context) {
	id, ok := pageID(c)
	if !ok || !ensurePage(c, id) {
		return
	}
	row, err := pages.GetSource(c.Request.Context(), id)
	handlePageResult(c, row, err)
}

func updatePageSource(c *gin.Context) {
	id, ok := pageID(c)
	if !ok || !ensurePage(c, id) {
		return
	}
	var input pages.SourceUpdateInput
	if !decodePageJSON(c, &input) {
		return
	}
	row, err := pages.UpdateSourceAs(c.Request.Context(), id, input, pageActor(c))
	handlePageResult(c, row, err)
}

func deletePageSource(c *gin.Context) {
	id, ok := pageID(c)
	if !ok || !ensurePage(c, id) {
		return
	}
	row, err := pages.DeleteSource(c.Request.Context(), id)
	handlePageResult(c, row, err)
}

func checkPageSource(c *gin.Context) {
	id, ok := pageID(c)
	if !ok || !ensurePage(c, id) {
		return
	}
	row, err := pages.DispatchSourceAction(c.Request.Context(), id, "check", pageActor(c), "")
	handlePageResult(c, row, err)
}

func syncPageSource(c *gin.Context) {
	id, ok := pageID(c)
	if !ok || !ensurePage(c, id) {
		return
	}
	var input pages.SourceSyncInput
	if !decodePageJSON(c, &input) {
		return
	}
	row, err := pages.DispatchSourceAction(c.Request.Context(), id, "sync", pageActor(c), input.ConfirmedRevision)
	handlePageResult(c, row, err)
}

func listPageDeployments(c *gin.Context) {
	id, ok := pageID(c)
	if !ok || !ensurePage(c, id) {
		return
	}
	row, err := pages.ListProjectDeployments(c.Request.Context(), id)
	handlePageResult(c, row, err)
}

func uploadPageDeployment(c *gin.Context) {
	id, ok := pageID(c)
	if !ok || !ensurePage(c, id) {
		return
	}
	file, err := c.FormFile("package")
	if err != nil {
		response.AbortBadRequest(c, "缺少部署包")
		return
	}
	row, err := pages.UploadDeployment(c.Request.Context(), id, file, pageActor(c))
	handlePageResult(c, row, err)
}

func uploadPageDeploymentFromURL(c *gin.Context) {
	id, ok := pageID(c)
	if !ok || !ensurePage(c, id) {
		return
	}
	var input pages.UploadFromURLInput
	if !bind(c, &input) {
		return
	}
	row, err := pages.UploadDeploymentFromURL(c.Request.Context(), id, input.URL, pageActor(c))
	handlePageResult(c, row, err)
}

func activatePageDeployment(c *gin.Context) {
	id, ok := pageID(c)
	if !ok || !ensurePage(c, id) {
		return
	}
	deployment, ok := deploymentID(c)
	if !ok {
		return
	}
	row, err := pages.ActivateDeploymentAs(c.Request.Context(), id, deployment, pageActor(c))
	handlePageResult(c, row, err)
}

func deletePageDeployment(c *gin.Context) {
	id, ok := pageID(c)
	if !ok || !ensurePage(c, id) {
		return
	}
	deployment, ok := deploymentID(c)
	if !ok {
		return
	}
	handlePageResult(c, nil, pages.DeleteDeployment(c.Request.Context(), id, deployment))
}

func listPageDeploymentFiles(c *gin.Context) {
	id, ok := deploymentID(c)
	if !ok {
		return
	}
	if !pageAdmin(c) {
		if _, err := repository.GetPagesProjectByDeploymentIDAndOwner(c.Request.Context(), id, userID(c)); err != nil {
			response.AbortNotFound(c, "部署不存在")
			return
		}
	}
	row, err := pages.ListDeploymentFiles(c.Request.Context(), id)
	handlePageResult(c, row, err)
}

func decodePageJSON(c *gin.Context, target any) bool {
	return bind(c, target)
}
