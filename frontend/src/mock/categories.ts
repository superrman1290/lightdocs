import type { Category } from '../types/category'

/**
 * LightDocs 分类 Mock 数据。
 *
 * 文章列表和编辑器共用这一份数据；后续接入 Go API 后，
 * 可以保留相同的 service 调用方式。
 */
export const mockCategories: Category[] = [
  { id: 1, name: '前端开发', parentId: null, sort: 1 },
  { id: 6, name: 'Vue', parentId: 1, sort: 1 },
  { id: 7, name: 'Vue Router', parentId: 1, sort: 2 },
  { id: 2, name: '后端开发', parentId: null, sort: 2 },
  { id: 8, name: 'Go', parentId: 2, sort: 1 },
  { id: 9, name: 'Gin', parentId: 2, sort: 2 },
  { id: 3, name: 'Docker', parentId: null, sort: 3 },
  { id: 10, name: 'Compose', parentId: 3, sort: 1 },
  { id: 4, name: '数据库', parentId: null, sort: 4 },
  { id: 11, name: 'PostgreSQL', parentId: 4, sort: 1 },
  { id: 5, name: '项目开发', parentId: null, sort: 5 },
  { id: 12, name: 'LightDocs', parentId: 5, sort: 1 },
]
