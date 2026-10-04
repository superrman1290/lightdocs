import { request } from './apiClient'
import type { RecycleItem, RecycleItemType } from '../types/recycle'

const getItems = async (type?: RecycleItemType, keyword?: string): Promise<RecycleItem[]> => {
  const params = new URLSearchParams({ page: '1', pageSize: '100' })
  if (type) params.set('type', type)
  if (keyword) params.set('keyword', keyword)
  const result = await request<{ list: RecycleItem[] }>(`/recycle-bin?${params}`)
  return result.list
}

const restoreItem = async (id: string): Promise<RecycleItem> => request<RecycleItem>(`/recycle-bin/${id}`, {
  method: 'PATCH',
  body: JSON.stringify({ action: 'restore' }),
})

const deletePermanently = async (ids: string[]): Promise<void> => {
  await request<void>(`/recycle-bin?ids=${ids.join(',')}`, { method: 'DELETE' })
}

export const recycleService = {
  getItems,
  restoreItem,
  deletePermanently,
}
