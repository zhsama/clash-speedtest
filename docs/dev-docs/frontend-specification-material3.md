# Clash SpeedTest 前端规范文档 - Material 3设计系统

## 文档概述

本文档为Clash SpeedTest项目前端界面的设计规范，基于Google Material 3设计原则，结合项目特色紫色主题，为开发团队提供统一的设计标准和实现指导。

## 1. 设计理念与原则

### 1.1 核心设计理念
- **用户至上**：以用户的实际使用场景为导向，提供直观易用的界面
- **信息层次**：通过视觉层次清晰传达信息重要性
- **一致性**：保持整个应用界面的视觉和交互一致性
- **可访问性**：确保所有用户都能轻松使用应用功能
- **响应式设计**：适配各种设备尺寸和使用场景

### 1.2 Material 3核心原则
- **个性化**：支持动态色彩和用户偏好
- **适应性**：响应式设计适配不同设备
- **表现力**：通过色彩、动效等元素增强用户体验
- **功能性**：设计服务于功能，提升使用效率

## 2. 色彩系统

### 2.1 主色系 - 薰衣草紫色调

基于项目现有的薰衣草紫色主题，结合Material 3动态色彩原则：

```css
/* 核心色彩调色板 */
:root {
  /* 主色系 - 薰衣草紫 */
  --primary-50: hsl(280, 44%, 98%);   /* 极浅紫 - 背景提亮 */
  --primary-100: hsl(280, 40%, 95%);  /* 浅紫 - 悬停状态 */
  --primary-200: hsl(280, 35%, 90%);  /* 轻紫 - 禁用状态 */
  --primary-300: hsl(280, 30%, 82%);  /* 中浅紫 - 边框 */
  --primary-400: hsl(280, 25%, 74%);  /* 标准紫 - 文本/图标 */
  --primary-500: hsl(280, 20%, 66%);  /* 主紫 - 主要交互元素 */
  --primary-600: hsl(280, 18%, 58%);  /* 深紫 - 按钮默认 */
  --primary-700: hsl(280, 20%, 48%);  /* 较深紫 - 按钮悬停 */
  --primary-800: hsl(280, 25%, 38%);  /* 暗紫 - 活跃状态 */
  --primary-900: hsl(280, 30%, 28%);  /* 深暗紫 - 强调元素 */
  --primary-950: hsl(280, 35%, 18%);  /* 极暗紫 - 文本 */

  /* 辅助色系 */
  --secondary-50: hsl(260, 20%, 98%);
  --secondary-400: hsl(260, 15%, 65%);
  --secondary-600: hsl(260, 12%, 50%);
  --secondary-800: hsl(260, 18%, 35%);

  /* 功能色系 */
  --success: hsl(142, 76%, 36%);      /* 成功状态 */
  --warning: hsl(45, 93%, 58%);       /* 警告状态 */
  --error: hsl(0, 84%, 60%);          /* 错误状态 */
  --info: hsl(217, 91%, 60%);         /* 信息状态 */
}
```

### 2.2 暗色主题适配

针对项目的暗色主题环境：

```css
.dark {
  /* 暗色模式色彩映射 */
  --surface: hsl(0, 0%, 6%);          /* 主要表面 */
  --surface-variant: hsl(0, 0%, 12%);  /* 变体表面 */
  --surface-container: hsl(0, 0%, 8%); /* 容器表面 */
  --outline: var(--primary-600);       /* 轮廓线 */
  --outline-variant: var(--primary-800); /* 变体轮廓 */
  
  /* 文本层级 */
  --on-surface: var(--primary-50);     /* 主要文本 */
  --on-surface-variant: var(--primary-300); /* 次要文本 */
  --on-surface-disabled: var(--primary-500); /* 禁用文本 */
}
```

### 2.3 语义化色彩应用

**状态色彩编码系统**：
- **速度测试结果**：
  - 优秀 (>50 Mbps): `hsl(142, 76%, 36%)` 绿色
  - 良好 (20-50 Mbps): `hsl(45, 93%, 58%)` 黄色  
  - 一般 (5-20 Mbps): `hsl(25, 95%, 53%)` 橙色
  - 较差 (<5 Mbps): `hsl(0, 84%, 60%)` 红色

