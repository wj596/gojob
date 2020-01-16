import request from '@/utils/request'

const runtimeApi = {}
runtimeApi.getRuntime = function () {
  return request({
    url: '/runtimes'
    , method: 'get'
  })
}
runtimeApi.getRunmode = function () {
  return request({
    url: '/runtimes/runmode'
    , method: 'get'
  })
}

export default runtimeApi