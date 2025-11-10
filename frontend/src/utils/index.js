import ClipboardJS from 'clipboard'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'

// 复制到剪贴板
export function copyToClipboard(text, message = '已复制到剪贴板') {
    const clipboard = new ClipboardJS('.copy-btn', {
        text: () => text
    })

    clipboard.on('success', () => {
        ElMessage.success(message)
        clipboard.destroy()
    })

    clipboard.on('error', () => {
        ElMessage.error('复制失败，请手动复制')
        clipboard.destroy()
    })

    // 触发点击
    document.querySelector('.copy-btn').click()
}

// 简单复制方法
export async function simpleCopy(text, message = '已复制到剪贴板') {
    try {
        await navigator.clipboard.writeText(text)
        ElMessage.success(message)
        return true
    } catch (err) {
        // 降级方案
        const textarea = document.createElement('textarea')
        textarea.value = text
        textarea.style.position = 'fixed'
        textarea.style.opacity = '0'
        document.body.appendChild(textarea)
        textarea.select()

        try {
            document.execCommand('copy')
            ElMessage.success(message)
            return true
        } catch (err) {
            ElMessage.error('复制失败，请手动复制')
            return false
        } finally {
            document.body.removeChild(textarea)
        }
    }
}

// 格式化日期
export function formatDate(date, format = 'YYYY-MM-DD HH:mm:ss') {
    return dayjs(date).format(format)
}

// 格式化数字
export function formatNumber(num) {
    if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M'
    } else if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K'
    }
    return num.toString()
}

// 验证URL
export function validateURL(url) {
    try {
        new URL(url)
        return true
    } catch {
        return false
    }
}

// 生成随机颜色
export function generateColors(count) {
    const colors = [
        '#5470c6',
        '#91cc75',
        '#fac858',
        '#ee6666',
        '#73c0de',
        '#3ba272',
        '#fc8452',
        '#9a60b4',
        '#ea7ccc'
    ]
    return colors.slice(0, count)
}

// 获取最近N天的日期
export function getRecentDays(days = 7) {
    const dates = []
    for (let i = days - 1; i >= 0; i--) {
        dates.push(dayjs().subtract(i, 'day').format('YYYY-MM-DD'))
    }
    return dates
}

// 获取今天的小时列表
export function getTodayHours() {
    const hours = []
    for (let i = 0; i < 24; i++) {
        hours.push(`${i.toString().padStart(2, '0')}:00`)
    }
    return hours
}