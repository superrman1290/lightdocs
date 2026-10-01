<template>
  <div class="category-page">
    <div class="page-header">
      <div>
        <h2>分类管理</h2>
        <p>管理网站的多级文章分类和显示顺序。</p>
      </div>

      <el-button
        type="primary"
        :icon="Plus"
        @click="handleCreate"
      >
        新建分类
      </el-button>
    </div>

    <el-card
      class="category-card"
      shadow="never"
    >
      <div class="category-toolbar">
        <div class="category-root-title">
          <span class="folder-badge">
            <el-icon>
              <Folder />
            </el-icon>
          </span>
          <span>全部分类</span>
          <span class="record-count">共 {{ filteredCategories.length }} 项</span>
        </div>

        <div class="category-toolbar-actions">
          <el-button
            v-if="selectedCategories.length > 0"
            type="danger"
            plain
            :icon="Delete"
            @click="handleBatchDelete"
          >
            删除所选（{{ selectedCategories.length }}）
          </el-button>

          <el-input
            v-model="searchKeyword"
            class="search-input"
            clearable
            placeholder="请输入分类名称"
          >
            <template #prefix>
              <el-icon>
                <Search />
              </el-icon>
            </template>
          </el-input>
        </div>
      </div>

      <div class="table-container">
        <el-table
          ref="tableRef"
          v-loading="loading"
          :data="treeCategories"
          row-key="id"
          :row-class-name="getRowClassName"
          :tree-props="{ children: 'children' }"
          :indent="24"
          default-expand-all
          @row-click="handleRowClick"
          @selection-change="handleSelectionChange"
          class="category-table"
          empty-text="暂无分类"
        >
          <el-table-column
            type="index"
            width="42"
            class-name="drag-column"
          >
            <template #default="{ row }">
              <button
                class="drag-handle"
                title="拖动调整排序"
                @click.stop
                @pointerdown="handlePointerDown(row.id, $event)"
              >
                <span class="drag-dots">
                  <i
                    v-for="dot in 6"
                    :key="dot"
                  />
                </span>
              </button>
            </template>
          </el-table-column>

          <el-table-column
            type="selection"
            width="54"
          />

          <el-table-column
            label="分类名称"
            min-width="320"
          >
            <template #default="{ row }">
              <div
                class="category-name"
                :style="{
                  paddingLeft: `${row.level * 28}px`,
                }"
              >
                <el-icon>
                  <Folder />
                </el-icon>
                <span>{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column
            label="父级分类"
            min-width="220"
          >
            <template #default="{ row }">
              <span class="parent-name">
                {{ getParentName(row) }}
              </span>
            </template>
          </el-table-column>

          <el-table-column
            label="操作"
            width="170"
            fixed="right"
          >
            <template #default="{ row }">
              <div class="row-actions">
                <el-button
                  link
                  type="primary"
                  :icon="Edit"
                  @click="handleEdit(row)"
                >
                  编辑
                </el-button>

                <el-button
                  link
                  type="danger"
                  :icon="Delete"
                  @click="handleDelete(row)"
                >
                  删除
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="table-footer">
        <span class="footer-count">共 {{ filteredCategories.length }} 条记录</span>

        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50]"
          :total="filteredCategories.length"
          layout="total, sizes, prev, pager, next"
          background
          hide-on-single-page
        />
      </div>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新建分类' : '编辑分类'"
      width="460px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="90px"
      >
        <el-form-item
          label="分类名称"
          prop="name"
        >
          <el-input
            v-model="form.name"
            maxlength="40"
            show-word-limit
            placeholder="请输入分类名称"
          />
        </el-form-item>

        <el-form-item label="父级分类">
          <el-select
            v-model="form.parentId"
            clearable
            class="dialog-control"
            placeholder="选择父级分类（可选）"
          >
            <el-option
              v-for="category in parentOptions"
              :key="category.id"
              :label="category.name"
              :value="category.id"
            />
          </el-select>
        </el-form-item>

      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">
          取消
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          @click="handleSubmit"
        >
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
} from 'vue'

import {
  ElMessage,
  ElMessageBox,
  type FormInstance,
  type FormRules,
} from 'element-plus'

import {
  Delete,
  Edit,
  Folder,
  Plus,
  Search,
} from '@element-plus/icons-vue'

