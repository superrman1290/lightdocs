import type { Article, ArticlePageResult, ArticleQuery, CreateArticleInput, UpdateArticleInput } from '../types/article'
import { apiBaseURL, getAccessToken, request } from './apiClient'

const getArticles = async (query: ArticleQuery = {}): Promise<ArticlePageResult> => {
  const params = new URLSearchParams()
  if (query.status) params.set('status', query.status)
  if (query.keyword) params.set('keyword', query.keyword)
  if (query.categoryId) params.set('categoryId', String(query.categoryId))
  params.set('page', String(query.page || 1))
  params.set('pageSize', String(query.pageSize || 10))
  return request<ArticlePageResult>(`/articles?${params}`)
}

const getArticleById = (id: number) => request<Article>(`/articles/${id}`)

const getPublishedArticleBySlug = async (slug: string) => {
  try {
    return await request<Article>(`/public/docs/${encodeURIComponent(slug)}`)
  } catch (error) {
    if ((error as { status?: number }).status === 404) return null
    throw error
  }
}

const createArticle = (data: CreateArticleInput) => request<Article>('/articles', {
  method: 'POST',
  body: JSON.stringify(data),
})

const updateArticle = (id: number, data: UpdateArticleInput) => request<Article>(`/articles/${id}`, {
  method: 'PATCH',
  body: JSON.stringify(data),
})

const deleteArticle = async (id: number): Promise<void> => {
  await request<void>(`/articles/${id}`, { method: 'DELETE' })
}

const restoreArticle = async (recycleId: string): Promise<Article> => {
  const item = await request<{ type: string; data: Article }>(`/recycle-bin/${recycleId}`, {
    method: 'PATCH',
    body: JSON.stringify({ action: 'restore' }),
  })
  if (item.type !== 'article') throw new Error('回收站项目类型错误')
  return item.data
}

const exportArticle = async (id: number, format: 'markdown' | 'html' | 'pdf'): Promise<void> => {
  const token = getAccessToken()
  const response = await fetch(`${apiBaseURL}/articles/${id}/export?format=${format}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  })
  if (!response.ok) throw new Error('导出文章失败')
  const blob = await response.blob()
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `article-${id}.${format === 'markdown' ? 'md' : format}`
  link.click()
  URL.revokeObjectURL(url)
}

export const articleService = {
  getArticles,
  getArticleById,
  getPublishedArticleBySlug,
  createArticle,
  updateArticle,
  deleteArticle,
  restoreArticle,
  exportArticle,
}
