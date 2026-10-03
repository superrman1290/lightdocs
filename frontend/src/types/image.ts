export type ImageSource = 'upload' | 'article' | 'system'

export interface ImageAsset {
  id: number
  name: string
  url: string
  size: number
  sizeBytes?: number
  mimeType?: string
  storageKey?: string
  width?: number
  height?: number
  source: ImageSource
  createdAt: string
  referenced?: boolean
}

export interface ImageQuery {
  keyword?: string
  source?: ImageSource
  page?: number
  pageSize?: number
}

export interface ImagePageResult {
  list: ImageAsset[]
  total: number
  page: number
  pageSize: number
}
