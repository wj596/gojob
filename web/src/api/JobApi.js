import request from '@/utils/request'

const jobApi = {}
jobApi.getJobs = function (_params) {
  return request({
    url: '/jobs'
    , method: 'get'
    , params: _params
  })
}
jobApi.getJobSelections = function (_name) {
  return request({
    url: '/jobs'
    , method: 'get'
    , params: { search_type:1,name:_name }
  })
}
jobApi.getSubJobSelections = function (_id) {
  return request({
    url: '/jobs'
    , method: 'get'
    , params: { id: _id,search_type:2 }
  })
}
jobApi.getJob = function (id) {
  return request({
    url: '/jobs/' + id
    , method: 'get'
  })
}
jobApi.postJob = function (_params) {
  return request({
    url: '/jobs'
    , method: 'post'
    , data: _params
  })
}
jobApi.putJob = function (_params) {
  return request({
    url: '/jobs'
    , method: 'put'
    , data: _params
  })
}
jobApi.updateStatus = function (id, status) {
  return request({
    url: '/jobs/' + '/update_status/' + id + '/' + status
    , method: 'put'
  })
}
jobApi.deleteJob = function (id) {
  return request({
    url: '/jobs/' + id
    , method: 'delete'
  })
}

jobApi.launchJob = function (id) {
  return request({
    url: '/jobs/' + id + '/launch'
    , method: 'get'
  })
}
jobApi.validateCron = function (spec) {
  return request({
    url: '/jobs/cron_validate?spec=' + encodeURIComponent(spec)
    , method: 'post'
  })
}
export default jobApi