import {
  categoryService,
} from '../../services/categoryService'

import type {
  Category,
  CategoryInput,
} from '../../types/category'

interface CategoryRow extends Category {
  level: number
  children?: CategoryRow[]
}

interface CategoryTableInstance {
  toggleRowExpansion: (row: CategoryRow, expanded?: boolean) => void
  clearSelection: () => void
}

const categories = ref<Category[]>([])
const loading = ref(false)
const saving = ref(false)
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const tableRef = ref<CategoryTableInstance | null>(null)
const draggingId = ref<number | null>(null)
const dropTargetId = ref<number | null>(null)
const dropPosition = ref<'before' | 'after' | null>(null)
const selectedCategories = ref<Category[]>([])

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()

const form = reactive<CategoryInput>({
  name: '',
  parentId: null,
  sort: 1,
})

const formRules: FormRules<CategoryInput> = {
  name: [
    {
      required: true,
      message: '请输入分类名称',
      trigger: 'blur',
    },
  ],
}

const categoryMap = computed(() => {
  return new Map(
    categories.value.map(category => [category.id, category]),
  )
})

const filteredCategories = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()

  if (!keyword) {
    return categories.value
  }

  const visibleIds = new Set<number>()

  for (const category of categories.value) {
    if (category.name.toLowerCase().includes(keyword)) {
      visibleIds.add(category.id)

      let parentId = category.parentId

      while (parentId !== null) {
        visibleIds.add(parentId)
        parentId = categoryMap.value.get(parentId)?.parentId ?? null
      }
    }
  }

  return categories.value.filter(category => visibleIds.has(category.id))
})

const treeCategories = computed<CategoryRow[]>(() => {
  const visible = filteredCategories.value

  const build = (
    parentId: number | null,
    level = 0,
  ): CategoryRow[] => {
    return visible
      .filter(category => category.parentId === parentId)
      .sort((a, b) => a.sort - b.sort)
      .map(category => ({
        ...category,
        level,
        children: build(category.id, level + 1),
      }))
  }

  return build(null)
})

const parentOptions = computed(() => {
  return categories.value.filter(category => {
    if (category.id === editingId.value) {
      return false
    }

    let parentId = category.parentId

    while (parentId !== null) {
      if (parentId === editingId.value) {
        return false
      }

      parentId = categoryMap.value.get(parentId)?.parentId ?? null
    }

    return true
  })
})

const getParentName = (category: Category) => {
  if (category.parentId === null) {
    return ''
  }

  return categoryMap.value.get(category.parentId)?.name ?? '未分类'
}

const getRowClassName = ({ row }: { row: CategoryRow }) => {
  const classes = [`category-row-${row.id}`]

  if (draggingId.value === row.id) {
    classes.push('is-dragging')
  }

  if (dropTargetId.value === row.id && dropPosition.value) {
    classes.push(`drop-${dropPosition.value}`)
  }

  return classes.join(' ')
}

const handleRowClick = (
  row: CategoryRow,
  _column: unknown,
  event: MouseEvent,
) => {
  const target = event.target as HTMLElement | null

  if (
    target?.closest(
      'button, input, .el-button, .el-checkbox, .drag-handle',
    )
  ) {
    return
  }

  if (row.children?.length) {
    tableRef.value?.toggleRowExpansion(row)
  }
}

const handleSelectionChange = (rows: Category[]) => {
  selectedCategories.value = rows
}

const getCategoryDepth = (category: Category) => {
  let depth = 0
  let parentId = category.parentId

  while (parentId !== null) {
    depth += 1
    parentId = categoryMap.value.get(parentId)?.parentId ?? null
  }

  return depth
}

