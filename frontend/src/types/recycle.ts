import type { Article } from './article'
import type { ImageAsset } from './image'

export type RecycleItemType = 'article' | 'image'

export interface RecycleArticleItem {
  id: string
  type: 'article'
  data: Article
  deletedAt: string
}

export interface RecycleImageItem {
  id: string
  type: 'image'
  data: ImageAsset
  deletedAt: string
}

export type RecycleItem = RecycleArticleItem | RecycleImageItem
