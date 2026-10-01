<template>
  <div class="dashboard">

    <!-- ==================== -->
    <!-- 欢迎区域 -->
    <!-- ==================== -->

    <div class="welcome-section">
      <div>
        <h2>欢迎回来，管理员</h2>
        <p>这是 LightDocs 的内容管理概览。</p>
      </div>
    </div>


    <!-- ==================== -->
    <!-- 数据统计 -->
    <!-- ==================== -->

    <el-row :gutter="20" class="statistics-row">

      <!-- 文章 -->
      <el-col :xs="24" :sm="12" :md="8">
        <el-card shadow="never" class="stat-card">
          <div class="stat-content">

            <div class="stat-icon article-icon">
              <el-icon :size="24">
                <Document />
              </el-icon>
            </div>

            <div class="stat-info">
              <div class="stat-label">
                文章
              </div>

              <div class="stat-number">
                {{ statistics.articles }}
              </div>
            </div>

          </div>
        </el-card>
      </el-col>


      <!-- 分类 -->
      <el-col :xs="24" :sm="12" :md="8">
        <el-card shadow="never" class="stat-card">
          <div class="stat-content">

            <div class="stat-icon category-icon">
              <el-icon :size="24">
                <Folder />
              </el-icon>
            </div>

            <div class="stat-info">
              <div class="stat-label">
                分类
              </div>

              <div class="stat-number">
                {{ statistics.categories }}
              </div>
            </div>

          </div>
        </el-card>
      </el-col>


      <!-- 图片 -->
      <el-col :xs="24" :sm="12" :md="8">
        <el-card shadow="never" class="stat-card">
          <div class="stat-content">

            <div class="stat-icon image-icon">
              <el-icon :size="24">
                <Picture />
              </el-icon>
            </div>

            <div class="stat-info">
              <div class="stat-label">
                图片
              </div>

              <div class="stat-number">
                {{ statistics.images }}
              </div>
            </div>

          </div>
        </el-card>
      </el-col>

    </el-row>


    <!-- ==================== -->
    <!-- 快捷操作 -->
    <!-- ==================== -->

    <el-card
      shadow="never"
      class="section-card"
    >

      <template #header>
        <div class="section-header">
          <span>快捷操作</span>
        </div>
      </template>

      <div class="quick-actions">

        <!-- 新建文章 -->
        <el-button
          type="primary"
          class="quick-action"
          @click="createArticle"
        >
          <el-icon>
            <Edit />
          </el-icon>

          <span>新建文章</span>
        </el-button>


        <!-- 新建分类 -->
        <el-button
          class="quick-action"
          @click="createCategory"
        >
          <el-icon>
            <FolderAdd />
          </el-icon>

          <span>新建分类</span>
        </el-button>


        <!-- 上传图片 -->
        <el-button
          class="quick-action"
          @click="uploadImage"
        >
          <el-icon>
            <Upload />
          </el-icon>

          <span>上传图片</span>
        </el-button>

      </div>

    </el-card>


    <!-- ==================== -->
    <!-- 最新文章 -->
    <!-- ==================== -->

    <el-card
      shadow="never"
      class="section-card"
    >

      <template #header>

        <div class="section-header">
          <span>最新文章</span>

          <el-button
            link
            type="primary"
            @click="viewAllArticles"
          >
            查看全部
          </el-button>
        </div>

      </template>


      <el-table
        :data="latestArticles"
        style="width: 100%"
        empty-text="暂无文章"
      >

        <!-- 标题 -->
        <el-table-column
          prop="title"
          label="标题"
          min-width="240"
        >
          <template #default="{ row }">

            <span class="article-title">
              {{ row.title }}
            </span>

          </template>
        </el-table-column>


        <!-- 分类 -->
        <el-table-column
          prop="category"
          label="分类"
          width="140"
        />


        <!-- 状态 -->
        <el-table-column
          label="状态"
          width="100"
        >
          <template #default="{ row }">

            <el-tag
              v-if="row.status === 'published'"
              type="success"
              size="small"
            >
              已发布
            </el-tag>

            <el-tag
              v-else
              type="warning"
              size="small"
            >
              草稿
            </el-tag>

          </template>
        </el-table-column>


        <!-- 标签 -->
        <el-table-column
          label="标签"
          min-width="180"
        >
          <template #default="{ row }">

            <el-tag
              v-for="tag in row.tags"
              :key="tag"
              size="small"
              effect="plain"
              class="article-tag"
            >
              {{ tag }}
            </el-tag>

          </template>
        </el-table-column>


        <!-- 更新时间 -->
        <el-table-column
          prop="updatedAt"
          label="更新时间"
          width="180"
        />


        <!-- 操作 -->
        <el-table-column
          label="操作"
          width="100"
          fixed="right"
        >

          <template #default="{ row }">

            <el-button
              link
              type="primary"
              @click="editArticle(row)"
            >
              编辑
            </el-button>

          </template>

        </el-table-column>

      </el-table>

    </el-card>

  </div>
