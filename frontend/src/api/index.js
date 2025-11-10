import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建axios实例
const request = axios.create({
    baseURL: '/api',
    timeout: 10000
})

// 请求拦截器
request.interceptors.request.use(
    config => {
        // 可以在这里添加token等认证信息
        const token = localStorage.getItem('token')
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    error => {
        console.error('请求错误:', error)
        return Promise.reject(error)
    }
)

// 响应拦截器
request.interceptors.response.use(
    response => {
        const res = response.data
        if (res.code !== 0) {
            ElMessage.error(res.message || '请求失败')
            return Promise.reject(new Error(res.message || 'Error'))
        }
        return res
    },
    error => {
        console.error('响应错误:', error)
        ElMessage.error(error.message || '网络错误')
        return Promise.reject(error)
    }
)

// API接口
export const api = {
    // 创建短链接
    createShortLink(data) {
        return request.post('/shorten', data)
    },

    // 批量创建短链接
    batchCreateShortLinks(urls) {
        return request.post('/batch/shorten', { urls })
    },

    // 获取短链接详情
    getShortLink(code) {
        return request.get(`/links/${code}`)
    },

    // 获取访问统计
    getStats(code) {
        return request.get(`/stats/${code}`)
    },

    // 获取访问日志
    getLogs(code, limit = 50) {
        return request.get(`/stats/${code}/logs`, {
            params: { limit }
        })
    },

    // 获取每日统计
    getDailyStats(code, startDate, endDate) {
        return request.get(`/analytics/daily/${code}`, {
            params: { start_date: startDate, end_date: endDate }
        })
    },

    // 获取每小时统计
    getHourlyStats(code, date) {
        return request.get(`/analytics/hourly/${code}`, {
            params: { date }
        })
    },

    // 获取浏览器统计
    getBrowserStats(code) {
        return request.get(`/analytics/browser/${code}`)
    },

    // 获取设备统计
    getDeviceStats(code) {
        return request.get(`/analytics/device/${code}`)
    },

    // 获取操作系统统计
    getOSStats(code) {
        return request.get(`/analytics/os/${code}`)
    }
}

export default request