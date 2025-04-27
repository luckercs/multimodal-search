<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import axios from 'axios'
import { useMilvusInstanceStore } from '@/stores/milvusInstance.js'
import { UploadFilled } from '@element-plus/icons-vue'
import { picSearchByTextUrl, picSearchByImgUrl, UploadUrl } from '@/api/constants.js'

const milvusInstanceStore = useMilvusInstanceStore()
const filesList = ref([])
const search_img_filename = ref('')
const searchImageUrl = ref('')
const search_text = ref('')
const search_topk = ref('3')
const imageUrlAndScores = reactive([])
const search_status = ref('')

const onPicSearchByText = () => {
  if (
    milvusInstanceStore.milvusInstance.MilvusServerName === '' ||
    milvusInstanceStore.milvusInstance.MilvusServerPort === '' ||
    milvusInstanceStore.milvusInstance.MilvusCollectionName === '' ||
    milvusInstanceStore.milvusInstance.ModelVecDim === '' ||
    milvusInstanceStore.milvusInstance.MilvusIndexName === '' ||
    milvusInstanceStore.milvusInstance.MilvusMetricType === '' ||
    milvusInstanceStore.milvusInstance.Model_API_KEY === '' ||
    search_text.value === '' ||
    search_topk.value === ''
  ) {
    ElMessage({
      showClose: true,
      message: '请输入milvus实例创建所有参数以及搜索参数',
      type: 'error',
    })
    return
  }
  search_status.value = '查询中...'
  search_img_filename.value = ''
  searchImageUrl.value = ''
  imageUrlAndScores.length = 0
  axios
    .post(picSearchByTextUrl, {
      milvus_server: milvusInstanceStore.milvusInstance.MilvusServerName,
      milvus_port: milvusInstanceStore.milvusInstance.MilvusServerPort,
      milvus_username: milvusInstanceStore.milvusInstance.MilvusServerUserName,
      milvus_pass: milvusInstanceStore.milvusInstance.MilvusServerPassWord,
      collection_name: milvusInstanceStore.milvusInstance.MilvusCollectionName,
      index_name: milvusInstanceStore.milvusInstance.MilvusIndexName,
      metric_type: milvusInstanceStore.milvusInstance.MilvusMetricType,
      embed_server_url: milvusInstanceStore.milvusInstance.ModelUrl,
      embed_server_apikey: milvusInstanceStore.milvusInstance.Model_API_KEY,
      search_text: search_text.value,
      search_topk: search_topk.value,
    })
    .then((response) => {
      search_status.value = ''
      if (response.status === 200) {
        ElMessage({ showClose: true, message: '查询成功', type: 'success' })
        imageUrlAndScores.push(...JSON.parse(response.data.data))
      } else {
        ElMessage({ showClose: true, message: '查询失败', type: 'error' })
      }
    })
    .catch((err) => {
      search_status.value = ''
      console.error('查询失败:', error)
      console.error(response.data.data)
      ElMessage({ showClose: true, message: '查询失败', type: 'error' })
    })
}

const customUpload = (options) => {
  search_img_filename.value = ''
  searchImageUrl.value = ''
  const formData = new FormData()
  formData.append('files', options.file) // 添加文件到表单数据中
  formData.append('collectionName', milvusInstanceStore.milvusInstance.MilvusCollectionName)
  axios
    .post(UploadUrl, formData, { headers: { 'Content-Type': 'multipart/form-data' } })
    .then((response) => {
      if (response.status === 200) {
        search_img_filename.value = options.file.name
        searchImageUrl.value = response.data.url
        ElMessage({ showClose: true, message: '上传成功', type: 'success' })
      } else {
        ElMessage({ showClose: true, message: '上传失败', type: 'error' })
      }
    })
    .catch((err) => {
      ElMessage({ showClose: true, message: '上传失败', type: 'error' })
    })
}

