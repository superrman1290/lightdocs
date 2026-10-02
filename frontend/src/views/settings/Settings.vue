<template>
  <div class="settings-page">
    <div class="page-header">
      <div>
        <h2>设置管理</h2>
        <p>管理网站的基本配置和系统设置。</p>
      </div>
    </div>

    <el-tabs
      v-model="activeTab"
      class="settings-tabs"
    >
      <el-tab-pane
        label="通用设置"
        name="general"
      >
        <el-card
          class="settings-card"
          shadow="never"
        >
          <template #header>
            <div class="card-title">
              <el-icon><Setting /></el-icon>
              <span>网站基本信息</span>
            </div>
          </template>

          <el-form
            label-position="left"
            label-width="120px"
          >
            <el-form-item label="网站名称">
              <el-input
                v-model="form.siteName"
                maxlength="30"
                placeholder="请输入网站名称"
              />
            </el-form-item>

            <el-form-item label="网站标题">
              <el-input
                v-model="form.siteTitle"
                maxlength="80"
                placeholder="请输入浏览器标签页标题"
              />
            </el-form-item>

            <el-form-item label="网站 Logo / Favicon">
              <div class="logo-upload-row">
                <div class="logo-preview">
                  <img
                    :src="form.logoUrl"
                    alt="网站 Logo"
                    @error="handleLogoError"
                  />
                </div>

                <input
                  ref="fileInput"
                  class="hidden-input"
                  type="file"
                  accept="image/png,image/jpeg,image/webp,image/svg+xml"
                  @change="handleLogoChange"
                />

                <div class="upload-actions">
                  <el-button
                    :icon="Upload"
                    @click="triggerUpload"
                  >
                    选择图片
                  </el-button>
                  <span class="upload-tip">
                    同一张图片用于网站 Logo 和浏览器图标，建议 128 × 128，支持 JPG、PNG、WEBP、SVG
                  </span>
                </div>
              </div>
            </el-form-item>

            <div class="form-footer">
              <el-button
                type="primary"
                :loading="saving"
                @click="handleSave"
              >
                保存设置
              </el-button>
            </div>
          </el-form>
        </el-card>
      </el-tab-pane>

      <el-tab-pane
        label="安全"
        name="security"
      >
        <el-card
          class="settings-card security-card"
          shadow="never"
        >
          <template #header>
            <div class="card-title">
              <el-icon><Lock /></el-icon>
              <span>管理员登录</span>
            </div>
          </template>

          <el-form
            label-position="left"
            label-width="150px"
          >
            <div class="security-section">
              <div class="security-heading">登录失败限制</div>

              <el-form-item label="登录失败次数">
                <el-input-number
                  v-model="securityForm.maxLoginFailures"
                  :min="1"
                  :max="20"
                  controls-position="right"
                  class="security-number"
                />
                <span class="unit">次</span>
              </el-form-item>

              <el-form-item label="锁定时间">
                <el-input-number
                  v-model="securityForm.lockMinutes"
                  :min="1"
                  :max="1440"
                  controls-position="right"
                  class="security-number"
                />
                <span class="unit">分钟</span>
              </el-form-item>
            </div>

            <div class="security-section">
              <div class="security-heading">会话安全</div>

              <el-form-item label="登录有效期">
                <el-input-number
                  v-model="securityForm.sessionDays"
                  :min="1"
                  :max="365"
                  controls-position="right"
                  class="security-number"
                />
                <span class="unit">天</span>
              </el-form-item>

              <el-form-item label="">
                <el-checkbox v-model="securityForm.rememberLogin">
                  登录后保持登录
                </el-checkbox>
              </el-form-item>
            </div>

            <div class="security-section">
              <div class="security-heading">管理员操作</div>

              <el-form-item label="">
                <el-checkbox
                  v-model="securityForm.logoutOtherDevicesOnPasswordChange"
                >
                  修改密码后注销其他设备
                </el-checkbox>
              </el-form-item>

              <el-form-item label="">
                <el-checkbox
                  v-model="securityForm.requirePasswordReauthOnSave"
                >
                  保存设置时要求重新验证密码
                </el-checkbox>
              </el-form-item>
            </div>

            <div class="security-section danger-section">
              <div class="security-heading">危险操作</div>

              <div class="danger-row">
                <span>清除所有登录会话</span>
                <el-button
                  type="danger"
                  plain
                  :icon="Delete"
                  @click="handleClearSessions"
                >
                  清除会话
                </el-button>
              </div>
            </div>

            <div class="form-footer">
              <el-button @click="restoreSecurityDefaults">
                恢复默认
              </el-button>
              <el-button
                type="primary"
                :loading="saving"
                @click="handleSaveSecurity"
              >
                保存设置
              </el-button>
            </div>
          </el-form>
        </el-card>
      </el-tab-pane>

      <el-tab-pane
        label="管理员"
        name="admin"
      >
        <el-card
          class="settings-card admin-card"
          shadow="never"
        >
          <template #header>
            <div class="card-title">
              <el-icon><User /></el-icon>
              <span>管理员账号</span>
            </div>
          </template>

          <div class="admin-summary">
            <div class="admin-summary-row">
              <span>管理员账号</span>
              <strong>{{ adminSettings.username }}</strong>
            </div>
            <div class="admin-summary-row">
              <span>登录密码</span>
              <strong>••••{{ passwordLastFour }}</strong>
            </div>
          </div>

          <div class="security-section admin-edit-section">
            <div class="security-heading">修改账号密码</div>

            <el-form
              label-position="left"
              label-width="120px"
            >
              <el-form-item label="新的账号">
                <el-input
                  v-model="adminForm.username"
                  maxlength="30"
                  autocomplete="username"
                  placeholder="请输入新的管理员账号"
                />
              </el-form-item>

              <el-form-item label="当前密码">
                <el-input
                  v-model="adminForm.currentPassword"
                  type="password"
                  show-password
                  autocomplete="current-password"
                  placeholder="请输入当前密码"
                />
              </el-form-item>

              <el-form-item label="新密码">
                <el-input
                  v-model="adminForm.newPassword"
                  type="password"
                  show-password
                  autocomplete="new-password"
                  placeholder="至少 8 位字符"
                />
              </el-form-item>

              <el-form-item label="确认新密码">
                <el-input
                  v-model="adminForm.confirmPassword"
                  type="password"
                  show-password
                  autocomplete="new-password"
                  placeholder="再次输入新密码"
                />
              </el-form-item>
            </el-form>
          </div>

          <div class="form-footer">
            <el-button
              type="primary"
              :loading="saving"
              @click="handleSaveAdmin"
            >
              保存设置
            </el-button>
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  onMounted,
  reactive,
  ref,
  watch,
} from 'vue'