- **延迟指示**：
  - 极低 (<50ms): `hsl(142, 76%, 36%)` 绿色
  - 低 (50-150ms): `hsl(60, 100%, 50%)` 黄绿
  - 中等 (150-300ms): `hsl(45, 93%, 58%)` 黄色
  - 高 (>300ms): `hsl(0, 84%, 60%)` 红色

- **解锁状态**：
  - 完全支持: `hsl(142, 76%, 36%)` 绿色
  - 部分支持: `hsl(45, 93%, 58%)` 黄色
  - 不支持: `hsl(0, 84%, 60%)` 红色
  - 检测中: `hsl(217, 91%, 60%)` 蓝色

## 3. 字体系统

### 3.1 字体家族

```css
/* 字体定义 */
.typography-system {
  /* 主要字体 - 系统字体栈 */
  --font-family-primary: -apple-system, BlinkMacSystemFont, "Segoe UI", 
                         "Roboto", "Helvetica Neue", Arial, sans-serif;
  
  /* 等宽字体 - 用于代码、IP地址等 */
  --font-family-mono: "SF Mono", "Monaco", "Inconsolata", 
                      "Roboto Mono", "Consolas", monospace;
                      
  /* 数字字体 - 用于数据展示 */
  --font-family-numeric: "SF Pro Display", -apple-system, sans-serif;
}
```

### 3.2 字体层级

基于Material 3的类型系统：

```css
/* Material 3 字体等级 */
.typography-scale {
  /* 显示级 - 页面标题 */
  --display-large: 57px/64px;    /* 主标题 */
  --display-medium: 45px/52px;   /* 次级标题 */
  --display-small: 36px/44px;    /* 卡片标题 */
  
  /* 标题级 - 组件标题 */
  --headline-large: 32px/40px;   /* 页面区块标题 */
  --headline-medium: 28px/36px;  /* 卡片大标题 */
  --headline-small: 24px/32px;   /* 卡片小标题 */
  
  /* 标签级 - 界面标签 */
  --title-large: 22px/28px;      /* 主要标签 */
  --title-medium: 16px/24px;     /* 标准标签 */
  --title-small: 14px/20px;      /* 次要标签 */
  
  /* 正文级 - 内容文本 */
  --body-large: 16px/24px;       /* 主要内容 */
  --body-medium: 14px/20px;      /* 标准内容 */
  --body-small: 12px/16px;       /* 辅助信息 */
  
  /* 标记级 - 按钮等 */
  --label-large: 14px/20px;      /* 大按钮文字 */
  --label-medium: 12px/16px;     /* 标准按钮 */
  --label-small: 11px/16px;      /* 小标签 */
}
```

### 3.3 字重系统

```css
.font-weights {
  --font-weight-light: 300;      /* 轻量文本 */
  --font-weight-regular: 400;    /* 标准文本 */
  --font-weight-medium: 500;     /* 中等强调 */
  --font-weight-semibold: 600;   /* 半粗体 */
  --font-weight-bold: 700;       /* 粗体强调 */
}
```

## 4. 组件规范

### 4.1 按钮组件

基于Material 3的按钮设计：

```css
/* 主要按钮 - Filled Button */
.button-filled {
  background: var(--primary-600);
  color: var(--primary-50);
  padding: 12px 24px;
  border-radius: 24px;
  border: none;
  font: var(--label-large);
  font-weight: var(--font-weight-medium);
  transition: all 200ms cubic-bezier(0.2, 0, 0, 1);
  min-height: 48px;
  min-width: 64px;
}

.button-filled:hover {
  background: var(--primary-700);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.button-filled:active {
  background: var(--primary-800);
  transform: scale(0.98);
}

/* 轮廓按钮 - Outlined Button */
.button-outlined {
  background: transparent;
  color: var(--primary-600);
  border: 1px solid var(--primary-600);
  padding: 12px 24px;
  border-radius: 24px;
  font: var(--label-large);
  font-weight: var(--font-weight-medium);
  transition: all 200ms cubic-bezier(0.2, 0, 0, 1);
  min-height: 48px;
  min-width: 64px;
}

.button-outlined:hover {
  background: hsla(var(--primary-600-hsl), 0.08);
  border-color: var(--primary-700);
}

/* 文本按钮 - Text Button */
.button-text {
  background: transparent;
  color: var(--primary-600);
  border: none;
  padding: 12px 16px;
  border-radius: 24px;
  font: var(--label-large);
  font-weight: var(--font-weight-medium);
  transition: all 200ms cubic-bezier(0.2, 0, 0, 1);
  min-height: 48px;
}

.button-text:hover {
  background: hsla(var(--primary-600-hsl), 0.08);
}
```

