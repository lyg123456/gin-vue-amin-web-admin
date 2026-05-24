import service from '@/utils/request'

/** 申请临时邮箱 */
export const createTempMailbox = () =>
  service({
    url: '/public/office/tempEmail/create',
    method: 'get',
    donNotShowLoading: true
  })

/** 收件箱 */
export const getTempEmailMessages = (params) =>
  service({
    url: '/public/office/tempEmail/messages',
    method: 'get',
    params,
    donNotShowLoading: true
  })

/** 读信 */
export const readTempEmailMessage = (params) =>
  service({
    url: '/public/office/tempEmail/message',
    method: 'get',
    params,
    donNotShowLoading: true
  })

/** 转换能力（是否已装 LibreOffice 等） */
export const getOfficeConvertCapabilities = () =>
  service({
    url: '/public/office/convert/capabilities',
    method: 'get',
    donNotShowLoading: true
  })

/** 上传文件转换，返回 Blob */
export const convertOfficeFile = (formData) =>
  service({
    url: '/public/office/convert/file',
    method: 'post',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' },
    responseType: 'blob',
    donNotShowLoading: false
  })

export const getMediaCapabilities = () =>
  service({ url: '/public/office/media/capabilities', method: 'get', donNotShowLoading: true })

export const processMediaVideo = (formData) =>
  service({
    url: '/public/office/media/video',
    method: 'post',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' },
    responseType: 'blob'
  })

export const compositeImages = (formData) =>
  service({
    url: '/public/office/media/composite',
    method: 'post',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' },
    responseType: 'blob'
  })

export const extractImageBackground = (formData) =>
  service({
    url: '/public/office/media/extractBackground',
    method: 'post',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' },
    responseType: 'blob'
  })

export const compressOfficeImage = (formData) =>
  service({
    url: '/public/office/compress/image',
    method: 'post',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' },
    responseType: 'blob'
  })

export const compressOfficeExcel = (formData) =>
  service({
    url: '/public/office/compress/excel',
    method: 'post',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' },
    responseType: 'blob'
  })

export const getTweetStyles = () =>
  service({ url: '/public/office/tweet/styles', method: 'get', donNotShowLoading: true })

export const rewriteTweet = (data) =>
  service({ url: '/public/office/tweet/rewrite', method: 'post', data, donNotShowLoading: false })

/** 下载网站前端页面 ZIP（同域爬取） */
export const downloadWebsitePagesZip = (data) =>
  service({
    url: '/public/office/web/downloadPages',
    method: 'post',
    data,
    responseType: 'blob'
  })

/** @deprecated 请用 downloadWebsitePagesZip */
export const generateWebStyleZip = downloadWebsitePagesZip

/** 商品爬取导出 Excel */
/** 抖音官方垂类列表 */
export const getDouyinOfficialCategories = () =>
  service({
    url: '/public/office/douyin/categories',
    method: 'get',
    donNotShowLoading: true
  })

/** 检测 Cookie 是否可用 */
export const verifyDouyinCookie = (data) =>
  service({
    url: '/public/office/douyin/verifyCookie',
    method: 'post',
    data,
    donNotShowLoading: true,
    timeout: 60000
  })

/** Cookie 按垂类抓取视频（每类限条数、热度下限、metric: auto|play|digg） */
export const crawlDouyinIndustryVideos = (data) =>
  service({
    url: '/public/office/douyin/crawl',
    method: 'post',
    data,
    donNotShowLoading: false,
    timeout: 180000
  })

/** 微信视频号官方垂类列表 */
export const getWechatOfficialCategories = () =>
  service({
    url: '/public/office/wechat/categories',
    method: 'get',
    donNotShowLoading: true
  })

/** 检测视频号助手 Cookie */
export const verifyWechatCookie = (data) =>
  service({
    url: '/public/office/wechat/verifyCookie',
    method: 'post',
    data,
    donNotShowLoading: true,
    timeout: 60000
  })

/** Cookie 按垂类抓取微信视频（每类限条数、热度下限、metric: auto|play|digg） */
export const crawlWechatIndustryVideos = (data) =>
  service({
    url: '/public/office/wechat/crawl',
    method: 'post',
    data,
    donNotShowLoading: false,
    timeout: 180000
  })

/** 小红书官方垂类列表 */
export const getXhsOfficialCategories = () =>
  service({
    url: '/public/office/xhs/categories',
    method: 'get',
    donNotShowLoading: true
  })

/** 检测小红书 Cookie */
export const verifyXhsCookie = (data) =>
  service({
    url: '/public/office/xhs/verifyCookie',
    method: 'post',
    data,
    donNotShowLoading: true,
    timeout: 90000
  })

/** Cookie 按垂类抓取小红书视频笔记 */
export const crawlXhsIndustryVideos = (data) =>
  service({
    url: '/public/office/xhs/crawl',
    method: 'post',
    data,
    donNotShowLoading: false,
    timeout: 180000
  })

/** 服务端代下视频（带 Referer/Cookie，解决直链 403） */
export const proxyOfficeMediaDownload = (data) =>
  service({
    url: '/public/office/media/download',
    method: 'post',
    data,
    responseType: 'blob',
    donNotShowLoading: false,
    timeout: 300000
  })

export const crawlWebProductsExcel = (data) =>
  service({
    url: '/public/office/web/crawlProducts',
    method: 'post',
    data,
    responseType: 'blob'
  })

export const speedPing = () =>
  service({ url: '/public/office/speed/ping', method: 'get', donNotShowLoading: true })

export const speedInfo = () =>
  service({ url: '/public/office/speed/info', method: 'get', donNotShowLoading: true })

export const speedUploadTest = (data) =>
  service({
    url: '/public/office/speed/upload',
    method: 'post',
    data,
    headers: { 'Content-Type': 'application/octet-stream' },
    donNotShowLoading: true
  })

export const getWatermarkCapabilities = () =>
  service({ url: '/public/office/watermark/capabilities', method: 'get', donNotShowLoading: true })

export const removeWatermark = (formData) =>
  service({
    url: '/public/office/watermark/remove',
    method: 'post',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' },
    responseType: 'blob'
  })
