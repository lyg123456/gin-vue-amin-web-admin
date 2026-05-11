package content

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"

type RouterGroup struct {
	ArticleRouter
}

var (
	articleApi         = api.ApiGroupApp.ContentApiGroup.ArticleApi
	articleCategoryApi = api.ApiGroupApp.ContentApiGroup.ArticleCategoryApi
)

