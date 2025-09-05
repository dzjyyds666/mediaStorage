// 全局类型声明文件

// Element Plus 相关类型
declare module 'element-plus' {
  import type { App } from 'vue'
  const ElementPlus: {
    install(app: App): void
  }
  export default ElementPlus
}

declare module '@element-plus/icons-vue' {
  import type { Component } from 'vue'
  const icons: Record<string, Component>
  export = icons
}

// ECharts 相关类型
declare module 'echarts' {
  export function init(dom: HTMLElement | null): any
  export const graphic: any
  export const color: any
}

// Vue ECharts 相关类型
declare module 'vue-echarts' {
  import type { DefineComponent } from 'vue'
  const VChart: DefineComponent<any, any, any>
  export default VChart
}