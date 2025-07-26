# SpeedTest组件集成示例

这个文件展示了如何将测试结果保存功能集成到现有的SpeedTest组件中。

## 1. 添加导入语句

在SpeedTest.tsx文件顶部的导入部分添加：

```typescript
// 在现有导入后添加
import { History } from "lucide-react"
import { useTestResultSaver } from '../hooks/useTestResultSaver';
import TestHistoryModal from './TestHistoryModal';
```

## 2. 添加状态管理

在SpeedTest组件函数内部，在现有状态声明后添加：

```typescript
export default function SpeedTest() {
  // ... 现有的状态声明 ...
  
  // 新增：历史记录相关状态
  const [showHistory, setShowHistory] = useState(false);
  const { saveTestSession } = useTestResultSaver();
  
  // ... 现有的useWebSocket和其他逻辑 ...
```

## 3. 添加自动保存逻辑

在现有的useEffect之后添加：

```typescript
  // 新增：自动保存测试完成结果
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

## 4. 修改UI结构

### 方案A：在页面顶部添加历史记录按钮

找到SpeedTest组件的return语句，在主要内容的顶部添加：

```typescript
return (
  <div className="container mx-auto p-6 space-y-6">
    {/* 新增：页面头部区域 */}
    <div className="flex justify-between items-center">
      <div>
        <h1 className="text-2xl font-bold text-foreground">Clash SpeedTest</h1>
        <p className="text-muted-foreground text-sm">智能代理速度测试工具</p>
      </div>
      
      <Button 
        onClick={() => setShowHistory(true)}
        className="btn-outlined"
      >
        <ClientIcon icon={History} className="h-4 w-4 mr-2" />
        历史记录
      </Button>
    </div>
    
    {/* 现有的TUN警告组件 */}
    <TUNWarning onTUNStatusChange={setIsTUNEnabled} />
    
    {/* ... 现有的其他组件 ... */}
```

### 方案B：在现有按钮区域添加历史记录按钮

如果你想在现有的测试控制按钮旁添加，找到测试按钮区域并添加：

```typescript
{/* 在现有的测试控制按钮区域 */}
<div className="flex items-center component-gap">
  {/* 现有的开始测试按钮 */}
  <Button
    onClick={handleStartTest}
    disabled={!isConnected || isTestRunning || !configPaths}
    className="btn-filled"
  >
    {/* ... 现有按钮内容 ... */}
  </Button>
  
  {/* 新增：历史记录按钮 */}
  <Button 
    onClick={() => setShowHistory(true)}
    className="btn-outlined"
  >
    <ClientIcon icon={History} className="h-4 w-4 mr-2" />
    历史记录
  </Button>
</div>
```

## 5. 添加历史记录模态框

在组件return语句的最后，在最外层div的结束标签之前添加：

```typescript
    {/* ... 现有的所有组件 ... */}
    
    {/* 新增：历史记录模态框 */}
    {showHistory && (
      <TestHistoryModal onClose={() => setShowHistory(false)} />
    )}
  </div>
); // 这是SpeedTest组件的结束
```

## 完整的修改总结

1. **导入添加** - 3个新的导入语句
2. **状态添加** - 2个新的状态变量
3. **效果添加** - 1个新的useEffect用于自动保存
4. **UI修改** - 1个历史记录按钮和1个模态框组件

## 测试步骤

1. 修改完成后，运行 `npm run dev` 启动开发服务器
2. 进行一次完整的速度测试
3. 测试完成后，点击"历史记录"按钮
4. 验证测试结果是否已自动保存并显示在历史记录中
5. 测试导出功能是否正常工作

## 注意事项

- 所有保存操作都是异步的，不会阻塞UI
- 如果localStorage空间不足，会自动清理旧数据
- 错误处理已内置，保存失败不会影响测试功能
- 历史记录按钮只是触发模态框显示，不会影响现有功能

这样集成后，用户完成测试就会自动保存结果，可以随时查看历史记录和导出数据！
