import axios from 'axios'
import { Message } from 'element-ui'
import { getToken } from "@/utils/accredit";
import { removeToken } from "@/utils/accredit";
import NProgress from 'nprogress'

/**
 * 封装axios
 **/
const myaxios = axios.create({
  baseURL: process.env.VUE_APP_BASE_URL //请求根地址
  , timeout: 50000 // 超时时间为5s
  , withCredentials: true// 发送请求时携带cookie
})

myaxios.interceptors.request.use(config => {
  config.headers = {
    'Content-Type': 'application/json;charset=UTF-8'
  }
  if (getToken()) {
    config.headers['Authorization'] = getToken()
  }
  NProgress.start() // start progress bar
  return config
}, error => { // 失败处理
  NProgress.done()
  console.log(error) // for debug
  Promise.reject(error)
})

// respone interceptor
myaxios.interceptors.response.use(response => {
  NProgress.done() // finish progress bar
  const res = response.data
  if (!res.succeed) {
    Message({
      message: res.message,
      type: 'error',
      duration: 5 * 1000
    })
    return Promise.reject('error')
  } else {
    return response.data
  }
}, error => {
  console.log('err:' + error)
  NProgress.done()
  if ('Network Error' === error.message) {
    Message({
      message: '网络错误，无法链接到后台服务',
      type: 'error',
      duration: 5 * 1000
    })
    return Promise.reject(error)
  }

  let status = error.response.status
  if (401 == status) {
    removeToken();
    location.reload();
  } else {
    let msg = error.message
    let d = error.response.data
    if (d) {
      msg = d.msg
    }
    Message({
      message: msg,
      type: 'error',
      duration: 5 * 1000
    })
  }
  return Promise.reject(error)
})

export default myaxios