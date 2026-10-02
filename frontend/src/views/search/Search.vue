<template>
  <div class="search-page">
    <h1>搜索文档</h1>
    <el-input v-model="keyword" autofocus placeholder="搜索标题、摘要、正文或标签" @keyup.enter="search" />
    <el-empty v-if="searched && results.length === 0" description="没有找到相关文章" />
    <div v-for="item in results" :key="item.id" class="result-item" @click="router.push(`/docs/${item.slug}`)">
      <h2>{{ item.title }}</h2><p>{{ item.summary }}</p><span>{{ item.category }} · {{ item.updatedAt }}</span>
    </div>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { request } from '../../services/apiClient'
import type { ArticlePageResult } from '../../types/article'
const route=useRoute();const router=useRouter();const keyword=ref(String(route.query.keyword||''));const results=ref<ArticlePageResult['list']>([]);const searched=ref(false)
const search=async()=>{if(!keyword.value.trim())return;const result=await request<ArticlePageResult>(`/public/search?keyword=${encodeURIComponent(keyword.value)}&page=1&pageSize=20`);results.value=result.list;searched.value=true}
onMounted(search)
</script>
<style scoped>
.search-page{max-width:900px;margin:0 auto;padding:48px 24px}.search-page h1{color:#1f2d5a}.search-page>.el-input{margin:20px 0 30px}.result-item{padding:18px 0;border-bottom:1px solid #e5e7eb;cursor:pointer}.result-item h2{margin:0 0 8px;color:#2563eb;font-size:20px}.result-item p{margin:0 0 8px;color:#64748b}.result-item span{color:#94a3b8;font-size:13px}.dark-mode .result-item{border-color:#374151}.dark-mode .result-item p,.dark-mode .result-item span{color:#9ca3af}.dark-mode .search-page h1{color:#f3f4f6}
</style>
