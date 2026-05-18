import service from '@/utils/request'

export const getPortalContactLeadList = (params) => {
  return service({
    url: '/contentPortalContactLead/getPortalContactLeadList',
    method: 'get',
    params
  })
}
