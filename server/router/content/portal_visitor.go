package content

import "github.com/gin-gonic/gin"

type PortalVisitorRouter struct{}

func (r *PortalVisitorRouter) InitPortalVisitorRouter(private *gin.RouterGroup) {
	g := private.Group("contentPortalVisitor")
	{
		g.GET("getPortalVisitorList", portalVisitorApi.GetPortalVisitorList)
		g.GET("getPortalVisitorSummary", portalVisitorApi.GetPortalVisitorSummary)
	}
}
