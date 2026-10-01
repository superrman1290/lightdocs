import { mockArticles } from './articles'
import { mockImages } from './images'

import type { RecycleItem } from '../types/recycle'

const cloneArticle = (article: typeof mockArticles[number]) => ({
  ...article,
  tags: [...article.tags],
})

const recycleDemoArticles = [
  ...mockArticles.slice(0, 7),
  {
    ...mockArticles[0],
    id: 100,
    title: 'Linux 系统基础命令详解',
    slug: 'linux-command-basics',
    categoryId: 2,
    category: '后端开发',
    summary: 'Linux 常用基础命令和文件管理操作。',
  },
]

const articleItems: RecycleItem[] = recycleDemoArticles
  .map((article, index) => ({
    id: `article-demo-${index + 1}`,
    type: 'article' as const,
    data: cloneArticle(article),
    deletedAt: article.updatedAt,
  }))

const imageItems: RecycleItem[] = mockImages
  .slice(0, 4)
  .map((image, index) => ({
    id: `image-demo-${index + 1}`,
    type: 'image' as const,
    data: { ...image },
    deletedAt: image.createdAt,
  }))

export const mockRecycleItems: RecycleItem[] = [
  ...articleItems,
  ...imageItems,
]
