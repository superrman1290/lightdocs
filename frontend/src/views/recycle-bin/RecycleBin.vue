<template>
  <div class="recycle-page">
    <div class="page-header">
      <div>
        <h2>回收站</h2>
        <p>已删除的文章和图片会暂时保存在这里，你可以选择恢复或永久删除。</p>
      </div>
    </div>

    <el-card
      class="recycle-card"
      shadow="never"
    >
      <div class="recycle-toolbar">
        <el-tabs
          v-model="activeType"
          class="type-tabs"
          @tab-change="clearSelection"
        >
          <el-tab-pane
            name="all"
            :label="`全部（${items.length}）`"
          />
          <el-tab-pane
            name="article"
            :label="`文章（${articleCount}）`"
          />
          <el-tab-pane
            name="image"
            :label="`图片（${imageCount}）`"
          />
        </el-tabs>

        <div class="toolbar-actions">
          <el-input
            v-model="searchKeyword"
            class="search-input"
            clearable
            placeholder="搜索回收站内容..."
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <el-icon>
                <Search />
              </el-icon>
            </template>
          </el-input>

          <el-button
            type="danger"
            plain
            :icon="Delete"
            @click="handleClear"
          >
            清空回收站
          </el-button>
        </div>
      </div>

      <div class="selection-toolbar">
        <el-checkbox
          :model-value="allVisibleSelected"
          :indeterminate="someVisibleSelected"
          @change="toggleSelectAll"
        />
        <span>已选择 {{ selectedIds.length }} 项</span>

        <el-button
          v-if="selectedIds.length > 0"
          type="danger"
          link
          :icon="Delete"
          @click="handleBatchDelete"
        >
          永久删除所选
        </el-button>
      </div>

      <section
        v-if="activeType !== 'image'"
        class="recycle-section"
      >
        <div class="section-heading">
          <div class="section-title">
            <el-icon><Document /></el-icon>
            <span>文章 {{ articleItems.length }} 项</span>
          </div>
        </div>

        <el-table
          v-loading="loading"
          :data="articleItems"
          class="recycle-table"
          empty-text="暂无已删除文章"
        >
          <el-table-column width="54">
            <template #default="{ row }">
              <el-checkbox
                :model-value="isSelected(row.id)"
                @change="toggleItem(row.id)"
              />
            </template>
          </el-table-column>

          <el-table-column
            label="标题"
            min-width="360"
          >
            <template #default="{ row }">
              <div class="article-recycle-title">
                <span class="item-icon">
                  <el-icon><Document /></el-icon>
                </span>
                <div>
                  <strong>{{ row.data.title }}</strong>
                  <p>{{ row.data.summary || row.data.content }}</p>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column
            label="分类"
            prop="data.category"
            width="180"
          />

          <el-table-column
            label="删除时间"
            prop="deletedAt"
            width="190"
          />

          <el-table-column
            label="操作"
            width="190"
            fixed="right"
          >
            <template #default="{ row }">
              <div class="row-actions">
                <el-button
                  link
                  type="primary"
                  :icon="RefreshLeft"
                  @click="handleRestore(row)"
                >
                  恢复
                </el-button>
                <el-button
                  link
                  type="danger"
                  :icon="Delete"
                  @click="handlePermanentDelete(row)"
                >
                  永久删除
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section
        v-if="activeType !== 'article'"
        class="recycle-section"
      >
        <div class="section-heading">
          <div class="section-title">
            <el-icon><Picture /></el-icon>
            <span>图片 {{ imageItems.length }} 项</span>
          </div>
        </div>

        <el-table
          v-loading="loading"
          :data="imageItems"
          class="recycle-table"
          empty-text="暂无已删除图片"
        >
          <el-table-column width="54">
            <template #default="{ row }">
              <el-checkbox
                :model-value="isSelected(row.id)"
                @change="toggleItem(row.id)"
              />
            </template>
          </el-table-column>

          <el-table-column
            label="图片名称"
            min-width="360"
          >
            <template #default="{ row }">
              <div class="image-recycle-title">
                <img
                  :src="row.data.url"
                  :alt="row.data.name"
                  @error="handleImageError"
                />
                <div>
                  <strong>{{ row.data.name }}</strong>
                  <span>{{ row.data.size }} KB</span>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column
            label="来源"
            width="180"
          >
            <template #default="{ row }">
              {{ formatImageSource(row.data.source) }}
            </template>
          </el-table-column>

          <el-table-column
            label="删除时间"
            prop="deletedAt"
            width="190"
          />

          <el-table-column
            label="操作"
            width="190"
            fixed="right"
          >
            <template #default="{ row }">
              <div class="row-actions">
                <el-button
                  link
                  type="primary"
                  :icon="RefreshLeft"
                  @click="handleRestore(row)"
                >
                  恢复
                </el-button>
                <el-button
                  link
                  type="danger"
                  :icon="Delete"
                  @click="handlePermanentDelete(row)"
                >
                  永久删除
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <el-empty
        v-if="!loading && items.length === 0"
        description="回收站为空"
      />

      <div class="table-footer">
        <span class="footer-count">共 {{ items.length }} 项</span>
        <el-pagination
          :total="items.length"
          :page-size="12"
          layout="total, prev, pager, next"
          background
          hide-on-single-page
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  onMounted,
  ref,
} from 'vue'