**按钮使用指南**：
- **Filled Button**: 主要操作（开始测试、导出结果）
- **Outlined Button**: 次要操作（刷新配置、停止测试）
- **Text Button**: 辅助操作（查看详情、切换选项）

### 4.2 卡片组件

Material 3风格的卡片系统：

```css
/* 主要卡片 - Elevated Card */
.card-elevated {
  background: var(--surface-container);
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15), 
              0 1px 3px rgba(0, 0, 0, 0.1);
  transition: all 200ms cubic-bezier(0.2, 0, 0, 1);
}

.card-elevated:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2), 
              0 2px 6px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

/* 轮廓卡片 - Outlined Card */
.card-outlined {
  background: var(--surface);
  border: 1px solid var(--outline-variant);
  border-radius: 16px;
  padding: 24px;
  transition: all 200ms cubic-bezier(0.2, 0, 0, 1);
}

.card-outlined:hover {
  border-color: var(--outline);
  background: var(--surface-variant);
}

/* 填充卡片 - Filled Card */
.card-filled {
  background: var(--surface-variant);
  border-radius: 16px;
  padding: 24px;
  transition: all 200ms cubic-bezier(0.2, 0, 0, 1);
}
```

### 4.3 表单控件

#### 4.3.1 输入框 (Text Fields)

```css
/* 轮廓输入框 - Outlined Text Field */
.textfield-outlined {
  position: relative;
  background: transparent;
}

.textfield-outlined input {
  width: 100%;
  padding: 16px;
  border: 1px solid var(--outline);
  border-radius: 8px;
  background: transparent;
  color: var(--on-surface);
  font: var(--body-large);
  transition: all 200ms cubic-bezier(0.2, 0, 0, 1);
}

.textfield-outlined input:focus {
  outline: none;
  border-color: var(--primary-600);
  border-width: 2px;
  padding: 15px; /* 补偿边框宽度变化 */
}

.textfield-outlined label {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--on-surface-variant);
  font: var(--body-large);
  pointer-events: none;
  transition: all 200ms cubic-bezier(0.2, 0, 0, 1);
  background: var(--surface);
  padding: 0 4px;
}

.textfield-outlined input:focus + label,
.textfield-outlined input:not(:placeholder-shown) + label {
  top: 0;
  transform: translateY(-50%);
  font: var(--body-small);
  color: var(--primary-600);
}
```

#### 4.3.2 选择器 (Dropdowns)

```css
.select-outlined {
  position: relative;
  width: 100%;
}

.select-outlined select {
  width: 100%;
  padding: 16px;
  border: 1px solid var(--outline);
  border-radius: 8px;
  background: var(--surface);
  color: var(--on-surface);
  font: var(--body-large);
  cursor: pointer;
  appearance: none;
}

.select-outlined::after {
  content: '';
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  width: 0;
  height: 0;
  border-left: 5px solid transparent;
  border-right: 5px solid transparent;
  border-top: 5px solid var(--on-surface-variant);
  pointer-events: none;
}
```

#### 4.3.3 滑块 (Sliders)

```css
.slider-container {
  position: relative;
  padding: 20px 0;
}

.slider-track {
  width: 100%;
  height: 4px;
  background: var(--outline-variant);
  border-radius: 2px;
  position: relative;
}

.slider-active {
  height: 100%;
  background: var(--primary-600);
  border-radius: 2px;
  transition: width 200ms cubic-bezier(0.2, 0, 0, 1);
}

.slider-thumb {
  width: 20px;
  height: 20px;
  background: var(--primary-600);
  border: 2px solid var(--surface);
  border-radius: 50%;
  position: absolute;
  top: 50%;
  transform: translate(-50%, -50%);
  cursor: pointer;
  transition: all 200ms cubic-bezier(0.2, 0, 0, 1);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.slider-thumb:hover {
  transform: translate(-50%, -50%) scale(1.2);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
}
```

### 4.4 数据展示组件

#### 4.4.1 表格

