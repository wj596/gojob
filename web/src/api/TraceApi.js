import request from '@/utils/request'

const traceApi = {}
traceApi.getTraces = function (_params) {
  return request({
    url: '/traces'
    ,method: 'get'
    ,params: _params
  })
}
traceApi.getTrace = function (traceId) {
  return request({
    url: '/traces/'+traceId
    ,method: 'get'
  })
}
traceApi.cleanTrace = function (_data) {
  return request({
    url: '/traces/clean'
    ,method: 'post'
    ,data: _data
  })
}
traceApi.statisticToday = function () {
  return request({
      url: '/statistic/today'
      , method: 'get'
  })
}
traceApi.statisticWeek = function () {
  return request({
      url: '/statistic/week'
      , method: 'get'
  })
}
traceApi.statisticMonth = function () {
  return request({
      url: '/statistic/month'
      , method: 'get'
  })
}
traceApi.statisticAll = function () {
  return request({
      url: '/statistic/all'
      , method: 'get'
  })
}
export default traceApi