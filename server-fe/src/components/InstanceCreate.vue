<script setup>

import axios from 'axios';
import {ElMessage} from "element-plus";
import {useMilvusInstanceStore} from "@/stores/milvusInstance";
import {ref} from "vue";
import {instanceCreateUrl} from '@/api/constants.js';

const create_status = ref('')
const milvusInstanceStore = useMilvusInstanceStore()

const onInstanceCreate = () => {
  if (milvusInstanceStore.milvusInstance.MilvusServerName === '' ||
      milvusInstanceStore.milvusInstance.MilvusServerPort === '' ||
      milvusInstanceStore.milvusInstance.MilvusCollectionName === '' ||
      milvusInstanceStore.milvusInstance.ModelVecDim === '' ||
      milvusInstanceStore.milvusInstance.MilvusIndexName === '' ||
      milvusInstanceStore.milvusInstance.MilvusMetricType === '' ||
      milvusInstanceStore.milvusInstance.Model_API_KEY === ''
  ) {
    ElMessage({showClose: true, message: '请输入所有参数', type: 'error'})
    return
  }
  create_status.value = '创建中...'
  axios.post(instanceCreateUrl, {
    milvus_server: milvusInstanceStore.milvusInstance.MilvusServerName,
    milvus_port: milvusInstanceStore.milvusInstance.MilvusServerPort,
    milvus_username: milvusInstanceStore.milvusInstance.MilvusServerUserName,
    milvus_pass: milvusInstanceStore.milvusInstance.MilvusServerPassWord,
    collection_name: milvusInstanceStore.milvusInstance.MilvusCollectionName,
    collection_dim: milvusInstanceStore.milvusInstance.ModelVecDim,
    index_name: milvusInstanceStore.milvusInstance.MilvusIndexName,
    metric_type: milvusInstanceStore.milvusInstance.MilvusMetricType,
  }).then(response => {
    create_status.value = ''
    if (response.status === 200) {
      ElMessage({showClose: true, message: '创建实例成功', type: 'success',})
    } else {
      ElMessage({showClose: true, message: '创建实例失败', type: 'error',})
    }
  }).catch(err => {
    create_status.value = ''
    console.error('请求失败:', error);
    ElMessage({showClose: true, message: '创建实例失败', type: 'error'})
  })
}

</script>

<template>
  <el-form label-width="auto">
    <el-row>
      <el-col :span="7">
        <el-form-item label="MilvusServerIp">
          <el-input v-model="milvusInstanceStore.milvusInstance.MilvusServerName" placeholder="<ip address>"/>
        </el-form-item>
      </el-col>
      <el-col :span="7" :offset="1">
        <el-form-item label="MilvusServerPort">
          <el-input v-model="milvusInstanceStore.milvusInstance.MilvusServerPort" placeholder="19530"/>
        </el-form-item>
      </el-col>
    </el-row>
    <el-row>
      <el-col :span="7">
        <el-form-item label="MilvusUserName">
          <el-input v-model="milvusInstanceStore.milvusInstance.MilvusServerUserName" placeholder="root"/>
        </el-form-item>
      </el-col>
      <el-col :span="7" :offset="1">
        <el-form-item label="MilvusPassWord">
          <el-input v-model="milvusInstanceStore.milvusInstance.MilvusServerPassWord"
                    placeholder="skip if no authentication" type="password" autocomplete="off"/>
        </el-form-item>
      </el-col>
    </el-row>
    <el-row>
      <el-col :span="7">
        <el-form-item label="Milvus集合实例名称">
          <el-input v-model="milvusInstanceStore.milvusInstance.MilvusCollectionName" placeholder="imageserch_1"/>
        </el-form-item>
      </el-col>
    </el-row>
    <el-row>
      <el-col :span="7">
        <el-form-item label="MilvusIndexName">
          <el-select v-model="milvusInstanceStore.milvusInstance.MilvusIndexName" placeholder="please select indexName">
            <el-option label="HNSW" value="HNSW"/>
            <el-option label="IVF_SQ8" value="IVF_SQ8"/>
            <el-option label="IVF_FLAT" value="IVF_FLAT"/>
            <el-option label="SCANN" value="SCANN"/>
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="7" :offset="1">
        <el-form-item label="MilvusMetricType">
          <el-select v-model="milvusInstanceStore.milvusInstance.MilvusMetricType"
                     placeholder="please select metricType">
            <el-option label="L2" value="L2"/>
            <el-option label="IP" value="IP"/>
            <el-option label="COSINE" value="COSINE"/>
          </el-select>
        </el-form-item>
      </el-col>
    </el-row>
    <el-divider></el-divider>
    <el-row>
      <el-col :span="7">
        <el-form-item label="Model_API_KEY">
          <el-input v-model="milvusInstanceStore.milvusInstance.Model_API_KEY" type="password" autocomplete="off"/>
        </el-form-item>
      </el-col>

      <el-col :span="2" :offset="1">
        <a href="https://bailian.console.aliyun.com/?tab=api#/api/?type=model&url=https%3A%2F%2Fhelp.aliyun.com%2Fdocument_detail%2F2712195.html&renderType=iframe"
           target="_blank">
          获取
        </a>
      </el-col>
    </el-row>
    <el-row>
      <el-col :span="2" :offset="1">
        <el-form-item>
          <el-button type="primary" @click="onInstanceCreate">点击创建实例</el-button>
        </el-form-item>
      </el-col>
    </el-row>
    <el-row>
      <el-col :span="2" :offset="1">
        <span>{{ create_status }}</span>
      </el-col>
    </el-row>
  </el-form>
</template>

<style scoped>

</style>