```css
.data-table {
  width: 100%;
  background: var(--surface-container);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.data-table thead {
  background: var(--surface-variant);
}

.data-table th {
  padding: 16px;
  text-align: left;
  font: var(--title-medium);
  font-weight: var(--font-weight-medium);
  color: var(--on-surface);
  border-bottom: 1px solid var(--outline-variant);
}

.data-table td {
  padding: 16px;
  font: var(--body-medium);
  color: var(--on-surface);
  border-bottom: 1px solid var(--outline-variant);
}

.data-table tbody tr:hover {
  background: var(--surface-variant);
}

.data-table tbody tr:last-child td {
  border-bottom: none;
}
```

#### 4.4.2 进度指示器

```css
/* 线性进度条 */
.progress-linear {
  width: 100%;
  height: 4px;
  background: var(--outline-variant);
  border-radius: 2px;
  overflow: hidden;
}

.progress-linear-bar {
  height: 100%;
  background: var(--primary-600);
  border-radius: 2px;
  transition: width 300ms cubic-bezier(0.4, 0, 0.2, 1);
}

/* 圆形进度指示器 */
.progress-circular {
  width: 48px;
  height: 48px;
  position: relative;
}

.progress-circular svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.progress-circular circle {
  fill: none;
  stroke-width: 4;
  stroke-linecap: round;
}

.progress-circular .background {
  stroke: var(--outline-variant);
}

.progress-circular .foreground {
  stroke: var(--primary-600);
  stroke-dasharray: 150.79;
  stroke-dashoffset: 150.79;
  transition: stroke-dashoffset 300ms cubic-bezier(0.4, 0, 0.2, 1);
}
```

## 5. 布局系统

### 5.1 栅格系统

基于Material 3的响应式栅格：

```css
.layout-grid {
  display: grid;
  gap: 24px;
  padding: 24px;
  max-width: 1440px;
  margin: 0 auto;
}

/* 断点系统 */
@media (min-width: 600px) {
  .layout-grid {
    grid-template-columns: repeat(8, 1fr);
    gap: 24px;
    padding: 32px;
  }
}

@media (min-width: 900px) {
  .layout-grid {
    grid-template-columns: repeat(12, 1fr);
    gap: 24px;
    padding: 32px;
  }
}

@media (min-width: 1200px) {
  .layout-grid {
    gap: 32px;
    padding: 40px;
  }
}
```

### 5.2 间距系统

统一的间距标准：

```css
:root {
  /* 基础间距单元 - 4px */
  --spacing-unit: 4px;
  
  /* 间距等级 */
  --spacing-xs: calc(var(--spacing-unit) * 1);    /* 4px */
  --spacing-sm: calc(var(--spacing-unit) * 2);    /* 8px */
  --spacing-md: calc(var(--spacing-unit) * 3);    /* 12px */
  --spacing-lg: calc(var(--spacing-unit) * 4);    /* 16px */
  --spacing-xl: calc(var(--spacing-unit) * 6);    /* 24px */
  --spacing-2xl: calc(var(--spacing-unit) * 8);   /* 32px */
  --spacing-3xl: calc(var(--spacing-unit) * 12);  /* 48px */
  --spacing-4xl: calc(var(--spacing-unit) * 16);  /* 64px */
  
  /* 语义化间距 */
  --component-gap: var(--spacing-xl);     /* 组件间距 */
  --section-gap: var(--spacing-2xl);     /* 区块间距 */
  --content-padding: var(--spacing-xl);  /* 内容内边距 */
}
```

### 5.3 表面层级

Material 3的表面层级系统：

```css
.surface-elevation {
  /* 基础表面 */
  --elevation-0: none;
  
  /* 轻微提升 */
  --elevation-1: 0 1px 3px rgba(0, 0, 0, 0.12), 
                 0 1px 2px rgba(0, 0, 0, 0.08);
  
  /* 标准提升 */
  --elevation-2: 0 2px 6px rgba(0, 0, 0, 0.16), 
                 0 1px 3px rgba(0, 0, 0, 0.12);
  
  /* 中等提升 */
  --elevation-3: 0 4px 12px rgba(0, 0, 0, 0.15), 
                 0 2px 6px rgba(0, 0, 0, 0.12);
  
  /* 高提升 */
  --elevation-4: 0 8px 24px rgba(0, 0, 0, 0.15), 
                 0 4px 12px rgba(0, 0, 0, 0.12);
  
  /* 最高提升 */
  --elevation-5: 0 16px 48px rgba(0, 0, 0, 0.15), 
                 0 8px 24px rgba(0, 0, 0, 0.12);
}
```

