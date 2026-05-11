package content

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	contentModel "github.com/flipped-aurora/gin-vue-admin/server/model/content"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ArticleCategoryApi struct{}

func (a *ArticleCategoryApi) GetCategoryTree(c *gin.Context) {
	tree, err := articleCategoryService.GetCategoryTree()
	if err != nil {
		global.GVA_LOG.Error("获取分类树失败", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"list": tree}, "获取成功", c)
}

func (a *ArticleCategoryApi) GetCategoryList(c *gin.Context) {
	list, err := articleCategoryService.GetCategoryListFlat()
	if err != nil {
		global.GVA_LOG.Error("获取分类列表失败", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list}, "获取成功", c)
}

func (a *ArticleCategoryApi) CreateCategory(c *gin.Context) {
	var in contentModel.ContentArticleCategory
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := articleCategoryService.CreateCategory(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(in, "创建成功", c)
}

func (a *ArticleCategoryApi) UpdateCategory(c *gin.Context) {
	var in contentModel.ContentArticleCategory
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := articleCategoryService.UpdateCategory(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *ArticleCategoryApi) DeleteCategory(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.ID <= 0 {
		response.FailWithMessage("ID 不能为空", c)
		return
	}
	if err := articleCategoryService.DeleteCategory(req.Uint()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
