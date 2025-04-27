<script setup>
import {ref} from "vue";
import {ElMessage} from "element-plus";
import axios from "axios";
import {useMilvusInstanceStore} from "@/stores/milvusInstance.js";
import {UploadUrl, PicImportUrl} from "@/api/constants.js"

const milvusInstanceStore = useMilvusInstanceStore()
const filesList = ref([])
const insert_status = ref('')

const customUpload = (options) => {
  const formData = new FormData();
  formData.append('files', options.file); // 添加文件到表单数据中
  formData.append('collectionName', milvusInstanceStore.milvusInstance.MilvusCollectionName);
  axios.post(UploadUrl, formData, {headers: {'Content-Type': 'multipart/form-data'}})
      .then(response => {
        if (response.status === 200) {
          ElMessage({showClose: true, message: '上传成功', type: 'success',})
        } else {
          ElMessage({showClose: true, message: '上传失败', type: 'error',})
        }
      })
      .catch(err => {
        ElMessage({showClose: true, message: '上传失败', type: 'error',})
      })
}

const beforeUpload = (file) => {
  const maxSize = 3 * 1024 * 1024;
  if (file.size > maxSize) {
    alert('文件大小不能超过 3MB，请重新选择！');
    return false;
  }
  return true;
}

const onPicImport = () => {
  if (milvusInstanceStore.milvusInstance.MilvusServerName === '' ||
      milvusInstanceStore.milvusInstance.MilvusServerPort === '' ||
      milvusInstanceStore.milvusInstance.MilvusCollectionName === '' ||
      milvusInstanceStore.milvusInstance.ModelVecDim === '' ||
      milvusInstanceStore.milvusInstance.MilvusIndexName === '' ||
      milvusInstanceStore.milvusInstance.MilvusMetricType === '' ||
      milvusInstanceStore.milvusInstance.Model_API_KEY === '') {
    ElMessage({showClose: true, message: '请输入milvus实例创建所有参数', type: 'error'})
    return
  }
  insert_status.value = "导入中..."
  axios.post(PicImportUrl, {
    milvus_server: milvusInstanceStore.milvusInstance.MilvusServerName,
    milvus_port: milvusInstanceStore.milvusInstance.MilvusServerPort,
    milvus_username: milvusInstanceStore.milvusInstance.MilvusServerUserName,
    milvus_pass: milvusInstanceStore.milvusInstance.MilvusServerPassWord,
    collection_name: milvusInstanceStore.milvusInstance.MilvusCollectionName,
    embed_server_url: milvusInstanceStore.milvusInstance.ModelUrl,
    embed_server_apikey: milvusInstanceStore.milvusInstance.Model_API_KEY
  }).then(response => {
    insert_status.value = ""
    if (response.status === 200) {
      ElMessage({showClose: true, message: '导入成功', type: 'success',})
    } else {
      ElMessage({showClose: true, message: '导入失败', type: 'error',})
    }
  }).catch(err => {
    insert_status.value = ""
    console.error('导入失败:', error);
    ElMessage({showClose: true, message: '导入失败', type: 'error'})
  })
}
</script>

<template>
  <el-form label-width="auto">
    <el-row>
      <el-col :span="7">
        <el-form-item label="Milvus集合实例名称">
          <el-input v-model="milvusInstanceStore.milvusInstance.MilvusCollectionName"/>
        </el-form-item>
      </el-col>
    </el-row>
    <el-row>
      <el-col :span="2" :offset="1">
        <el-form-item>
          <el-upload
              class="upload-demo"
              action=""
              multiple
              list-type="picture"
              v-model:file-list="filesList"
              :show-file-list="false"
              :http-request="customUpload"
              :before-upload="beforeUpload"
              acccept=".jpg,.jpeg,.png,.bmp"
          >
            <el-button type="primary">点击上传图片</el-button>
          </el-upload>
        </el-form-item>
      </el-col>
      <el-col :span="2" :offset="1">
        <el-form-item>
          <el-button type="primary" @click="onPicImport">导入到实例中</el-button>
        </el-form-item>
      </el-col>
    </el-row>
    <el-row>
      <el-col :span="2" :offset="1">
        <span>{{ insert_status }}</span>
      </el-col>
    </el-row>
  </el-form>
</template>

<style scoped>

</style>