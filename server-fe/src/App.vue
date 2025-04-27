<script setup>
import HeaderNavbar from '@/components/HeaderNavbar.vue'
import { ref } from 'vue'

const navigationItems = ref([
  { name: '实例创建', route: '/instanceCreate' },
  { name: '图片导入', route: '/picImport' },
  { name: '图片检索', route: '/picSearch' },
  { name: '实例删除', route: '/instanceDelete' },
])
</script>

<template>
  <div id="header">
    <HeaderNavbar :navItems="navigationItems" />
  </div>
  <div id="content">
    <RouterView />
  </div>
</template>

<style scoped>
#content {
  padding: 10px;
}
</style>