import { useRoute } from 'vue-router'

import {
  ElMessage,
} from 'element-plus'

import {
  Delete,
  Lock,
  Setting,
  User,
  Upload,
} from '@element-plus/icons-vue'

import fallbackLogo from '../../assets/hero.png'

import {
  defaultSecuritySettings,
  loadSiteSettings,
  saveSiteSettings,
  siteSettings,
} from '../../stores/siteSettings'

import { request } from '../../services/apiClient'

import type {
  AdminSettings,
  SecuritySettings,
} from '../../stores/siteSettings'

const activeTab = ref('general')
const route = useRoute()
const saving = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const form = reactive({
  siteName: siteSettings.siteName,
  siteTitle: siteSettings.siteTitle,
  logoUrl: siteSettings.logoUrl,
})

const securityForm = reactive<SecuritySettings>({
  ...siteSettings.security,
})

const adminSettings = siteSettings.admin

const adminForm = reactive<AdminSettings & {
  currentPassword: string
  newPassword: string
  confirmPassword: string
}>({
  username: siteSettings.admin.username,
  password: siteSettings.admin.password,
  currentPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const passwordLastFour = computed(() => {
  return adminSettings.password.slice(-4)
})

const triggerUpload = () => {
  fileInput.value?.click()
}

const handleLogoChange = (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]

  if (!file) {
    return
  }

  const reader = new FileReader()

  reader.onload = () => {
    if (typeof reader.result === 'string') {
      form.logoUrl = reader.result
    }
  }

  reader.readAsDataURL(file)
  input.value = ''
}

const handleLogoError = (event: Event) => {
  const image = event.target as HTMLImageElement

  if (image.src !== fallbackLogo) {
    image.src = fallbackLogo
  }
}

const handleSave = async () => {
  if (!form.siteName.trim() || !form.siteTitle.trim()) {
    ElMessage.warning('网站名称和网站标题不能为空')
    return
  }

  saving.value = true

  try {
    saveSiteSettings({
      siteName: form.siteName.trim(),
      siteTitle: form.siteTitle.trim(),
      logoUrl: form.logoUrl,
      security: {
        ...securityForm,
      },
      admin: {
        ...adminSettings,
      },
    })

    ElMessage.success('设置已保存')
  } catch (error) {
    console.error('保存网站设置失败：', error)
    ElMessage.error('保存网站设置失败')
  } finally {
    saving.value = false
  }
}

const handleSaveSecurity = () => {
  saving.value = true

  try {
    void request('/settings/security', {
      method: 'PATCH',
      body: JSON.stringify(securityForm),
    })

    saveSiteSettings({
      siteName: siteSettings.siteName,
      siteTitle: siteSettings.siteTitle,
      logoUrl: siteSettings.logoUrl,
      security: {
        ...securityForm,
      },
      admin: {
        ...adminSettings,
      },
    })

    ElMessage.success('安全设置已保存')
  } catch (error) {
    console.error('保存安全设置失败：', error)
    ElMessage.error('保存安全设置失败')
  } finally {
    saving.value = false
  }
}

const restoreSecurityDefaults = () => {
  Object.assign(
    securityForm,
    defaultSecuritySettings,
  )

  ElMessage.success('已恢复默认安全设置，请保存后生效')
}

const handleSaveAdmin = () => {
  const username = adminForm.username.trim()

  if (!username) {
    ElMessage.warning('管理员账号不能为空')
    return
  }

  if (adminForm.currentPassword !== adminSettings.password) {
    ElMessage.error('当前密码不正确')
    return
  }

  if (adminForm.newPassword.length < 8) {
    ElMessage.warning('新密码至少需要 8 位字符')
    return
  }

  if (adminForm.newPassword !== adminForm.confirmPassword) {
    ElMessage.error('两次输入的新密码不一致')
    return
  }

  saving.value = true

  try {
    const nextAdmin = {
      username,
      password: adminForm.newPassword,
    }

    void request('/settings/admin', {
      method: 'PATCH',
      body: JSON.stringify({
        username,
        currentPassword: adminForm.currentPassword,
        newPassword: adminForm.newPassword,
        confirmPassword: adminForm.confirmPassword,
      }),
    })

    saveSiteSettings({
      siteName: siteSettings.siteName,
      siteTitle: siteSettings.siteTitle,
      logoUrl: siteSettings.logoUrl,
      security: {
        ...securityForm,
      },
      admin: nextAdmin,
    })

    adminForm.currentPassword = ''
    adminForm.newPassword = ''
    adminForm.confirmPassword = ''
    ElMessage.success('管理员账号密码已更新')
  } catch (error) {
    console.error('保存管理员设置失败：', error)
    ElMessage.error('保存管理员设置失败')
  } finally {
    saving.value = false
  }
}

const handleClearSessions = () => {
  if (typeof window !== 'undefined') {
    window.localStorage.removeItem('lightdocs-session')
  }

  ElMessage.success('登录会话已清除')
}

onMounted(() => {
  loadSiteSettings()
  if (route.query.tab === 'admin') {
    activeTab.value = 'admin'
  }

  form.siteName = siteSettings.siteName
  form.siteTitle = siteSettings.siteTitle
  form.logoUrl = siteSettings.logoUrl
  Object.assign(securityForm, siteSettings.security)
  adminForm.username = siteSettings.admin.username
  adminForm.password = siteSettings.admin.password
})

watch(
  () => route.query.tab,
  tab => {
    if (tab === 'admin') {
      activeTab.value = 'admin'
    }
  },
)
</script>

<style scoped>
.settings-page {
  width: 100%;
  max-width: 960px;
}

.page-header {
  margin-bottom: 18px;
}

.page-header h2 {
  margin: 0 0 6px;
  color: #1f2d5a;
  font-size: 24px;
  font-weight: 600;
}

.page-header p {
  margin: 0;
  color: #8b96ad;
  font-size: 13px;
}

.settings-tabs {
  width: 100%;
}

.settings-card {
  border: 1px solid #e7edf7;
  border-radius: 8px;
}

.security-card {
  max-width: 820px;
}

.admin-card {
  max-width: 820px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1d376a;
  font-size: 14px;
  font-weight: 600;
}

.card-title .el-icon {
  color: #2784f2;
  font-size: 18px;
}

.settings-card :deep(.el-input) {
  max-width: 680px;
}

.logo-upload-row {
  display: flex;
  align-items: center;
  gap: 14px;
}

.logo-preview {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  flex: 0 0 64px;
  overflow: hidden;
  border: 1px solid #dbe7f6;
  border-radius: 10px;
  background: #f4f8ff;
}

.logo-preview img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.upload-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.upload-tip {
  color: #8b96ad;
  font-size: 12px;
  line-height: 1.5;
}

.hidden-input {
  display: none;
}

.security-section {
  padding: 4px 0 16px;
  border-bottom: 1px solid #eef2f8;
}

.security-section + .security-section {
  padding-top: 20px;
}

.security-heading {
  margin-bottom: 16px;
  color: #294474;
  font-size: 14px;
  font-weight: 600;
}

.admin-summary {
  margin-bottom: 22px;
  border: 1px solid #e5edf8;
  border-radius: 6px;
  background: #f8fbff;
}

.admin-summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 14px 16px;
  color: #6d80a4;
  font-size: 13px;
}

.admin-summary-row + .admin-summary-row {
  border-top: 1px solid #e5edf8;
}

.admin-summary-row strong {
  color: #294474;
  font-weight: 600;
  letter-spacing: 0.08em;
}

.admin-edit-section {
  border-bottom: 0;
}

.security-number {
  width: 150px;
}

.unit {
  margin-left: 8px;
  color: #7d8eaf;
  font-size: 13px;
}

.danger-section {
  border-bottom: 0;
}

.danger-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  color: #53698f;
  font-size: 13px;
}

.form-footer {
  display: flex;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid #eef2f8;
}

@media (max-width: 680px) {
  .logo-upload-row,
  .upload-actions {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
