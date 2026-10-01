import type { Article } from '../types/article'


/**
 * LightDocs 文章 Mock 数据
 *
 * 当前阶段用于前端开发。
 * 后续接入 Go API 后，这部分可以移除。
 */
export const mockArticles: Article[] = [
  {
    id: 1,
    title: 'Vue 3 项目搭建指南',
    slug: 'vue-3-project-guide',
    categoryId: 1,
    category: '前端开发',
    status: 'published',
    tags: ['Vue', '前端'],
    content: '# Vue 3 项目搭建指南',
    summary: '介绍如何从零搭建 Vue 3 项目。',
    createdAt: '2026-09-28 10:00',
    updatedAt: '2026-09-30 10:32',
  },

  {
    id: 2,
    title: 'Docker Compose 入门',
    slug: 'docker-compose-guide',
    categoryId: 3,
    category: 'Docker',
    status: 'published',
    tags: ['Docker', '部署'],
    content: '# Docker Compose 入门',
    summary: '介绍 Docker Compose 的基本使用方法。',
    createdAt: '2026-09-27 09:20',
    updatedAt: '2026-09-29 16:20',
  },

  {
    id: 3,
    title: 'Go Gin Web 开发',
    slug: 'go-gin-web-development',
    categoryId: 2,
    category: '后端开发',
    status: 'draft',
    tags: ['Go', 'Gin'],
    content: '# Go Gin Web 开发',
    summary: '使用 Go 和 Gin 构建 Web 应用。',
    createdAt: '2026-09-26 14:00',
    updatedAt: '2026-09-28 14:15',
  },

  {
    id: 4,
    title: 'PostgreSQL 基础使用',
    slug: 'postgresql-basic',
    categoryId: 4,
    category: '数据库',
    status: 'published',
    tags: ['PostgreSQL', '数据库'],
    content: '# PostgreSQL 基础使用',
    summary: '介绍 PostgreSQL 数据库的基本操作。',
    createdAt: '2026-09-25 10:00',
    updatedAt: '2026-09-27 11:08',
  },

  {
    id: 5,
    title: 'LightDocs 项目架构设计',
    slug: 'lightdocs-architecture',
    categoryId: 5,
    category: '项目开发',
    status: 'published',
    tags: ['LightDocs', '架构'],
    content: '# LightDocs 项目架构设计',
    summary: '介绍 LightDocs 的整体项目架构。',
    createdAt: '2026-09-24 15:00',
    updatedAt: '2026-09-26 18:42',
  },

  {
    id: 6,
    title: 'Vue Router 路由配置',
    slug: 'vue-router-guide',
    categoryId: 1,
    category: '前端开发',
    status: 'draft',
    tags: ['Vue', 'Router'],
    content: '# Vue Router 路由配置',
    summary: '介绍 Vue Router 的基本配置方法。',
    createdAt: '2026-09-23 09:00',
    updatedAt: '2026-09-25 09:30',
  },

  {
    id: 7,
    title: 'Element Plus 使用指南',
    slug: 'element-plus-guide',
    categoryId: 1,
    category: '前端开发',
    status: 'published',
    tags: ['Vue', 'Element Plus'],
    content: '# Element Plus 使用指南',
    summary: '介绍 Element Plus 常用组件的使用方法。',
    createdAt: '2026-09-22 13:00',
    updatedAt: '2026-09-24 15:12',
  },
]