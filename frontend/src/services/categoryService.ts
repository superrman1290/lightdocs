import type { Category, CategoryInput } from '../types/category'
import { request } from './apiClient'

const getCategories = () => request<Category[]>('/categories')

const getCategoryById = async (id: number): Promise<Category | null> => {
  const categories = await getCategories()
  return categories.find(category => category.id === id) ?? null
}

const createCategory = (data: CategoryInput) => request<Category>('/categories', {
  method: 'POST',
  body: JSON.stringify(data),
})

const updateCategory = (id: number, data: CategoryInput) => request<Category>(`/categories/${id}`, {
  method: 'PATCH',
  body: JSON.stringify(data),
})

const deleteCategory = async (id: number): Promise<void> => {
  await request<void>(`/categories/${id}`, { method: 'DELETE' })
}

export const categoryService = {
  getCategories,
  getCategoryById,
  createCategory,
  updateCategory,
  deleteCategory,
}