## 6. 动效系统

### 6.1 过渡时长

Material 3标准过渡时长：

```css
:root {
  /* 标准过渡 */
  --duration-short1: 50ms;     /* 即时反馈 */
  --duration-short2: 100ms;    /* 快速过渡 */
  --duration-short3: 150ms;    /* 标准交互 */
  --duration-short4: 200ms;    /* 元素移动 */
  
  /* 中等过渡 */
  --duration-medium1: 250ms;   /* 内容变化 */
  --duration-medium2: 300ms;   /* 状态变化 */
  --duration-medium3: 350ms;   /* 复杂动画 */
  --duration-medium4: 400ms;   /* 页面过渡 */
  
  /* 长过渡 */
  --duration-long1: 450ms;     /* 大幅移动 */
  --duration-long2: 500ms;     /* 复杂状态 */
  --duration-long3: 550ms;     /* 完整过渡 */
  --duration-long4: 600ms;     /* 最长标准 */
}
```

### 6.2 缓动函数

```css
:root {
  /* 标准缓动 */
  --easing-standard: cubic-bezier(0.2, 0, 0, 1);
  
  /* 强调缓动 */
  --easing-emphasized: cubic-bezier(0.05, 0.7, 0.1, 1);
  
  /* 减速缓动 */
  --easing-decelerated: cubic-bezier(0, 0, 0.2, 1);
  
  /* 加速缓动 */
  --easing-accelerated: cubic-bezier(0.4, 0, 1, 1);
}
```

### 6.3 常用动效模式

```css
/* 悬停效果 */
.hover-lift {
  transition: transform var(--duration-short4) var(--easing-standard),
              box-shadow var(--duration-short4) var(--easing-standard);
}

.hover-lift:hover {
  transform: translateY(-2px);
  box-shadow: var(--elevation-3);
}

/* 按压效果 */
.press-scale {
  transition: transform var(--duration-short2) var(--easing-standard);
}

.press-scale:active {
  transform: scale(0.98);
}

/* 淡入效果 */
.fade-in {
  opacity: 0;
  animation: fadeIn var(--duration-medium2) var(--easing-standard) forwards;
}

@keyframes fadeIn {
  to {
    opacity: 1;
  }
}

/* 滑入效果 */
.slide-in-up {
  transform: translateY(20px);
  opacity: 0;
  animation: slideInUp var(--duration-medium3) var(--easing-standard) forwards;
}

@keyframes slideInUp {
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
```

## 7. 图标系统

### 7.1 图标规范

使用Lucide React图标库，保持一致的视觉风格：

```tsx
// 图标大小标准
const IconSizes = {
  small: 16,    // 小图标 - 表格、标签
  medium: 20,   // 标准图标 - 按钮、导航
  large: 24,    // 大图标 - 标题、主要功能
  xlarge: 32,   // 超大图标 - 状态显示
} as const;

// 图标使用示例
import { Play, Download, Settings, AlertTriangle } from 'lucide-react';

// 功能分类图标
const SystemIcons = {
  // 操作类
  play: Play,
  pause: Pause,
  stop: Square,
  refresh: RefreshCw,
  download: Download,
  upload: Upload,
  
  // 状态类
  success: CheckCircle,
  warning: AlertTriangle,
  error: XCircle,
  info: Info,
  loading: Loader2,
  
  // 导航类
  menu: Menu,
  close: X,
  back: ArrowLeft,
  forward: ArrowRight,
  expand: ChevronDown,
  collapse: ChevronUp,
  
  // 功能类
  settings: Settings,
  filter: Filter,
  search: Search,
  export: FileDown,
  import: FileUp,
};
```

### 7.2 图标使用原则

- **一致性**：同一功能使用同一图标
- **可识别性**：选择通用、易懂的图标
- **适配性**：确保在不同背景下都清晰可见
- **比例协调**：图标大小与文字匹配

## 8. 数据可视化

### 8.1 速度测试结果展示

