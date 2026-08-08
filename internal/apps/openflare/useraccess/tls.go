package useraccess

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Rain-kl/Wavelet/internal/apps/openflare/tls"
	"github.com/Rain-kl/Wavelet/internal/repository"
	"github.com/Rain-kl/Wavelet/internal/shared/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerTLSRoutes(group *gin.RouterGroup) {
	group.GET("/tls-certificates", listUserCertificates)
	group.GET("/tls-certificates/dns-accounts", listUserDNSAccounts)
	group.POST("/tls-certificates", createUserCertificate)
	group.POST("/tls-certificates/apply", applyUserCertificate)
	group.POST("/tls-certificates/import-file", importUserCertificateFile)
	group.GET("/tls-certificates/acme-account/default", getDefaultUserAcmeAccount)
	group.GET("/tls-certificates/:id", getUserCertificate)
	group.GET("/tls-certificates/:id/content", getUserCertificateContent)
	group.POST("/tls-certificates/:id/update", updateUserCertificate)
	group.POST("/tls-certificates/:id/update-acme", updateUserACMECertificate)
	group.POST("/tls-certificates/:id/convert-acme", convertUserCertificate)
	group.POST("/tls-certificates/:id/renew", renewUserCertificate)
	group.POST("/tls-certificates/:id/delete", deleteUserCertificate)
}

func parseCertificateID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.AbortBadRequest(c, "证书 ID 无效")
		return 0, false
	}
	return uint(id), true
}

func certificateResult(c *gin.Context, value any, err error) {
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.AbortNotFound(c, "证书不存在")
		} else {
			response.AbortBadRequest(c, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, response.OK(value))
}

// listUserCertificates lists certificates owned by the current user.
// @Summary 获取我的 TLS 证书
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Success 200 {object} response.Any{data=[]model.TLSCertificate}
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/resources/tls-certificates [get]
func listUserCertificates(c *gin.Context) {
	rows, err := tls.ListCertificatesOwned(c.Request.Context(), userID(c))
	certificateResult(c, rows, err)
}

// listUserDNSAccounts lists administrator-configured DNS accounts for ACME use.
// @Summary 获取可用 DNS 账号
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Success 200 {object} response.Any{data=[]model.DNSAccount}
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/resources/tls-certificates/dns-accounts [get]
func listUserDNSAccounts(c *gin.Context) {
	rows, err := tls.ListDNSAccounts(c.Request.Context())
	certificateResult(c, rows, err)
}

// createUserCertificate creates an uploaded certificate owned by the current user.
// @Summary 创建我的 TLS 证书
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param body body tls.CertificateInput true "证书参数"
// @Success 200 {object} response.Any{data=model.TLSCertificate}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/resources/tls-certificates [post]
func createUserCertificate(c *gin.Context) {
	var input tls.CertificateInput
	if !bind(c, &input) {
		return
	}
	row, err := tls.CreateCertificateOwned(c.Request.Context(), userID(c), input)
	certificateResult(c, row, err)
}

// applyUserCertificate starts an ACME certificate request for the current user.
// @Summary 申请我的 ACME 证书
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param body body tls.ApplyInput true "申请参数"
// @Success 200 {object} response.Any{data=model.TLSCertificate}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/resources/tls-certificates/apply [post]
func applyUserCertificate(c *gin.Context) {
	var input tls.ApplyInput
	if !bind(c, &input) {
		return
	}
	row, err := tls.ApplyCertificateOwned(c.Request.Context(), userID(c), input)
	certificateResult(c, row, err)
}

// importUserCertificateFile imports a certificate and key for the current user.
// @Summary 导入我的 TLS 证书
// @Tags custom-resources
// @Accept multipart/form-data
// @Produce json
// @Security SessionCookie
// @Param name formData string false "证书名称"
// @Param remark formData string false "备注"
// @Param cert_file formData file true "证书文件"
// @Param key_file formData file true "私钥文件"
// @Success 200 {object} response.Any{data=model.TLSCertificate}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/resources/tls-certificates/import-file [post]
func importUserCertificateFile(c *gin.Context) {
	certFile, err := c.FormFile("cert_file")
	if err != nil {
		response.AbortBadRequest(c, "缺少证书文件")
		return
	}
	keyFile, err := c.FormFile("key_file")
	if err != nil {
		response.AbortBadRequest(c, "缺少私钥文件")
		return
	}
	row, err := tls.CreateCertificateFromFilesOwned(c.Request.Context(), userID(c), c.PostForm("name"), certFile, keyFile, c.PostForm("remark"))
	certificateResult(c, row, err)
}

// getDefaultUserAcmeAccount returns the configured default ACME account.
// @Summary 获取默认 ACME 账号
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Success 200 {object} response.Any{data=model.AcmeAccount}
// @Failure 401 {object} response.Any
// @Router /api/v1/custom/resources/tls-certificates/acme-account/default [get]
func getDefaultUserAcmeAccount(c *gin.Context) {
	row, err := tls.GetDefaultAcmeAccount(c.Request.Context())
	certificateResult(c, row, err)
}

