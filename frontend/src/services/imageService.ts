import { request, resolveApiURL } from './apiClient'
import type { ImageAsset, ImagePageResult, ImageQuery } from '../types/image'

const getImages = async (query: ImageQuery = {}): Promise<ImagePageResult> => {
  const params = new URLSearchParams()
  if (query.keyword) params.set('keyword', query.keyword)
  if (query.source) params.set('source', query.source)
  params.set('page', String(query.page || 1))
  params.set('pageSize', String(query.pageSize || 15))
  const result = await request<ImagePageResult>(`/images?${params}`)
  return { ...result, list: result.list.map(image => ({ ...image, url: resolveApiURL(image.url) })) }
}

const deleteImages = async (ids: number[]): Promise<void> => {
  await request<void>(`/images?ids=${ids.join(',')}`, { method: 'DELETE' })
}

const restoreImage = async (recycleId: string): Promise<ImageAsset> => {
  const item = await request<{ type: string; data: ImageAsset }>(`/recycle-bin/${recycleId}`, {
    method: 'PATCH',
    body: JSON.stringify({ action: 'restore' }),
  })
  if (item.type !== 'image') throw new Error('回收站项目类型错误')
  return { ...item.data, url: resolveApiURL(item.data.url) }
}

const uploadImages = async (files: File[]): Promise<ImageAsset[]> => {
  const form = new FormData()
  files.forEach(file => form.append('files', file))
  const result = await request<ImageAsset[]>('/images', { method: 'POST', body: form })
  return result.map(image => ({ ...image, url: resolveApiURL(image.url) }))
}

export const imageService = {
  getImages,
  deleteImages,
  restoreImage,
  uploadImages,
}
