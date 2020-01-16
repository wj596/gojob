import request from '@/utils/request'
/**
 * cluster API 封装
 */
const clusterApi = {}
clusterApi.getNodes = function () {
  return request({
    url: '/cluster/nodes'
    ,method: 'get'
  })
}
clusterApi.removeNode = function (_params) {
  return request({
    url: '/cluster/remove/'+_params
    ,method: 'get'
  })
}
export default clusterApi