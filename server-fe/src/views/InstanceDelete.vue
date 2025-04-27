<script setup>
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useMilvusInstanceStore } from '@/stores/milvusInstance.js'
import { ref } from 'vue'
import { instanceDeleteUrl } from '@/api/constants.js'

const milvusInstanceStore = useMilvusInstanceStore()
const delete_status = ref('')
const onInstanceDelete = () => {
  if (
    milvusInstanceStore.milvusInstance.MilvusServerName === '' ||
    milvusInstanceStore.milvusInstance.MilvusServerPort === '' ||
    milvusInstanceStore.milvusInstance.MilvusCollectionName === '' ||
    milvusInstanceStore.milvusInstance.ModelVecDim === '' ||
    milvusInstanceStore.milvusInstance.MilvusIndexName === '' ||
    milvusInstanceStore.milvusInstance.MilvusMetricType === '' ||
    milvusInstanceStore.milvusInstance.Model_API_KEY === ''
  ) {
    ElMessage({ showClose: true, message: '请输入milvus实例创建所有参数', type: 'error' })
    return
  }
  delete_status.value = '删除中...'
  axios
    .post(instanceDeleteUrl, {
      milvus_server: milvusInstanceStore.milvusInstance.MilvusServerName,
      milvus_port: milvusInstanceStore.milvusInstance.MilvusServerPort,
      milvus_username: milvusInstanceStore.milvusInstance.MilvusServerUserName,
      milvus_pass: milvusInstanceStore.milvusInstance.MilvusServerPassWord,
      collection_name: milvusInstanceStore.milvusInstance.MilvusCollectionName,
    })
    .then((response) => {
      delete_status.value = ''
      if (response.status === 200) {
        milvusInstanceStore.milvusInstance.MilvusCollectionName = ''
        ElMessage({ showClose: true, message: '删除实例成功', type: 'success' })
      } else {
        ElMessage({ showClose: true, message: '删除实例失败', type: 'error' })
      }
    })
    .catch((err) => {
      delete_status.value = ''
      console.error('删除失败:', error)
      ElMessage({ showClose: true, message: '删除实例失败', type: 'error' })
    })
}
</script>

<template>
  <el-form>
    <el-row>
      <el-col :span="7">
        <el-form-item label="Milvus集合实例名称">
          <el-input v-model="milvusInstanceStore.milvusInstance.MilvusCollectionName" />
        </el-form-item>
      </el-col>
    </el-row>
    <el-row>
      <el-col :span="2" :offset="1">
        <el-form-item>
          <el-button type="primary" @click="onInstanceDelete">点击删除实例</el-button>
        </el-form-item>
      </el-col>
    </el-row>
    <el-row>
      <el-col :span="2" :offset="1">
        <span>{{ delete_status }}</span>
      </el-col>
    </el-row>
  </el-form>
</template>

<style scoped>
.el-form {
  min-width: 1px;
}
</style>
