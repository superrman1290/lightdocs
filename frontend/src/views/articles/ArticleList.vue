<template>
  <div class="article-list">

    <!-- =========================
         页面标题
    ========================== -->
    <div class="page-header">

      <div class="page-title">
        <h2>文章</h2>
        <p>管理所有文章内容</p>
      </div>

      <el-button
        type="primary"
        :icon="Plus"
        @click="handleCreate"
      >
        新建文章
      </el-button>

    </div>


    <!-- =========================
         筛选区域
    ========================== -->
    <el-card
      class="filter-card"
      shadow="never"
    >

      <!-- 状态 Tabs -->
      <el-tabs
        v-model="activeStatus"
        @tab-change="handleStatusChange"
      >

        <el-tab-pane
          label="全部"
          name="all"
        />

        <el-tab-pane
          label="已发布"
          name="published"
        />

        <el-tab-pane
          label="草稿"
          name="draft"
        />

      </el-tabs>


      <!-- 搜索 / 分类 -->
      <div class="filter-row">

        <!-- 搜索文章 -->
        <el-input
          v-model="searchKeyword"
          class="search-input"
          placeholder="搜索文章标题或标签"
          clearable
          @keyup.enter="handleSearch"
        >

          <template #prefix>
            <el-icon>
              <Search />
            </el-icon>
          </template>

        </el-input>


        <!-- 分类 -->
        <el-select
          v-model="selectedCategory"
          class="category-select"
          placeholder="选择分类"
          clearable
        >

          <el-option
            v-for="category in categories"
            :key="category.id"
            :label="category.name"
            :value="category.id"
          />

        </el-select>


        <!-- 搜索 -->
        <el-button
          type="primary"
          @click="handleSearch"
        >
          搜索
        </el-button>


        <!-- 重置 -->
        <el-button
          @click="resetFilters"
        >
          重置
        </el-button>

      </div>

    </el-card>


    <!-- =========================
         文章列表
    ========================== -->
    <el-card
      class="table-card"
      shadow="never"
    >

      <el-table
        v-loading="loading"
        :data="articles"
        style="width: 100%"
        empty-text="暂无文章"
      >

        <!-- 标题 -->
        <el-table-column
          label="标题"
          min-width="280"
        >

          <template #default="{ row }">

            <div
              class="article-title"
              @click="handleEdit(row)"
            >
              {{ row.title }}
            </div>

          </template>

        </el-table-column>


        <!-- 分类 -->
        <el-table-column
          prop="category"
          label="分类"
          width="150"
        />


        <!-- 状态 -->
        <el-table-column
          label="状态"
          width="100"
        >

          <template #default="{ row }">

            <el-select
              :model-value="row.status"
              size="small"
              class="status-select"
              :loading="updatingStatusIds.has(row.id)"
              @click.stop
              @change="handleArticleStatusChange(row, $event)"
            >
              <el-option
                label="已发布"
                value="published"
              />
              <el-option
                label="草稿"
                value="draft"
              />
            </el-select>

          </template>

        </el-table-column>


        <!-- 标签 -->
        <el-table-column
          label="标签"
          min-width="200"
        >

          <template #default="{ row }">

            <div class="tag-list">

              <el-tag
                v-for="tag in row.tags"
                :key="tag"
                size="small"
                effect="plain"
              >
                {{ tag }}
              </el-tag>

            </div>

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
          width="150"
          fixed="right"
        >

          <template #default="{ row }">

            <el-dropdown>

              <el-button
                link
                type="primary"
              >
                操作

                <el-icon class="el-icon--right">
                  <ArrowDown />
                </el-icon>

              </el-button>


              <template #dropdown>

                <el-dropdown-menu>

                  <!-- 编辑 -->
                  <el-dropdown-item
                    @click="handleEdit(row)"
                  >
                    编辑
                  </el-dropdown-item>


                  <!-- 导出 -->
                  <el-dropdown-item
                    @click="handleExport(row)"
                  >
                    导出
                  </el-dropdown-item>


                  <!-- 删除 -->
                  <el-dropdown-item
                    divided
                    @click="handleDelete(row)"
                  >
                    删除
                  </el-dropdown-item>

                </el-dropdown-menu>

              </template>

            </el-dropdown>

          </template>

        </el-table-column>

      </el-table>


      <!-- =========================
           分页
      ========================== -->
      <div class="pagination">

        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />

      </div>

    </el-card>

  </div>
</template>


<script setup lang="ts">

/**
 * Vue
 */
