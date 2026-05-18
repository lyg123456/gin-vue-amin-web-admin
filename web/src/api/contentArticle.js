import service from '@/utils/request'

export const createContentArticle = (data) => {
  return service({
    url: '/contentArticle/createArticle',
    method: 'post',
    data
  })
}

export const updateContentArticle = (data) => {
  return service({
    url: '/contentArticle/updateArticle',
    method: 'put',
    data
  })
}

export const deleteContentArticle = (data) => {
  return service({
    url: '/contentArticle/deleteArticle',
    method: 'delete',
    data
  })
}

export const findContentArticle = (params) => {
  return service({
    url: '/contentArticle/findArticle',
    method: 'get',
    params
  })
}

export const getContentArticleList = (params) => {
  return service({
    url: '/contentArticle/getArticleList',
    method: 'get',
    params
  })
}

export const publishContentArticle = (data) => {
  return service({
    url: '/contentArticle/publishArticle',
    method: 'post',
    data
  })
}

/** 百度文心：根据标题/关键词生成 Markdown 正文（服务端读 AK/SK） */
export const generateArticleByBaidu = (data) => {
  return service({
    url: '/contentArticle/generateArticleByBaidu',
    method: 'post',
    data,
    timeout: 600000
  })
}

/** 诊断：配置 → token → 最小对话（用于定位错误 6 等） */
export const diagnoseBaiduWenxin = () => {
  return service({
    url: '/contentArticle/diagnoseBaiduWenxin',
    method: 'get',
    timeout: 120000
  })
}

// public: SEO 公开访问（不需要 token，仍然走同一个 baseURL）
export const getPublishedArticleBySlug = (slug) => {
  return service({
    url: `/public/article/${encodeURIComponent(slug)}`,
    method: 'get'
  })
}

