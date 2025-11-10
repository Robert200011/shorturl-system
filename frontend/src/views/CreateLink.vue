<template>
  <div class="create-link">
    <div class="card">
      <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-width="120px"
          style="max-width: 800px"
      >
        <el-form-item label="原始链接" prop="original_url">
          <el-input
              v-model="form.original_url"
              placeholder="请输入要缩短的网址，如: https://www.example.com"
          />
        </el-form-item>

        <el-form-item label="自定义短码" prop="custom_code">
          <el-input
              v-model="form.custom_code"
              placeholder="可选，留空将自动生成"
          >
            <template #prepend>http://localhost:8002/</template>
          </el-input>
          <div class="form-tip">只能包含字母、数字，长度3-20个字符</div>
        </el-form-item>

        <el-form-item label="链接标题" prop="title">
          <el-input
              v-model="form.title"
              placeholder="可选，便于识别和管理"
          />
        </el-form-item>

        <el-form-item label="链接描述" prop="description">
          <el-input
              v-model="form.description"
              type="textarea"
              :rows="3"
              placeholder="可选，添加详细描述"
          />
        </el-form-item>

        <el-form-item label="过期时间" prop="expire_at">
          <el-date-picker
              v-model="form.expire_at"
              type="datetime"
              placeholder="可选，不设置则永久有效"
              format="YYYY-MM-DD HH:mm:ss"
              value-format="YYYY-MM-DD HH:mm:ss"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="loading">
            生成短链接
          </el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 生成结果 -->
    <div v-if="result" class="card result-card">
      <h3>生成成功！</h3>
      <div class="result-info">
        <div class="info-item">
          <label>短链接：</label>
          <div class="info-value">
            <a :href="result.short_url" target="_blank" class="short-link">
              {{ result.short_url }}
            </a>
            <el-button
                type="primary"
                text
                @click="copyLink(result.short_url)"
            >
              <el-icon><CopyDocument /></el-icon>
              复制
            </el-button>
          </div>
        </div>

        <div class="info-item">
          <label>短链码：</label>
          <div class="info-value">{{ result.short_code }}</div>
        </div>

        <div class="info-item">
          <label>原始链接：</label>
          <div class="info-value">
            <el-tooltip :content="result.original_url" placement="top">
              <span class="original-url">{{ result.original_url }}</span>
            </el-tooltip>
          </div>
        </div>

        <div class="info-item">
          <label>二维码：</label>
          <div class="info-value">
            <canvas ref="qrcodeCanvas" id="qrcode"></canvas>
            <el-button
                type="primary"
                text
                size="small"
                @click="downloadQRCode"
            >
              下载二维码
            </el-button>
          </div>
        </div>
      </div>

      <div class="result-actions">
        <el-button @click="createAnother">创建另一个</el-button>
        <el-button type="primary" @click="viewAnalytics">查看统计</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import QRCode from 'qrcode'
import { api } from '@/api'
import { simpleCopy } from '@/utils'

const router = useRouter()
const formRef = ref()
const qrcodeCanvas = ref()

const form = ref({
  original_url: '',
  custom_code: '',
  title: '',
  description: '',
  expire_at: null
})

const rules = {
  original_url: [
    { required: true, message: '请输入原始链接', trigger: 'blur' },
    { type: 'url', message: '请输入有效的网址', trigger: 'blur' }
  ],
  custom_code: [
    {
      pattern: /^[a-zA-Z0-9]{3,20}$/,
      message: '只能包含字母、数字，长度3-20个字符',
      trigger: 'blur'
    }
  ]
}

const loading = ref(false)
const result = ref(null)

// 提交表单
const handleSubmit = async () => {
  const valid = await formRef.value.validate()
  if (!valid) return

  loading.value = true
  try {
    const data = {
      original_url: form.value.original_url,
      custom_code: form.value.custom_code || undefined,
      title: form.value.title || undefined,
      description: form.value.description || undefined,
      expire_at: form.value.expire_at || undefined
    }

    const res = await api.createShortLink(data)
    result.value = res.data

    ElMessage.success('短链接创建成功！')

    // 生成二维码
    await nextTick()
    await generateQRCode(res.data.short_url)
  } catch (error) {
    console.error('创建失败:', error)
    ElMessage.error(error.response?.data?.message || '创建失败')
  } finally {
    loading.value = false
  }
}

// 重置表单
const handleReset = () => {
  formRef.value.resetFields()
  result.value = null
}

// 复制链接
const copyLink = (url) => {
  simpleCopy(url, '链接已复制到剪贴板')
}

// 生成二维码
const generateQRCode = async (url) => {
  if (!qrcodeCanvas.value) return

  try {
    await QRCode.toCanvas(qrcodeCanvas.value, url, {
      width: 200,
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
  link.download = `qrcode-${result.value.short_code}.png`
  link.href = canvas.toDataURL()
  link.click()
}

// 创建另一个
const createAnother = () => {
  handleReset()
  window.scrollTo(0, 0)
}

// 查看统计
const viewAnalytics = () => {
  router.push({
    path: '/analytics',
    query: { code: result.value.short_code }
  })
}
</script>

<style lang="scss" scoped>
.create-link {
  .form-tip {
    font-size: 12px;
    color: #909399;
    margin-top: 5px;
  }

  .result-card {
    margin-top: 20px;

    h3 {
      color: #67c23a;
      margin-bottom: 20px;
    }

    .result-info {
      .info-item {
        display: flex;
        margin-bottom: 15px;
        align-items: flex-start;

        label {
          width: 100px;
          color: #606266;
          flex-shrink: 0;
        }

        .info-value {
          flex: 1;
          color: #303133;
          word-break: break-all;

          .short-link {
            color: #409eff;
            text-decoration: none;
            font-weight: bold;
            margin-right: 10px;

            &:hover {
              text-decoration: underline;
            }
          }

          .original-url {
            display: inline-block;
            max-width: 500px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
          }

          #qrcode {
            display: block;
            margin: 10px 0;
          }
        }
      }
    }

    .result-actions {
      margin-top: 30px;
      padding-top: 20px;
      border-top: 1px solid #ebeef5;
    }
  }
}
</style>