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
	}

	// public：用于 SEO 收录的公开访问入口
	articlePublic := public.Group("public")
	{
		// 列表路径不可使用 article/xxx 形式，否则会被 article/:slug 抢占（如 xxx=list）
		articlePublic.GET("contentArticles", articleApi.GetPublishedList)
		articlePublic.GET("article/:slug", articleApi.GetPublishedBySlug)
	}
}

