<template>
  <div class="analytics">
    <!-- 选择短链接 -->
    <div class="card">
      <el-form :inline="true">
        <el-form-item label="选择短链接">
          <el-select
              v-model="selectedCode"
              placeholder="请选择或输入短链码"
              filterable
              allow-create
              style="width: 300px"
              @change="handleCodeChange"
          >
            <el-option
                v-for="item in linkOptions"
                :key="item.code"
                :label="`${item.code} - ${item.url}`"
                :value="item.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              @change="handleDateChange"
          />
        </el-form-item>
      </el-form>
    </div>

    <template v-if="selectedCode">
      <!-- 统计概览 -->
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="stat-card">
            <div class="stat-value">{{ stats.totalVisits }}</div>
            <div class="stat-label">总访问量</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card">
            <div class="stat-value">{{ stats.uniqueVisitors }}</div>
            <div class="stat-label">独立访客</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card">
            <div class="stat-value">{{ stats.todayVisits }}</div>
            <div class="stat-label">今日访问</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card">
            <div class="stat-value">{{ stats.avgDaily }}次/天</div>
            <div class="stat-label">日均访问</div>
          </div>
        </el-col>
      </el-row>

      <!-- 图表区域 -->
      <el-row :gutter="20" style="margin-top: 20px">
        <!-- 访问趋势图 -->
        <el-col :span="24">
          <div class="card">
            <h3>访问趋势</h3>
            <div class="chart-container">
              <canvas ref="trendChart"></canvas>
            </div>
          </div>
        </el-col>
      </el-row>

      <el-row :gutter="20" style="margin-top: 20px">
        <!-- 设备统计 -->
        <el-col :span="8">
          <div class="card">
            <h3>设备类型</h3>
            <div class="chart-container">
              <canvas ref="deviceChart"></canvas>
            </div>
          </div>
        </el-col>

        <!-- 浏览器统计 -->
        <el-col :span="8">
          <div class="card">
            <h3>浏览器分布</h3>
            <div class="chart-container">
              <canvas ref="browserChart"></canvas>
            </div>
          </div>
        </el-col>

        <!-- 操作系统统计 -->
        <el-col :span="8">
          <div class="card">
            <h3>操作系统</h3>
            <div class="chart-container">
              <canvas ref="osChart"></canvas>
            </div>
          </div>
        </el-col>
      </el-row>

      <!-- 访问日志 -->
      <div class="card" style="margin-top: 20px">
        <h3>最近访问记录</h3>
        <el-table :data="visitLogs" style="width: 100%">
          <el-table-column label="访问时间" width="180">
            <template #default="{ row }">
              {{ formatDate(row.visited_at) }}
            </template>
          </el-table-column>
          <el-table-column prop="ip" label="IP地址" width="150" />
          <el-table-column prop="device_type" label="设备" width="100" />
          <el-table-column prop="browser" label="浏览器" width="150" />
          <el-table-column prop="os" label="操作系统" width="150" />
          <el-table-column prop="referer" label="来源" min-width="200">
            <template #default="{ row }">
              <el-tooltip :content="row.referer || '直接访问'" placement="top">
                <span class="ellipsis">{{ row.referer || '直接访问' }}</span>
              </el-tooltip>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </template>

    <!-- 无数据提示 -->
    <div v-else class="empty-state">
      <el-empty description="请选择一个短链接查看统计数据" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Chart, registerables } from 'chart.js'
import { api } from '@/api'
import { formatDate, getRecentDays, generateColors } from '@/utils'

Chart.register(...registerables)

const route = useRoute()

// 数据
const selectedCode = ref('')
const dateRange = ref([])
const linkOptions = ref([])

const stats = ref({
  totalVisits: 0,
  uniqueVisitors: 0,
  todayVisits: 0,
  avgDaily: 0
})

const visitLogs = ref([])

// 图表实例
let trendChartInstance = null
let deviceChartInstance = null
let browserChartInstance = null
let osChartInstance = null

// 图表引用
const trendChart = ref()
const deviceChart = ref()
const browserChart = ref()
const osChart = ref()

// 处理短链码变化
const handleCodeChange = () => {
  loadAnalyticsData()
}

// 处理日期变化
const handleDateChange = () => {
  if (selectedCode.value) {
    loadAnalyticsData()
  }
}

