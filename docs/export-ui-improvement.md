# 导出按钮位置优化

## 改进说明

将测试结果导出按钮从独立的卡片区域移动到了**测试结果标题栏的右侧**，让界面更加紧凑和直观。

## 改进前后对比

### 改进前
```
测试进度卡片
测试完成统计卡片

测试结果表格

导出测试结果卡片
├─ 标题: 导出测试结果  
└─ 按钮: [导出 Markdown] [导出 CSV]
```

### 改进后  
```
测试进度卡片
测试完成统计卡片

测试结果卡片
├─ 测试结果 ················ [导出 Markdown] [导出 CSV]

测试结果表格
```

## 技术实现

### 1. 修改 RealTimeProgressTable 组件

**增加了 Props**：
```typescript
interface RealTimeProgressTableProps {
  // ... 原有 props
  // 导出功能相关
  onExportMarkdown?: () => void
  onExportCSV?: () => void
  showExportButtons?: boolean
}
```

**添加了标题栏**：
```tsx
{results.length > 0 && (
  <Card className="card-dark p-4">
    <div className="flex items-center justify-between">
      <h4 className="text-lg font-semibold text-lavender-50 flex items-center gap-2">
        <ClientIcon icon={Download} className="h-5 w-5 text-lavender-400" />
        测试结果
      </h4>
      {showExportButtons && (onExportMarkdown || onExportCSV) && (
        <div className="flex gap-3">
          {/* 导出按钮 */}
        </div>
      )}
    </div>
  </Card>
)}
```

### 2. 更新 SpeedTest 组件

**移除了独立的导出卡片**，将导出功能通过 props 传递给 RealTimeProgressTable：

```tsx
<RealTimeProgressTable
  // ... 原有 props
  onExportMarkdown={exportToMarkdown}
  onExportCSV={exportToCSV}
  showExportButtons={Boolean((testCompleteData || testCancelledData) && testResults.length > 0)}
/>
```

## 用户体验改进

### ✅ 优势

1. **空间节省**：减少了一个独立卡片的占用空间
2. **操作便捷**：导出按钮紧邻测试结果，操作更直观
3. **视觉整洁**：界面层次更清晰，不会有冗余的卡片
4. **逻辑相关**：导出功能与测试结果在同一视觉区域

### 🎯 显示逻辑

- **测试进行中**：不显示导出按钮
- **测试完成**：显示导出按钮在测试结果标题右侧
- **有测试结果**：才显示整个标题栏和导出按钮
- **无测试结果**：不显示标题栏

### 💡 设计细节

1. **标题与按钮对齐**：使用 `justify-between` 让标题和按钮分别左右对齐
2. **视觉层次**：标题使用较大字体，按钮使用 `outline` 样式不抢夺注意力
3. **间距处理**：按钮之间有适当间距（`gap-3`）
4. **图标一致性**：标题和按钮都使用相应的图标保持视觉一致

## 样式效果

最终效果如下：

```
┌─────────────────────────────────────────────────────────────┐
│ 📊 测试结果                    [📄 导出 Markdown] [📊 导出 CSV] │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ 测试结果表格内容...                                          │
│ ┌─────────┬─────────┬─────────┬─────────┐                  │
│ │ 节点名称 │ 延迟     │ 下载速度 │ 状态     │                  │
│ ├─────────┼─────────┼─────────┼─────────┤                  │
│ │ 香港01   │ 45ms    │ 54.3MB/s│ 成功     │                  │
│ └─────────┴─────────┴─────────┴─────────┘                  │
└─────────────────────────────────────────────────────────────┘
```

这个改进让界面更加紧凑实用，符合用户的操作习惯和期望。