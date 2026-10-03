<template>
  <div class="docs-page" :class="{ 'dark-mode': isDarkMode }">
    <header class="docs-header">
      <RouterLink to="/docs/vue-3-project-guide" class="docs-brand">
        <span class="brand-mark"><img :src="siteSettings.logoUrl" :alt="siteSettings.siteName" /></span>
        <strong>{{ siteSettings.siteName }}</strong>
        <span class="brand-subtitle">LightDoc</span>
      </RouterLink>
      <div class="header-actions">
        <form class="docs-search" @submit.prevent="handleSearch">
          <Search :size="18" /><input v-model="searchKeyword" type="search" placeholder="搜索教程..." aria-label="搜索教程" />
        </form>
        <button class="theme-toggle" :title="isDarkMode ? '切换亮色模式' : '切换暗色模式'" @click="toggleTheme">
          <Sunny v-if="!isDarkMode" :size="18" /><Moon v-else :size="18" />
        </button>
      </div>
    </header>

    <div class="docs-layout">
      <aside class="catalog-sidebar">
        <div class="sidebar-heading">教程目录</div>
        <nav>
          <section v-for="group in categoryGroups" :key="group.id" class="catalog-group">
            <button class="catalog-group-title" @click="toggleGroup(group.id)">
              <span class="group-label"><FolderOpened :size="18" />{{ group.name }}</span>
              <ArrowDown :size="16" :class="{ collapsed: !expandedGroups.has(group.id) }" />
            </button>
            <div v-show="expandedGroups.has(group.id)" class="catalog-items">
              <RouterLink v-for="item in group.children" :key="item.id" :to="`/docs/${item.slug}`" class="catalog-item" :class="{ active: item.slug === route.params.slug }">{{ item.title }}</RouterLink>
            </div>
          </section>
        </nav>
      </aside>

      <main class="document-main">
        <div v-if="!route.params.slug" class="document-blank" />
        <div v-else-if="loading" class="document-loading">正在加载文档...</div>
        <div v-else-if="!article" class="document-empty"><h1>文档不存在</h1><p>这篇文档可能已被删除，或尚未公开发布。</p></div>
        <article v-else class="document-content">
          <div class="breadcrumb"><RouterLink to="/docs/vue-3-project-guide"><House :size="16" /></RouterLink><span>/</span><span>{{ article.category }}</span><span>/</span><span>{{ article.title }}</span></div>
          <header class="article-header">
            <h1>{{ article.title }}</h1>
            <div class="article-meta"><span><Calendar :size="16" />更新于 {{ formatDate(article.updatedAt) }}</span><span><Clock :size="16" />阅读时间约 {{ readingMinutes }} 分钟</span></div>
          </header>
          <div class="article-body">
            <template v-for="block in blocks" :key="block.id">
              <h2 v-if="block.type === 'heading' && block.level === 2" :id="block.id">{{ block.text }}</h2>
              <h3 v-else-if="block.type === 'heading'" :id="block.id">{{ block.text }}</h3>
              <img v-else-if="block.type === 'image'" class="article-image" :src="block.src" :alt="block.alt" />
              <p v-else-if="block.type === 'paragraph'" v-html="block.html" />
              <ul v-else-if="block.type === 'list'"><li v-for="item in block.items" :key="item">{{ item }}</li></ul>
              <div v-else class="code-block"><div class="code-header"><span>{{ block.language || 'text' }}</span><button @click="copyCode(block.id, block.code)"><Check v-if="copiedCode === block.id" :size="14" /><DocumentCopy v-else :size="14" />{{ copiedCode === block.id ? '已复制' : '复制' }}</button></div><pre><code>{{ block.code }}</code></pre></div>
            </template>
          </div>
        </article>
      </main>

      <aside class="toc-sidebar"><div class="toc-heading">本页目录</div><nav v-if="headings.length" class="toc-list"><a v-for="heading in headings" :key="heading.id" :href="`#${heading.id}`" :class="`toc-level-${heading.level}`">{{ heading.text }}</a></nav><div v-else class="toc-empty">暂无目录</div></aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowDown, Calendar, Check, Clock, DocumentCopy, FolderOpened, House, Moon, Search, Sunny } from '@element-plus/icons-vue'
import { request, resolveApiURL } from '../../services/apiClient'
import { articleService } from '../../services/articleService'
import { categoryService } from '../../services/categoryService'
import { loadSiteSettings, siteSettings } from '../../stores/siteSettings'
import type { Article } from '../../types/article'

interface CategoryGroup { id: number; name: string; children: Array<{ id: number; slug: string; title: string }> }
type ContentBlock = { id: string; type: 'heading'; level: number; text: string } | { id: string; type: 'paragraph'; html: string } | { id: string; type: 'image'; src: string; alt: string } | { id: string; type: 'list'; items: string[] } | { id: string; type: 'code'; language: string; code: string }

