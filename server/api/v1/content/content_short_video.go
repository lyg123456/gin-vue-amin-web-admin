package content

import (
	"errors"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	contentModel "github.com/flipped-aurora/gin-vue-admin/server/model/content"
	svccontent "github.com/flipped-aurora/gin-vue-admin/server/service/content"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ShortVideoApi struct{}

type generateShortVideoScriptBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PromptText  string `json:"promptText"`
	DurationSec uint   `json:"durationSec"`
}

type createShortVideoWithAIBody struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	PromptText   string `json:"promptText"`
	DurationSec  uint   `json:"durationSec"`
	Script       string `json:"script"`
	CoverImage    string `json:"coverImage"`
	FirstFrameUrl string `json:"firstFrameUrl"`
	LastFrameUrl  string `json:"lastFrameUrl"`
	SourceImages  string `json:"sourceImages"` // 兼容：逗号分隔，取前两位作为首尾帧
	VideoURL      string `json:"videoUrl"`
	AutoGenerate  bool   `json:"autoGenerate"`
}

type regenerateShortVideoBody struct {
	ID           uint   `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	PromptText   string `json:"promptText"`
	Script       string `json:"script"`
	DurationSec  uint   `json:"durationSec"`
	CoverImage    string `json:"coverImage"`
	FirstFrameUrl string `json:"firstFrameUrl"`
	LastFrameUrl  string `json:"lastFrameUrl"`
	SourceImages  string `json:"sourceImages"`
	RegenScript   bool   `json:"regenScript"`
	RegenVideo    bool   `json:"regenVideo"`
}

type generateShortVideoBody struct {
	ID            uint   `json:"id"`
	FirstFrameUrl string `json:"firstFrameUrl"`
	LastFrameUrl  string `json:"lastFrameUrl"`
}

func (a *ShortVideoApi) CreateShortVideo(c *gin.Context) {
	var v contentModel.ContentShortVideo
	if err := c.ShouldBindJSON(&v); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	v.AuthorID = utils.GetUserID(c)
	if err := shortVideoService.Create(&v); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(v, "创建成功", c)
}

func (a *ShortVideoApi) UpdateShortVideo(c *gin.Context) {
	var v contentModel.ContentShortVideo
	if err := c.ShouldBindJSON(&v); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := shortVideoService.Update(&v); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *ShortVideoApi) DeleteShortVideo(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := shortVideoService.Delete(uint(req.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *ShortVideoApi) FindShortVideo(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	v, err := shortVideoService.Find(uint(req.ID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailWithMessage("记录不存在", c)
			return
		}
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(v, "获取成功", c)
}

func (a *ShortVideoApi) GetShortVideoList(c *gin.Context) {
	var page request.PageInfo
	if err := c.ShouldBindQuery(&page); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	status := c.Query("status")
	list, total, err := shortVideoService.GetList(page, status)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: page.Page, PageSize: page.PageSize}, "获取成功", c)
}

func (a *ShortVideoApi) PublishShortVideo(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := shortVideoService.Publish(uint(req.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("发布成功", c)
}

func (a *ShortVideoApi) GenerateShortVideoScript(c *gin.Context) {
	var body generateShortVideoScriptBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	script, err := volcArkShortVideoService.GenerateScript(svccontent.ShortVideoScriptReq{
		Title:       body.Title,
		Description: body.Description,
		PromptText:  body.PromptText,
		DurationSec: body.DurationSec,
	})
	if err != nil {
		global.GVA_LOG.Warn("生成短视频脚本失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"script": script}, "生成成功", c)
}

func (a *ShortVideoApi) CreateShortVideoWithAI(c *gin.Context) {
	var body createShortVideoWithAIBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	script := body.Script
	if script == "" {
		var err error
		script, err = volcArkShortVideoService.GenerateScript(svccontent.ShortVideoScriptReq{
			Title:       body.Title,
			Description: body.Description,
			PromptText:  body.PromptText,
			DurationSec: body.DurationSec,
		})
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
	}
	v := contentModel.ContentShortVideo{
		AuthorID:     utils.GetUserID(c),
		Title:        body.Title,
		Description:  body.Description,
		PromptText:   body.PromptText,
		Script:       script,
		DurationSec:  body.DurationSec,
		CoverImage:    body.CoverImage,
		FirstFrameURL: body.FirstFrameUrl,
		LastFrameURL:  body.LastFrameUrl,
		SourceImages:  body.SourceImages,
		VideoURL:      strings.TrimSpace(body.VideoURL),
		Status:       "draft",
		AiProvider:   "volc",
	}
	if v.VideoURL != "" {
		v.Status = "ready"
	}
	svccontent.SyncShortVideoFrames(&v)
	if err := shortVideoService.Create(&v); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if body.AutoGenerate {
		if err := shortVideoService.RunVideoGeneration(v.ID); err != nil {
			v.GenerationError = err.Error()
			v.Status = "failed"
			_ = shortVideoService.Update(&v)
		} else {
			v, _ = shortVideoService.Find(v.ID)
		}
	}
	response.OkWithDetailed(v, "已入库", c)
}

func (a *ShortVideoApi) GenerateShortVideo(c *gin.Context) {
	var body generateShortVideoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if body.ID == 0 {
		response.FailWithMessage("ID 不能为空", c)
		return
	}
	if body.FirstFrameUrl != "" || body.LastFrameUrl != "" {
		v, err := shortVideoService.Find(body.ID)
		if err != nil {
			response.FailWithMessage("记录不存在", c)
			return
		}
		if body.FirstFrameUrl != "" {
			v.FirstFrameURL = body.FirstFrameUrl
		}
		if body.LastFrameUrl != "" {
			v.LastFrameURL = body.LastFrameUrl
		}
		svccontent.SyncShortVideoFrames(&v)
		if err := shortVideoService.Update(&v); err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
	}
	if err := shortVideoService.RunVideoGeneration(body.ID); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	v, _ := shortVideoService.Find(body.ID)
	response.OkWithDetailed(v, "已提交生成", c)
}

func (a *ShortVideoApi) RegenerateShortVideo(c *gin.Context) {
	var body regenerateShortVideoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if body.ID == 0 {
		response.FailWithMessage("ID 不能为空", c)
		return
	}
	v, err := shortVideoService.Find(body.ID)
	if err != nil {
		response.FailWithMessage("记录不存在", c)
		return
	}
	if body.Title != "" {
		v.Title = body.Title
	}
	if body.Description != "" {
		v.Description = body.Description
	}
	if body.PromptText != "" {
		v.PromptText = body.PromptText
	}
	if body.DurationSec > 0 {
		v.DurationSec = body.DurationSec
	}
	if body.CoverImage != "" {
		v.CoverImage = body.CoverImage
	}
	if body.FirstFrameUrl != "" {
		v.FirstFrameURL = body.FirstFrameUrl
	}
	if body.LastFrameUrl != "" {
		v.LastFrameURL = body.LastFrameUrl
	}
	if body.SourceImages != "" {
		v.SourceImages = body.SourceImages
	}
	svccontent.SyncShortVideoFrames(&v)
	if body.RegenScript {
		script, err := volcArkShortVideoService.GenerateScript(svccontent.ShortVideoScriptReq{
			Title:       v.Title,
			Description: v.Description,
			PromptText:  v.PromptText,
			DurationSec: v.DurationSec,
		})
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
		v.Script = script
	} else if body.Script != "" {
		v.Script = body.Script
	}
	if err := shortVideoService.Update(&v); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if body.RegenVideo {
		if err := shortVideoService.RunVideoGeneration(v.ID); err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
		v, _ = shortVideoService.Find(v.ID)
	}
	response.OkWithDetailed(v, "操作成功", c)
}

func (a *ShortVideoApi) GetPublishedShortVideoList(c *gin.Context) {
	var page request.PageInfo
	if err := c.ShouldBindQuery(&page); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := shortVideoService.GetPublishedList(page)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: page.Page, PageSize: page.PageSize}, "获取成功", c)
}

func (a *ShortVideoApi) GetPublishedShortVideoBySlug(c *gin.Context) {
	slug := c.Param("slug")
	v, err := shortVideoService.GetPublishedBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailWithMessage("短视频不存在或未发布", c)
			return
		}
		response.FailWithMessage(err.Error(), c)
		return
	}
	_ = shortVideoService.IncViewBySlug(slug)
	response.OkWithDetailed(v, "获取成功", c)
}
