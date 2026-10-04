<template>
  <div class="article-editor-page">
    <!-- =========================
         页面主体
    ========================== -->
    <div class="editor-body">

      <!-- =========================
           中间文档区域
      ========================== -->
      <main class="document-container">
        <div class="document-content">

          <!-- =========================
               文章标题
          ========================== -->
          <input
            v-model="form.title"
            class="document-title"
            placeholder="请输入文章标题"
            maxlength="200"
          />

          <!-- =========================
               文章信息
          ========================== -->
          <div class="document-meta">
            <span>
              {{ categoryName }}
            </span>

            <span class="meta-divider">
              ·
            </span>

            <span>
              {{
                form.tags.length > 0
                  ? form.tags.join('、')
                  : '暂无标签'
              }}
            </span>
          </div>

          <div class="article-fields">
            <el-select
              v-model="form.categoryId"
              class="article-field category-field"
              placeholder="选择分类"
              :disabled="categories.length === 0"
            >
              <el-option
                v-for="category in categories"
                :key="category.id"
                :label="category.name"
                :value="category.id"
              />
            </el-select>

            <el-select
              v-model="form.tags"
              class="article-field tags-field"
              multiple
              filterable
              allow-create
              default-first-option
              collapse-tags
              collapse-tags-tooltip
              placeholder="输入标签，回车确认（每个最多 8 个字）"
              @change="handleTagsChange"
            >
              <el-option
                v-for="tag in form.tags"
                :key="tag"
                :label="tag"
                :value="tag"
              />
            </el-select>
          </div>

          <!-- =========================
               文档编辑器
          ========================== -->
          <DocumentEditor
            ref="documentEditor"
            v-model="form.content"
          />

        </div>
      </main>

    </div>

    <!-- =========================
         底部操作栏
    ========================== -->
    <div class="bottom-actions">
      <div class="bottom-left">
        <span class="word-count">
          {{ wordCount }} 字
        </span>
      </div>

      <div class="bottom-right">
        <el-button
          @click="handleBack"
        >
          取消
        </el-button>

        <el-button
          :loading="saving"
          @click="handleSaveDraft"
        >
          保存草稿
        </el-button>

        <el-button
          type="primary"
          :loading="saving"
          @click="handlePublish"
        >
          {{ isEditMode ? '保存并发布' : '发布文章' }}
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  nextTick,
  onMounted,
  reactive,
  ref,
} from 'vue'

import { useRouter } from 'vue-router'

import {
  ElMessage,
  ElMessageBox,
} from 'element-plus'

import DocumentEditor from './DocumentEditor.vue'

import {
  articleService,
} from '../../services/articleService'

import {
  categoryService,
} from '../../services/categoryService'

import type {
  Article,
  ArticleForm,
} from '../../types/article'

import type {
  Category,
} from '../../types/category'


/* =========================================================
   Props
========================================================= */

interface Props {
  /**
   * 新建文章：
   * article 不传
   *
   * 编辑文章：
   * article 传入已有文章
   */
  article?: Article
}

const props = defineProps<Props>()


/* =========================================================
   Router
========================================================= */

const router = useRouter()


/* =========================================================
   页面状态
========================================================= */

/**
 * 是否为编辑模式
 */
const isEditMode = computed(() => {
  return !!props.article
})

/**
 * 是否正在保存
 */
const saving = ref(false)

/**
 * 是否已经保存
 */
const saved = ref(false)


/* =========================================================
   DocumentEditor
========================================================= */

/**
 * 获取 DocumentEditor 组件实例
 *
 * 用于在编辑页初始化时聚焦正文编辑器，
 * 后续也可以继续扩展编辑器能力。
 */
const documentEditor =
  ref<InstanceType<typeof DocumentEditor> | null>(null)

const categories = ref<Category[]>([])


/* =========================================================
   文章表单
========================================================= */

const form = reactive<ArticleForm>({
  title: props.article?.title ?? '',

  slug: props.article?.slug ?? '',

  categoryId:
    props.article?.categoryId ?? 0,

  tags:
    props.article?.tags
      ? [...props.article.tags]
      : [],

  content:
    props.article?.content ?? '',

  summary:
    props.article?.summary ?? '',
})

const categoryName = computed(() => {
  return categories.value.find(
    category => category.id === form.categoryId,
  )?.name ?? props.article?.category ?? '未分类'
})

