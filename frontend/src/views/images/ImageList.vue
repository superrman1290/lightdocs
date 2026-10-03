<template>
  <div class="image-page">
    <div class="page-header">
      <div>
        <h2>图片管理</h2>
        <p>管理网站文章中的图片资源。</p>
      </div>

      <input
        ref="uploadInput"
        class="hidden-input"
        type="file"
        accept="image/*"
        multiple
        @change="handleUpload"
      />
    </div>

    <el-card
      class="image-card-container"
      shadow="never"
    >
      <div class="image-toolbar">
        <div class="image-root-title">
          <span class="folder-badge">
            <el-icon>
              <Folder />
            </el-icon>
          </span>
          <span>全部图片</span>
          <span class="record-count">共 {{ total }} 项</span>
        </div>

        <div class="toolbar-actions">
          <el-input
            v-model="searchKeyword"
            class="search-input"
            clearable
            placeholder="搜索图片名称..."
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <el-icon>
                <Search />
              </el-icon>
            </template>
          </el-input>

          <el-select
            v-model="selectedSource"
            class="source-select"
            placeholder="全部来源"
            @change="handleSearch"
          >
            <el-option
              label="全部来源"
              value=""
            />
            <el-option
              label="文章内容"
              value="article"
            />
            <el-option
              label="上传图片"
              value="upload"
            />
            <el-option
              label="系统图片"
              value="system"
            />
          </el-select>

          <el-button
            type="primary"
            :icon="UploadFilled"
            @click="triggerUpload"
          >
            上传图片
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
          批量删除
        </el-button>
      </div>

      <div
        v-loading="loading"
        class="image-grid"
      >
        <article
          v-for="image in images"
          :key="image.id"
          class="image-item"
        >
          <div
            class="image-preview"
            @click="handlePreview(image)"
          >
            <img
              :src="image.url"
              :alt="image.name"
              @error="handleImageError"
            />

            <el-checkbox
              class="image-checkbox"
              :model-value="isSelected(image.id)"
              :disabled="image.referenced"
              @click.stop
              @change="toggleImage(image.id)"
            />

            <el-dropdown
              class="image-menu"
              trigger="click"
              @command="handleImageCommand($event, image)"
            >
              <el-button
                text
                circle
                :icon="MoreFilled"
                title="图片操作"
                @click.stop
              />

              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="preview">
                    预览
                  </el-dropdown-item>
                  <el-dropdown-item
                    command="delete"
                    divided
                    :disabled="image.referenced"
                  >
                    {{ image.referenced ? '文章引用中' : '删除' }}
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>

          <div class="image-info">
            <div
              class="image-name"
              :title="image.name"
            >
              {{ image.name }}
            </div>

            <div class="image-meta">
              <span>
                <el-icon><Document /></el-icon>
                {{ formatSize(image.size) }}
              </span>
              <span>
                <el-icon><Calendar /></el-icon>
                {{ image.createdAt }}
              </span>
            </div>
          </div>
        </article>
      </div>

      <el-empty
        v-if="!loading && images.length === 0"
        description="暂无图片"
      />

      <div class="table-footer">
        <span class="footer-count">共 {{ total }} 张图片</span>

        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 15, 30, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          background
          @current-change="loadImages"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>

    <el-dialog
      v-model="previewVisible"
      :title="previewImage?.name ?? '图片预览'"
      width="820px"
      align-center
    >
      <img
        v-if="previewImage"
        class="preview-image"
        :src="previewImage.url"
        :alt="previewImage.name"
        @error="handleImageError"
      />
    </el-dialog>
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
  Calendar,
  Delete,
  Document,
  Folder,
  MoreFilled,
  Search,
  UploadFilled,
} from '@element-plus/icons-vue'

import fallbackImage from '../../assets/hero.png'

import {
  imageService,
} from '../../services/imageService'

import type {
  ImageAsset,
  ImageSource,
} from '../../types/image'

const images = ref<ImageAsset[]>([])
const total = ref(0)
const loading = ref(false)
const searchKeyword = ref('')
const selectedSource = ref<ImageSource | ''>('')
const currentPage = ref(1)
const pageSize = ref(15)
const selectedIds = ref<number[]>([])
const uploadInput = ref<HTMLInputElement | null>(null)
const previewVisible = ref(false)
const previewImage = ref<ImageAsset | null>(null)

const visibleDeletableIds = computed(() => images.value
  .filter(image => !image.referenced)
  .map(image => image.id))

const allVisibleSelected = computed(() => {
  return visibleDeletableIds.value.length > 0 && visibleDeletableIds.value.every(
    id => selectedIds.value.includes(id),
  )
})

const someVisibleSelected = computed(() => {
  const selectedCount = visibleDeletableIds.value.filter(
    id => selectedIds.value.includes(id),
  ).length

  return selectedCount > 0 && selectedCount < visibleDeletableIds.value.length
})

