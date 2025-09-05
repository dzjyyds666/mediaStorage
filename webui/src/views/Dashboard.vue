<template>
  <div class="dashboard">
    <!-- 顶部导航栏 -->
    <el-header class="header">
      <div class="header-content">
        <h1 class="title">
          <el-icon><Monitor /></el-icon>
          媒体存储服务 Dashboard
        </h1>
        <div class="header-actions">
          <el-button type="primary" @click="refreshData">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>
    </el-header>

    <!-- 主要内容区域 -->
    <el-container class="main-container">
      <!-- 侧边栏 -->
      <el-aside class="sidebar" width="250px">
        <el-menu
          :default-active="activeMenu"
          class="sidebar-menu"
          @select="handleMenuSelect"
        >
          <el-menu-item index="dashboard">
            <el-icon><Monitor /></el-icon>
            <span>仪表板</span>
          </el-menu-item>
          <el-menu-item index="storage">
            <el-icon><FolderOpened /></el-icon>
            <span>存储管理</span>
          </el-menu-item>
          <el-menu-item index="monitoring">
            <el-icon><TrendCharts /></el-icon>
            <span>系统监控</span>
          </el-menu-item>
          <el-menu-item index="settings">
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主内容区 -->
      <el-main class="content">
        <!-- 统计卡片 -->
        <div class="stats-grid">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon storage">
                <el-icon><FolderOpened /></el-icon>
              </div>
              <div class="stat-info">
                <h3>{{ formatBytes(storageStats.totalSize) }}</h3>
                <p>总存储空间</p>
              </div>
            </div>
          </el-card>

          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon files">
                <el-icon><Document /></el-icon>
              </div>
              <div class="stat-info">
                <h3>{{ storageStats.fileCount.toLocaleString() }}</h3>
                <p>文件总数</p>
              </div>
            </div>
          </el-card>

          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon upload">
                <el-icon><Upload /></el-icon>
              </div>
              <div class="stat-info">
                <h3>{{ formatBytes(storageStats.uploadToday) }}</h3>
                <p>今日上传</p>
              </div>
            </div>
          </el-card>

          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon status">
                <el-icon><CircleCheck /></el-icon>
              </div>
              <div class="stat-info">
                <h3>{{ systemStatus.status }}</h3>
                <p>系统状态</p>
              </div>
            </div>
          </el-card>
        </div>

        <!-- 图表区域 -->
        <div class="charts-grid">
          <el-card class="chart-card">
            <template #header>
              <div class="card-header">
                <span>存储使用趋势</span>
              </div>
            </template>
            <div class="chart-container" ref="storageChart"></div>
          </el-card>

          <el-card class="chart-card">
            <template #header>
              <div class="card-header">
                <span>文件类型分布</span>
              </div>
            </template>
            <div class="chart-container" ref="fileTypeChart"></div>
          </el-card>
        </div>

        <!-- 最近活动 -->
        <el-card class="activity-card">
          <template #header>
            <div class="card-header">
              <span>最近活动</span>
            </div>
          </template>
          <el-table :data="recentActivities" style="width: 100%">
            <el-table-column prop="time" label="时间" width="180" />
            <el-table-column prop="action" label="操作" width="120" />
            <el-table-column prop="file" label="文件" />
            <el-table-column prop="size" label="大小" width="120" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="scope">
                <el-tag :type="scope.row.status === '成功' ? 'success' : 'danger'">
                  {{ scope.row.status }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'

// 路由
const router = useRouter()
const activeMenu = ref('dashboard')

// 数据状态
const storageStats = ref({
  totalSize: 1024 * 1024 * 1024 * 500, // 500GB
  fileCount: 12580,
  uploadToday: 1024 * 1024 * 256, // 256MB
})

const systemStatus = ref({
  status: '正常',
  uptime: '15天 8小时',
  cpu: 45,
  memory: 68,
  disk: 72
})

const recentActivities = ref([
  {
    time: '2024-01-15 14:30:25',
    action: '上传',
    file: 'video_sample.mp4',
    size: '125MB',
    status: '成功'
  },
  {
    time: '2024-01-15 14:28:15',
    action: '删除',
    file: 'old_backup.zip',
    size: '2.3GB',
    status: '成功'
  },
  {
    time: '2024-01-15 14:25:10',
    action: '上传',
    file: 'document.pdf',
    size: '5.2MB',
    status: '失败'
  }
])

// 图表引用
const storageChart = ref<HTMLElement>()
const fileTypeChart = ref<HTMLElement>()

// 方法
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const handleMenuSelect = (key: string) => {
  activeMenu.value = key
  if (key !== 'dashboard') {
    router.push(`/${key}`)
  }
}

const refreshData = () => {
  // 模拟刷新数据
  console.log('刷新数据...')
}

const initCharts = () => {
  // 存储使用趋势图
  if (storageChart.value) {
    const chart1 = echarts.init(storageChart.value)
    const option1 = {
      title: {
        text: '存储使用量 (GB)',
        left: 'center',
        textStyle: { fontSize: 14 }
      },
      tooltip: {
        trigger: 'axis'
      },
      xAxis: {
        type: 'category',
        data: ['1月', '2月', '3月', '4月', '5月', '6月']
      },
      yAxis: {
        type: 'value'
      },
      series: [{
        data: [120, 200, 150, 280, 350, 400],
        type: 'line',
        smooth: true,
        areaStyle: {
          opacity: 0.3
        }
      }]
    }
    chart1.setOption(option1)
  }

  // 文件类型分布图
  if (fileTypeChart.value) {
    const chart2 = echarts.init(fileTypeChart.value)
    const option2 = {
      title: {
        text: '文件类型分布',
        left: 'center',
        textStyle: { fontSize: 14 }
      },
      tooltip: {
        trigger: 'item'
      },
      series: [{
        type: 'pie',
        radius: '60%',
        data: [
          { value: 1048, name: '视频文件' },
          { value: 735, name: '图片文件' },
          { value: 580, name: '文档文件' },
          { value: 484, name: '音频文件' },
          { value: 300, name: '其他文件' }
        ],
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }]
    }
    chart2.setOption(option2)
  }
}

onMounted(() => {
  nextTick(() => {
    initCharts()
  })
})
</script>

<style scoped>
.dashboard {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  background: #fff;
  border-bottom: 1px solid #e6e6e6;
  padding: 0 20px;
  height: 60px !important;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.title {
  margin: 0;
  font-size: 20px;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 8px;
}

.main-container {
  flex: 1;
  height: calc(100vh - 60px);
}

.sidebar {
  background: #fff;
  border-right: 1px solid #e6e6e6;
}

.sidebar-menu {
  border: none;
  height: 100%;
}

.content {
  padding: 20px;
  background: #f5f5f5;
  overflow-y: auto;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.stat-card {
  border: none;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 15px;
}

.stat-icon {
  width: 50px;
  height: 50px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: white;
}

.stat-icon.storage { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.stat-icon.files { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
.stat-icon.upload { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
.stat-icon.status { background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); }

.stat-info h3 {
  margin: 0;
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.stat-info p {
  margin: 5px 0 0 0;
  color: #909399;
  font-size: 14px;
}

.charts-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 20px;
}

.chart-card {
  border: none;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.chart-container {
  height: 300px;
}

.activity-card {
  border: none;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.card-header {
  font-weight: bold;
  color: #303133;
}

@media (max-width: 768px) {
  .charts-grid {
    grid-template-columns: 1fr;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>