import {
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from 'vue'


/**
 * Vue Router
 */
import {
  useRouter,
} from 'vue-router'


/**
 * Element Plus 图标
 */
import {
  Plus,
  Search,
  ArrowDown,
} from '@element-plus/icons-vue'


/**
 * Element Plus 消息 / 弹窗
 */
import {
  ElMessage,
  ElMessageBox,
} from 'element-plus'


/**
 * Article Service
 */
import {
  articleService,
} from '../../services/articleService'

import {
  categoryService,
} from '../../services/categoryService'


/**
 * Article Type
 */
import type {
  Article,
  ArticleStatus,
} from '../../types/article'

import type {
  Category,
} from '../../types/category'


/* =====================================================
   Router
===================================================== */

const router = useRouter()


/* =====================================================
   页面状态
===================================================== */

/**
 * 当前状态筛选
 *
 * all       = 全部
 * published = 已发布
 * draft     = 草稿
 */
const activeStatus = ref('all')


/**
 * 搜索关键词
 */
const searchKeyword = ref('')
let searchTimer: ReturnType<typeof setTimeout> | undefined

watch(searchKeyword, () => {
  if (searchTimer) {
    clearTimeout(searchTimer)
  }

  searchTimer = setTimeout(() => {
    void handleSearch()
  }, 500)
})


/**
 * 当前选择的分类
 */
const selectedCategory = ref<number | ''>('')


/**
 * 文章列表
 */
const articles = ref<Article[]>([])


/**
 * 总文章数量
 */
const total = ref(0)


/**
 * 加载状态
 */
const loading = ref(false)


/**
 * 当前页
 */
const currentPage = ref(1)


/**
 * 每页数量
 */
const pageSize = ref(10)
const updatingStatusIds = ref(new Set<number>())


/* =====================================================
   分类
===================================================== */

const categories = ref<Category[]>([])

const loadCategories = async () => {
  try {
    categories.value = await categoryService.getCategories()
  } catch (error) {
    console.error(
      '获取分类列表失败：',
      error,
    )

    ElMessage.error(
      '获取分类列表失败',
    )
  }
}


/* =====================================================
   获取文章列表
===================================================== */

/**
 * 从 Article Service 获取文章。
 *
 * 当前：
 *
 * ArticleList
 *     ↓
 * articleService
 *     ↓
 * API
 *
 * 未来：
 *
 * ArticleList
 *     ↓
 * articleService
 *     ↓
 * HTTP API
 *     ↓
 * Go
 *     ↓
 * PostgreSQL
 */
const loadArticles = async () => {

  loading.value = true

  try {

    const result =
      await articleService.getArticles({

        /**
         * 状态
         */
        status:
          activeStatus.value === 'all'
            ? undefined
            : activeStatus.value as
                'published' | 'draft',


        /**
         * 搜索关键词
         */
        keyword:
          searchKeyword.value.trim(),


        /**
         * 分类
         */
        categoryId:
          selectedCategory.value === ''
            ? undefined
            : selectedCategory.value,


        /**
         * 当前页
         */
        page:
          currentPage.value,


        /**
         * 每页数量
         */
        pageSize:
          pageSize.value,

      })


    /**
     * 更新列表
     */
    articles.value =
      result.list


    /**
     * 更新总数量
     */
    total.value =
      result.total

  } catch (error) {

    console.error(
      '获取文章列表失败：',
      error,
    )

    ElMessage.error(
      '获取文章列表失败',
    )

  } finally {

    loading.value = false

  }

}


/* =====================================================
   状态 Tab 切换
===================================================== */

const handleStatusChange = async () => {

  /**
   * 切换状态后回到第一页
   */
  currentPage.value = 1

  await loadArticles()

}

const handleArticleStatusChange = async (
  article: Article,
  status: ArticleStatus,
) => {
  const previousStatus = article.status
  if (status === previousStatus) return

  updatingStatusIds.value.add(article.id)
  try {
    const updated = await articleService.updateArticle(article.id, { status })
    Object.assign(article, updated)
    ElMessage.success(status === 'published' ? '文章已发布' : '文章已设为草稿')
  } catch (error) {
    article.status = previousStatus
    console.error('更新文章状态失败：', error)
    ElMessage.error(error instanceof Error && error.message ? error.message : '更新文章状态失败')
  } finally {
    updatingStatusIds.value.delete(article.id)
  }
}


/* =====================================================
   搜索
===================================================== */

const handleSearch = async () => {

  if (searchTimer) {
    clearTimeout(searchTimer)
    searchTimer = undefined
  }

  /**
   * 搜索时回到第一页
   */
  currentPage.value = 1

  await loadArticles()

}


/* =====================================================
   重置筛选
===================================================== */

const resetFilters = async () => {

  /**
   * 清空搜索
   */
  searchKeyword.value = ''


  /**
   * 清空分类
   */
  selectedCategory.value = ''


  /**
   * 恢复全部状态
   */
  activeStatus.value = 'all'


  /**
   * 回到第一页
   */
  currentPage.value = 1


  /**
   * 重新加载
   */
  await loadArticles()

}


/* =====================================================
   分页：切换页码
===================================================== */

const handleCurrentChange = async (
  page: number,
) => {

  currentPage.value = page

  await loadArticles()

}


/* =====================================================
   分页：修改每页数量
===================================================== */

const handleSizeChange = async (
  size: number,
) => {

  pageSize.value = size

  currentPage.value = 1

  await loadArticles()

}


/* =====================================================
   新建文章
===================================================== */

const handleCreate = () => {

  router.push(
    '/admin/articles/new',
  )

}


/* =====================================================
   编辑文章
===================================================== */

const handleEdit = (
  article: Article,
) => {

  router.push(
    `/admin/articles/${article.id}/edit`,
  )

}


/* =====================================================
   删除文章
===================================================== */

const handleDelete = async (
  article: Article,
) => {

  try {

    /**
     * 删除确认
     */
    await ElMessageBox.confirm(

      `确定要删除文章「${article.title}」吗？`,

      '删除文章',

      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',

        type: 'warning',
      },

    )


    /**
     * 调用 Service 删除
     */
    await articleService.deleteArticle(
      article.id,
    )


    /**
     * 提示
     */
    ElMessage.success(
      '文章已删除',
    )


    /**
     * 重新加载列表
     */
    await loadArticles()

  } catch (error) {

    /**
     * 用户点击取消时，
     * Element Plus 会进入这里。
     *
     * 所以这里不提示错误。
     */
    if (error !== 'cancel') {

      console.error(
        '删除文章失败：',
        error,
      )

    }

  }

}


