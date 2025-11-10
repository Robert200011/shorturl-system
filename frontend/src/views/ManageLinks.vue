<template>
  <div class="manage-links">
    <!-- 搜索和筛选 -->
    <div class="card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="短链码">
          <el-input v-model="searchForm.shortCode" placeholder="输入短链码" />
        </el-form-item>
        <el-form-item label="原始链接">
          <el-input v-model="searchForm.url" placeholder="输入原始链接" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部">
            <el-option label="全部" value="" />
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 链接列表 -->
    <div class="card">
      <el-table
          :data="tableData"
          style="width: 100%"
          v-loading="loading"
      >
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="expand-content">
              <div class="expand-item">
                <label>标题：</label>
                <span>{{ row.title || '无' }}</span>
              </div>
              <div class="expand-item">
                <label>描述：</label>
                <span>{{ row.description || '无' }}</span>
              </div>
              <div class="expand-item">
                <label>过期时间：</label>
                <span>{{ row.expire_at ? formatDate(row.expire_at) : '永久有效' }}</span>
              </div>
              <div class="expand-item">
                <label>更新时间：</label>
                <span>{{ formatDate(row.updated_at) }}</span>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="short_code" label="短链码" width="120">
          <template #default="{ row }">
            <el-button
                type="primary"
                text
                @click="copyShortUrl(row.short_code)"
            >
              {{ row.short_code }}
            </el-button>
          </template>
        </el-table-column>

        <el-table-column label="原始链接" min-width="300">
          <template #default="{ row }">
            <el-tooltip :content="row.original_url" placement="top">
              <div class="url-cell">
                <span class="url-text">{{ row.original_url }}</span>
                <el-button
                    type="primary"
                    text
                    size="small"
                    @click="openUrl(row.original_url)"
                >
                  <el-icon><Link /></el-icon>
                </el-button>
              </div>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column prop="visit_count" label="访问次数" width="100">
          <template #default="{ row }">
            <el-tag type="info">{{ row.visit_count }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button
                type="primary"
                text
                size="small"
                @click="viewDetails(row)"
            >
              详情
            </el-button>
            <el-button
                type="primary"
                text
                size="small"
                @click="viewAnalytics(row.short_code)"
            >
              统计
            </el-button>
            <el-button
                type="primary"
                text
                size="small"
                @click="showQRCode(row)"
            >
              二维码
            </el-button>
            <el-button
                :type="row.status === 1 ? 'danger' : 'success'"
                text
                size="small"
                @click="toggleStatus(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
          style="margin-top: 20px"
      />
    </div>

    <!-- 详情对话框 -->
    <el-dialog v-model="detailDialog" title="短链接详情" width="600px">
      <div v-if="currentLink" class="detail-content">
        <div class="detail-item">
          <label>短链码：</label>
          <span>{{ currentLink.short_code }}</span>
        </div>
        <div class="detail-item">
          <label>短链接：</label>
          <a :href="getFullUrl(currentLink.short_code)" target="_blank">
            {{ getFullUrl(currentLink.short_code) }}
          </a>
        </div>
        <div class="detail-item">
          <label>原始链接：</label>
          <span>{{ currentLink.original_url }}</span>
        </div>
        <div class="detail-item">
          <label>标题：</label>
          <span>{{ currentLink.title || '无' }}</span>
        </div>
        <div class="detail-item">
          <label>描述：</label>
          <span>{{ currentLink.description || '无' }}</span>
        </div>
        <div class="detail-item">
          <label>访问次数：</label>
          <span>{{ currentLink.visit_count }}</span>
        </div>
        <div class="detail-item">
          <label>状态：</label>
          <el-tag :type="currentLink.status === 1 ? 'success' : 'danger'">
            {{ currentLink.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </div>
        <div class="detail-item">
          <label>过期时间：</label>
          <span>{{ currentLink.expire_at ? formatDate(currentLink.expire_at) : '永久有效' }}</span>
        </div>
        <div class="detail-item">
          <label>创建时间：</label>
          <span>{{ formatDate(currentLink.created_at) }}</span>
        </div>
      </div>
    </el-dialog>

    <!-- 二维码对话框 -->
    <el-dialog v-model="qrcodeDialog" title="二维码" width="400px" center>
      <div class="qrcode-content">
        <canvas ref="qrcodeCanvas"></canvas>
        <div class="qrcode-url">{{ currentQRUrl }}</div>
        <el-button type="primary" @click="downloadQRCode">下载二维码</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Link } from '@element-plus/icons-vue'
import QRCode from 'qrcode'
import { api } from '@/api'
import { simpleCopy, formatDate } from '@/utils'

const router = useRouter()

// 搜索表单
const searchForm = ref({
  shortCode: '',
  url: '',
  status: ''
})

// 表格数据
const tableData = ref([])
const loading = ref(false)

// 分页
const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0
})

