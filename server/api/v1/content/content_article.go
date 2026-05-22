package content

import (
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	contentModel "github.com/flipped-aurora/gin-vue-admin/server/model/content"
	svccontent "github.com/flipped-aurora/gin-vue-admin/server/service/content"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"sync"
)

type ArticleApi struct{}

 
var countMutex sync.Mutex
 

func (a *ArticleApi) CountWebView(c *gin.Context) {

 
	countMutex.Lock()
	defer countMutex.Unlock()

	var stats contentModel.SiteStats
	// 固定查第一条数据
	err := global.GVA_DB.First(&stats).Error
	if err != nil {
		// 没有就创建
		stats = contentModel.SiteStats{
			Total: 0,
			Today: 0,
		}
		global.GVA_DB.Create(&stats)
	}

	stats.Total++
	stats.Today++
	global.GVA_DB.Save(&stats)

	if err := portalVisitorService.RecordVisit(utils.ClientIP(c)); err != nil {
		global.GVA_LOG.Warn("记录门户访客失败", zap.Error(err))
	}

	response.OkWithDetailed(gin.H{
		"total": stats.Total,
		"today": stats.Today,
	}, "统计成功", c)
}

func (a *ArticleApi) GetWebViewCount(c *gin.Context) {
	countMutex.Lock()
	defer countMutex.Unlock()

	var stats contentModel.SiteStats
	global.GVA_DB.FirstOrCreate(&stats, contentModel.SiteStats{})

	response.OkWithDetailed(gin.H{
		"total": stats.Total,
		"today": stats.Today,
	}, "获取成功", c)
}

