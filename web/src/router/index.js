import Vue from 'vue'
import Router from 'vue-router'

Vue.use(Router)

export default new Router({
    routes: [
        {
            path: '/login'
            , component: () => import('@/views/Login.vue')
        },
        {
            path: '/'
            , redirect: '/dashboard'
            , component: () => import('@/components/layouts/Layout.vue')
            , children: [
                {
                    path: '/dashboard'
                    , component: () => import('@/views/Dashboard.vue')
                    , meta: { title: "运行分析", closeAble: false }
                }, {
                    path: '/job'
                    , component: () => import('@/views/job/JobList.vue')
                    , meta: { title: "任务管理", closeAble: true }
                }, {
                    path: '/trace'
                    , component: () => import('@/views/trace/TraceList.vue')
                    , meta: { title: "调度日志", closeAble: true }
                }, {
                    path: '/user'
                    , component: () => import('@/views/user/UserList.vue')
                    , meta: { title: "用户管理", closeAble: true }
                }, {
                    path: '/alarm'
                    , component: () => import('@/views/alarm/AlarmList.vue')
                    , meta: { title: "告警设置", closeAble: true }
                }, {
                    path: '/cluster'
                    , component: () => import('@/views/cluster/NodeList.vue')
                    , meta: { title: "集群管理", closeAble: true }
                }, {
                    path: '/about'
                    , component: () => import('@/views/About.vue')
                    , meta: { title: "关于系统", closeAble: true }
                }
            ]
        }]
})