// 加载分析数据
const loadAnalyticsData = async () => {
  if (!selectedCode.value) return

  try {
    // 加载统计数据
    const statsRes = await api.getStats(selectedCode.value)
    if (statsRes.data) {
      stats.value = {
        totalVisits: statsRes.data.total_visits || 0,
        uniqueVisitors: statsRes.data.unique_visits || 0,
        todayVisits: statsRes.data.today_visits || 0,
        avgDaily: Math.round((statsRes.data.total_visits || 0) / 7)
      }
    }

    // 加载访问日志
    const logsRes = await api.getLogs(selectedCode.value, 20)
    visitLogs.value = logsRes.data || []

    // 更新图表
    await nextTick()
    updateCharts()
  } catch (error) {
    console.error('加载分析数据失败:', error)
    // 使用模拟数据
    loadMockData()
  }
}

// 加载模拟数据
const loadMockData = () => {
  stats.value = {
    totalVisits: 1234,
    uniqueVisitors: 567,
    todayVisits: 89,
    avgDaily: 176
  }

  visitLogs.value = [
    {
      visited_at: new Date(),
      ip: '192.168.1.1',
      device_type: 'Desktop',
      browser: 'Chrome 120',
      os: 'Windows 10',
      referer: 'https://www.google.com'
    },
    {
      visited_at: new Date(),
      ip: '10.0.0.1',
      device_type: 'Mobile',
      browser: 'Safari 17',
      os: 'iOS 17',
      referer: null
    }
  ]

  updateCharts()
}

// 更新图表
const updateCharts = () => {
  updateTrendChart()
  updateDeviceChart()
  updateBrowserChart()
  updateOSChart()
}

// 更新趋势图
const updateTrendChart = () => {
  if (!trendChart.value) return

  const dates = getRecentDays(7)
  const data = [65, 78, 90, 81, 56, 55, 89]

  if (trendChartInstance) {
    trendChartInstance.destroy()
  }

  trendChartInstance = new Chart(trendChart.value, {
    type: 'line',
    data: {
      labels: dates,
      datasets: [{
        label: '访问量',
        data: data,
        borderColor: '#409EFF',
        backgroundColor: 'rgba(64, 158, 255, 0.1)',
        tension: 0.3
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        }
      },
      scales: {
        y: {
          beginAtZero: true
        }
      }
    }
  })
}

// 更新设备图表
const updateDeviceChart = () => {
  if (!deviceChart.value) return

  if (deviceChartInstance) {
    deviceChartInstance.destroy()
  }

  deviceChartInstance = new Chart(deviceChart.value, {
    type: 'pie',
    data: {
      labels: ['Desktop', 'Mobile', 'Tablet'],
      datasets: [{
        data: [65, 30, 5],
        backgroundColor: ['#409EFF', '#67C23A', '#E6A23C']
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false
    }
  })
}

// 更新浏览器图表
const updateBrowserChart = () => {
  if (!browserChart.value) return

  if (browserChartInstance) {
    browserChartInstance.destroy()
  }

  browserChartInstance = new Chart(browserChart.value, {
    type: 'doughnut',
    data: {
      labels: ['Chrome', 'Firefox', 'Safari', 'Edge', 'Others'],
      datasets: [{
        data: [45, 20, 18, 12, 5],
        backgroundColor: generateColors(5)
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false
    }
  })
}

// 更新操作系统图表
const updateOSChart = () => {
  if (!osChart.value) return

  if (osChartInstance) {
    osChartInstance.destroy()
  }

  osChartInstance = new Chart(osChart.value, {
    type: 'bar',
    data: {
      labels: ['Windows', 'macOS', 'Linux', 'iOS', 'Android'],
      datasets: [{
        label: '使用量',
        data: [40, 25, 10, 15, 10],
        backgroundColor: '#409EFF'
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        }
      },
      scales: {
        y: {
          beginAtZero: true
        }
      }
    }
  })
}

// 初始化链接选项
const initLinkOptions = () => {
  linkOptions.value = [
    { code: 'abc123', url: 'google.com' },
    { code: 'xyz789', url: 'github.com' },
    { code: 'test456', url: 'element-plus.org' }
  ]
}

onMounted(() => {
  initLinkOptions()

  // 如果URL中有code参数，自动选择
  if (route.query.code) {
    selectedCode.value = route.query.code
    loadAnalyticsData()
  }
})

// 监听路由变化
watch(() => route.query.code, (newCode) => {
  if (newCode) {
    selectedCode.value = newCode
    loadAnalyticsData()
  }
})
</script>

<style lang="scss" scoped>
.analytics {
  h3 {
    margin-bottom: 20px;
    color: #303133;
  }

  .chart-container {
    height: 300px;
    position: relative;
  }

  .ellipsis {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: block;
  }

  .empty-state {
    padding: 100px 0;
    text-align: center;
  }
}
</style>