const handleTagsChange = (tags: string[]) => {
  const normalized = tags
    .map(tag => tag.trim())
    .filter(Boolean)

  const invalid = normalized.find(tag => [...tag].length > 8)

  if (invalid) {
    ElMessage.warning(`标签「${invalid}」不能超过 8 个字`)
  }

  form.tags = Array.from(
    new Set(normalized.filter(tag => [...tag].length <= 8)),
  )
}

const validateArticleFields = () => {
  if (!form.categoryId) {
    ElMessage.warning('请选择文章分类')
    return false
  }

  const invalid = form.tags.find(tag => [...tag].length > 8)

  if (invalid) {
    ElMessage.warning(`标签「${invalid}」不能超过 8 个字`)
    return false
  }

  return true
}

const loadCategories = async () => {
  categories.value = await categoryService.getCategories()

  if (!props.article && categories.value.length > 0) {
    form.categoryId = categories.value[0].id
  }
}


/* =========================================================
   字数统计
========================================================= */

const wordCount = computed(() => {
  return form.content
    .replace(/\s/g, '')
    .length
})


/* =========================================================
   返回文章列表
========================================================= */

const handleBack = async () => {
  /**
   * 如果已经保存，则直接返回
   */
  if (saved.value) {
    router.push('/admin/articles')
    return
  }

  /**
   * 新文章完全没有内容
   * 直接返回
   */
  if (
    !form.title.trim() &&
    !form.content.trim()
  ) {
    router.push('/admin/articles')
    return
  }

  /**
   * 有内容但没有保存
   */
  try {
    await ElMessageBox.confirm(
      '当前文章可能还有未保存的内容，确定离开吗？',
      '提示',
      {
        confirmButtonText: '离开',
        cancelButtonText: '继续编辑',
        type: 'warning',
      },
    )
  } catch {
    return
  }

  router.push('/admin/articles')
}


/* =========================================================
   生成 Slug
========================================================= */

const generateSlug = (
  title: string,
) => {
  const slug = title
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '-')
    .replace(
      /[^a-z0-9-]/g,
      '',
    )

  if (slug) {
    return slug
  }

  const suffix = globalThis.crypto?.randomUUID?.().slice(0, 8)
    ?? Date.now().toString(36)

  return `article-${suffix}`
}


/* =========================================================
   保存草稿
========================================================= */

const handleSaveDraft = async () => {
  /**
   * 标题不能为空
   */
  if (!form.title.trim()) {
    ElMessage.warning(
      '请输入文章标题',
    )

    return
  }

  if (!validateArticleFields()) return

  saving.value = true
  saved.value = false

  try {
    /**
     * 如果没有 slug，
     * 根据标题自动生成。
     */
    if (!form.slug.trim()) {
      form.slug = generateSlug(form.title)
    }

    /**
     * 编辑已有文章
     */
    if (
      isEditMode.value &&
      props.article
    ) {
      await articleService.updateArticle(
        props.article.id,
        {
          title: form.title,

          slug: form.slug,

          categoryId:
            form.categoryId,

          tags:
            form.tags,

          content:
            form.content,

          summary:
            form.summary,

          status: 'draft',
        },
      )
    }

    /**
     * 新建文章
     */
    else {
      await articleService.createArticle({
        title: form.title,

        slug: form.slug,

        categoryId:
          form.categoryId,

        tags:
          form.tags,

        content:
          form.content,

        summary:
          form.summary,

        status: 'draft',
      })
    }

    saved.value = true

    ElMessage.success(
      '草稿已保存',
    )

    await router.push(
      '/admin/articles',
    )
  } catch (error) {
    console.error(
      '保存草稿失败：',
      error,
    )

    ElMessage.error(
      error instanceof Error && error.message
        ? `保存草稿失败：${error.message}`
        : '保存草稿失败',
    )
  } finally {
    saving.value = false
  }
}


/* =========================================================
   发布文章
========================================================= */

