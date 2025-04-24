import { defineStore } from 'pinia';
import {reactive} from "vue";

export const useMilvusInstanceStore = defineStore('milvusInstance', () => {
    const milvusInstance = reactive({
        MilvusServerName: '',
        MilvusServerPort: '',
        MilvusServerUserName: '',
        MilvusServerPassWord: '',
        MilvusCollectionName: '',
        MilvusIndexName: 'IVF_SQ8',
        MilvusMetricType: 'IP',
        ModelUrl: 'http://localhost:8010',
        Model_API_KEY: '',
        ModelVecDim: '1024'
    })

    return { milvusInstance }
})