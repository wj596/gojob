import request from '@/utils/request'

const alarmApi = {}
alarmApi.getAlarmConfig = function () {
  return request({
    url: '/alarm_configs'
    , method: 'get'
  })
}

alarmApi.putAlarmConfig = function (_params) {
  return request({
    url: '/alarm_configs'
    , method: 'put'
    , data: _params
  })
}
alarmApi.testAlarmConfig = function (_params) {
    return request({
      url: '/alarm_configs/test'
      , method: 'post'
      , data: _params
    })
  }
export default alarmApi