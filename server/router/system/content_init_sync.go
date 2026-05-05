package system

import "github.com/gin-gonic/gin"

type ContentInitRouter struct{}

func (r *ContentInitRouter) InitContentInitRouter(private *gin.RouterGroup) {
	contentInit := private.Group("contentInit")
	{
		contentInit.POST("sync", contentInitApi.SyncContentInit)
	}
}