/* =====================================================
   导出文章
===================================================== */

const handleExport = async (
  article: Article,
) => {

  try {

    /**
     * 当前默认导出 Markdown。
     *
     * 后续可以改成：
     *
     * Markdown
     * HTML
     * PDF
     */
    await articleService.exportArticle(
      article.id,
      'markdown',
    )


    ElMessage.success(
      `已准备导出：${article.title}`,
    )

  } catch (error) {

    console.error(
      '导出文章失败：',
      error,
    )

    ElMessage.error(
      '导出文章失败',
    )

  }

}


/* =====================================================
   页面初始化
===================================================== */

onMounted(async () => {
  await Promise.all([
    loadCategories(),
    loadArticles(),
  ])
})

onBeforeUnmount(() => {
  if (searchTimer) {
    clearTimeout(searchTimer)
  }
})

</script>


<style scoped>

/* =====================================================
   页面
===================================================== */

.article-list {
  width: 100%;
}

.status-select {
  width: 88px;
}


/* =====================================================
   页面标题
===================================================== */

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;

  margin-bottom: 20px;
}


.page-title h2 {
  margin: 0 0 6px;

  font-size: 22px;
  font-weight: 600;

  color: #303133;
}


.page-title p {
  margin: 0;

  font-size: 14px;

  color: #909399;
}


/* =====================================================
   筛选卡片
===================================================== */

.filter-card {
  margin-bottom: 20px;

  border: 1px solid #ebeef5;
}


/* =====================================================
   Tabs
===================================================== */

:deep(.el-tabs__header) {
  margin-bottom: 18px;
}


/* =====================================================
   筛选行
===================================================== */

.filter-row {
  display: flex;
  align-items: center;

  gap: 12px;
}


.search-input {
  width: 320px;
}


.category-select {
  width: 180px;
}


/* =====================================================
   表格卡片
===================================================== */

.table-card {
  border: 1px solid #ebeef5;
}


/* =====================================================
   文章标题
===================================================== */

.article-title {
  color: #303133;

  font-weight: 500;

  cursor: pointer;

  transition:
    color 0.2s ease;
}


.article-title:hover {
  color: #409eff;
}


/* =====================================================
   标签
===================================================== */

.tag-list {
  display: flex;
  flex-wrap: wrap;

  gap: 6px;
}


/* =====================================================
   分页
===================================================== */

.pagination {
  display: flex;

  justify-content: flex-end;

  margin-top: 20px;
}


/* =====================================================
   响应式
===================================================== */

@media (max-width: 768px) {

  .page-header {
    align-items: flex-start;

    gap: 12px;
  }


  .filter-row {
    flex-wrap: wrap;
  }


  .search-input {
    width: 100%;
  }


  .category-select {
    width: 100%;
  }


  .pagination {
    justify-content: center;

    overflow-x: auto;
  }

}

</style>
