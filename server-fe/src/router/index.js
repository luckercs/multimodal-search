import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      redirect: { name: 'instanceCreate' },
    },
    {
      path: '/instanceCreate',
      name: 'instanceCreate',
      component: () => import('../views/InstanceCreate.vue'),
    },
    {
      path: '/picImport',
      name: 'picImport',
      component: () => import('../views/PicImport.vue'),
    },
    {
      path: '/picSearch',
      name: 'picSearch',
      component: () => import('../views/PicSearch.vue'),
    },
    {
      path: '/instanceDelete',
      name: 'instanceDelete',
      component: () => import('../views/InstanceDelete.vue'),
    },
  ],
})

export default router