import {
  ElMessage,
  ElMessageBox,
} from 'element-plus'

import {
  Delete,
  Document,
  Picture,
  RefreshLeft,
  Search,
} from '@element-plus/icons-vue'

import fallbackImage from '../../assets/hero.png'

import {
  articleService,
} from '../../services/articleService'

import {
  imageService,
} from '../../services/imageService'

import {
  recycleService,
} from '../../services/recycleService'

import type {
  ImageSource,
} from '../../types/image'

import type {
  RecycleItem,
} from '../../types/recycle'

const activeType = ref<'all' | 'article' | 'image'>('all')
const items = ref<RecycleItem[]>([])
const loading = ref(false)
const searchKeyword = ref('')
const selectedIds = ref<string[]>([])

const articleCount = computed(() => {
  return items.value.filter(item => item.type === 'article').length
})

const imageCount = computed(() => {
  return items.value.filter(item => item.type === 'image').length
})

const articleItems = computed(() => {
  return items.value.filter(item => item.type === 'article')
})

const imageItems = computed(() => {
  return items.value.filter(item => item.type === 'image')
})

const visibleItems = computed(() => {
  if (activeType.value === 'article') {
    return articleItems.value
  }

  if (activeType.value === 'image') {
    return imageItems.value
  }

  return items.value
})

const allVisibleSelected = computed(() => {
  return visibleItems.value.length > 0 && visibleItems.value.every(
    item => selectedIds.value.includes(item.id),
  )
})

const someVisibleSelected = computed(() => {
  const selectedCount = visibleItems.value.filter(
    item => selectedIds.value.includes(item.id),
  ).length

  return selectedCount > 0 && selectedCount < visibleItems.value.length
})

