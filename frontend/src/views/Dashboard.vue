<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20">
      <el-col :span="6">
        <div class="stat-card">
          <el-icon class="stat-icon" color="#409EFF"><Link /></el-icon>
          <div class="stat-value">{{ stats.totalLinks }}</div>
          <div class="stat-label">总链接数</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <el-icon class="stat-icon" color="#67C23A"><View /></el-icon>
          <div class="stat-value">{{ stats.totalVisits }}</div>
          <div class="stat-label">总访问量</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <el-icon class="stat-icon" color="#E6A23C"><TrendCharts /></el-icon>
          <div class="stat-value">{{ stats.todayVisits }}</div>
          <div class="stat-label">今日访问</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <el-icon class="stat-icon" color="#F56C6C"><Timer /></el-icon>
          <div class="stat-value">{{ stats.avgVisits }}</div>
          <div class="stat-label">平均访问</div>
        </div>
      </el-col>
    </el-row>

    <!-- 快速创建 -->
    <div class="card quick-create">
      <h3>快速创建短链接</h3>
      <el-form :model="quickForm" :rules="rules" ref="quickFormRef">
        <el-row :gutter="20">
          <el-col :span="18">
            <el-form-item prop="url">
              <el-input
                  v-model="quickForm.url"
                  placeholder="请输入要缩短的网址"
                  size="large"
              >
                <template #prefix>
                  <el-icon><Link /></el-icon>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-button
                type="primary"
                size="large"
                style="width: 100%"
                :loading="creating"
                @click="handleQuickCreate"
            >
              生成短链接
            </el-button>
          </el-col>
        </el-row>
      </el-form>

      <!-- 生成结果 -->
      <div v-if="shortUrl" class="result-box">
        <el-alert type="success" :closable="false">
          <div class="result-content">
            <span>短链接：</span>
            <a :href="shortUrl" target="_blank" class="short-url">{{ shortUrl }}</a>
            <el-button
                type="primary"
                text
                @click="handleCopy(shortUrl)"
                class="copy-btn"
            >
              <el-icon><CopyDocument /></el-icon>
              复制
            </el-button>
          </div>
        </el-alert>
      </div>
    </div>

    <!-- 最近创建的链接 -->
    <div class="card">
      <h3>最近创建的链接</h3>
      <el-table :data="recentLinks" style="width: 100%">
        <el-table-column prop="short_code" label="短链码" width="120" />
        <el-table-column label="原始链接" min-width="300">
          <template #default="{ row }">
            <el-tooltip :content="row.original_url" placement="top">
              <span class="ellipsis">{{ row.original_url }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="visit_count" label="访问次数" width="100" />
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
                type="primary"
                text
                size="small"
                @click="handleCopy(getFullUrl(row.short_code))"
            >
              复制链接
            </el-button>
            <el-button
                type="primary"
                text
                size="small"
                @click="viewAnalytics(row.short_code)"
            >
              查看分析
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Link,
  View,
  TrendCharts,
  Timer,
  CopyDocument
} from '@element-plus/icons-vue'
import { api } from '@/api'
import { simpleCopy, formatDate } from '@/utils'

const router = useRouter()

// 统计数据
const stats = ref({
  totalLinks: 0,
  totalVisits: 0,
  todayVisits: 0,
  avgVisits: 0
})

// 快速创建表单
const quickFormRef = ref()
const quickForm = ref({
  url: ''
})
const creating = ref(false)
const shortUrl = ref('')

const rules = {
  url: [
    { required: true, message: '请输入网址', trigger: 'blur' },
    { type: 'url', message: '请输入有效的网址', trigger: 'blur' }
  ]
}

// 最近链接
const recentLinks = ref([])

// 快速创建短链接
const handleQuickCreate = async () => {
  const valid = await quickFormRef.value.validate()
  if (!valid) return

  creating.value = true
  try {
    const res = await api.createShortLink({
      original_url: quickForm.value.url
    })

    shortUrl.value = res.data.short_url
    ElMessage.success('短链接创建成功')

    // 刷新列表
    loadRecentLinks()

    // 清空表单
    quickForm.value.url = ''
  } catch (error) {
    console.error('创建失败:', error)
  } finally {
    creating.value = false
  }
}

// 复制链接
const handleCopy = (url) => {
  simpleCopy(url, '链接已复制到剪贴板')
}

// 获取完整URL
const getFullUrl = (code) => {
  return `http://localhost:8002/${code}`
}

// 查看分析
const viewAnalytics = (code) => {
  router.push({
    path: '/analytics',
    query: { code }
  })
}

// 加载统计数据
const loadStats = async () => {
  // 这里模拟数据，实际应该从API获取
  stats.value = {
    totalLinks: 128,
    totalVisits: 15432,
    todayVisits: 523,
    avgVisits: 120
  }
}

// 加载最近链接
const loadRecentLinks = async () => {
  // 模拟数据，实际应该调用API
  recentLinks.value = [
    {
      short_code: 'abc123',
      original_url: 'https://www.google.com/search?q=vue3+composition+api',
      visit_count: 42,
      created_at: new Date()
    },
    {
      short_code: 'xyz789',
      original_url: 'https://github.com/vuejs/vue-next',
      visit_count: 18,
      created_at: new Date()
    }
  ]
}

onMounted(() => {
  loadStats()
  loadRecentLinks()
})
</script>

<style lang="scss" scoped>
.dashboard {
  .quick-create {
    margin: 20px 0;

    h3 {
      margin-bottom: 20px;
      color: #303133;
    }

    .result-box {
      margin-top: 20px;

      .result-content {
        display: flex;
        align-items: center;
        gap: 10px;

        .short-url {
          color: #409EFF;
          text-decoration: none;
          font-weight: bold;

          &:hover {
            text-decoration: underline;
          }
        }
      }
    }
  }

  .ellipsis {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: block;
  }
}
</style>