const handlePublish = async () => {
  /**
   * 标题不能为空
   */
  if (!form.title.trim()) {
    ElMessage.warning(
      '请输入文章标题',
    )

    return
  }

  if (!validateArticleFields()) return

  /**
   * 内容不能为空
   */
  if (!form.content.trim()) {
    ElMessage.warning(
      '请输入文章内容',
    )

    return
  }

  saving.value = true
  saved.value = false

  try {
    /**
     * 自动生成 Slug
     */
    if (!form.slug.trim()) {
      form.slug = generateSlug(form.title)
    }

    /**
     * 编辑已有文章
     */
    if (
      isEditMode.value &&
      props.article
    ) {
      await articleService.updateArticle(
        props.article.id,
        {
          title:
            form.title,

          slug:
            form.slug,

          categoryId:
            form.categoryId,

          tags:
            form.tags,

          content:
            form.content,

          summary:
            form.summary,

          status:
            'published',
        },
      )
    }

    /**
     * 新建文章
     */
    else {
      await articleService.createArticle({
        title:
          form.title,

        slug:
          form.slug,

        categoryId:
          form.categoryId,

        tags:
          form.tags,

        content:
          form.content,

        summary:
          form.summary,

        status:
          'published',
      })
    }

    saved.value = true

    ElMessage.success(
      isEditMode.value
        ? '文章已保存并发布'
        : '文章已发布',
    )

    /**
     * 发布成功后返回文章列表
     */
    await router.push(
      '/admin/articles',
    )
  } catch (error) {
    console.error(
      '发布文章失败：',
      error,
    )

    ElMessage.error(
      error instanceof Error && error.message
        ? `发布文章失败：${error.message}`
        : '发布文章失败',
    )
  } finally {
    saving.value = false
  }
}


/* =========================================================
   页面初始化
========================================================= */

onMounted(async () => {
  try {
    await loadCategories()
  } catch (error) {
    console.error(
      '获取分类列表失败：',
      error,
    )

    ElMessage.error(
      '获取分类列表失败',
    )
  }

  nextTick(() => {
    documentEditor.value?.editor
      ?.chain()
      .focus()
      .run()
  })
})
</script>


<style scoped>
/* =========================================================
   页面整体
========================================================= */

.article-editor-page {
  min-height: 100vh;

  background: #ffffff;

  color: #1f2329;
}


/* =========================================================
   主体
========================================================= */

.editor-body {
  display: flex;

  max-width: 1440px;

  min-height: 100vh;

  margin: 0 auto;
}


/* =========================================================
   中间文档区域
========================================================= */

.document-container {
  flex: 1;

  min-width: 0;

  padding:
    64px
    80px
    140px;
}


/* 文档最大宽度 */

.document-content {
  max-width: 900px;

  margin: 0 auto;
}


/* =========================================================
   文章标题
========================================================= */

.document-title {
  display: block;

  width: 100%;

  padding: 0;

  margin: 0;

  border: none;

  outline: none;

  background: transparent;

  color: #1f2329;

  font-size: 40px;

  line-height: 1.35;

  font-weight: 700;

  letter-spacing: -0.5px;
}


/* 标题 Placeholder */

.document-title::placeholder {
  color: #b7bdc6;
}


/* =========================================================
   文章信息
========================================================= */

.document-meta {
  display: flex;

  align-items: center;

  margin-top: 16px;

  margin-bottom: 28px;

  font-size: 13px;

  color: #8f959e;
}


.meta-divider {
  margin:
    0 8px;
}

.article-fields {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: -10px 0 22px;
}

.article-field {
  min-width: 0;
}

.category-field {
  width: 190px;
}

.tags-field {
  flex: 1;
}

.article-fields :deep(.el-select__wrapper) {
  min-height: 34px;
}

@media (max-width: 620px) {
  .article-fields {
    align-items: stretch;
    flex-direction: column;
  }

  .category-field,
  .tags-field {
    width: 100%;
  }
}


/* =========================================================
   底部操作栏
========================================================= */

.bottom-actions {
  position: fixed;

  left: 0;

  right: 0;

  bottom: 0;

  z-index: 90;

  display: flex;

  align-items: center;

  justify-content: space-between;

  padding:
    12px
    24px;

  background: rgba(
    255,
    255,
    255,
    0.96
  );

  border-top:
    1px solid #ebeef0;

  backdrop-filter: blur(8px);
}


/* 底部左侧 */

.bottom-left {
  display: flex;

  align-items: center;
}


/* 字数 */

.word-count {
  font-size: 12px;

  color: #b0b5bd;
}


/* 底部右侧 */

.bottom-right {
  display: flex;

  align-items: center;

  gap: 10px;
}


/* =========================================================
   响应式
========================================================= */

@media (max-width: 1100px) {
  .document-container {
    padding-left: 50px;

    padding-right: 40px;
  }

}


@media (max-width: 760px) {
  .document-container {
    padding:
      40px
      24px
      120px;
  }

  .document-title {
    font-size: 32px;
  }

  .bottom-actions {
    padding:
      10px
      16px;
  }

  .bottom-left {
    display: none;
  }

  .bottom-right {
    width: 100%;

    justify-content: flex-end;
  }
}
</style>