const loadItems = async () => {
  loading.value = true

  try {
    items.value = await recycleService.getItems(
      undefined,
      searchKeyword.value,
    )
    selectedIds.value = []
  } catch (error) {
    console.error('获取回收站失败：', error)
    ElMessage.error('获取回收站失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = async () => {
  await loadItems()
}

const clearSelection = () => {
  selectedIds.value = []
}

const isSelected = (id: string) => selectedIds.value.includes(id)

const toggleItem = (id: string) => {
  if (isSelected(id)) {
    selectedIds.value = selectedIds.value.filter(item => item !== id)
  } else {
    selectedIds.value = [...selectedIds.value, id]
  }
}

const toggleSelectAll = (value: string | number | boolean) => {
  if (Boolean(value)) {
    selectedIds.value = Array.from(
      new Set([
        ...selectedIds.value,
        ...visibleItems.value.map(item => item.id),
      ]),
    )
  } else {
    const visibleIds = new Set(visibleItems.value.map(item => item.id))
    selectedIds.value = selectedIds.value.filter(id => !visibleIds.has(id))
  }
}

const handleRestore = async (item: RecycleItem) => {
  try {
    if (item.type === 'article') {
      await articleService.restoreArticle(item.id)
    } else {
      await imageService.restoreImage(item.id)
    }

    ElMessage.success('项目已恢复')
    await loadItems()
  } catch (error) {
    console.error('恢复项目失败：', error)
    ElMessage.error('恢复项目失败')
  }
}

const handlePermanentDelete = async (item: RecycleItem) => {
  try {
    await ElMessageBox.confirm(
      '永久删除后将无法恢复，确定继续吗？',
      '永久删除',
      {
        confirmButtonText: '永久删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )

    await recycleService.deletePermanently([item.id])
    ElMessage.success('项目已永久删除')
    await loadItems()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      console.error('永久删除失败：', error)
      ElMessage.error('永久删除失败')
    }
  }
}

const handleBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要永久删除选中的 ${selectedIds.value.length} 项吗？`,
      '永久删除所选',
      {
        confirmButtonText: '永久删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )

    await recycleService.deletePermanently([...selectedIds.value])
    ElMessage.success('选中项目已永久删除')
    await loadItems()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      console.error('批量永久删除失败：', error)
      ElMessage.error('批量永久删除失败')
    }
  }
}

const handleClear = async () => {
  if (items.value.length === 0) {
    ElMessage.info('回收站已经为空')
    return
  }

  try {
    await ElMessageBox.confirm(
      '清空后所有项目都将永久删除，确定继续吗？',
      '清空回收站',
      {
        confirmButtonText: '清空',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )

    await recycleService.clear()
    ElMessage.success('回收站已清空')
    await loadItems()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      console.error('清空回收站失败：', error)
      ElMessage.error('清空回收站失败')
    }
  }
}

const formatImageSource = (source: ImageSource) => {
  const labels: Record<ImageSource, string> = {
    article: '文章内容',
    upload: '上传图片',
    system: '系统图片',
  }

  return labels[source]
}

const handleImageError = (event: Event) => {
  const image = event.target as HTMLImageElement

  if (image.src !== fallbackImage) {
    image.src = fallbackImage
  }
}

onMounted(loadItems)
</script>

<style scoped>
.recycle-page {
  width: 100%;
}

.page-header {
  margin-bottom: 20px;
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

.recycle-card {
  border: 1px solid #e7edf7;
  border-radius: 8px;
}

.recycle-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding-bottom: 10px;
}

.type-tabs {
  min-width: 360px;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.search-input {
  width: 280px;
}

.selection-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 42px;
  padding: 8px 0;
  border-top: 1px solid #eef2f8;
  border-bottom: 1px solid #eef2f8;
  color: #7383a1;
  font-size: 13px;
}

.recycle-section {
  margin-top: 18px;
  overflow: hidden;
  border: 1px solid #e5edf8;
  border-radius: 8px;
}

.section-heading {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: #f8fbff;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1d376a;
  font-size: 14px;
  font-weight: 600;
}

.section-title .el-icon {
  color: #2784f2;
  font-size: 17px;
}

.recycle-table {
  color: #58709d;
}

.article-recycle-title,
.image-recycle-title {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.item-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 38px;
  flex: 0 0 34px;
  border-radius: 6px;
  color: #5c9df3;
  background: #edf5ff;
  font-size: 20px;
}

.article-recycle-title > div,
.image-recycle-title > div {
  min-width: 0;
}

.article-recycle-title strong,
.image-recycle-title strong {
  display: block;
  overflow: hidden;
  color: #294474;
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.article-recycle-title p {
  max-width: 420px;
  overflow: hidden;
  margin: 4px 0 0;
  color: #9aa9c3;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-recycle-title img {
  width: 76px;
  height: 42px;
  flex: 0 0 76px;
  border-radius: 5px;
  object-fit: cover;
}

.image-recycle-title span {
  display: block;
  margin-top: 4px;
  color: #9aa9c3;
  font-size: 11px;
}

.row-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.table-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 18px;
}

.footer-count {
  color: #7d8eaf;
  font-size: 13px;
}

@media (max-width: 1000px) {
  .recycle-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .type-tabs {
    min-width: 0;
  }

  .toolbar-actions {
    justify-content: flex-start;
  }
}

@media (max-width: 620px) {
  .toolbar-actions,
  .table-footer {
    align-items: stretch;
    flex-direction: column;
  }

  .search-input {
    width: 100%;
  }
}
</style>
