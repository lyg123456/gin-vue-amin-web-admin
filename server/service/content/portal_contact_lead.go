package content

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	contentModel "github.com/flipped-aurora/gin-vue-admin/server/model/content"
)

type PortalContactLeadService struct{}

func (s *PortalContactLeadService) Submit(lead *contentModel.PortalContactLead) error {
	lead.Phone = strings.TrimSpace(lead.Phone)
	lead.Remark = strings.TrimSpace(lead.Remark)
	if lead.Phone == "" {
		return errors.New("请填写电话号码")
	}
	if len(lead.Phone) < 6 || len(lead.Phone) > 32 {
		return errors.New("电话号码长度不合法")
	}
	if utf8.RuneCountInString(lead.Remark) > 2000 {
		return errors.New("备注过长")
	}
	return global.GVA_DB.Create(lead).Error
}

func (s *PortalContactLeadService) GetList(info request.PageInfo) (list []contentModel.PortalContactLead, total int64, err error) {
	db := global.GVA_DB.Model(&contentModel.PortalContactLead{})
	if info.Keyword != "" {
		kw := "%" + strings.TrimSpace(info.Keyword) + "%"
		db = db.Where("phone LIKE ? OR remark LIKE ?", kw, kw)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Scopes(info.Paginate()).Order("id DESC").Find(&list).Error
	return
}
