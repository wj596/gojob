import request from '@/utils/request'
/**
 * user API 封装
 */
const userApi = {}
userApi.getUsers = function (_params) {
  return request({
    url: '/users'
    ,method: 'get'
    ,params: _params
  })
}
userApi.getUsersForMailSelect = function () {
  return request({
    url: '/users'
    ,method: 'get'
    ,params: { has_email: "true" }
  })
}
userApi.getUser = function (_params) {
  return request({
    url: '/users/name/'+_params
    ,method: 'get'
  })
}
userApi.postUser = function (_data) {
  return request({
    url: '/users'
    ,method: 'post'
    ,data: _data
  })
}
userApi.putUser = function (_data) {
  return request({
    url: '/users'
    ,method: 'put'
    ,data: _data
  })
}
userApi.deleteUser = function (id) {
  return request({
    url: '/users/'+id
    ,method: 'delete'
  })
}
userApi.login = function (_params) {
  return request({
    url: '/users/login'
    ,method: 'post'
    ,data: _params
  })
}
userApi.authorised = function () {
  return request({
    url: 'users/authorised'
    ,method: 'get'
  })
}
userApi.logout = function () {
  return request({
    url: '/users/logout'
    ,method: 'get'
  })
}
export default userApi