const handleBatchDelete = async () => {
  const selected = [...selectedCategories.value]
  const selectedIds = new Set(selected.map(category => category.id))
  const blocked = selected.filter(category => {
    return categories.value.some(
      child => child.parentId === category.id && !selectedIds.has(child.id),
    )
  })

  if (blocked.length > 0) {
    ElMessage.warning(
      `请先选择分类「${blocked.map(category => category.name).join('、')}」的全部子分类，或单独处理子分类`,
    )
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selected.length} 个分类吗？删除后无法恢复。`,
      '批量删除分类',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )

    const ordered = selected.sort(
      (a, b) => getCategoryDepth(b) - getCategoryDepth(a),
    )

    for (const category of ordered) {
      await categoryService.deleteCategory(category.id)
    }

    tableRef.value?.clearSelection()
    selectedCategories.value = []
    ElMessage.success(`已删除 ${selected.length} 个分类`)
    await loadCategories()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      console.error('批量删除分类失败：', error)
      ElMessage.error(error instanceof Error ? error.message : '批量删除分类失败')
    }
  }
}

const getRowId = (element: Element | null) => {
  const row = element?.closest('tr')
  const className = row?.getAttribute('class') ?? ''
  const categoryClass = className
    .split(' ')
    .find(value => value.startsWith('category-row-'))

  return {
    row,
    id: categoryClass
      ? Number(categoryClass.replace('category-row-', ''))
      : NaN,
  }
}

const updateDropTarget = (
  element: Element | null,
  clientY: number,
) => {
  const { row, id: targetId } = getRowId(element)
  const source = categories.value.find(
    category => category.id === draggingId.value,
  )
  const target = categories.value.find(category => category.id === targetId)

  if (
    !row ||
    !source ||
    !target ||
    source.id === target.id ||
    source.parentId !== target.parentId
  ) {
    dropTargetId.value = null
    dropPosition.value = null
    return
  }

  const rect = row.getBoundingClientRect()

  dropTargetId.value = target.id
  dropPosition.value = clientY < rect.top + rect.height / 2
    ? 'before'
    : 'after'
}

const clearDragState = () => {
  window.removeEventListener('pointermove', handlePointerMove)
  window.removeEventListener('pointerup', handlePointerUp)
  window.removeEventListener('pointercancel', handlePointerCancel)

  draggingId.value = null
  dropTargetId.value = null
  dropPosition.value = null
}

const handlePointerDown = (id: number, event: PointerEvent) => {
  if (event.button !== 0) {
    return
  }

  event.preventDefault()
  draggingId.value = id
  dropTargetId.value = null
  dropPosition.value = null

  const target = event.currentTarget as HTMLElement | null
  target?.setPointerCapture?.(event.pointerId)

  window.addEventListener('pointermove', handlePointerMove, {
    passive: false,
  })
  window.addEventListener('pointerup', handlePointerUp)
  window.addEventListener('pointercancel', handlePointerCancel)
}

const handlePointerMove = (event: PointerEvent) => {
  if (draggingId.value === null) {
    return
  }

  event.preventDefault()
  updateDropTarget(
    document.elementFromPoint(event.clientX, event.clientY),
    event.clientY,
  )
}

const persistReorder = async (
  sourceId: number,
  targetId: number,
  position: 'before' | 'after',
) => {
  const source = categories.value.find(category => category.id === sourceId)
  const target = categories.value.find(category => category.id === targetId)

  if (!source || !target || source.parentId !== target.parentId) {
    ElMessage.info('只能调整同级分类的顺序')
    return
  }

  const siblings = categories.value
    .filter(category => category.parentId === source.parentId)
    .sort((a, b) => a.sort - b.sort)

  const sourceIndex = siblings.findIndex(category => category.id === sourceId)
  const targetIndex = siblings.findIndex(category => category.id === targetId)

  if (sourceIndex === -1 || targetIndex === -1) {
    return
  }

  const [moved] = siblings.splice(sourceIndex, 1)
  const insertIndex = position === 'after'
    ? targetIndex + 1
    : targetIndex
  const adjustedIndex = sourceIndex < insertIndex
    ? insertIndex - 1
    : insertIndex

  siblings.splice(adjustedIndex, 0, moved)

  try {
    await Promise.all(
      siblings.map((category, index) => {
        return categoryService.updateCategory(category.id, {
          name: category.name,
          parentId: category.parentId,
          sort: index + 1,
        })
      }),
    )

    await loadCategories()
    ElMessage.success('排序已更新')
  } catch (error) {
    console.error('更新分类排序失败：', error)
    ElMessage.error('更新分类排序失败')
  }
}

const finishPointerDrag = async () => {
  const sourceId = draggingId.value
  const targetId = dropTargetId.value
  const position = dropPosition.value

  clearDragState()

  if (!sourceId || !targetId || !position || sourceId === targetId) {
    return
  }

  await persistReorder(sourceId, targetId, position)
}

const handlePointerUp = () => {
  void finishPointerDrag()
}

const handlePointerCancel = () => {
  clearDragState()
}

const loadCategories = async () => {
  loading.value = true

  try {
    categories.value = await categoryService.getCategories()
  } catch (error) {
    console.error('获取分类失败：', error)
    ElMessage.error('获取分类失败')
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.name = ''
  form.parentId = null
  form.sort = categories.value.filter(category => category.parentId === null).length + 1
}

const handleCreate = () => {
  dialogMode.value = 'create'
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (category: Category) => {
  dialogMode.value = 'edit'
  editingId.value = category.id
  form.name = category.name
  form.parentId = category.parentId
  form.sort = category.sort
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate()

  if (!valid) {
    return
  }

  saving.value = true

  try {
    if (dialogMode.value === 'create') {
      await categoryService.createCategory({ ...form })
      ElMessage.success('分类已创建')
    } else if (editingId.value !== null) {
      await categoryService.updateCategory(editingId.value, { ...form })
      ElMessage.success('分类已更新')
    }

    dialogVisible.value = false
    await loadCategories()
  } catch (error) {
    console.error('保存分类失败：', error)
    ElMessage.error(error instanceof Error ? error.message : '保存分类失败')
  } finally {
    saving.value = false
  }
}

const handleDelete = async (category: Category) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除分类「${category.name}」吗？`,
      '删除分类',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )

    await categoryService.deleteCategory(category.id)
    ElMessage.success('分类已删除')
    await loadCategories()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      console.error('删除分类失败：', error)
      ElMessage.error(error instanceof Error ? error.message : '删除分类失败')
    }
  }
}

