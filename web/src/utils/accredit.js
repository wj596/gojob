/**
 * 用户登陆凭证、资源工具
 */
import Cookies from 'js-cookie'
import store from '../store'
import userApi from "@/api/UserApi";

const TokenKey = 'USER-TOKEN' // 常量用户令牌名称
function getCookie(key) {
    return Cookies.get(key)
}
function setCookie(key, value) {
    return Cookies.set(key, value)
}
function removeCookie(key) {
    return Cookies.remove(key)
}
// 判断用户是否已经登录
export function isLogined() {
    let token = getCookie(TokenKey);
    if (token) {
        if (!store.getters.getUserName) {// 用户刷新了页面
            userApi.authorised().then(accredit => {
                store.dispatch('setUserName', accredit.userName)
            })
        }
        return true
    }
    return false
}

// 判断用户是否已经登录
export function getToken() {
    let token = getCookie(TokenKey);
    if (token) {
        return token
    }
    return null
}

export function removeToken() {
    removeCookie(TokenKey)
}

// 用户登陆成功后缓存相关凭证和资源
export function storeAccredit(accredit) {
    setCookie(TokenKey, accredit.token)
    store.dispatch('setUserName', accredit.userName)
}

// 用户退出成功后清理相关凭证和资源
export function clearAccredit() {
    removeCookie(TokenKey)
    store.dispatch('removeUserName')
    store.dispatch('removeKeepAlivePages')
}