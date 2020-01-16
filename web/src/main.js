import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import Element from 'element-ui'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import 'element-ui/lib/theme-chalk/index.css'

Vue.config.productionTip = false 
Vue.use(Element)
NProgress.configure({ showSpinner: false })
// 引入 ECharts 主模块
var echarts = require("echarts/lib/echarts");
// 引入柱状图
require("echarts/lib/chart/bar");
// 引入提示框和标题组件
require("echarts/lib/component/tooltip");
require("echarts/lib/component/title");
Vue.prototype.$echarts = echarts

import {isLogined} from '@/utils/accredit'

// 路由前置拦截
router.beforeEach((to, from, next) => {
  NProgress.start()
  if(!isLogined() && to.path!='/login'){// 登陆拦截
      next('/login')
      NProgress.done()
  }else{
      next()
  }
})
// 路由后置拦截
router.afterEach(() => {
  NProgress.done()
})

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')