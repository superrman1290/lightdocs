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
  await delay()

  const index = mockRecycleItems.findIndex(item => item.id === id)

  if (index === -1) {
    throw new Error('回收站项目不存在')
  }

  const [item] = mockRecycleItems.splice(index, 1)
  return item
}

const deletePermanently = async (ids: string[]): Promise<void> => {
  await delay()

  const selectedIds = new Set(ids)

  for (let index = mockRecycleItems.length - 1; index >= 0; index -= 1) {
    if (selectedIds.has(mockRecycleItems[index].id)) {
      mockRecycleItems.splice(index, 1)
    }
  }
}

const clear = async (): Promise<void> => {
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
