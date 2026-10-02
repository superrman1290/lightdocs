<template>
  <div
    class="admin-layout"
    :class="{ 'dark-mode': isDarkMode }"
  >
    <!-- 左侧菜单 -->
    <aside class="sidebar">
      <!-- Logo -->
      <div class="logo">
        <span class="logo-icon">
          <img
            v-if="siteSettings.logoUrl"
            :src="siteSettings.logoUrl"
            :alt="siteSettings.siteName"
          />
          <span v-else>L</span>
        </span>
        <span class="logo-text">{{ siteSettings.siteName }}</span>
      </div>

      <!-- 菜单 -->
      <el-menu
        :default-active="currentRoute"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/admin">
          <el-icon>
            <House />
          </el-icon>
          <span>仪表盘</span>
        </el-menu-item>

        <el-menu-item index="/admin/articles">
          <el-icon>
            <Document />
          </el-icon>
          <span>文章</span>
        </el-menu-item>

        <el-menu-item index="/admin/categories">
          <el-icon>
            <Folder />
          </el-icon>
          <span>分类</span>
        </el-menu-item>

        <el-menu-item index="/admin/images">
          <el-icon>
            <Picture />
          </el-icon>
          <span>图片</span>
        </el-menu-item>

        <el-menu-item index="/admin/recycle-bin">
          <el-icon>
            <Delete />
          </el-icon>
          <span>回收站</span>
        </el-menu-item>

        <el-menu-item index="/admin/settings">
          <el-icon>
            <Setting />
          </el-icon>
          <span>设置</span>
        </el-menu-item>

        <el-menu-item index="/docs">
          <el-icon>
            <Reading />
          </el-icon>
          <span>用户页面</span>
        </el-menu-item>
      </el-menu>
    </aside>

    <!-- 右侧区域 -->
    <div class="main-container">

      <!-- 顶部栏 -->
      <header class="header">

        <!-- 左侧 -->
        <div class="header-left">
          <span class="page-title">
            {{ pageTitle }}
          </span>
        </div>

        <!-- 右侧 -->
        <div class="header-right">

          <!-- 深色模式 -->
          <el-button
            text
            circle
            :title="isDarkMode ? '切换亮色模式' : '切换暗色模式'"
            @click="toggleDarkMode"
          >
            <el-icon :size="20">
              <Moon v-if="isDarkMode" />
              <Sunny v-else />
            </el-icon>
          </el-button>

          <!-- 管理员 -->
          <el-dropdown>
            <span class="admin-user">
              <el-avatar :size="32">
                A
              </el-avatar>

              <span class="username">
                {{ siteSettings.admin.username }}
              </span>

              <el-icon>
                <ArrowDown />
              </el-icon>
            </span>

            <template #dropdown>
              <el-dropdown-menu>
              <el-dropdown-item @click="handleProfile">
                个人资料
              </el-dropdown-item>

              <el-dropdown-item divided>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

        </div>
      </header>

      <!-- 页面内容 -->
      <main class="content">
        <router-view />
      </main>

    </div>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  onMounted,
  ref,
} from 'vue'
import {
  useRoute,
  useRouter,
} from 'vue-router'

import {
  House,
  Document,
  Folder,
  Picture,
  Delete,
  Setting,
  Reading,
  Moon,
  Sunny,
  ArrowDown,
} from '@element-plus/icons-vue'

import {
  loadSiteSettings,
  siteSettings,
} from '../stores/siteSettings'

const route = useRoute()
const router = useRouter()
const isDarkMode = ref(false)

onMounted(() => {
  loadSiteSettings()
  isDarkMode.value = document.documentElement.classList.contains('dark-mode')
})

/**
 * 当前路由
 */
const currentRoute = computed(() => {
  return route.path
})

/**
 * 页面标题
 */
const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    '/admin': '仪表盘',
    '/admin/articles': '文章',
    '/admin/categories': '分类',
    '/admin/images': '图片',
    '/admin/recycle-bin': '回收站',
    '/admin/settings': '设置',
  }

  return titles[route.path] || 'LightDocs'
})

/**
 * 深色模式
 */
const toggleDarkMode = () => {
  isDarkMode.value = !isDarkMode.value
  document.documentElement.classList.toggle(
    'dark-mode',
    isDarkMode.value,
  )
}

const handleProfile = () => {
  router.push({
    path: '/admin/settings',
    query: { tab: 'admin' },
  })
}
</script>

<style scoped>
.admin-layout {
  display: flex;
  width: 100%;
  height: 100vh;
  overflow: hidden;
  background: #f5f7fa;
}

/* ==================== */
/* 左侧菜单 */
/* ==================== */

.sidebar {
  width: 220px;
  height: 100vh;
  flex-shrink: 0;
  background: #ffffff;
  border-right: 1px solid #e5e7eb;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  border-bottom: 1px solid #f0f0f0;
}

.logo-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 10px;
  border-radius: 8px;
  background: #409eff;
  color: #ffffff;
  font-weight: 600;
  font-size: 18px;
}

.logo-icon img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.sidebar-menu {
  border-right: none;
  padding: 12px 8px;
}

.sidebar-menu .el-menu-item {
  height: 44px;
  line-height: 44px;
  margin-bottom: 4px;
  border-radius: 6px;
}

.sidebar-menu .el-menu-item.is-active {
  background: #ecf5ff;
  color: #409eff;
}

/* ==================== */
/* 右侧 */
/* ==================== */

.main-container {
  flex: 1;
  min-width: 0;
  height: 100vh;
  display: flex;
  flex-direction: column;
}

/* ==================== */
/* 顶部栏 */
/* ==================== */

.header {
  height: 64px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  background: #ffffff;
  border-bottom: 1px solid #e5e7eb;
}

.page-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 4px;
}

.admin-user {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  cursor: pointer;
  border-radius: 6px;
}

.admin-user:hover {
  background: #f5f7fa;
}

.username {
  font-size: 14px;
  color: #606266;
}

/* ==================== */
/* 页面内容 */
/* ==================== */

.content {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 24px;
}
</style>