```css
/* 速度等级色彩编码 */
.speed-excellent { color: hsl(142, 76%, 36%); }    /* >50 Mbps */
.speed-good { color: hsl(60, 100%, 40%); }         /* 20-50 Mbps */
.speed-fair { color: hsl(45, 93%, 58%); }          /* 5-20 Mbps */
.speed-poor { color: hsl(0, 84%, 60%); }           /* <5 Mbps */

/* 延迟等级色彩编码 */
.latency-excellent { color: hsl(142, 76%, 36%); }  /* <50ms */
.latency-good { color: hsl(60, 100%, 40%); }       /* 50-150ms */
.latency-fair { color: hsl(45, 93%, 58%); }        /* 150-300ms */
.latency-poor { color: hsl(0, 84%, 60%); }         /* >300ms */

/* 解锁状态图标 */
.unlock-supported::before { 
  content: "✅"; 
  margin-right: 4px; 
}

.unlock-partial::before { 
  content: "⚠️"; 
  margin-right: 4px; 
}

.unlock-blocked::before { 
  content: "❌"; 
  margin-right: 4px; 
}

.unlock-testing::before { 
  content: "🔄"; 
  margin-right: 4px; 
  animation: spin 1s linear infinite;
}
```

### 8.2 进度可视化

```css
/* 测试进度条 */
.test-progress {
  width: 100%;
  height: 8px;
  background: var(--outline-variant);
  border-radius: 4px;
  overflow: hidden;
  position: relative;
}

.test-progress-bar {
  height: 100%;
  background: linear-gradient(90deg, 
    var(--primary-600) 0%, 
    var(--primary-500) 100%);
  border-radius: 4px;
  transition: width 300ms var(--easing-standard);
  position: relative;
}

.test-progress-bar::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(90deg, 
    transparent 0%, 
    rgba(255, 255, 255, 0.3) 50%, 
    transparent 100%);
  animation: progressShimmer 2s ease-in-out infinite;
}

@keyframes progressShimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}
```

### 8.3 统计数据展示

```css
/* 统计卡片 */
.stats-card {
  background: var(--surface-container);
  border-radius: 16px;
  padding: 24px;
  text-align: center;
  box-shadow: var(--elevation-1);
  transition: all var(--duration-short4) var(--easing-standard);
}

.stats-card:hover {
  box-shadow: var(--elevation-2);
  transform: translateY(-2px);
}

.stats-number {
  font: var(--display-small);
  font-weight: var(--font-weight-bold);
  color: var(--primary-600);
  font-family: var(--font-family-numeric);
}

.stats-label {
  font: var(--body-medium);
  color: var(--on-surface-variant);
  margin-top: 8px;
}

.stats-change {
  font: var(--label-small);
  margin-top: 4px;
  font-weight: var(--font-weight-medium);
}

.stats-change.positive {
  color: var(--success);
}

.stats-change.negative {
  color: var(--error);
}
```

## 9. 响应式设计

### 9.1 断点系统

```css
:root {
  /* Material 3 断点 */
  --breakpoint-compact: 0px;      /* 手机竖屏 */
  --breakpoint-medium: 600px;     /* 手机横屏/小平板 */
  --breakpoint-expanded: 840px;   /* 平板 */
  --breakpoint-large: 1200px;     /* 桌面 */
  --breakpoint-xlarge: 1600px;    /* 大桌面 */
}

/* 响应式容器 */
.responsive-container {
  width: 100%;
  max-width: 1440px;
  margin: 0 auto;
  padding: 16px;
}

@media (min-width: 600px) {
  .responsive-container {
    padding: 24px;
  }
}

@media (min-width: 840px) {
  .responsive-container {
    padding: 32px;
  }
}

@media (min-width: 1200px) {
  .responsive-container {
    padding: 40px;
  }
}
```

### 9.2 自适应组件

```css
/* 自适应卡片网格 */
.card-grid {
  display: grid;
  gap: 16px;
  grid-template-columns: 1fr;
}

@media (min-width: 600px) {
  .card-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 24px;
  }
}

@media (min-width: 840px) {
  .card-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 1200px) {
  .card-grid {
    grid-template-columns: repeat(4, 1fr);
    gap: 32px;
  }
}

/* 自适应表格 */
.responsive-table {
  width: 100%;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

@media (max-width: 600px) {
  .responsive-table table {
    min-width: 640px;
  }
  
  .responsive-table th,
  .responsive-table td {
    padding: 8px 12px;
    font-size: 14px;
  }
}
```

### 9.3 移动端优化

