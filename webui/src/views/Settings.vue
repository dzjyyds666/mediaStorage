<template>
  <div class="settings-page">
    <el-page-header @back="goBack" content="系统设置" />
    
    <div class="settings-content">
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="存储配置" name="storage">
          <el-form :model="storageConfig" label-width="120px">
            <el-form-item label="存储路径">
              <el-input v-model="storageConfig.path" />
            </el-form-item>
            <el-form-item label="最大文件大小">
              <el-input v-model="storageConfig.maxFileSize" />
            </el-form-item>
            <el-form-item label="允许的文件类型">
              <el-input v-model="storageConfig.allowedTypes" type="textarea" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveStorageConfig">保存配置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="系统配置" name="system">
          <el-form :model="systemConfig" label-width="120px">
            <el-form-item label="服务端口">
              <el-input v-model="systemConfig.port" />
            </el-form-item>
            <el-form-item label="日志级别">
              <el-select v-model="systemConfig.logLevel">
                <el-option label="DEBUG" value="debug" />
                <el-option label="INFO" value="info" />
                <el-option label="WARN" value="warn" />
                <el-option label="ERROR" value="error" />
              </el-select>
            </el-form-item>
            <el-form-item label="启用HTTPS">
              <el-switch v-model="systemConfig.enableHttps" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveSystemConfig">保存配置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="用户管理" name="users">
          <div class="users-section">
            <el-button type="primary" @click="addUserDialogVisible = true">
              添加用户
            </el-button>
            <el-table :data="users" style="width: 100%; margin-top: 20px;">
              <el-table-column prop="username" label="用户名" />
              <el-table-column prop="role" label="角色" />
              <el-table-column prop="lastLogin" label="最后登录" />
              <el-table-column label="操作" width="200">
                <template #default="scope">
                  <el-button size="small" @click="editUser(scope.row)">编辑</el-button>
                  <el-button size="small" type="danger" @click="deleteUser(scope.row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 添加用户对话框 -->
    <el-dialog v-model="addUserDialogVisible" title="添加用户" width="400px">
      <el-form :model="newUser" label-width="80px">
        <el-form-item label="用户名">
          <el-input v-model="newUser.username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="newUser.password" type="password" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="newUser.role">
            <el-option label="管理员" value="admin" />
            <el-option label="用户" value="user" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addUserDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="addUser">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const activeTab = ref('storage')
const addUserDialogVisible = ref(false)

const storageConfig = ref({
  path: '/data/storage',
  maxFileSize: '100MB',
  allowedTypes: 'jpg,png,gif,mp4,avi,pdf,doc,docx'
})

const systemConfig = ref({
  port: '8080',
  logLevel: 'info',
  enableHttps: false
})

const users = ref([
  {
    username: 'admin',
    role: '管理员',
    lastLogin: '2024-01-15 14:30:25'
  },
  {
    username: 'user1',
    role: '用户',
    lastLogin: '2024-01-15 10:15:30'
  }
])

const newUser = ref({
  username: '',
  password: '',
  role: 'user'
})

const goBack = () => {
  router.push('/')
}

const saveStorageConfig = () => {
  console.log('保存存储配置:', storageConfig.value)
}

const saveSystemConfig = () => {
  console.log('保存系统配置:', systemConfig.value)
}

const addUser = () => {
  users.value.push({
    username: newUser.value.username,
    role: newUser.value.role === 'admin' ? '管理员' : '用户',
    lastLogin: '从未登录'
  })
  addUserDialogVisible.value = false
  newUser.value = { username: '', password: '', role: 'user' }
}

const editUser = (user: any) => {
  console.log('编辑用户:', user)
}

const deleteUser = (user: any) => {
  console.log('删除用户:', user)
}
</script>

<style scoped>
.settings-page {
  padding: 20px;
}

.settings-content {
  margin-top: 20px;
}

.users-section {
  padding: 20px;
}
</style>