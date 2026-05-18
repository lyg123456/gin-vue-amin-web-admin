package content

import "github.com/gin-gonic/gin"

type ArticleRouter struct{}

func (r *ArticleRouter) InitContentArticleRouter(private *gin.RouterGroup, public *gin.RouterGroup) {
	articlePrivate := private.Group("contentArticle")
	{
		articlePrivate.POST("createArticle", articleApi.CreateArticle)
		articlePrivate.PUT("updateArticle", articleApi.UpdateArticle)
		articlePrivate.DELETE("deleteArticle", articleApi.DeleteArticle)
		articlePrivate.GET("findArticle", articleApi.FindArticle)
		articlePrivate.GET("getArticleList", articleApi.GetArticleList)
		articlePrivate.POST("publishArticle", articleApi.PublishArticle)
		articlePrivate.POST("generateArticleByBaidu", articleApi.GenerateArticleByBaidu)
		articlePrivate.GET("diagnoseBaiduWenxin", articleApi.DiagnoseBaiduWenxin)
	}

	catPrivate := private.Group("contentArticleCategory")
	{
		catPrivate.GET("getCategoryTree", articleCategoryApi.GetCategoryTree)
		catPrivate.GET("getCategoryList", articleCategoryApi.GetCategoryList)
		catPrivate.POST("createCategory", articleCategoryApi.CreateCategory)
		catPrivate.PUT("updateCategory", articleCategoryApi.UpdateCategory)
		catPrivate.DELETE("deleteCategory", articleCategoryApi.DeleteCategory)
	}

	// public：用于 SEO 收录的公开访问入口
	articlePublic := public.Group("public")
	{
		// 列表路径不可使用 article/xxx 形式，否则会被 article/:slug 抢占（如 xxx=list）
		articlePublic.GET("contentArticles", articleApi.GetPublishedList)
		articlePublic.GET("article/:slug", articleApi.GetPublishedBySlug)

		articlePublic.GET("web/stats", articleApi.CountWebView)          // 统计+返回
		articlePublic.GET("web/stats/info", articleApi.GetWebViewCount) // 仅查询
		articlePublic.GET("homeCarousel", articleApi.GetHomeCarousel)   // 门户首页轮播
		articlePublic.POST("portalContactLead", portalContactLeadApi.Submit) // 门户留资提交
	}

	leadPrivate := private.Group("contentPortalContactLead")
	{
		leadPrivate.GET("getPortalContactLeadList", portalContactLeadApi.GetList)
	}
}