// getUserCertificate returns one certificate owned by the current user.
// @Summary 获取我的 TLS 证书详情
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Param id path int true "证书 ID"
// @Success 200 {object} response.Any{data=model.TLSCertificate}
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/tls-certificates/{id} [get]
func getUserCertificate(c *gin.Context) {
	id, ok := parseCertificateID(c)
	if !ok {
		return
	}
	row, err := repository.GetOwnedTLSCertificateByID(c.Request.Context(), id, userID(c))
	certificateResult(c, row, err)
}

// getUserCertificateContent returns PEM content for an owned certificate.
// @Summary 获取我的 TLS 证书内容
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Param id path int true "证书 ID"
// @Success 200 {object} response.Any{data=tls.CertificateContent}
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/tls-certificates/{id}/content [get]
func getUserCertificateContent(c *gin.Context) {
	id, ok := parseCertificateID(c)
	if !ok {
		return
	}
	row, err := tls.GetCertificateContentOwned(c.Request.Context(), id, userID(c))
	certificateResult(c, row, err)
}

// updateUserCertificate updates an uploaded certificate owned by the current user.
// @Summary 更新我的 TLS 证书
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param id path int true "证书 ID"
// @Param body body tls.CertificateInput true "证书参数"
// @Success 200 {object} response.Any{data=model.TLSCertificate}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/tls-certificates/{id}/update [post]
func updateUserCertificate(c *gin.Context) {
	id, ok := parseCertificateID(c)
	if !ok {
		return
	}
	var input tls.CertificateInput
	if !bind(c, &input) {
		return
	}
	row, err := tls.UpdateCertificateOwned(c.Request.Context(), id, userID(c), input)
	certificateResult(c, row, err)
}

// updateUserACMECertificate updates an owned ACME certificate request.
// @Summary 更新我的 ACME 证书配置
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param id path int true "证书 ID"
// @Param body body tls.ApplyInput true "申请参数"
// @Success 200 {object} response.Any{data=model.TLSCertificate}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/tls-certificates/{id}/update-acme [post]
func updateUserACMECertificate(c *gin.Context) {
	id, ok := parseCertificateID(c)
	if !ok {
		return
	}
	certificate, err := repository.GetOwnedTLSCertificateByID(c.Request.Context(), id, userID(c))
	if err != nil {
		certificateResult(c, nil, err)
		return
	}
	if certificate.Provider != "acme" {
		certificateResult(c, nil, errors.New("仅 ACME 证书支持更新申请配置"))
		return
	}
	var input tls.ApplyInput
	if !bind(c, &input) {
		return
	}
	row, err := tls.UpdateACMECertificateOwned(c.Request.Context(), id, userID(c), input)
	certificateResult(c, row, err)
}

// convertUserCertificate converts an owned uploaded certificate to ACME management.
// @Summary 转换我的 TLS 证书为 ACME
// @Tags custom-resources
// @Accept json
// @Produce json
// @Security SessionCookie
// @Param id path int true "证书 ID"
// @Param body body tls.ApplyInput true "申请参数"
// @Success 200 {object} response.Any{data=model.TLSCertificate}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/tls-certificates/{id}/convert-acme [post]
func convertUserCertificate(c *gin.Context) {
	id, ok := parseCertificateID(c)
	if !ok {
		return
	}
	certificate, err := repository.GetOwnedTLSCertificateByID(c.Request.Context(), id, userID(c))
	if err != nil {
		certificateResult(c, nil, err)
		return
	}
	if certificate.Provider != "upload" {
		certificateResult(c, nil, errors.New("仅手动证书支持转换为 ACME"))
		return
	}
	var input tls.ApplyInput
	if !bind(c, &input) {
		return
	}
	row, err := tls.ConvertCertificateToACMEOwned(c.Request.Context(), id, userID(c), input)
	certificateResult(c, row, err)
}

// renewUserCertificate queues renewal for an owned ACME certificate.
// @Summary 续期我的 ACME 证书
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Param id path int true "证书 ID"
// @Success 200 {object} response.Any{data=model.TLSCertificate}
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/tls-certificates/{id}/renew [post]
func renewUserCertificate(c *gin.Context) {
	id, ok := parseCertificateID(c)
	if !ok {
		return
	}
	row, err := tls.RenewCertificateOwned(c.Request.Context(), id, userID(c))
	certificateResult(c, row, err)
}

// deleteUserCertificate deletes an owned certificate when it is unreferenced.
// @Summary 删除我的 TLS 证书
// @Tags custom-resources
// @Produce json
// @Security SessionCookie
// @Param id path int true "证书 ID"
// @Success 200 {object} response.Any
// @Failure 400 {object} response.Any
// @Failure 401 {object} response.Any
// @Failure 404 {object} response.Any
// @Router /api/v1/custom/resources/tls-certificates/{id}/delete [post]
func deleteUserCertificate(c *gin.Context) {
	id, ok := parseCertificateID(c)
	if !ok {
		return
	}
	certificateResult(c, nil, tls.DeleteCertificateOwned(c.Request.Context(), id, userID(c)))
}
