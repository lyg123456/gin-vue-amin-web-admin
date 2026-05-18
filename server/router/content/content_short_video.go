package content

import "github.com/gin-gonic/gin"

type ShortVideoRouter struct{}

func (r *ShortVideoRouter) InitContentShortVideoRouter(private *gin.RouterGroup, public *gin.RouterGroup) {
	svPrivate := private.Group("contentShortVideo")
	{
		svPrivate.POST("createShortVideo", shortVideoApi.CreateShortVideo)
		svPrivate.PUT("updateShortVideo", shortVideoApi.UpdateShortVideo)
		svPrivate.DELETE("deleteShortVideo", shortVideoApi.DeleteShortVideo)
		svPrivate.GET("findShortVideo", shortVideoApi.FindShortVideo)
		svPrivate.GET("getShortVideoList", shortVideoApi.GetShortVideoList)
		svPrivate.POST("publishShortVideo", shortVideoApi.PublishShortVideo)
		svPrivate.POST("generateShortVideoScript", shortVideoApi.GenerateShortVideoScript)
		svPrivate.POST("createShortVideoWithAI", shortVideoApi.CreateShortVideoWithAI)
		svPrivate.POST("generateShortVideo", shortVideoApi.GenerateShortVideo)
		svPrivate.POST("regenerateShortVideo", shortVideoApi.RegenerateShortVideo)
	}

	svPublic := public.Group("public")
	{
		svPublic.GET("shortVideos", shortVideoApi.GetPublishedShortVideoList)
		svPublic.GET("shortVideo/:slug", shortVideoApi.GetPublishedShortVideoBySlug)
	}
}
