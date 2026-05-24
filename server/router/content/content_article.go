package content

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

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

		office := articlePublic.Group("office")
		office.Use(middleware.OfficeToolsDataPolicy())
		{
			office.GET("tempEmail/create", officeToolsApi.CreateTempMailbox)
			office.GET("tempEmail/messages", officeToolsApi.GetTempEmailMessages)
			office.GET("tempEmail/message", officeToolsApi.ReadTempEmailMessage)
			office.GET("convert/capabilities", officeToolsApi.GetConvertCapabilities)
			office.GET("convert/libreofficeStatus", officeToolsApi.GetLibreOfficeStatus)
			office.POST("convert/file", officeToolsApi.ConvertOfficeFile)
			office.GET("media/capabilities", officeToolsApi.GetMediaCapabilities)
			office.POST("media/video", officeToolsApi.ProcessMedia)
			office.POST("media/composite", officeToolsApi.CompositeImages)
			office.POST("media/extractBackground", officeToolsApi.ExtractImageBackground)
			office.POST("compress/image", officeToolsApi.CompressImage)
			office.POST("compress/excel", officeToolsApi.CompressExcel)
			office.GET("tweet/styles", officeToolsApi.GetTweetStyles)
			office.POST("tweet/rewrite", officeToolsApi.RewriteTweet)
			office.POST("web/downloadPages", officeToolsApi.GenerateWebStyleZip)
			office.POST("web/styleTemplates", officeToolsApi.GenerateWebStyleZip) // 兼容旧路径
			office.POST("web/crawlProducts", officeToolsApi.CrawlWebProductsExcel)
			office.GET("speed/ping", officeToolsApi.SpeedPing)
			office.GET("speed/info", officeToolsApi.SpeedInfo)
			office.GET("speed/download", officeToolsApi.SpeedDownload)
			office.POST("speed/upload", officeToolsApi.SpeedUpload)
			office.GET("watermark/capabilities", officeToolsApi.GetWatermarkCapabilities)
			office.POST("watermark/remove", officeToolsApi.RemoveWatermark)
			office.GET("douyin/categories", officeToolsApi.GetDouyinOfficialCategories)
			office.POST("douyin/verifyCookie", officeToolsApi.VerifyDouyinCookie)
			office.POST("douyin/crawl", officeToolsApi.CrawlDouyinIndustryVideos)
			office.GET("wechat/categories", officeToolsApi.GetWechatOfficialCategories)
			office.POST("wechat/verifyCookie", officeToolsApi.VerifyWechatCookie)
			office.POST("wechat/crawl", officeToolsApi.CrawlWechatIndustryVideos)
			office.GET("xhs/categories", officeToolsApi.GetXhsOfficialCategories)
			office.POST("xhs/verifyCookie", officeToolsApi.VerifyXhsCookie)
			office.POST("xhs/crawl", officeToolsApi.CrawlXhsIndustryVideos)
			office.POST("media/download", officeToolsApi.ProxyMediaDownload)
		}
	}

	leadPrivate := private.Group("contentPortalContactLead")
	{
		leadPrivate.GET("getPortalContactLeadList", portalContactLeadApi.GetList)
	}
}

