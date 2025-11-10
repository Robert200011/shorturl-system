<template>
  <div class="batch-create">
    <div class="card">
      <h3>批量创建短链接</h3>
      <el-alert
          type="info"
          :closable="false"
          style="margin-bottom: 20px"
      >
        <template #default>
          请输入要批量创建短链接的URL列表，每行一个URL，最多支持100个
        </template>
      </el-alert>

      <el-form ref="formRef" :model="form" :rules="rules">
        <el-form-item prop="urls">
          <el-input
              v-model="form.urls"
              type="textarea"
              :rows="10"
              placeholder="请输入URL列表，每行一个，例如：&#10;https://www.example1.com&#10;https://www.example2.com&#10;https://www.example3.com"
          />
        </el-form-item>

        <el-form-item>
          <div class="form-actions">
            <div>
              <el-button type="primary" @click="handleSubmit" :loading="loading">
                批量生成
              </el-button>
              <el-button @click="handleClear">清空</el-button>
              <el-button @click="loadExample">加载示例</el-button>
            </div>
            <div class="url-count">
              已输入 {{ urlCount }} 个URL
            </div>
          </div>
        </el-form-item>
      </el-form>
    </div>

    <!-- 生成结果 -->
    <div v-if="results.length > 0" class="card">
      <h3>生成结果</h3>
      <div class="result-summary">
        <el-tag type="success">成功: {{ successCount }}</el-tag>
        <el-tag type="danger">失败: {{ failedCount }}</el-tag>
        <el-button
            type="primary"
            text
            @click="exportResults"
            style="margin-left: 20px"
        >
          <el-icon><Download /></el-icon>
          导出结果
        </el-button>
        <el-button
            type="primary"
            text
            @click="copyAllLinks"
        >
          <el-icon><CopyDocument /></el-icon>
          复制所有短链接
        </el-button>
      </div>

      <el-table :data="results" style="width: 100%; margin-top: 20px">
        <el-table-column type="index" label="序号" width="60" />

        <el-table-column label="原始链接" min-width="300">
          <template #default="{ row }">
            <el-tooltip :content="row.original_url" placement="top">
              <span class="ellipsis">{{ row.original_url }}</span>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column label="短链接" width="300">
          <template #default="{ row }">
            <div v-if="row.success" class="link-cell">
              <a :href="row.short_url" target="_blank">{{ row.short_url }}</a>
              <el-button
                  type="primary"
                  text
                  size="small"
                  @click="copyLink(row.short_url)"
              >
                复制
              </el-button>
            </div>
            <el-tag v-else type="danger">生成失败</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="短链码" width="120">
          <template #default="{ row }">
            {{ row.short_code || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.success ? 'success' : 'danger'">
              {{ row.success ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Download, CopyDocument } from '@element-plus/icons-vue'
import { api } from '@/api'
import { simpleCopy, validateURL } from '@/utils'

const formRef = ref()
const form = ref({
  urls: ''
})

const rules = {
  urls: [
    { required: true, message: '请输入URL列表', trigger: 'blur' }
  ]
}

const loading = ref(false)
const results = ref([])

// 计算URL数量
const urlCount = computed(() => {
  if (!form.value.urls) return 0
  const urls = form.value.urls.split('\n').filter(url => url.trim())
  return urls.length
})

// 计算成功和失败数量
const successCount = computed(() => {
  return results.value.filter(r => r.success).length
})

const failedCount = computed(() => {
  return results.value.filter(r => !r.success).length
})

// 提交批量创建
const handleSubmit = async () => {
  const valid = await formRef.value.validate()
  if (!valid) return

  const urls = form.value.urls
      .split('\n')
      .filter(url => url.trim())
      .map(url => url.trim())

  if (urls.length === 0) {
    ElMessage.warning('请至少输入一个URL')
    return
  }

  if (urls.length > 100) {
    ElMessage.warning('最多支持100个URL')
    return
  }

  // 验证URL格式
  const invalidUrls = urls.filter(url => !validateURL(url))
  if (invalidUrls.length > 0) {
    ElMessage.error(`存在${invalidUrls.length}个无效的URL格式`)
    return
  }

  loading.value = true
  try {
    const res = await api.batchCreateShortLinks(urls)

    // 处理结果
    results.value = res.data.results.map((item, index) => ({
      original_url: urls[index],
      short_url: item.short_url,
      short_code: item.short_code,
      success: true
    }))

    // 添加失败的URL
    const successUrls = res.data.results.map(r => r.original_url)
    const failedUrls = urls.filter(url => !successUrls.includes(url))

    failedUrls.forEach(url => {
      results.value.push({
        original_url: url,
        short_url: '',
        short_code: '',
        success: false
      })
    })

    ElMessage.success(`批量创建完成：成功 ${successCount.value} 个，失败 ${failedCount.value} 个`)
  } catch (error) {
    console.error('批量创建失败:', error)
    ElMessage.error('批量创建失败')
  } finally {
    loading.value = false
  }
}

// 清空
const handleClear = () => {
  form.value.urls = ''
  results.value = []
}

// 加载示例
const loadExample = () => {
  form.value.urls = `https://www.google.com
https://www.github.com
https://www.stackoverflow.com
https://www.youtube.com
https://www.twitter.com`
}

// 复制链接
const copyLink = (url) => {
  simpleCopy(url, '链接已复制')
}

// 复制所有短链接
const copyAllLinks = () => {
  const links = results.value
      .filter(r => r.success)
      .map(r => r.short_url)
      .join('\n')

  if (links) {
    simpleCopy(links, '所有短链接已复制')
  } else {
    ElMessage.warning('没有可复制的短链接')
  }
}

// 导出结果
const exportResults = () => {
  const csvContent = [
    ['序号', '原始链接', '短链接', '短链码', '状态'].join(','),
    ...results.value.map((row, index) => [
      index + 1,
      `"${row.original_url}"`,
      row.short_url,
      row.short_code,
      row.success ? '成功' : '失败'
    ].join(','))
  ].join('\n')

  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `batch_shortlinks_${Date.now()}.csv`
  link.click()

  ElMessage.success('导出成功')
}
</script>

<style lang="scss" scoped>
.batch-create {
  h3 {
    margin-bottom: 20px;
    color: #303133;
  }

  .form-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;

    .url-count {
      color: #909399;
      font-size: 14px;
    }
  }

  .result-summary {
    display: flex;
    align-items: center;
    gap: 10px;
    padding-bottom: 15px;
    border-bottom: 1px solid #ebeef5;
  }

  .link-cell {
    display: flex;
    align-items: center;
    gap: 10px;

    a {
      color: #409eff;
      text-decoration: none;

      &:hover {
        text-decoration: underline;
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