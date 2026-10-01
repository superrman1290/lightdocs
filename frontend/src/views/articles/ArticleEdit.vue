<template>

  <ArticleEditor
    v-if="article"
    :article="article"
  />

  <div
    v-else
    class="loading-container"
  >
    <el-skeleton
      :rows="8"
      animated
    />
  </div>

</template>


<script setup lang="ts">

import {
  onMounted,
  ref,
} from 'vue'

import {
  useRoute,
  useRouter,
} from 'vue-router'

import {
  ElMessage,
} from 'element-plus'

import ArticleEditor from '../../components/article/ArticleEditor.vue'

import {
  articleService,
} from '../../services/articleService'

import type {
  Article,
} from '../../types/article'


/* =====================================================
   Router
===================================================== */

const route = useRoute()

const router = useRouter()


/* =====================================================
   页面状态
===================================================== */

const loading = ref(true)

const article = ref<Article | null>(null)


/* =====================================================
   获取文章
===================================================== */

const loadArticle = async () => {

  const id =
    Number(route.params.id)


  if (!id) {

    ElMessage.error(
      '文章 ID 无效',
    )

    router.push(
      '/admin/articles',
    )

    return

  }


  try {

    const result =
      await articleService.getArticleById(
        id,
      )


    if (!result) {

      ElMessage.error(
        '文章不存在',
      )

      router.push(
        '/admin/articles',
      )

      return

    }


    article.value = result

  } catch (error) {

    console.error(
      '获取文章失败：',
      error,
    )

    ElMessage.error(
      '获取文章失败',
    )

    router.push(
      '/admin/articles',
    )

  } finally {

    loading.value = false

  }

}


/* =====================================================
   初始化
===================================================== */

onMounted(() => {

  loadArticle()

})

</script>


<style scoped>

.loading-container {
  width: 100%;

  padding: 20px;

  box-sizing: border-box;
}

</style>
