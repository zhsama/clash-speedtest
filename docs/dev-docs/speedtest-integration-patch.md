# SpeedTest.tsx 集成补丁

## 需要添加的导入

在文件顶部添加：

```typescript
import { History } from "lucide-react"
import { useTestResultSaver } from '../hooks/useTestResultSaver';
import TestHistoryModal from './TestHistoryModal';
```

## 需要添加的状态

在组件内部添加：

```typescript
// 历史记录相关状态
const [showHistory, setShowHistory] = useState(false);
const { saveTestSession } = useTestResultSaver();
```

## 自动保存逻辑

在现有的useEffect后添加：

```typescript
// 自动保存测试完成结果
useEffect(() => {
  if (testCompleteData && testStartData && testResults.length > 0) {
    // 异步保存，不阻塞UI
    saveTestSession(
      testStartData,
      testResults,
      testCompleteData
    ).catch(console.error);
  }
}, [testCompleteData, testStartData, testResults, saveTestSession]);
```

## UI修改

### 1. 在页面标题旁添加历史记录按钮

找到现有的页面标题区域，添加历史记录按钮：

```typescript
{/* 在现有的页面头部区域添加 */}
<div className="flex justify-between items-center mb-6">
  <div>
    <h1 className="text-2xl font-bold text-lavender-50">Clash SpeedTest</h1>
    <p className="text-lavender-300 text-sm">智能代理速度测试工具</p>
  </div>
  
  {/* 新增历史记录按钮 */}
  <Button 
    onClick={() => setShowHistory(true)}
    className="btn-outlined"
  >
    <ClientIcon icon={History} className="h-4 w-4 mr-2" />
    历史记录
  </Button>
</div>
```

### 2. 在组件返回的JSX最后添加历史记录模态框

```typescript
{/* 在return语句的最后，</div>之前添加 */}
{showHistory && (
  <TestHistoryModal onClose={() => setShowHistory(false)} />
)}
```

## 完整的修改说明

### 修改位置1：导入部分
在文件开头的导入语句中添加：
- `History` 图标
- `useTestResultSaver` Hook
- `TestHistoryModal` 组件

### 修改位置2：组件状态
在现有状态声明后添加历史记录相关状态。

### 修改位置3：效果监听
在现有useEffect后添加自动保存逻辑。

### 修改位置4：UI结构
在页面标题区域添加历史记录按钮，在组件结尾添加模态框。

## 最小侵入性

这个集成方案确保：

1. **不影响现有功能** - 所有现有的测试逻辑保持不变
2. **自动无感保存** - 测试完成后自动保存，用户无需操作
3. **可选功能** - 历史记录查看是可选功能，不影响主流程
4. **错误隔离** - 保存失败不会影响测试结果显示

## 文件依赖

确保以下文件存在：
- `/hooks/useTestResultSaver.ts` - 保存功能Hook
- `/components/TestHistoryModal.tsx` - 历史记录组件
- `/components/ui/dialog.tsx` - 对话框组件（如果不存在需要添加）

这样就完成了基于现有WebSocket数据结构的测试结果保存功能！