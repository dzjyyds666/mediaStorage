<template>
  <div class="monitoring-page">
    <el-page-header @back="goBack" content="系统监控" />
    
    <div class="monitoring-content">
      <div class="metrics-grid">
        <el-card class="metric-card">
          <template #header>
            <span>CPU 使用率</span>
          </template>
          <div class="metric-value">{{ systemMetrics.cpu }}%</div>
          <el-progress :percentage="systemMetrics.cpu" :color="getProgressColor(systemMetrics.cpu)" />
        </el-card>

        <el-card class="metric-card">
          <template #header>
            <span>内存使用率</span>
          </template>
          <div class="metric-value">{{ systemMetrics.memory }}%</div>
          <el-progress :percentage="systemMetrics.memory" :color="getProgressColor(systemMetrics.memory)" />
        </el-card>

        <el-card class="metric-card">
          <template #header>
            <span>磁盘使用率</span>
          </template>
          <div class="metric-value">{{ systemMetrics.disk }}%</div>
          <el-progress :percentage="systemMetrics.disk" :color="getProgressColor(systemMetrics.disk)" />
        </el-card>

        <el-card class="metric-card">
          <template #header>
            <span>网络流量</span>
          </template>
          <div class="metric-value">{{ systemMetrics.network }} MB/s</div>
          <div class="network-info">
            <div>上行: {{ systemMetrics.networkUp }} MB/s</div>
            <div>下行: {{ systemMetrics.networkDown }} MB/s</div>
          </div>
        </el-card>
      </div>

      <el-card class="chart-card">
        <template #header>
          <span>系统性能趋势</span>
        </template>
        <div class="chart-container" ref="performanceChart"></div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'

const router = useRouter()
const performanceChart = ref<HTMLElement>()

const systemMetrics = ref({
  cpu: 45,
  memory: 68,
  disk: 72,
  network: 12.5,
  networkUp: 5.2,
  networkDown: 7.3
})

const goBack = () => {
  router.push('/')
}

const getProgressColor = (percentage: number) => {
  if (percentage < 50) return '#67c23a'
  if (percentage < 80) return '#e6a23c'
  return '#f56c6c'
}

const initChart = () => {
  if (performanceChart.value) {
    const chart = echarts.init(performanceChart.value)
    const option = {
      title: {
        text: '系统性能监控',
        left: 'center'
      },
      tooltip: {
        trigger: 'axis'
      },
      legend: {
        data: ['CPU', '内存', '磁盘'],
        top: 30
      },
      xAxis: {
        type: 'category',
        data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00', '24:00']
      },
      yAxis: {
        type: 'value',
        max: 100,
        axisLabel: {
          formatter: '{value}%'
        }
      },
      series: [
        {
          name: 'CPU',
          type: 'line',
          data: [20, 25, 45, 60, 55, 40, 35],
          smooth: true
        },
        {
          name: '内存',
          type: 'line',
          data: [50, 55, 65, 70, 68, 65, 60],
          smooth: true
        },
        {
          name: '磁盘',
          type: 'line',
          data: [70, 71, 72, 75, 73, 72, 70],
          smooth: true
        }
      ]
    }
    chart.setOption(option)
  }
}

onMounted(() => {
  nextTick(() => {
    initChart()
  })
})
</script>

<style scoped>
.monitoring-page {
  padding: 20px;
}

.monitoring-content {
  margin-top: 20px;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.metric-card {
  text-align: center;
}

.metric-value {
  font-size: 32px;
  font-weight: bold;
  color: #409eff;
  margin: 10px 0;
}

.network-info {
  margin-top: 10px;
  font-size: 14px;
  color: #666;
}

.chart-card {
  margin-top: 20px;
}

.chart-container {
  height: 400px;
}
</style>