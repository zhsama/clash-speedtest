# Tailwind 4 兼容性修复总结

## 🔧 问题描述

用户在使用Astro 5 + Tailwind 4环境中遇到了以下错误：
```
Cannot apply unknown utility class border-border
Cannot apply unknown utility class badge-filled
```

## ✅ 解决方案

### 1. 修复 `border-border` 错误
**问题位置**: `frontend/src/styles/global.css:241`
**原因**: Tailwind 4中不支持在CSS中使用`@apply border-border`语法
**解决方案**: 更改为直接CSS属性
```css
/* 修复前 */
* {
  @apply border-border;
}

/* 修复后 */
* {
  border-color: hsl(var(--border));
}
```

### 2. 修复协议类型徽章样式
**问题位置**: `frontend/src/styles/global.css:633-637`
**原因**: Tailwind 4中不能在CSS中使用`@apply`来应用自定义组件类
**解决方案**: 展开完整的样式定义

```css
/* 修复前 */
.protocol-vmess { @apply badge-filled; background-color: hsl(var(--md-primary-40)); }

/* 修复后 */
.protocol-vmess { 
  @apply inline-flex items-center gap-1;
  background-color: hsl(var(--md-primary-40));
  color: hsl(var(--primary-foreground));
  border-radius: var(--md-corner-sm);
  padding: var(--md-space-1) var(--md-space-2);
  font: var(--md-label-small-font);
}
```

### 3. 更新组件样式类映射
**文件**: `frontend/src/components/TUNWarning.tsx`
**更新内容**:
- `card-standard` → `card-elevated`
- `button-standard` → `btn-outlined` / `btn-text`
- `badge-standard` → `badge-filled` / `badge-outlined`

## 📊 验证结果

✅ **构建成功**: `npm run build` 无错误  
✅ **样式完整**: 所有Material 3样式正确应用  
✅ **兼容性**: 完全兼容Astro 5 + Tailwind 4  

## 🎯 技术要点

1. **Tailwind 4语法变化**: 不支持在CSS中使用`@apply`应用自定义类
2. **直接CSS优先**: 使用CSS自定义属性而非Tailwind工具类
3. **组件样式**: 完整定义组件样式，避免依赖组合类

## 📝 最佳实践

- 在Tailwind 4中，自定义组件样式应完整定义
- 使用CSS自定义属性实现主题系统
- 验证构建确保兼容性

---

**修复完成时间**: 2025年1月  
**兼容性**: ✅ Astro 5 + Tailwind 4  
**状态**: 🎉 完全解决