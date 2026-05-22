import service from '@/utils/request'

export const getPortalVisitorList = (params) => {
  return service({
    url: '/contentPortalVisitor/getPortalVisitorList',
    method: 'get',
    params
  })
}

export const getPortalVisitorSummary = (params) => {
  return service({
    url: '/contentPortalVisitor/getPortalVisitorSummary',
    method: 'get',
    params
  })
}