```css
/* 触摸友好的交互区域 */
.touch-target {
  min-height: 48px;
  min-width: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 移动端表单优化 */
@media (max-width: 600px) {
  .form-group {
    margin-bottom: 24px;
  }
  
  .form-input {
    font-size: 16px; /* 防止iOS缩放 */
    padding: 16px;
  }
  
  .form-button {
    width: 100%;
    padding: 16px;
    font-size: 16px;
  }
}

/* 移动端导航优化 */
.mobile-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: var(--surface-container);
  border-top: 1px solid var(--outline-variant);
  padding: 8px 0;
}

@media (min-width: 840px) {
  .mobile-nav {
    display: none;
  }
}
```

## 10. 可访问性规范

### 10.1 颜色对比度

确保所有文本都满足WCAG 2.1 AA标准：

```css
/* 文本对比度检查 */
.text-high-contrast {
  /* 确保至少4.5:1的对比度 */
  color: var(--on-surface);
  background: var(--surface);
}

.text-medium-contrast {
  /* 大文本至少3:1的对比度 */
  color: var(--on-surface-variant);
  background: var(--surface);
}

/* 状态色彩的可访问性版本 */
.accessible-success {
  color: hsl(142, 76%, 28%); /* 深化以提高对比度 */
}

.accessible-warning {
  color: hsl(45, 93%, 35%); /* 深化以提高对比度 */
}

.accessible-error {
  color: hsl(0, 84%, 45%); /* 深化以提高对比度 */
}
```

### 10.2 焦点管理

```css
/* 焦点样式 */
.focus-visible {
  outline: 2px solid var(--primary-600);
  outline-offset: 2px;
}

/* 跳过链接 */
.skip-link {
  position: absolute;
  top: -40px;
  left: 6px;
  background: var(--primary-600);
  color: var(--on-primary);
  padding: 8px;
  text-decoration: none;
  border-radius: 4px;
  z-index: 1000;
}

.skip-link:focus {
  top: 6px;
}

/* 焦点陷阱 */
.focus-trap {
  position: relative;
}

.focus-trap::before,
.focus-trap::after {
  content: '';
  position: absolute;
  width: 1px;
  height: 1px;
  opacity: 0;
  pointer-events: none;
}
```

### 10.3 屏幕阅读器支持

```css
/* 屏幕阅读器专用文本 */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

/* 聚焦时显示 */
.sr-only-focusable:focus {
  position: static;
  width: auto;
  height: auto;
  padding: inherit;
  margin: inherit;
  overflow: visible;
  clip: auto;
  white-space: normal;
}
```

## 11. 性能优化

### 11.1 CSS优化

```css
/* 使用CSS自定义属性减少重复 */
:root {
  --shadow-color: 0 0% 0%;
  --shadow-elevation-low: 0.3px 0.5px 0.7px hsl(var(--shadow-color) / 0.34),
                         0.4px 0.8px 1px -1.2px hsl(var(--shadow-color) / 0.34),
                         1px 2px 2.5px -2.5px hsl(var(--shadow-color) / 0.34);
  --shadow-elevation-medium: 0.3px 0.5px 0.7px hsl(var(--shadow-color) / 0.36),
                            0.8px 1.6px 2px -0.8px hsl(var(--shadow-color) / 0.36),
                            2.1px 4.1px 5.2px -1.7px hsl(var(--shadow-color) / 0.36),
                            5px 10px 12.6px -2.5px hsl(var(--shadow-color) / 0.36);
}

/* 优化动画性能 */
.optimized-animation {
  will-change: transform, opacity;
  transform: translateZ(0); /* 启用硬件加速 */
}

.optimized-animation.complete {
  will-change: auto; /* 动画完成后释放 */
}

/* 使用contain提升渲染性能 */
.isolated-component {
  contain: layout style paint;
}
```

### 11.2 资源加载优化

```css
/* 关键CSS内联，非关键CSS延迟加载 */
.critical-above-fold {
  /* 首屏关键样式 */
}

/* 使用CSS Grid和Flexbox减少JavaScript依赖 */
.layout-efficient {
  display: grid;
  grid-template-areas: 
    "header header"
    "sidebar main"
    "footer footer";
  grid-template-columns: 300px 1fr;
  grid-template-rows: auto 1fr auto;
}
```

## 12. 主题系统

### 12.1 动态主题切换