// CreateArticle
// @Tags      ContentArticle
// @Summary   创建文章
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      contentModel.ContentArticle      true  "文章信息"
// @Success   200   {object}  response.Response{msg=string}    "创建文章"
// @Router    /contentArticle/createArticle [post]
func (a *ArticleApi) CreateArticle(c *gin.Context) {
	var article contentModel.ContentArticle
	if err := c.ShouldBindJSON(&article); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := utils.Verify(article, utils.ContentArticleVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	article.AuthorID = utils.GetUserID(c)
	if err := articleService.CreateArticle(&article); err != nil {
		global.GVA_LOG.Error("创建文章失败!", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithDetailed(article, "创建成功", c)
}

// UpdateArticle
// @Tags      ContentArticle
// @Summary   更新文章
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      contentModel.ContentArticle      true  "文章信息"
// @Success   200   {object}  response.Response{msg=string}    "更新文章"
// @Router    /contentArticle/updateArticle [put]
func (a *ArticleApi) UpdateArticle(c *gin.Context) {
	var article contentModel.ContentArticle
	if err := c.ShouldBindJSON(&article); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := utils.Verify(article.GVA_MODEL, utils.IdVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := utils.Verify(article, utils.ContentArticleVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	article.AuthorID = utils.GetUserID(c)
	if err := articleService.UpdateArticle(&article); err != nil {
		global.GVA_LOG.Error("更新文章失败!", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteArticle
// @Tags      ContentArticle
// @Summary   删除文章
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                true  "文章ID"
// @Success   200   {object}  response.Response{msg=string}  "删除文章"
// @Router    /contentArticle/deleteArticle [delete]
func (a *ArticleApi) DeleteArticle(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.ID <= 0 {
		response.FailWithMessage("ID 不能为空", c)
		return
	}
	if err := articleService.DeleteArticle(req.Uint(), utils.GetUserID(c)); err != nil {
		global.GVA_LOG.Error("删除文章失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// FindArticle
// @Tags      ContentArticle
// @Summary   获取单篇文章(作者自己的)
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id   query     int                                                true  "文章ID"
// @Success   200  {object}  response.Response{data=contentModel.ContentArticle,msg=string}  "获取单篇文章"
// @Router    /contentArticle/findArticle [get]
func (a *ArticleApi) FindArticle(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.ID <= 0 {
		response.FailWithMessage("ID 不能为空", c)
		return
	}
	data, err := articleService.FindArticle(req.Uint(), utils.GetUserID(c))
	if err != nil {
		global.GVA_LOG.Error("获取文章失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(data, "获取成功", c)
}

// GetArticleList
// @Tags      ContentArticle
// @Summary   分页获取文章列表(作者自己的)
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     page     query     int     true  "页码"
// @Param     pageSize query     int     true  "每页大小"
// @Param     keyword  query     string  false "关键字"
// @Param     status   query     string  false "状态 draft/published/archived"
// @Success   200      {object}  response.Response{data=response.PageResult,msg=string}  "分页获取文章列表"
// @Router    /contentArticle/getArticleList [get]
func (a *ArticleApi) GetArticleList(c *gin.Context) {
	var pageInfo request.PageInfo
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := utils.Verify(pageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	status := c.Query("status")
	list, total, err := articleService.GetArticleList(utils.GetUserID(c), pageInfo, status)
	if err != nil {
		global.GVA_LOG.Error("获取文章列表失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

// PublishArticle
// @Tags      ContentArticle
// @Summary   发布文章
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                true  "文章ID"
// @Success   200   {object}  response.Response{msg=string}  "发布文章"
// @Router    /contentArticle/publishArticle [post]
func (a *ArticleApi) PublishArticle(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.ID <= 0 {
		response.FailWithMessage("ID 不能为空", c)
		return
	}
	if err := articleService.PublishArticle(req.Uint(), utils.GetUserID(c)); err != nil {
		global.GVA_LOG.Error("发布文章失败!", zap.Error(err))
		response.FailWithMessage("发布失败", c)
		return
	}
	response.OkWithMessage("发布成功", c)
}

// GetPublishedList
// @Tags      ContentArticle
// @Summary   公开访问：分页获取已发布文章（不含正文）
// @accept    application/json
// @Produce   application/json
// @Param     page     query     int     true  "页码"
// @Param     pageSize query     int     true  "每页大小"
// @Param     keyword  query     string  false "关键字"
// @Success   200      {object}  response.Response{data=response.PageResult,msg=string}  "列表"
// @Router    /public/contentArticles [get]
func (a *ArticleApi) GetPublishedList(c *gin.Context) {
	var pageInfo request.PageInfo
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := utils.Verify(pageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := articleService.GetPublishedList(pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取公开文章列表失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

// GetPublishedBySlug
// @Tags      ContentArticle
// @Summary   公开访问：通过 slug 获取已发布文章
// @accept    application/json
// @Produce   application/json
// @Param     slug  path      string                                              true  "文章slug"
// @Success   200   {object}  response.Response{data=contentModel.ContentArticle,msg=string}  "获取已发布文章"
// @Router    /public/article/{slug} [get]
func (a *ArticleApi) GetPublishedBySlug(c *gin.Context) {
	slug := c.Param("slug")
	data, err := articleService.GetPublishedBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailWithMessage("文章不存在或未发布", c)
			return
		}
		global.GVA_LOG.Error("获取公开文章失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	_ = articleService.IncViewBySlug(slug)
	response.OkWithDetailed(data, "获取成功", c)
}

type generateArticleByBaiduBody struct {
	Title            string  `json:"title"`
	Keywords         string  `json:"keywords"`
	Rules            string  `json:"rules"`
	RulesFileContent string  `json:"rulesFileContent"`
	WordCount        int     `json:"wordCount"`
	Temperature      float64 `json:"temperature"`
}

// GenerateArticleByBaidu 生成 Markdown 正文：若配置 volc-ark.api-key 则走火山豆包；否则走百度千帆（V2 或 OAuth）
func (a *ArticleApi) GenerateArticleByBaidu(c *gin.Context) {
	var body generateArticleByBaiduBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	req := svccontent.BaiduGenerateArticleReq{
		Title:            body.Title,
		Keywords:         body.Keywords,
		Rules:            body.Rules,
		RulesFileContent: body.RulesFileContent,
		WordCount:        body.WordCount,
		Temperature:      body.Temperature,
	}
	var text string
	var err error
	if volcArkArticleService.Enabled() {
		text, err = volcArkArticleService.GenerateArticle(req)
	} else {
		text, err = baiduWenxinArticleService.GenerateArticle(req)
	}
	if err != nil {
		global.GVA_LOG.Warn("AI 生成文章失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"content": text}, "生成成功", c)
}

// DiagnoseBaiduWenxin 诊断 AI 写文章链路：优先火山方舟；否则百度（V2 / OAuth）
func (a *ArticleApi) DiagnoseBaiduWenxin(c *gin.Context) {
	if volcArkArticleService.Enabled() {
		response.OkWithDetailed(volcArkArticleService.Diagnose(), "诊断完成", c)
		return
	}
	response.OkWithDetailed(baiduWenxinArticleService.Diagnose(), "诊断完成", c)
}