const loadImages = async () => {
  loading.value = true

  try {
    const result = await imageService.getImages({
      keyword: searchKeyword.value,
      source: selectedSource.value || undefined,
      page: currentPage.value,
      pageSize: pageSize.value,
    })

    images.value = result.list
    total.value = result.total
    selectedIds.value = []
  } catch (error) {
    console.error('获取图片失败：', error)
    ElMessage.error('获取图片失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = async () => {
  currentPage.value = 1
  await loadImages()
}

const handleSizeChange = async (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  await loadImages()
}

const isSelected = (id: number) => selectedIds.value.includes(id)

const toggleImage = (id: number) => {
  if (isSelected(id)) {
    selectedIds.value = selectedIds.value.filter(item => item !== id)
  } else {
    selectedIds.value = [...selectedIds.value, id]
  }
}

const toggleSelectAll = (value: string | number | boolean) => {
  const shouldSelect = Boolean(value)

  if (shouldSelect) {
    selectedIds.value = Array.from(
      new Set([...selectedIds.value, ...visibleDeletableIds.value]),
    )
  } else {
    selectedIds.value = selectedIds.value.filter(
      id => !visibleDeletableIds.value.includes(id),
    )
  }
}

const deleteImages = async (ids: number[]) => {
  const deletableIds = ids.filter(id => {
    return !images.value.find(image => image.id === id)?.referenced
  })

  if (deletableIds.length === 0) {
    ElMessage.warning('图片正在被文章引用，请先移除文章中的图片')
    return
  }

  await imageService.deleteImages(deletableIds)
  selectedIds.value = []
  await loadImages()
}

const handleBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedIds.value.length} 张图片吗？删除后无法恢复。`,
      '批量删除图片',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )

    await deleteImages([...selectedIds.value])
    ElMessage.success('图片已删除')
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      console.error('批量删除图片失败：', error)
      ElMessage.error('批量删除图片失败')
    }
  }
}

const handleSingleDelete = async (image: ImageAsset) => {
  if (image.referenced) {
    ElMessage.warning('图片正在被文章引用，请先移除文章中的图片')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除图片「${image.name}」吗？`,
      '删除图片',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )

    await deleteImages([image.id])
    ElMessage.success('图片已删除')
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      console.error('删除图片失败：', error)
      ElMessage.error('删除图片失败')
    }
  }
}

const handleImageCommand = (
  command: string,
  image: ImageAsset,
) => {
  if (command === 'preview') {
    handlePreview(image)
  }

  if (command === 'delete') {
    void handleSingleDelete(image)
  }
}

const handlePreview = (image: ImageAsset) => {
  previewImage.value = image
  previewVisible.value = true
}

const triggerUpload = () => {
  uploadInput.value?.click()
}

const handleUpload = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files ?? [])

  if (files.length === 0) {
    return
  }

  try {
    await imageService.uploadImages(files)
    ElMessage.success(`已上传 ${files.length} 张图片`)
    currentPage.value = 1
    await loadImages()
  } catch (error) {
    console.error('上传图片失败：', error)
    ElMessage.error('上传图片失败')
  } finally {
    input.value = ''
  }
}

const formatSize = (size: number) => {
  return `${size} KB`
}

const handleImageError = (event: Event) => {
  const image = event.target as HTMLImageElement

  if (image.src !== fallbackImage) {
    image.src = fallbackImage
  }
}

onMounted(loadImages)
</script>

<style scoped>
.image-page {
  width: 100%;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
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

.hidden-input {
  display: none;
}

.image-card-container {
  border: 1px solid #e7edf7;
  border-radius: 8px;
}

.image-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding-bottom: 18px;
}

.image-root-title {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #18284f;
  font-size: 15px;
  font-weight: 600;
  white-space: nowrap;
}

.folder-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  color: #2784f2;
  background: #eaf3ff;
}

.record-count {
  color: #9aa6bb;
  font-size: 12px;
  font-weight: 400;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  min-width: 0;
}

.search-input {
  width: 260px;
}

.source-select {
  width: 130px;
}

.selection-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  padding: 10px 0;
  border-top: 1px solid #eef2f8;
  border-bottom: 1px solid #eef2f8;
  color: #7383a1;
  font-size: 13px;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 16px;
  min-height: 250px;
  padding: 20px 0;
}

.image-item {
  overflow: hidden;
  border: 1px solid #dfe8f5;
  border-radius: 8px;
  background: #ffffff;
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
}

.image-item:hover {
  border-color: #9fc7ff;
  box-shadow: 0 8px 20px rgba(31, 89, 165, 0.1);
}

.image-preview {
  position: relative;
  aspect-ratio: 1.72;
  overflow: hidden;
  cursor: zoom-in;
  background: #eef4fb;
}

.image-preview img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
}

.image-checkbox {
  position: absolute;
  top: 10px;
  left: 10px;
  z-index: 2;
}

.image-menu {
  position: absolute;
  top: 5px;
  right: 5px;
  z-index: 2;
}

.image-menu :deep(.el-button) {
  color: #ffffff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
}

.image-info {
  padding: 10px 12px 12px;
}

.image-name {
  overflow: hidden;
  color: #1d376a;
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-meta {
  display: flex;
  flex-direction: column;
  gap: 5px;
  margin-top: 8px;
  color: #8999b7;
  font-size: 11px;
}

.image-meta span {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.table-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 4px;
}

.footer-count {
  color: #7d8eaf;
  font-size: 13px;
}

.preview-image {
  display: block;
  width: 100%;
  max-height: 70vh;
  object-fit: contain;
}

@media (max-width: 1200px) {
  .image-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (max-width: 900px) {
  .image-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .image-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .toolbar-actions {
    width: 100%;
    justify-content: flex-start;
  }
}

@media (max-width: 620px) {
  .image-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .toolbar-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .search-input,
  .source-select {
    width: 100%;
  }

  .table-footer {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
