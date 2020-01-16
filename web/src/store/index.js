import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)
// 全局缓存
const globalStore = {
    state: {
        scollapse: false// 侧边栏展开状态
        , userName: ""// 用户名称
        , keepAlivePages: []// 保活的页面（页面渲染的结果保存在内存中，不用每次都重新渲染）
    }
    , mutations: {
        TOGGLE_SCOLLAPSE(state) {
            state.scollapse = !state.scollapse
        }
        , CLOSE_SCOLLAPSE(state) {
            state.scollapse = false
        }
        , SET_USER_NAME: (state, name) => {
            state.userName = name
        }
        , SET_KEEP_ALIVE_PAGES: (state, keepAlivePage) => {
            state.keepAlivePages = keepAlivePage
        }
    }
    , actions: {
        toggleScollapse: (context) => {
            context.commit('TOGGLE_SCOLLAPSE')
        }
        , closeScollapse: (context) => {
            context.commit('CLOSE_SCOLLAPSE')
        }
        , setUserName: ({ commit }, info) => {
            commit('SET_USER_NAME', info)
        }
        , setKeepAlivePages: ({ commit }, keepAlivePages) => {
            commit('SET_KEEP_ALIVE_PAGES', keepAlivePages)
        }
        , removeUserName: ({ commit }) => {
            commit('SET_USER_NAME', "")
        }
        , removeKeepAlivePages: ({ commit }) => {
            commit('SET_KEEP_ALIVE_PAGES', null)
        }
    }
}
// --- 全局缓存结束
// get方法
const getters = {
    getScollapse: state => state.globalStore.scollapse
    , getUserName: state => state.globalStore.userName
    , getKeepAlivePages: state => state.globalStore.keepAlivePages
}

const store = new Vuex.Store({
    modules: {
        globalStore
    }
    , getters
})

export default store