package content

import (
	"errors"
	"sort"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	contentModel "github.com/flipped-aurora/gin-vue-admin/server/model/content"
	"gorm.io/gorm"
)

type ArticleCategoryService struct{}

// CascaderNode 供前端 el-cascader
type CascaderNode struct {
	Value    uint            `json:"value"`
	Label    string          `json:"label"`
	Children []*CascaderNode `json:"children,omitempty"`
}

func (s *ArticleCategoryService) listEnabledOrdered() ([]contentModel.ContentArticleCategory, error) {
	var list []contentModel.ContentArticleCategory
	err := global.GVA_DB.Model(&contentModel.ContentArticleCategory{}).
		Where("enabled = ?", true).
		Order("parent_id asc, sort asc, id asc").
		Find(&list).Error
	return list, err
}

// GetCategoryTree 级联选择：仅启用分类，二级树
func (s *ArticleCategoryService) GetCategoryTree() ([]CascaderNode, error) {
	list, err := s.listEnabledOrdered()
	if err != nil {
		return nil, err
	}
	return buildCascaderTree(list), nil
}

func buildCascaderTree(flat []contentModel.ContentArticleCategory) []CascaderNode {
	byParent := map[uint][]contentModel.ContentArticleCategory{}
	for _, c := range flat {
		pid := c.ParentID
		byParent[pid] = append(byParent[pid], c)
	}
	var roots []contentModel.ContentArticleCategory
	roots = append(roots, byParent[0]...)
	sort.Slice(roots, func(i, j int) bool {
		if roots[i].Sort != roots[j].Sort {
			return roots[i].Sort < roots[j].Sort
		}
		return roots[i].ID < roots[j].ID
	})
	out := make([]CascaderNode, 0, len(roots))
	for _, r := range roots {
		node := CascaderNode{Value: r.ID, Label: r.Name}
		children := byParent[r.ID]
		sort.Slice(children, func(i, j int) bool {
			if children[i].Sort != children[j].Sort {
				return children[i].Sort < children[j].Sort
			}
			return children[i].ID < children[j].ID
		})
		for _, ch := range children {
			node.Children = append(node.Children, &CascaderNode{Value: ch.ID, Label: ch.Name})
		}
		out = append(out, node)
	}
	return out
}

// GetCategoryListFlat 管理端平铺列表（含禁用的）
func (s *ArticleCategoryService) GetCategoryListFlat() ([]contentModel.ContentArticleCategory, error) {
	var list []contentModel.ContentArticleCategory
	err := global.GVA_DB.Model(&contentModel.ContentArticleCategory{}).
		Order("parent_id asc, sort asc, id asc").
		Find(&list).Error
	return list, err
}

// depthOf 返回 1 或 2；非法层级返回 error
func (s *ArticleCategoryService) depthOf(cat *contentModel.ContentArticleCategory) (int, error) {
	if cat.ParentID == 0 {
		return 1, nil
	}
	var parent contentModel.ContentArticleCategory
	if err := global.GVA_DB.First(&parent, "id = ?", cat.ParentID).Error; err != nil {
		return 0, err
	}
	if parent.ParentID != 0 {
		return 0, errors.New("分类层级超过二级")
	}
	return 2, nil
}

// ValidateCategoryIDForArticle 0 通过；否则须存在且深度≤2
func (s *ArticleCategoryService) ValidateCategoryIDForArticle(id uint) error {
	if id == 0 {
		return nil
	}
	var cat contentModel.ContentArticleCategory
	if err := global.GVA_DB.First(&cat, "id = ? AND enabled = ?", id, true).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("分类不存在或未启用")
		}
		return err
	}
	if _, err := s.depthOf(&cat); err != nil {
		return err
	}
	return nil
}

func (s *ArticleCategoryService) CreateCategory(in *contentModel.ContentArticleCategory) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return errors.New("分类名称不能为空")
	}
	if in.ParentID > 0 {
		var parent contentModel.ContentArticleCategory
		if err := global.GVA_DB.First(&parent, "id = ?", in.ParentID).Error; err != nil {
			return errors.New("父级分类不存在")
		}
		if parent.ParentID != 0 {
			return errors.New("最多支持二级分类，不能在二级下再建子类")
		}
	}
	return global.GVA_DB.Create(in).Error
}

func (s *ArticleCategoryService) UpdateCategory(in *contentModel.ContentArticleCategory) error {
	if in.ID == 0 {
		return errors.New("ID 不能为空")
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return errors.New("分类名称不能为空")
	}
	if in.ParentID == in.ID {
		return errors.New("不能将自身设为父级")
	}
	if in.ParentID > 0 {
		var parent contentModel.ContentArticleCategory
		if err := global.GVA_DB.First(&parent, "id = ?", in.ParentID).Error; err != nil {
			return errors.New("父级分类不存在")
		}
		if parent.ParentID != 0 {
			return errors.New("最多支持二级分类")
		}
	}
	return global.GVA_DB.Model(&contentModel.ContentArticleCategory{}).Where("id = ?", in.ID).Updates(map[string]any{
		"parent_id": in.ParentID,
		"name":      in.Name,
		"sort":      in.Sort,
		"enabled":   in.Enabled,
	}).Error
}

func (s *ArticleCategoryService) DeleteCategory(id uint) error {
	if id == 0 {
		return errors.New("ID 不能为空")
	}
	var n int64
	if err := global.GVA_DB.Model(&contentModel.ContentArticleCategory{}).Where("parent_id = ?", id).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return errors.New("请先删除子分类")
	}
	if err := global.GVA_DB.Model(&contentModel.ContentArticle{}).Where("category_id = ?", id).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return errors.New("该分类下仍有文章，无法删除")
	}
	return global.GVA_DB.Delete(&contentModel.ContentArticleCategory{}, "id = ?", id).Error
}
