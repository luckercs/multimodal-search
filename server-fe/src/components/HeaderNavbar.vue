<template>
  <nav class="nav-bar">
    <RouterLink v-for="(item, index) in navItems" :key="index" :to="item.route" class="nav-link">
      {{ item.name }}
    </RouterLink>
  </nav>
</template>

<script setup>
import { defineProps } from 'vue'

const props = defineProps({
  navItems: {
    type: Array,
    default: () => [
      { name: 'name1', route: '/path1' },
      { name: 'name2', route: '/path2' },
      { name: 'name3', route: '/path3' },
    ],
  },
})
</script>

<style scoped>
.nav-bar {
  top: 0;
  width: 100%;
  height: 50px;
  background-color: rgba(13, 139, 255, 255);
  display: flex;
  align-items: center;
}

.nav-link {
  color: white;
  text-decoration: none;
  padding: 0 15px;
}

.nav-link:hover {
  text-decoration: underline;
}
</style>