const onPicSearchByImg = () => {
  if (
    milvusInstanceStore.milvusInstance.MilvusServerName === '' ||
    milvusInstanceStore.milvusInstance.MilvusServerPort === '' ||
    milvusInstanceStore.milvusInstance.MilvusCollectionName === '' ||
    milvusInstanceStore.milvusInstance.ModelVecDim === '' ||
    milvusInstanceStore.milvusInstance.MilvusIndexName === '' ||
    milvusInstanceStore.milvusInstance.MilvusMetricType === '' ||
    milvusInstanceStore.milvusInstance.Model_API_KEY === '' ||
    search_img_filename.value === '' ||
    search_topk.value === ''
  ) {
    ElMessage({
      showClose: true,
      message: '请输入milvus实例创建所有参数以及搜索参数',
      type: 'error',
    })
    return
  }
  search_status.value = '查询中...'
  search_text.value = ''
  imageUrlAndScores.length = 0
  axios
    .post(picSearchByImgUrl, {
      milvus_server: milvusInstanceStore.milvusInstance.MilvusServerName,
      milvus_port: milvusInstanceStore.milvusInstance.MilvusServerPort,
      milvus_username: milvusInstanceStore.milvusInstance.MilvusServerUserName,
      milvus_pass: milvusInstanceStore.milvusInstance.MilvusServerPassWord,
      collection_name: milvusInstanceStore.milvusInstance.MilvusCollectionName,
      index_name: milvusInstanceStore.milvusInstance.MilvusIndexName,
      metric_type: milvusInstanceStore.milvusInstance.MilvusMetricType,
      embed_server_url: milvusInstanceStore.milvusInstance.ModelUrl,
      embed_server_apikey: milvusInstanceStore.milvusInstance.Model_API_KEY,
      search_img: search_img_filename.value,
      search_topk: search_topk.value,
    })
    .then((response) => {
      search_status.value = ''
      if (response.status === 200) {
        ElMessage({ showClose: true, message: '查询成功', type: 'success' })
        imageUrlAndScores.push(...JSON.parse(response.data.data))
      } else {
        ElMessage({ showClose: true, message: '查询失败', type: 'error' })
      }
    })
    .catch((err) => {
      search_status.value = ''
      console.error('查询失败:', error)
      console.error(response.data.data)
      ElMessage({ showClose: true, message: '查询失败', type: 'error' })
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
      <el-col :span="7" :offset="1">
        <el-form-item label="search_topk">
          <el-input v-model="search_topk" />
        </el-form-item>
      </el-col>
    </el-row>
    <el-divider></el-divider>
    <el-row>
      <el-col :span="7">
        <el-form-item label="search_text">
          <el-input v-model="search_text" placeholder="Please enter text to search" />
        </el-form-item>
      </el-col>
      <el-col :span="2" :offset="1">
        <el-form-item>
          <el-button type="primary" @click="onPicSearchByText">以文搜图</el-button>
        </el-form-item>
      </el-col>
    </el-row>
    <el-row>
      <el-col :span="7">
        <el-form-item label="search_img">
          <el-upload
            class="upload-demo"
            drag
            action=""
            list-type="picture"
            v-model:file-list="filesList"
            :show-file-list="false"
            :http-request="customUpload"
          >
            <el-icon class="el-icon--upload" style="height: 5px">
              <upload-filled />
            </el-icon>
            <div class="el-upload__text">拖拽图片 or <em>点击上传</em></div>
            <div v-if="searchImageUrl">
              <img :src="searchImageUrl" alt="Img" style="max-width: 100%; height: auto" />
            </div>
            {{ search_img_filename }}
          </el-upload>
        </el-form-item>
      </el-col>
      <el-col :span="2" :offset="1">
        <el-form-item>
          <el-button type="primary" @click="onPicSearchByImg">以图搜图</el-button>
        </el-form-item>
      </el-col>
    </el-row>
  </el-form>
  <el-divider></el-divider>
  <div>
    <h4>图片查询结果</h4>
    <span>{{ search_status }}</span>
    <el-divider></el-divider>
    <div class="image-container">
      <div v-for="(urlAndScore, index) in imageUrlAndScores" :key="index" class="image-item">
        <img
          :src="urlAndScore.url"
          :alt="`score: ${urlAndScore.url}`"
          :title="`score: ${urlAndScore.score}`"
        />
        <p>{{ urlAndScore.filename }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.image-container {
  display: flex;
  flex-wrap: wrap;
}

.image-item {
  flex: 0 0 20%;
  box-sizing: border-box;
  padding: 5px;
}

.image-item img {
  width: 50%;
  height: auto;
}

.el-form {
  min-width: 1px;
}
</style>