// 对话框
const detailDialog = ref(false)
const qrcodeDialog = ref(false)
const currentLink = ref(null)
const currentQRUrl = ref('')
const qrcodeCanvas = ref()

// 搜索
const handleSearch = () => {
  pagination.value.page = 1
  loadData()
}

// 重置
const handleReset = () => {
  searchForm.value = {
    shortCode: '',
    url: '',
    status: ''
  }
  handleSearch()
}

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    // 模拟数据
    tableData.value = [
      {
        short_code: 'abc123',
        original_url: 'https://www.google.com/search?q=vue3+composition+api',
        title: 'Vue3 文档',
        description: 'Vue3 Composition API 官方文档',
        visit_count: 142,
        status: 1,
        expire_at: null,
        created_at: new Date(),
        updated_at: new Date()
      },
      {
        short_code: 'xyz789',
        original_url: 'https://github.com/vuejs/vue-next',
        title: 'Vue GitHub',
        description: 'Vue.js 3.0 源代码仓库',
        visit_count: 89,
        status: 1,
        expire_at: new Date('2024-12-31'),
        created_at: new Date(),
        updated_at: new Date()
      },
      {
        short_code: 'test456',
        original_url: 'https://element-plus.org/',
        title: 'Element Plus',
        description: 'Element Plus 组件库官网',
        visit_count: 56,
        status: 0,
        expire_at: null,
        created_at: new Date(),
        updated_at: new Date()
      }
    ]
    pagination.value.total = 3
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// 分页变化
const handleSizeChange = (val) => {
  pagination.value.pageSize = val
  loadData()
}

const handlePageChange = (val) => {
  pagination.value.page = val
  loadData()
}

// 复制短链接
const copyShortUrl = (code) => {
  const url = getFullUrl(code)
  simpleCopy(url, '短链接已复制')
}

// 获取完整URL
const getFullUrl = (code) => {
  return `http://localhost:8002/${code}`
}

// 打开URL
const openUrl = (url) => {
  window.open(url, '_blank')
}

// 查看详情
const viewDetails = (row) => {
  currentLink.value = row
  detailDialog.value = true
}

// 查看统计
const viewAnalytics = (code) => {
  router.push({
    path: '/analytics',
    query: { code }
  })
}

// 显示二维码
const showQRCode = async (row) => {
  currentLink.value = row
  currentQRUrl.value = getFullUrl(row.short_code)
  qrcodeDialog.value = true

  await nextTick()
  await generateQRCode(currentQRUrl.value)
}

// 生成二维码
const generateQRCode = async (url) => {
  if (!qrcodeCanvas.value) return

  try {
    await QRCode.toCanvas(qrcodeCanvas.value, url, {
      width: 300,
      margin: 2,
      color: {
        dark: '#000000',
        light: '#ffffff'
      }
    })
  } catch (error) {
    console.error('生成二维码失败:', error)
  }
}

// 下载二维码
const downloadQRCode = () => {
  const canvas = qrcodeCanvas.value
  if (!canvas) return

  const link = document.createElement('a')
  link.download = `qrcode-${currentLink.value.short_code}.png`
  link.href = canvas.toDataURL()
  link.click()
}

// 切换状态
const toggleStatus = async (row) => {
  const action = row.status === 1 ? '禁用' : '启用'

  try {
    await ElMessageBox.confirm(
        `确定要${action}这个短链接吗？`,
        '提示',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
    )

    // 这里应该调用API更新状态
    row.status = row.status === 1 ? 0 : 1
    ElMessage.success(`${action}成功`)
  } catch {
    // 用户取消
  }
}

onMounted(() => {
  loadData()
})
</script>

<style lang="scss" scoped>
.manage-links {
  .expand-content {
    padding: 10px 50px;

    .expand-item {
      margin-bottom: 10px;

      label {
        display: inline-block;
        width: 80px;
        color: #909399;
      }

      span {
        color: #606266;
      }
    }
  }

  .url-cell {
    display: flex;
    align-items: center;
    gap: 5px;

    .url-text {
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .detail-content {
    .detail-item {
      margin-bottom: 15px;

      label {
        display: inline-block;
        width: 100px;
        color: #909399;
      }

      span, a {
        color: #606266;
      }

      a {
        color: #409eff;
        text-decoration: none;

        &:hover {
          text-decoration: underline;
        }
      }
    }
  }

  .qrcode-content {
    text-align: center;

    canvas {
      margin: 20px auto;
    }

    .qrcode-url {
      margin: 10px 0;
      color: #606266;
      word-break: break-all;
    }
  }
}
</style>