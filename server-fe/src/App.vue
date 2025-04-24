<script setup>

import {ref} from "vue";
import InstanceCreate from './components/InstanceCreate.vue'
import PicImport from './components/PicImport.vue'
import PicSearch from './components/PicSearch.vue'
import InstanceDelete from './components/InstanceDelete.vue'

const activeTagName = ref('instanceCreate')

const handleClick = (tab, event) => {
  console.log("changed tab: " + activeTagName.value)
}

</script>

<template>
  <div>
    <el-row >
      <el-col :span="24">
        <div class="center-text"><h2>Multimodal-Search</h2></div>
        <el-divider>
        </el-divider>
      </el-col>
    </el-row>
    <el-row>
      <el-col :span="24">
        <el-tabs v-model="activeTagName" class="main-tabs" @tab-click="handleClick">
          <el-tab-pane label="实例创建" name="instanceCreate"><InstanceCreate></InstanceCreate></el-tab-pane>
          <el-tab-pane label="图片导入" name="picImport"><PicImport></PicImport></el-tab-pane>
          <el-tab-pane label="图片检索" name="picSearch"><PicSearch></PicSearch></el-tab-pane>
          <el-tab-pane label="实例删除" name="instanceDelete"><InstanceDelete></InstanceDelete></el-tab-pane>
        </el-tabs>
      </el-col>
    </el-row>
  </div>

</template>

<style scoped>
.main-tabs > .el-tabs__content {
  padding: 32px;
  color: #6b778c;
  font-size: 32px;
  font-weight: 600;
}

.center-text {
  text-align: center;
}
</style>
