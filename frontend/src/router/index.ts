import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),

  routes: [
    {
      path: '/',
      redirect: '/login',
    },

    {
      path: '/login',
      component: () => import('../views/auth/Login.vue'),
    },

    // =========================
    // 后台管理
    // =========================
    {
      path: '/admin',
      component: () => import('../layouts/AdminLayout.vue'),

      children: [
        {
          path: '',
          component: () =>
            import('../views/dashboard/Dashboard.vue'),
        },

        {
          path: 'articles',
          component: () =>
            import('../views/articles/ArticleList.vue'),
        },

        {
          path: 'articles/new',
          component: () =>
            import('../components/article/ArticleEditor.vue'),
        },

        {
          path: 'articles/:id/edit',
          component: () =>
            import('../views/articles/ArticleEdit.vue'),
        },
        
        {
          path: 'categories',
          component: () =>
            import('../views/categories/CategoryList.vue'),
        },

        {
          path: 'images',
          component: () =>
            import('../views/images/ImageList.vue'),
        },

        {
          path: 'recycle-bin',
          component: () =>
            import('../views/recycle-bin/RecycleBin.vue'),
        },

        {
          path: 'settings',
          component: () =>
            import('../views/settings/Settings.vue'),
        },
      ],
    },

    // =========================
    // 前台
    // =========================
    {
      path: '/docs',
      component: () =>
        import('../views/docs/DocumentDetail.vue'),
    },

    {
      path: '/docs/:slug',
      component: () =>
        import('../views/docs/DocumentDetail.vue'),
    },

    {
      path: '/search',
      component: () =>
        import('../views/search/Search.vue'),
    },
  ],
})

router.beforeEach((to) => {
  if (to.path === '/login') {
    return true
  }

  if (to.path.startsWith('/admin')) {
    const authenticated = sessionStorage.getItem('lightdocs-access-token')

    if (!authenticated) {
      return {
        path: '/login',
        query: { redirect: to.fullPath },
      }
    }
  }

  return true
})

export default router
