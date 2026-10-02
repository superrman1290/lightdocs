import { mockImages } from '../mock/images'
import { recycleService } from './recycleService'
import { request } from './apiClient'

import type {
  ImageAsset,
  ImagePageResult,
  ImageQuery,
} from '../types/image'

const delay = (ms = 160) => {
  return new Promise(resolve => {
    setTimeout(resolve, ms)
  })
}

const getImages = async (
  query: ImageQuery = {},
): Promise<ImagePageResult> => {
  try { const params = new URLSearchParams(); if(query.keyword) params.set('keyword',query.keyword); if(query.source) params.set('source',query.source); params.set('page',String(query.page||1)); params.set('pageSize',String(query.pageSize||15)); return await request<ImagePageResult>(`/images?${params}`) } catch (error) { if ((error as {status?:number}).status !== 404) throw error }
  await delay()

  let result = [...mockImages]

  if (query.source) {
    result = result.filter(image => image.source === query.source)
  }

  if (query.keyword?.trim()) {
    const keyword = query.keyword.trim().toLowerCase()

    result = result.filter(image => {
      return image.name.toLowerCase().includes(keyword)
    })
  }

  const page = query.page && query.page > 0 ? query.page : 1
  const pageSize = query.pageSize && query.pageSize > 0
    ? query.pageSize
    : 15
  const total = result.length
  const start = (page - 1) * pageSize

  return {
    list: result.slice(start, start + pageSize),
    total,
    page,
    pageSize,
  }
}

const deleteImages = async (ids: number[]): Promise<void> => {
  try { await request<void>(`/images?ids=${ids.join(',')}`, { method: 'DELETE' }); return } catch (error) { if ((error as {status?:number}).status !== 404) throw error }
  await delay()

  const selectedIds = new Set(ids)

  for (let index = mockImages.length - 1; index >= 0; index -= 1) {
    if (selectedIds.has(mockImages[index].id)) {
      const [image] = mockImages.splice(index, 1)
      await recycleService.addImage(image)
    }
  }
}

const restoreImage = async (
  recycleId: string,
): Promise<ImageAsset> => {
  const item = await recycleService.restoreItem(recycleId)

  if (item.type !== 'image') {
    throw new Error('回收站项目类型错误')
  }

  mockImages.unshift(item.data)
  return item.data
}

const uploadImages = async (files: File[]): Promise<ImageAsset[]> => {
  try { const form = new FormData(); files.forEach(file => form.append('files',file)); return await request<ImageAsset[]>('/images',{method:'POST',body:form}) } catch (error) { if ((error as {status?:number}).status !== 404) throw error }
  await delay()

  const now = new Date()
    .toISOString()
    .slice(0, 16)
    .replace('T', ' ')

  const uploaded = files.map((file, index) => ({
    id: Math.max(0, ...mockImages.map(image => image.id)) + index + 1,
    name: file.name,
    url: URL.createObjectURL(file),
    size: Math.max(1, Math.round(file.size / 1024)),
    source: 'upload' as const,
    createdAt: now,
  }))

  mockImages.unshift(...uploaded)

  return uploaded
}

export const imageService = {
  getImages,
  deleteImages,
  restoreImage,
  uploadImages,
}
