import { mockCategories } from '../mock/categories'

import type { Category } from '../types/category'
import type { CategoryInput } from '../types/category'

const delay = (ms = 100) => {
  return new Promise(resolve => {
    setTimeout(resolve, ms)
  })
}

/**
 * 分类数据访问层。
 *
 * 当前使用 Mock 数据，后续可替换为 Go API 请求。
 */
const getCategories = async (): Promise<Category[]> => {
  await delay()

  return [...mockCategories]
}

const getCategoryById = async (
  id: number,
): Promise<Category | null> => {
  await delay()

  return (
    mockCategories.find(category => category.id === id)
    ?? null
  )
}

const createCategory = async (
  data: CategoryInput,
): Promise<Category> => {
  await delay()

  const name = data.name.trim()

  if (!name) {
    throw new Error('分类名称不能为空')
  }

  const category: Category = {
    ...data,
    name,
    id: Math.max(0, ...mockCategories.map(item => item.id)) + 1,
  }

  mockCategories.push(category)

  return category
}

const updateCategory = async (
  id: number,
  data: CategoryInput,
): Promise<Category> => {
  await delay()

  const index = mockCategories.findIndex(
    category => category.id === id,
  )

  if (index === -1) {
    throw new Error('分类不存在')
  }

  const name = data.name.trim()

  if (!name) {
    throw new Error('分类名称不能为空')
  }

  mockCategories[index] = {
    ...mockCategories[index],
    ...data,
    name,
  }

  return mockCategories[index]
}

const deleteCategory = async (id: number): Promise<void> => {
  await delay()

  if (mockCategories.some(category => category.parentId === id)) {
    throw new Error('请先删除或移动子分类')
  }

  const index = mockCategories.findIndex(
    category => category.id === id,
  )

  if (index === -1) {
    throw new Error('分类不存在')
  }

  mockCategories.splice(index, 1)
}

export const categoryService = {
  getCategories,
  getCategoryById,
  createCategory,
  updateCategory,
  deleteCategory,
}