const route = useRoute(); const router = useRouter(); const article = ref<Article | null>(null); const loading = ref(true); const searchKeyword = ref(''); const isDarkMode = ref(false); const copiedCode = ref(''); const categories = ref<Awaited<ReturnType<typeof categoryService.getCategories>>>([]); const navigationArticles = ref<Article[]>([]); const expandedGroups = reactive(new Set<number>())
const categoryGroups = computed<CategoryGroup[]>(() => categories.value.filter(category => category.parentId === null).map(root => ({ id: root.id, name: root.name, children: navigationArticles.value.filter(article => article.categoryId === root.id).map(article => ({ id: article.id, slug: article.slug, title: article.title })) })))
const fallbackMarkdown = (current: Article) => current.slug === 'docker-compose-guide' ? `## 一、准备环境

本文介绍如何使用 Docker Compose 快速部署 Nginx。你只需要准备一台已经安装 Docker 的服务器。

## 二、创建配置文件

创建 \`docker-compose.yml\` 文件：

\`\`\`yaml
services:
  nginx:
    image: nginx:latest
    ports:
      - "80:80"
    restart: unless-stopped
\`\`\`

## 三、启动服务

执行下面的命令即可启动：

\`\`\`bash
docker compose up -d
\`\`\`

## 四、验证部署

打开浏览器访问服务器地址，看到 Nginx 欢迎页面即表示部署完成。` : `## 一、文档概览

${current.summary || '这是一篇 LightDocs 公开文档。'}

## 二、正文内容

${current.content.replace(/^#\s+.*$/m, '').trim() || '文档内容正在整理中。'}

## 三、总结

感谢阅读本文。`
const blocks = computed<ContentBlock[]>(() => parseMarkdown(article.value ? fallbackMarkdown(article.value) : ''))
const headings = computed(() => blocks.value.filter(block => block.type === 'heading') as Array<Extract<ContentBlock, { type: 'heading' }>>)
const readingMinutes = computed(() => Math.max(1, Math.ceil(blocks.value.reduce((total, block) => total + (block.type === 'code' ? block.code.length : block.type === 'list' ? block.items.join('').length : block.type === 'heading' ? block.text.length : block.type === 'image' ? 100 : block.html.replace(/<[^>]+>/g, '').length), 0) / 420)))
const loadArticle = async () => {
  if (!route.params.slug) {
    article.value = null
    loading.value = false
    return
  }

  loading.value = true
  try { article.value = await request<Article>(`/public/docs/${encodeURIComponent(String(route.params.slug))}`) } catch { article.value = null } finally { loading.value = false }
}
const loadCategories = async () => { const [categoryResult, articleResult] = await Promise.all([categoryService.getCategories(), articleService.getArticles({ status: 'published', page: 1, pageSize: 100 })]); categories.value = categoryResult; navigationArticles.value = articleResult.list; categories.value.filter(category => category.parentId === null).forEach(category => expandedGroups.add(category.id)) }
const toggleGroup = (id: number) => expandedGroups.has(id) ? expandedGroups.delete(id) : expandedGroups.add(id)
const toggleTheme = () => { isDarkMode.value = !isDarkMode.value; document.documentElement.classList.toggle('dark-mode', isDarkMode.value) }
const handleSearch = () => { const keyword = searchKeyword.value.trim(); if (keyword) router.push({ path: '/search', query: { keyword } }) }
const formatDate = (value: string) => value.replace('T', ' ').slice(0, 10)
const copyCode = async (id: string, code: string) => { await navigator.clipboard?.writeText(code); copiedCode.value = id; window.setTimeout(() => { copiedCode.value = '' }, 1400) }
const escapeHtml = (value: string) => value.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/`([^`]+)`/g, '<code>$1</code>')
function parseMarkdown(markdown: string): ContentBlock[] {
  const lines = markdown.split('\n')
  const result: ContentBlock[] = []
  let index = 0
  let paragraph: string[] = []
  const flush = () => { if (paragraph.length) { result.push({ id: `paragraph-${result.length}`, type: 'paragraph', html: escapeHtml(paragraph.join(' ')) }); paragraph = [] } }
  while (index < lines.length) {
    const line = lines[index]
    if (line.startsWith('```')) { flush(); const language = line.slice(3).trim() || 'text'; const codeLines: string[] = []; index += 1; while (index < lines.length && !lines[index].startsWith('```')) { codeLines.push(lines[index]); index += 1 }; result.push({ id: `code-${result.length}`, type: 'code', language, code: codeLines.join('\n') }); index += 1; continue }
    const heading = /^(#{2,3})\s+(.+)$/.exec(line)
    if (heading) { flush(); result.push({ id: `heading-${result.length}`, type: 'heading', level: heading[1].length, text: heading[2] }); index += 1; continue }
    const image = /^!\[([^\]]*)\]\(([^)]+)\)$/.exec(line.trim())
    if (image) { flush(); result.push({ id: `image-${result.length}`, type: 'image', alt: image[1], src: resolveApiURL(image[2]) }); index += 1; continue }
    if (/^[-*]\s+/.test(line)) { flush(); const items: string[] = []; while (index < lines.length && /^[-*]\s+/.test(lines[index])) { items.push(lines[index].replace(/^[-*]\s+/, '')); index += 1 }; result.push({ id: `list-${result.length}`, type: 'list', items }); continue }
    if (!line.trim()) flush(); else paragraph.push(line.trim())
    index += 1
  }
  flush()
  return result
}
onMounted(async () => { loadSiteSettings(); isDarkMode.value = document.documentElement.classList.contains('dark-mode'); await Promise.all([loadCategories(), loadArticle()]) }); watch(() => route.params.slug, loadArticle)
</script>

<style scoped>
.docs-page svg {
  flex: 0 0 auto;
}

.document-blank {
  min-height: calc(100vh - 160px);
}

.article-body h2,
.article-body h3 {
  scroll-margin-top: 96px;
}

.article-image {
  display: block;
  max-width: 100%;
  max-height: 560px;
  margin: 24px auto;
  border-radius: 8px;
  object-fit: contain;
  box-shadow: 0 8px 24px rgba(30, 45, 66, 0.12);
}

.toc-sidebar {
  position: sticky;
  top: 72px;
  align-self: start;
  height: calc(100vh - 72px);
  overflow-y: auto;
}

.docs-search > svg {
  width: 18px !important;
  height: 18px !important;
  flex: 0 0 18px;
}

.theme-toggle > svg {
  width: 18px !important;
  height: 18px !important;
}

.catalog-group-title > svg {
  width: 16px !important;
  height: 16px !important;
  flex: 0 0 16px;
}

.group-label > svg {
  width: 18px !important;
  height: 18px !important;
  flex: 0 0 18px;
}

.breadcrumb svg,
.article-meta svg {
  width: 16px !important;
  height: 16px !important;
  flex: 0 0 16px;
}

.code-header button svg {
  width: 14px !important;
  height: 14px !important;
  flex: 0 0 14px;
}

.docs-page{min-height:100vh;color:#1f2d45;background:#f8fafc}.docs-header{position:sticky;top:0;z-index:20;display:flex;align-items:center;justify-content:space-between;height:72px;padding:0 42px;border-bottom:1px solid #e5ebf3;background:rgba(255,255,255,.94);backdrop-filter:blur(12px)}.docs-brand,.header-actions,.docs-search,.theme-toggle,.breadcrumb,.article-meta,.article-meta span,.catalog-group-title,.group-label{display:flex;align-items:center}.docs-brand{gap:10px;color:#13233d}.docs-brand strong{font-size:22px}.brand-subtitle{color:#8290a7;font-size:16px}.brand-mark{display:grid;width:34px;height:34px;place-items:center;overflow:hidden;border-radius:7px;background:#2e83ed}.brand-mark img{width:26px;height:26px;object-fit:contain}.header-actions{gap:18px}.docs-search{width:305px;height:40px;gap:10px;padding:0 14px;border:1px solid #dbe4ef;border-radius:12px;color:#72809a;background:#f7f9fc}.docs-search input{width:100%;border:0;outline:0;color:#253a59;background:transparent;font-size:14px}.docs-search input::placeholder{color:#9aa8bc}.theme-toggle{justify-content:center;width:42px;height:42px;border:1px solid #dbe4ef;border-radius:50%;color:#263a58;background:#fff;cursor:pointer}.docs-layout{display:grid;grid-template-columns:300px minmax(0,1fr) 300px;min-height:calc(100vh - 72px)}.catalog-sidebar,.toc-sidebar{padding:30px 32px;background:#f8fafc}.catalog-sidebar{border-right:1px solid #e1e8f1}.toc-sidebar{border-left:1px solid #e1e8f1}.sidebar-heading,.toc-heading{margin-bottom:24px;color:#152943;font-size:20px;font-weight:700}.catalog-group{margin-bottom:24px}.catalog-group-title{width:100%;justify-content:space-between;padding:8px 0;border:0;color:#172a47;background:transparent;cursor:pointer;font-size:16px;font-weight:700}.group-label{gap:10px}.catalog-group-title svg:last-child{transition:transform .18s ease}.catalog-group-title svg.collapsed{transform:rotate(-90deg)}.catalog-items{display:flex;flex-direction:column;gap:2px;margin:4px 0 0 8px;padding-left:16px;border-left:1px solid #dce5f0}.catalog-item{padding:9px 12px;border-radius:6px;color:#52647f;font-size:14px}.catalog-item:hover,.catalog-item.active{color:#1674e8;background:#eaf3ff}.document-main{min-width:0;padding:28px 30px 100px;background:#fff}.document-content{max-width:900px;margin:0 auto}.breadcrumb{gap:9px;color:#70819b;font-size:14px}.breadcrumb a{display:inline-flex;color:#61738e}.article-header{padding:30px 0 28px;border-bottom:1px solid #e2e8f0}.article-header h1{margin:0 0 20px;color:#10233f;font-size:42px;line-height:1.2}.article-meta{gap:22px;color:#7788a1;font-size:14px}.article-meta span{gap:7px}.article-body{padding-top:36px;color:#40526d;font-size:17px;line-height:1.85}.article-body h2{margin:35px 0 13px;color:#10233f;font-size:28px;line-height:1.35}.article-body h3{margin:26px 0 12px;color:#1c3455;font-size:22px}.article-body p{margin:0 0 22px}.article-body :deep(code){padding:2px 6px;border-radius:4px;color:#263b59;background:#eef3f8;font-size:.9em}.article-body ul{margin:0 0 22px;padding-left:26px}.code-block{margin:22px 0 30px;overflow:hidden;border-radius:7px;color:#d8e4f4;background:#202a38;box-shadow:0 10px 24px rgba(30,45,66,.14)}.code-header{display:flex;align-items:center;justify-content:space-between;padding:9px 14px;color:#d8e4f4;background:#303b4b;font-size:13px}.code-header button{display:inline-flex;align-items:center;gap:5px;border:0;color:#d8e4f4;background:transparent;cursor:pointer}.code-block pre{margin:0;padding:17px 20px 20px;overflow-x:auto;font-family:Consolas,Monaco,monospace;font-size:14px;line-height:1.65}.toc-list{display:flex;flex-direction:column;gap:8px;padding-left:14px;border-left:1px solid #dce5f0}.toc-list a{padding:2px 0 2px 8px;color:#536783;font-size:14px;line-height:1.6}.toc-list a:hover,.toc-level-2{color:#1674e8!important}.toc-level-3{padding-left:18px!important}.toc-empty,.document-loading,.document-empty{color:#8493a9}.document-empty{padding:120px 20px;text-align:center}.document-empty h1{color:#263b59}.docs-page.dark-mode{color:#dbeafe;background:#0b1220}.docs-page.dark-mode .docs-header,.docs-page.dark-mode .catalog-sidebar,.docs-page.dark-mode .toc-sidebar{border-color:#293a53;background:#111b2b}.docs-page.dark-mode .docs-brand,.docs-page.dark-mode .sidebar-heading,.docs-page.dark-mode .toc-heading,.docs-page.dark-mode .catalog-group-title,.docs-page.dark-mode .article-header h1,.docs-page.dark-mode .article-body h2,.docs-page.dark-mode .article-body h3{color:#f1f5f9}.docs-page.dark-mode .brand-subtitle,.docs-page.dark-mode .article-meta,.docs-page.dark-mode .article-body,.docs-page.dark-mode .catalog-item,.docs-page.dark-mode .toc-list a{color:#a9b8cc}.docs-page.dark-mode .docs-search,.docs-page.dark-mode .theme-toggle{border-color:#3b4d68;color:#dbeafe;background:#1b293c}.docs-page.dark-mode .docs-search input{color:#e2e8f0}.docs-page.dark-mode .catalog-item:hover,.docs-page.dark-mode .catalog-item.active{color:#93c5fd;background:#1b3a62}.docs-page.dark-mode .document-main{background:#0f172a}.docs-page.dark-mode .breadcrumb,.docs-page.dark-mode .breadcrumb a{color:#91a4bf}.docs-page.dark-mode .article-header{border-color:#2b3b55}.docs-page.dark-mode .article-body{color:#c8d5e6}.docs-page.dark-mode .article-body :deep(code){color:#bfdbfe;background:#1e3049}.docs-page.dark-mode .toc-list{border-color:#334761}.docs-page.dark-mode .toc-list a:hover,.docs-page.dark-mode .toc-level-2{color:#7db4ff!important}@media (max-width:1100px){.docs-layout{grid-template-columns:240px minmax(0,1fr)}.toc-sidebar{display:none}}@media (max-width:720px){.docs-header{height:auto;min-height:66px;gap:14px;padding:12px 18px}.brand-subtitle{display:none}.docs-search{width:180px}.docs-layout{display:block}.catalog-sidebar{display:none}.document-main{padding:20px 18px 70px}.article-header h1{font-size:32px}}
</style>
