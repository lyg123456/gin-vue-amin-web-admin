import service from '@/utils/request'

export const getContentArticleCategoryTree = () => {
  return service({
    url: '/contentArticleCategory/getCategoryTree',
    method: 'get',
    donNotShowLoading: true
  })
}

export const getContentArticleCategoryList = () => {
  return service({
    url: '/contentArticleCategory/getCategoryList',
    method: 'get',
    donNotShowLoading: true
  })
}

export const createContentArticleCategory = (data) => {
  return service({
    url: '/contentArticleCategory/createCategory',
    method: 'post',
    data
  })
}

export const updateContentArticleCategory = (data) => {
  return service({
    url: '/contentArticleCategory/updateCategory',
    method: 'put',
    data
  })
}

export const deleteContentArticleCategory = (data) => {
  return service({
    url: '/contentArticleCategory/deleteCategory',
    method: 'delete',
    data
  })
}