</template>


<script setup lang="ts">

import { useRouter } from 'vue-router'

import {
  Document,
  Folder,
  Picture,
  Edit,
  FolderAdd,
  Upload,
} from '@element-plus/icons-vue'


const router = useRouter()


/**
 * Dashboard 统计数据
 *
 * 目前使用 Mock 数据
 */
const statistics = {
  articles: 24,
  categories: 8,
  images: 126,
}


/**
 * 最新文章 Mock 数据
 */
const latestArticles = [
  {
    id: 1,
    title: 'Vue 3 项目搭建指南',
    category: '前端开发',
    status: 'published',
    tags: ['Vue', '前端'],
    updatedAt: '2026-09-30 10:32',
  },

  {
    id: 2,
    title: 'Docker Compose 入门',
    category: 'Docker',
    status: 'published',
    tags: ['Docker', '部署'],
    updatedAt: '2026-09-29 16:20',
  },

  {
    id: 3,
    title: 'Go Gin Web 开发',
    category: '后端开发',
    status: 'draft',
    tags: ['Go', 'Gin'],
    updatedAt: '2026-09-28 14:15',
  },

  {
    id: 4,
    title: 'PostgreSQL 基础使用',
    category: '数据库',
    status: 'published',
    tags: ['PostgreSQL', '数据库'],
    updatedAt: '2026-09-27 11:08',
  },

  {
    id: 5,
    title: 'LightDocs 项目架构设计',
    category: '项目开发',
    status: 'published',
    tags: ['LightDocs', '架构'],
    updatedAt: '2026-09-26 18:42',
  },
]


/**
 * 新建文章
 */
const createArticle = () => {
  router.push('/admin/articles/new')
}


/**
 * 新建分类
 */
const createCategory = () => {
  router.push('/admin/categories')
}


/**
 * 上传图片
 */
const uploadImage = () => {
  router.push('/admin/images')
}


/**
 * 查看全部文章
 */
const viewAllArticles = () => {
  router.push('/admin/articles')
}


/**
 * 编辑文章
 */
const editArticle = (article: any) => {
  router.push(`/admin/articles/${article.id}/edit`)
}

</script>


<style scoped>

/* =========================
   Dashboard
   ========================= */

.dashboard {
  width: 100%;
}


/* =========================
   欢迎区域
   ========================= */

.welcome-section {
  margin-bottom: 20px;
}

.welcome-section h2 {
  margin: 0 0 6px;
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}

.welcome-section p {
  margin: 0;
  font-size: 14px;
  color: #909399;
}


/* =========================
   统计卡片
   ========================= */

.statistics-row {
  margin-bottom: 20px;
}

.stat-card {
  border: 1px solid #ebeef5;
}

.stat-content {
  display: flex;
  align-items: center;
  min-height: 90px;
}

.stat-icon {
  width: 52px;
  height: 52px;

  display: flex;
  align-items: center;
  justify-content: center;

  margin-right: 16px;

  border-radius: 10px;
}


/* 文章 */

.article-icon {
  color: #409eff;
  background: #ecf5ff;
}


/* 分类 */

.category-icon {
  color: #67c23a;
  background: #f0f9eb;
}


/* 图片 */

.image-icon {
  color: #e6a23c;
  background: #fdf6ec;
}


.stat-label {
  margin-bottom: 6px;

  font-size: 14px;
  color: #909399;
}


.stat-number {
  font-size: 28px;
  line-height: 1;

  font-weight: 600;
  color: #303133;
}


/* =========================
   通用区块
   ========================= */

.section-card {
  margin-bottom: 20px;
  border: 1px solid #ebeef5;
}


.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;

  font-size: 16px;
  font-weight: 600;
  color: #303133;
}


/* =========================
   快捷操作
   ========================= */

.quick-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}


.quick-action {
  height: 40px;
  padding: 0 18px;
}


/* =========================
   最新文章
   ========================= */

.article-title {
  color: #303133;
  font-weight: 500;
}


.article-tag {
  margin-right: 6px;
}


/* =========================
   响应式
   ========================= */

@media (max-width: 768px) {

  .welcome-section h2 {
    font-size: 20px;
  }

  .stat-content {
    min-height: 70px;
  }

  .quick-action {
    flex: 1;
    min-width: 120px;
  }

}

</style>