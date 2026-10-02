import { mockRecycleItems } from '../mock/recycle'

import type {
  Article,
} from '../types/article'

import type {
  ImageAsset,
} from '../types/image'

import type {
  RecycleArticleItem,
  RecycleImageItem,
  RecycleItem,
  RecycleItemType,
} from '../types/recycle'
import { request } from './apiClient'

const delay = (ms = 120) => {
  return new Promise(resolve => {
    setTimeout(resolve, ms)
  })
}

const now = () => new Date()
  .toISOString()
  .slice(0, 16)
  .replace('T', ' ')

const addArticle = async (article: Article): Promise<RecycleArticleItem> => {
  await delay()

  const item: RecycleArticleItem = {
    id: `article-${article.id}-${Date.now()}`,
    type: 'article',
    data: {
      ...article,
      tags: [...article.tags],
    },
    deletedAt: now(),
  }

  mockRecycleItems.unshift(item)
  return item
}

const addImage = async (image: ImageAsset): Promise<RecycleImageItem> => {
  await delay()

  const item: RecycleImageItem = {
    id: `image-${image.id}-${Date.now()}`,
    type: 'image',
    data: { ...image },
    deletedAt: now(),
  }

  mockRecycleItems.unshift(item)
  return item
}

const getItems = async (
  type?: RecycleItemType,
  keyword?: string,
): Promise<RecycleItem[]> => {
  try { const params = new URLSearchParams(); if(type) params.set('type',type); if(keyword) params.set('keyword',keyword); params.set('page','1'); params.set('pageSize','100'); const result=await request<{list:RecycleItem[]}>('/recycle-bin?'+params); return result.list } catch(error) { if ((error as {status?:number}).status !== 404) throw error }
  await delay()

  const normalizedKeyword = keyword?.trim().toLowerCase()

  return mockRecycleItems.filter(item => {
    if (type && item.type !== type) {
      return false
    }

    if (!normalizedKeyword) {
      return true
    }

    const text = item.type === 'article'
      ? `${item.data.title} ${item.data.category}`
      : `${item.data.name} ${item.data.source}`

    return text.toLowerCase().includes(normalizedKeyword)
  })
}

const restoreItem = async (id: string): Promise<RecycleItem> => {
  try { return await request<RecycleItem>(`/recycle-bin/${id}`,{method:'PATCH',body:JSON.stringify({action:'restore'})}) } catch(error) { if ((error as {status?:number}).status !== 404) throw error }
  await delay()

  const index = mockRecycleItems.findIndex(item => item.id === id)

  if (index === -1) {
    throw new Error('回收站项目不存在')
  }

  const [item] = mockRecycleItems.splice(index, 1)
  return item
}

const deletePermanently = async (ids: string[]): Promise<void> => {
  try { await request<void>(`/recycle-bin?ids=${ids.join(',')}`,{method:'DELETE'}); return } catch(error) { if ((error as {status?:number}).status !== 404) throw error }
  await delay()

  const selectedIds = new Set(ids)

  for (let index = mockRecycleItems.length - 1; index >= 0; index -= 1) {
    if (selectedIds.has(mockRecycleItems[index].id)) {
      mockRecycleItems.splice(index, 1)
    }
  }
}

const clear = async (): Promise<void> => {
  try { await request<void>('/recycle-bin',{method:'DELETE'}); return } catch(error) { if ((error as {status?:number}).status !== 404) throw error }
  await delay()
  mockRecycleItems.splice(0, mockRecycleItems.length)
}

export const recycleService = {
  addArticle,
  addImage,
  getItems,
  restoreItem,
  deletePermanently,
  clear,
}
