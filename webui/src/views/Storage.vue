<template>
  <div class="storage-page">
    <el-page-header @back="goBack" content="存储管理" />
    
    <div class="storage-content">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>文件管理</span>
            <el-button type="primary" @click="uploadDialogVisible = true">
              <el-icon><Upload /></el-icon>
              上传文件
            </el-button>
          </div>
        </template>
        
        <el-table :data="fileList" style="width: 100%">
          <el-table-column prop="name" label="文件名" />
          <el-table-column prop="size" label="大小" width="120" />
          <el-table-column prop="type" label="类型" width="100" />
          <el-table-column prop="uploadTime" label="上传时间" width="180" />
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button size="small" @click="downloadFile(scope.row)">下载</el-button>
              <el-button size="small" type="danger" @click="deleteFile(scope.row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 上传对话框 -->
    <el-dialog v-model="uploadDialogVisible" title="上传文件" width="500px">
      <el-upload
        class="upload-demo"
        drag
        action="/api/upload"
        multiple
        :on-success="handleUploadSuccess"
      >
        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
      </el-upload>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const uploadDialogVisible = ref(false)

const fileList = ref([
  {
    name: 'sample_video.mp4',
    size: '125MB',
    type: '视频',
    uploadTime: '2024-01-15 14:30:25'
  },
  {
    name: 'document.pdf',
    size: '5.2MB',
    type: '文档',
    uploadTime: '2024-01-15 14:25:10'
  }
])

const goBack = () => {
  router.push('/')
}

const downloadFile = (file: any) => {
  console.log('下载文件:', file.name)
}

const deleteFile = (file: any) => {
  console.log('删除文件:', file.name)
}

const handleUploadSuccess = () => {
  uploadDialogVisible.value = false
  console.log('上传成功')
}
</script>

<style scoped>
.storage-page {
  padding: 20px;
}

.storage-content {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>