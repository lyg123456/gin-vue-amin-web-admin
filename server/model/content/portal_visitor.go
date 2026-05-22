package content

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// PortalVisitorDaily 门户访客按日统计（IP + 当天访问次数）
type PortalVisitorDaily struct {
	global.GVA_MODEL

	ClientIP   string `json:"clientIp" gorm:"type:varchar(64);uniqueIndex:idx_portal_visitor_ip_date;comment:访客IP"`
	VisitDate  string `json:"visitDate" gorm:"type:char(10);uniqueIndex:idx_portal_visitor_ip_date;comment:访问日期YYYY-MM-DD"`
	VisitCount uint   `json:"visitCount" gorm:"default:1;comment:当日访问次数(PV)"`
}

func (PortalVisitorDaily) TableName() string {
	return "portal_visitor_daily"
}