onMounted(loadCategories)
onBeforeUnmount(clearDragState)
</script>

<style scoped>
.category-page {
  width: 100%;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
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

.category-card {
  border: 1px solid #e7edf7;
  border-radius: 8px;
}

.category-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 0 0 18px;
}

.category-toolbar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.category-root-title {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #18284f;
  font-size: 15px;
  font-weight: 600;
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

.search-input {
  width: 260px;
}

.category-table {
  color: #24365f;
}

.category-table :deep(.el-table__expand-icon) {
  display: none;
}

.category-table :deep(.el-table__indent),
.category-table :deep(.el-table__placeholder) {
  display: none;
}

.category-table :deep(td.drag-column .cell) {
  padding: 0;
  font-size: 0;
  text-align: center;
}

.category-table :deep(.el-table__body tr td) {
  transition:
    background-color 0.18s ease,
    box-shadow 0.18s ease,
    border-color 0.18s ease;
}

.category-table :deep(.el-table__body tr.is-dragging td) {
  visibility: hidden;
  background: #f6f9ff;
}

.category-table :deep(.el-table__body tr.drop-before td) {
  border-top: 6px solid #f0f6ff;
  box-shadow: inset 0 2px 0 #2784f2;
}

.category-table :deep(.el-table__body tr.drop-after td) {
  border-bottom: 6px solid #f0f6ff;
  box-shadow: inset 0 -2px 0 #2784f2;
}

.table-container {
  width: 100%;
}

.drag-handle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 28px;
  padding: 0;
  border: 0;
  color: #2d7ff0;
  background: transparent;
  cursor: grab;
  touch-action: none;
  user-select: none;
}

.drag-handle:active {
  cursor: grabbing;
}

.drag-handle:hover {
  color: #155fd0;
  background: #edf5ff;
}

.drag-dots {
  display: grid;
  grid-template-columns: repeat(2, 3px);
  grid-template-rows: repeat(3, 3px);
  gap: 3px;
}

.drag-dots i {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: currentColor;
}

.category-name {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #1d376a;
  font-weight: 500;
}

.category-name .el-icon {
  color: #2784f2;
  font-size: 17px;
}

.parent-name {
  color: #64769e;
}

.row-actions {
  display: flex;
  align-items: center;
  gap: 2px;
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

.dialog-control {
  width: 100%;
}

@media (max-width: 760px) {
  .page-header,
  .category-toolbar,
  .category-toolbar-actions,
  .table-footer {
    align-items: stretch;
    flex-direction: column;
  }

  .search-input {
    width: 100%;
  }

  .table-footer :deep(.el-pagination) {
    justify-content: center;
  }
}
</style>
