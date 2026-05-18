package content

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// PortalContactLead 门户访客留资（电话 + 备注），公开提交、后台查看
type PortalContactLead struct {
	global.GVA_MODEL

	Phone    string `json:"phone" form:"phone" gorm:"type:varchar(32);index;comment:联系电话"`
	Remark   string `json:"remark" form:"remark" gorm:"type:varchar(2000);comment:留言备注"`
	ClientIP string `json:"clientIp" form:"clientIp" gorm:"type:varchar(64);comment:提交IP"`
}

func (PortalContactLead) TableName() string {
	return "portal_contact_leads"
}
