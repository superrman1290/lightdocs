import type {
  Article,
  ArticlePageResult,
  ArticleQuery,
  CreateArticleInput,
  UpdateArticleInput,
} from '../types/article'

import { mockArticles } from '../mock/articles'
import { categoryService } from './categoryService'
import { recycleService } from './recycleService'
import { request } from './apiClient'


/**
 * Article Service
 *
 * 负责处理文章相关的数据操作。
 *
 * 当前：
 * Service → Mock Data
 *
 * 未来：
 * Service → HTTP API → Go → PostgreSQL
 */


/**
 * 模拟网络请求
 *
 * 让 Mock Service 的行为更接近真实 API。
 */
const delay = (ms = 200) => {
  return new Promise(resolve => {
    setTimeout(resolve, ms)
  })
}

const getNow = () => new Date()
  .toISOString()
  .slice(0, 19)
  .replace('T', ' ')


/**
 * 获取文章列表
 */
const getArticles = async (
  query: ArticleQuery = {},
): Promise<ArticlePageResult> => {

  const params = new URLSearchParams()
  if (query.status) params.set('status', query.status)
  if (query.keyword) params.set('keyword', query.keyword)
  if (query.categoryId) params.set('categoryId', String(query.categoryId))
  params.set('page', String(query.page || 1))
  params.set('pageSize', String(query.pageSize || 10))

  try {
    return await request<ArticlePageResult>(`/articles?${params.toString()}`)
  } catch (error) {
    if ((error as { status?: number }).status !== 0 && (error as { status?: number }).status !== undefined) throw error
  }

  await delay()

  let result = [...mockArticles]


  /* 状态筛选 */

  if (query.status) {

    result = result.filter(
      article =>
        article.status === query.status
    )

  }


  /* 分类筛选 */

  if (query.categoryId) {

    result = result.filter(
      article =>
        article.categoryId === query.categoryId
    )

  }


  /* 关键词搜索 */

  if (query.keyword?.trim()) {

    const keyword =
      query.keyword
        .trim()
        .toLowerCase()

    result = result.filter(article => {

      const titleMatch =
        article.title
          .toLowerCase()
          .includes(keyword)

      const summaryMatch =
        article.summary
          .toLowerCase()
          .includes(keyword)

      const contentMatch =
        article.content
          .toLowerCase()
          .includes(keyword)

      const tagMatch =
        article.tags.some(tag =>
          tag.toLowerCase().includes(keyword)
        )

      return titleMatch || summaryMatch || contentMatch || tagMatch

    })

  }


  /* 分页 */

  const page =
    query.page && query.page > 0
      ? query.page
      : 1

  const pageSize =
    query.pageSize && query.pageSize > 0
      ? query.pageSize
      : 10

  const total = result.length

  const start =
    (page - 1) * pageSize

  const end =
    start + pageSize

  const list =
    result.slice(start, end)


  return {
    list,
    total,
    page,
    pageSize,
  }
}


/**
 * 根据 ID 获取文章
 */
const getArticleById = async (
  id: number,
): Promise<Article | null> => {

  try {
    return await request<Article>(`/articles/${id}`)
  } catch (error) {
    if ((error as { status?: number }).status !== 404) throw error
  }

  await delay()

  return (
    mockArticles.find(
      article =>
        article.id === id
    ) ?? null
  )

}


/**
 * 创建文章
 */
const createArticle = async (
  data: CreateArticleInput,
): Promise<Article> => {

  try {
    return await request<Article>('/articles', { method: 'POST', body: JSON.stringify(data) })
  } catch (error) {
    if ((error as { status?: number }).status !== 404) throw error
  }

  await delay()

  const category = await categoryService.getCategoryById(
    data.categoryId,
  )

  if (!category) {
    throw new Error('分类不存在')
  }

  const now =
    new Date()
      .toISOString()
      .slice(0, 19)
      .replace('T', ' ')


  const newArticle: Article = {

    ...data,

    publishedAt:
      data.status === 'published'
        ? now
        : undefined,

    category: category.name,

    id:
      Math.max(
        0,
        ...mockArticles.map(
          article => article.id
        ),
      ) + 1,

    createdAt: now,
    updatedAt: now,

  }


  mockArticles.unshift(newArticle)

  return newArticle
}


/**
 * 更新文章
 */
const updateArticle = async (
  id: number,
  data: UpdateArticleInput,
): Promise<Article> => {

  try {
    return await request<Article>(`/articles/${id}`, { method: 'PATCH', body: JSON.stringify(data) })
  } catch (error) {
    if ((error as { status?: number }).status !== 404) throw error
  }

  await delay()

  const index =
    mockArticles.findIndex(
      article =>
        article.id === id
    )

  if (index === -1) {
    throw new Error('文章不存在')
  }

  const category =
    data.categoryId === undefined
      ? null
      : await categoryService.getCategoryById(
          data.categoryId,
        )

  if (data.categoryId !== undefined && !category) {
    throw new Error('分类不存在')
  }


  mockArticles[index] = {
    ...mockArticles[index],
    ...data,
    publishedAt:
      data.status === 'published'
        ? getNow()
        : data.status === 'draft'
          ? undefined
          : mockArticles[index].publishedAt,
    ...(category ? { category: category.name } : {}),
    updatedAt: getNow(),
  }


  return mockArticles[index]
}


/**
 * 删除文章
 */
const deleteArticle = async (
  id: number,
): Promise<void> => {

  try {
    await request<void>(`/articles/${id}`, { method: 'DELETE' })
    return
  } catch (error) {
    if ((error as { status?: number }).status !== 404) throw error
  }

  await delay()

  const index =
    mockArticles.findIndex(
      article =>
        article.id === id
    )

  if (index === -1) {
    throw new Error('文章不存在')
  }


  const [article] = mockArticles.splice(index, 1)

  await recycleService.addArticle(article)
}

const getPublishedArticleBySlug = async (
  slug: string,
): Promise<Article | null> => {
  await delay()

  return (
    mockArticles.find(article => {
      return article.slug === slug && article.status === 'published'
    }) ?? null
  )
}

const restoreArticle = async (
  recycleId: string,
): Promise<Article> => {
  const item = await recycleService.restoreItem(recycleId)

  if (item.type !== 'article') {
    throw new Error('回收站项目类型错误')
  }

  mockArticles.unshift(item.data)
  return item.data
}


/**
 * 导出文章
 *
 * 当前阶段只保留接口。
 */
const exportArticle = async (
  id: number,
  format: 'markdown' | 'html' | 'pdf',
): Promise<void> => {

  await delay()

  console.log(
    `Export article ${id} as ${format}`,
  )
}


/**
 * 统一导出 Article Service
 */
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
