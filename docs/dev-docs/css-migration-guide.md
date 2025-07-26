# CSS迁移指南 - 从旧样式到Material 3

## 📋 概述

本指南帮助开发团队将Clash SpeedTest项目的现有CSS样式迁移到新的Material 3设计系统。新系统提供了更现代、一致、可访问的用户界面。

## 🔄 样式类映射表

### 按钮样式迁移

| 旧样式类 | 新样式类 | 说明 |
|---------|---------|------|
| `.button-standard` | `.btn-filled` | 主要操作按钮 |
| `.button-gradient` | `.btn-filled` | 填充样式按钮 |
| `variant="outline"` | `.btn-outlined` | 次要操作按钮 |
| 无对应 | `.btn-text` | 辅助操作按钮 |

### 卡片样式迁移

| 旧样式类 | 新样式类 | 说明 |
|---------|---------|------|
| `.card-standard` | `.card-elevated` | 带阴影的卡片 |
| `.glass-morphism` | `.card-filled` | 填充背景卡片 |
| 无对应 | `.card-outlined` | 描边卡片 |

### 输入组件迁移

| 旧样式类 | 新样式类 | 说明 |
|---------|---------|------|
| `.input-standard` | `.input-outlined` | 描边输入框 |
| `.input-dark` | `.input-outlined` | 描边输入框 |
| 无对应 | `.input-filled` | 填充输入框 |

### 表格样式迁移

| 旧样式类 | 新样式类 | 说明 |
|---------|---------|------|
| `.table-scroll-container` | `.table-container` | 表格容器 |
| `.table-standard` | `.table-modern` | 现代表格样式 |
| `.table-dark` | `.table-modern` | 现代表格样式 |
| `.table-wrapper` | `.table-container.scrollbar-modern` | 带滚动条的表格 |

### 徽章样式迁移

| 旧样式类 | 新样式类 | 说明 |
|---------|---------|------|
| `.badge-standard` | `.badge-filled` | 填充徽章 |
| `.badge-dark` | `.badge-filled` | 填充徽章 |
| 无对应 | `.badge-outlined` | 描边徽章 |

### 状态指示器迁移

| 旧样式类 | 新样式类 | 说明 |
|---------|---------|------|
| `.status-dot.success` | `.status-success .status-dot` | 成功状态 |
| `.status-dot.error` | `.status-error .status-dot` | 错误状态 |
| `.status-dot.warning` | `.status-warning .status-dot` | 警告状态 |
| 无对应 | `.status-info .status-dot` | 信息状态 |

## 🎨 色彩Token迁移

### 旧色彩变量 → 新Token

```css
/* 旧变量 */
--lavender-600 → --primary
--lavender-700 → --primary (hover状态)
--lavender-400 → --ring
--lavender-50 → --foreground
--lavender-800 → --muted
--lavender-500 → --border

/* 语义化色彩 */
自定义红色 → --destructive
自定义绿色 → --success
自定义黄色 → --warning
自定义蓝色 → --info
```

## 📝 组件迁移示例

### 1. 按钮组件迁移

**旧代码：**
```tsx
<Button className="button-standard">
  开始测试
</Button>
```

**新代码：**
```tsx
<Button className="btn-filled">
  开始测试
</Button>
```

### 2. 卡片组件迁移

**旧代码：**
```tsx
<Card className="card-standard">
  <div className="form-element">
    内容
  </div>
</Card>
```

**新代码：**
```tsx
<Card className="card-elevated">
  <div style={{ marginBottom: 'var(--md-space-4)' }}>
    内容
  </div>
</Card>
```

### 3. 表格组件迁移

**旧代码：**
```tsx
<div className="table-scroll-container">
  <div className="table-scroll-content">
    <Table className="table-standard">
      {/* 表格内容 */}
    </Table>
  </div>
</div>
```

**新代码：**
```tsx
<div className="table-container scrollbar-modern">
  <Table className="table-modern">
    {/* 表格内容 */}
  </Table>
</div>
```

### 4. 输入组件迁移

**旧代码：**
```tsx
<Input className="input-standard" placeholder="输入内容..." />
```

**新代码：**
```tsx
<Input className="input-outlined" placeholder="输入内容..." />
```

### 5. 徽章组件迁移

**旧代码：**
```tsx
<span className="badge-standard">
  vmess
</span>
```

**新代码：**
```tsx
<span className="badge-filled protocol-vmess">
  vmess
</span>
```

## 🔧 特殊样式迁移

### 1. 进度指示器

**旧代码：**
```tsx
<div className="progress-indicator" style={{ width: '60%' }} />
```

**新代码：**
```tsx
<div className="progress-linear">
  <div className="progress-linear-indicator" style={{ width: '60%' }} />
</div>
```

### 2. 状态指示器

**旧代码：**
```tsx
<div className="status-indicator">
  <div className="status-dot success" />
  <span>已连接</span>
</div>
```

**新代码：**
```tsx
<div className="status-indicator status-success">
  <div className="status-dot" />
  <span>已连接</span>
</div>
```

### 3. 动画效果

**旧样式：**
```css
.animate-pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}
```

**新样式：**
```css
.animate-pulse-gentle {
  animation: pulseGentle 2s var(--md-motion-easing-standard) infinite;
}
```

## 📦 渐进式迁移策略

### 阶段1：基础组件迁移（第1周）
1. 更新全局CSS文件
2. 迁移按钮组件
3. 迁移卡片组件
4. 迁移输入组件

### 阶段2：表格和数据展示（第2周）
1. 迁移表格组件
2. 迁移徽章组件
3. 迁移状态指示器
4. 迁移进度指示器

### 阶段3：高级组件和动效（第3周）
1. 实现surface层级系统
2. 添加Material 3动效
3. 优化响应式设计
4. 完善无障碍访问

### 阶段4：测试和优化（第4周）
1. 跨浏览器测试
2. 性能优化
3. 无障碍访问测试
4. 视觉回归测试

## ⚠️ 注意事项

### 1. 兼容性考虑
- 新样式系统需要现代浏览器支持
- CSS自定义属性需要IE11+支持
- 如需兼容旧浏览器，考虑使用CSS后备值

### 2. 性能考虑
- 新系统使用了更多CSS自定义属性，可能略微影响性能
- 建议在生产环境使用CSS压缩和优化
- 考虑移除未使用的旧样式

### 3. 测试清单
- [ ] 所有组件在暗色模式下正常显示
- [ ] 响应式设计在所有断点正常工作
- [ ] 动效在减少动画偏好下被禁用
- [ ] 高对比度模式下文本清晰可读
- [ ] 键盘导航功能正常
- [ ] 屏幕阅读器兼容性

## 🎯 质量保证

### 视觉回归测试
使用以下工具进行视觉测试：
- Percy或Chromatic进行视觉回归
- 手动测试不同设备和浏览器
- 验证色彩对比度符合WCAG标准

### 性能测试
- 使用Lighthouse检查性能分数
- 测试CSS加载时间
- 验证没有未使用的CSS

### 无障碍测试
- 使用axe-core进行自动化测试
- 使用屏幕阅读器进行手动测试
- 验证键盘导航功能

## 📚 参考资源

- [Material Design 3官方文档](https://m3.material.io/)
- [CSS自定义属性MDN文档](https://developer.mozilla.org/en-US/docs/Web/CSS/--*)
- [WCAG 2.1无障碍指南](https://www.w3.org/WAI/WCAG21/quickref/)
- [前端规范文档](./frontend-specification-material3.md)

---

通过遵循这个迁移指南，团队可以有序地将现有界面更新到现代化的Material 3设计系统，提供更好的用户体验和开发者体验。