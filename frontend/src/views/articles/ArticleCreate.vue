<template>
  <ArticleEditor />
</template>

<script setup lang="ts">
/**
 * 兼容保留：新建文章路由当前直接使用 ArticleEditor。
 *
 * 保留该包装组件可避免旧引用重新引入另一套表单和数据逻辑。
 */
import ArticleEditor from '../../components/article/ArticleEditor.vue'
</script>
