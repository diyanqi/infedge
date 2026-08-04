package useraccess

import (
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
	group.POST("/tls-certificates", createUserCertificate)
	group.GET("/tls-certificates/:id", getUserCertificate)
	group.POST("/tls-certificates/:id/update", updateUserCertificate)
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
		if err == gorm.ErrRecordNotFound {
			response.AbortNotFound(c, "证书不存在")
		} else {
			response.AbortBadRequest(c, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, response.OK(value))
}

func listUserCertificates(c *gin.Context) {
	rows, err := tls.ListCertificatesOwned(c.Request.Context(), userID(c))
	certificateResult(c, rows, err)
}

func createUserCertificate(c *gin.Context) {
	var input tls.CertificateInput
	if !bind(c, &input) {
		return
	}
	row, err := tls.CreateCertificateOwned(c.Request.Context(), userID(c), input)
	certificateResult(c, row, err)
}

func getUserCertificate(c *gin.Context) {
	id, ok := parseCertificateID(c)
	if !ok {
		return
	}
	row, err := repository.GetOwnedTLSCertificateByID(c.Request.Context(), id, userID(c))
	certificateResult(c, row, err)
}

func updateUserCertificate(c *gin.Context) {
	id, ok := parseCertificateID(c)
	if !ok {
		return
	}
	if _, err := repository.GetOwnedTLSCertificateByID(c.Request.Context(), id, userID(c)); err != nil {
		certificateResult(c, nil, err)
		return
	}
	var input tls.CertificateInput
	if !bind(c, &input) {
		return
	}
	row, err := tls.UpdateCertificate(c.Request.Context(), id, input)
	certificateResult(c, row, err)
}

func deleteUserCertificate(c *gin.Context) {
	id, ok := parseCertificateID(c)
	if !ok {
		return
	}
	if _, err := repository.GetOwnedTLSCertificateByID(c.Request.Context(), id, userID(c)); err != nil {
		certificateResult(c, nil, err)
		return
	}
	certificateResult(c, nil, tls.DeleteCertificate(c.Request.Context(), id))
}
