package content

import (
	"fmt"
	"net/http"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OfficeToolsApi struct{}

// CreateTempMailbox 申请临时邮箱
func (a *OfficeToolsApi) CreateTempMailbox(c *gin.Context) {
	result, err := officeTempEmailService.CreateMailbox()
	if err != nil {
		global.GVA_LOG.Warn("申请临时邮箱失败", zap.Error(err))
		response.FailWithMessage("申请失败: "+err.Error(), c)
		return
	}
	if result == nil || result.Mailbox == "" {
		response.FailWithMessage("邮箱地址格式异常", c)
		return
	}
	response.OkWithDetailed(gin.H{
		"mailbox":  result.Mailbox,
		"login":    result.Login,
		"domain":   result.Domain,
		"token":    result.Token,
		"provider": result.Provider,
	}, "ok", c)
}

// GetTempEmailMessages 收件箱
func (a *OfficeToolsApi) GetTempEmailMessages(c *gin.Context) {
	token := c.Query("token")
	login := c.Query("login")
	domain := c.Query("domain")
	list, err := officeTempEmailService.ListMessages(token, login, domain)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "ok", c)
}

// ReadTempEmailMessage 读信
func (a *OfficeToolsApi) ReadTempEmailMessage(c *gin.Context) {
	token := c.Query("token")
	login := c.Query("login")
	domain := c.Query("domain")
	id := c.Query("id")
	msg, err := officeTempEmailService.ReadMessage(token, login, domain, id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(msg, "ok", c)
}

// GetConvertCapabilities 文件转换能力
func (a *OfficeToolsApi) GetConvertCapabilities(c *gin.Context) {
	response.OkWithDetailed(officeFileConvertService.Capabilities(), "ok", c)
}

// ConvertOfficeFile 上传并转换文件
func (a *OfficeToolsApi) ConvertOfficeFile(c *gin.Context) {
	target := c.PostForm("target")
	text := c.PostForm("text")
	fh, _ := c.FormFile("file")
	data, filename, mime, err := officeFileConvertService.Convert(fh, target, text)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, mime, data)
}