```css
/* 主题变量定义 */
[data-theme="light"] {
  --surface: hsl(0, 0%, 100%);
  --on-surface: hsl(0, 0%, 10%);
  --primary: hsl(280, 18%, 58%);
}

[data-theme="dark"] {
  --surface: hsl(0, 0%, 6%);
  --on-surface: hsl(280, 44%, 98%);
  --primary: hsl(280, 25%, 74%);
}

[data-theme="auto"] {
  --surface: hsl(0, 0%, 100%);
  --on-surface: hsl(0, 0%, 10%);
}

@media (prefers-color-scheme: dark) {
  [data-theme="auto"] {
    --surface: hsl(0, 0%, 6%);
    --on-surface: hsl(280, 44%, 98%);
  }
}

/* 主题切换动画 */
* {
  transition: background-color var(--duration-medium2) var(--easing-standard),
              color var(--duration-medium2) var(--easing-standard),
              border-color var(--duration-medium2) var(--easing-standard);
}
```

### 12.2 个性化定制

```css
/* 用户偏好支持 */
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}

@media (prefers-contrast: high) {
  :root {
    --outline: hsl(0, 0%, 0%);
    --outline-variant: hsl(0, 0%, 30%);
  }
}

/* 字体大小偏好 */
.text-size-small { font-size: 0.875em; }
.text-size-normal { font-size: 1em; }
.text-size-large { font-size: 1.125em; }
.text-size-xlarge { font-size: 1.25em; }
```

## 13. 实施指南

### 13.1 逐步迁移策略

1. **阶段一：基础系统**
   - 建立色彩系统和CSS自定义属性
   - 实现字体系统和间距标准
   - 更新基础组件（按钮、输入框）

2. **阶段二：组件升级**
   - 升级卡片和表格组件
   - 实现新的表单控件
   - 添加动效系统

3. **阶段三：高级特性**
   - 实现响应式优化
   - 添加可访问性特性
   - 性能优化

### 13.2 代码组织

```
src/
├── styles/
│   ├── foundations/
│   │   ├── colors.css          # 色彩系统
│   │   ├── typography.css      # 字体系统
│   │   ├── spacing.css         # 间距系统
│   │   └── elevation.css       # 层级系统
│   ├── components/
│   │   ├── buttons.css         # 按钮组件
│   │   ├── cards.css           # 卡片组件
│   │   ├── forms.css           # 表单组件
│   │   └── tables.css          # 表格组件
│   ├── layout/
│   │   ├── grid.css            # 栅格系统
│   │   └── responsive.css      # 响应式
│   ├── utilities/
│   │   ├── animations.css      # 动效工具
│   │   └── accessibility.css   # 可访问性
│   └── themes/
│       ├── light.css           # 浅色主题
│       └── dark.css            # 深色主题
└── components/
    ├── ui/                     # 基础UI组件
    └── features/               # 功能组件
```

### 13.3 质量保证

1. **设计审查清单**
   - [ ] 颜色对比度符合WCAG标准
   - [ ] 组件在不同屏幕尺寸下正常显示
   - [ ] 交互元素有适当的反馈
   - [ ] 动效不会引起视觉疲劳
   - [ ] 支持键盘导航

2. **性能检查清单**
   - [ ] CSS文件大小优化
   - [ ] 减少不必要的重排和重绘
   - [ ] 动画使用GPU加速
   - [ ] 响应式图片优化

3. **可访问性检查清单**
   - [ ] 所有交互元素都可以用键盘访问
   - [ ] 表单字段有适当的标签
   - [ ] 色彩不是传达信息的唯一方式
   - [ ] 支持屏幕阅读器

## 14. 总结

本规范文档基于Material 3设计原则，结合Clash SpeedTest项目的特色薰衣草紫主题，提供了完整的前端设计系统。通过统一的色彩、字体、组件和布局标准，确保用户界面的一致性和专业性。

关键特色：
- **紫色主题**：保持项目特色的薰衣草紫色系
- **Material 3**：采用最新的Material Design 3.0规范
- **响应式设计**：适配各种设备和屏幕尺寸
- **可访问性**：确保所有用户都能良好使用
- **性能优化**：注重加载速度和渲染性能

实施时请遵循渐进式迁移策略，确保在升级过程中不影响现有功能。同时建议建立设计审查流程，确保所有新增功能都符合本规范要求。