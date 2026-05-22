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
