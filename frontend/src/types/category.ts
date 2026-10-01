/**
 * 文章分类。
 */
export interface Category {
  id: number
  name: string
  parentId: number | null
  sort: number
}

export interface CategoryInput {
  name: string
  parentId: number | null
  sort: number
}
