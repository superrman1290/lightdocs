/**
 * 文章状态
 */
export type ArticleStatus = 'published' | 'draft'


/**
 * 文章
 */
export interface Article {
  id: number

  /**
   * 文章标题
   */
  title: string

  /**
   * 文章唯一标识
   *
   * 后续用于前台 URL：
   * /docs/vue-3-guide
   */
  slug: string

  /**
   * 分类 ID
   */
  categoryId: number

  /**
   * 分类名称
   *
   * 接口返回的分类名称，方便列表直接显示
   */
  category: string

  /**
   * 文章状态
   */
  status: ArticleStatus

  /**
   * 标签
   */
  tags: string[]

  /**
   * Markdown 正文
   *
   * 当前阶段可以为空
   */
  content: string

  /**
   * 摘要
   */
  summary: string

  /**
   * 创建时间
   */
  createdAt: string

  /**
   * 更新时间
   */
  updatedAt: string

  /**
   * 最近一次发布时间；草稿为空。
   */
  publishedAt?: string
}

/**
 * 文章新建和编辑共用的表单模型。
 *
 * 该模型只包含用户可编辑字段，不包含服务端生成的字段。
 */
export interface ArticleForm {
  title: string
  slug: string
  categoryId: number
  tags: string[]
  content: string
  summary: string
}

/**
 * 新建文章时提交的数据。
 */
export interface CreateArticleInput extends ArticleForm {
  status: ArticleStatus
}

/**
 * 编辑文章时提交的数据。
 */
export interface UpdateArticleInput extends Partial<ArticleForm> {
  status?: ArticleStatus
}

/**
 * 文章列表查询条件。
 */
export interface ArticleQuery {
  status?: ArticleStatus
  keyword?: string
  categoryId?: number
  page?: number
  pageSize?: number
}

/**
 * 文章列表分页结果。
 */
export interface ArticlePageResult {
  list: Article[]
  total: number
  page: number
  pageSize: number
}
