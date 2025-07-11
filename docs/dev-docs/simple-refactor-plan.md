# 最简单的代码拆分方案

## 目标
为个人项目设计一个最简单实用的代码拆分方案，确保代码易于阅读和维护，同时避免过度工程化。

## 设计原则
1. **最小改动原则** - 只移动代码，不改变逻辑
2. **高内聚原则** - 相关代码放在一起
3. **简单直观原则** - 文件结构一目了然
4. **渐进式原则** - 可以分步实施

## 拆分方案

### 1. backend/ 目录结构调整

#### 现状
```
backend/
├── main.go (1,347行)
└── speedtester/
    └── speedtester.go (1,276行)
```

#### 拆分后
```
backend/
├── main.go        # 主程序和核心逻辑 (~600行)
├── types.go       # 数据结构定义 (~200行)
├── handlers.go    # HTTP处理器 (~500行)
└── speedtester/
    ├── speedtester.go  # 核心测试逻辑 (~500行)
    ├── types.go        # 数据结构 (~150行)
    ├── loader.go       # 配置加载 (~300行)
    └── errors.go       # 错误处理 (~100行)
```

### 2. 具体拆分内容

#### 2.1 backend/types.go
移入所有数据结构定义：
- `TestRequest` - API请求结构
- `TestResponse` - API响应结构
- `TestTask` - 任务管理结构
- `ProtocolsResponse` - 协议列表响应
- `NodeInfo` - 节点信息
- `NodesResponse` - 节点列表响应
- `loggingResponseWriter` - 日志响应写入器

#### 2.2 backend/handlers.go
移入所有HTTP处理函数：
- `handleHealth` - 健康检查
- `handleTestAsync` - 异步测试
- `handleTest` - 同步测试
- `handleTestWithWebSocket` - WebSocket测试
- `handleGetProtocols` - 获取协议列表
- `handleGetNodes` - 获取节点列表
- `handleExportResults` - 导出结果
- `handleLogManagement` - 日志管理
- `sendError` / `sendSuccess` - 响应辅助函数

#### 2.3 backend/speedtester/types.go
移入测速相关的数据结构：
- `Config` - 测速配置
- `Result` - 测试结果
- `CProxy` - 代理配置
- `RawConfig` - 原始配置
- `VlessTestError` - VLESS错误
- 所有相关常量

#### 2.4 backend/speedtester/loader.go
移入配置加载相关功能：
- `LoadProxies` - 主加载函数
- 配置文件解析
- URL订阅处理
- Base64解码
- 代理过滤逻辑

#### 2.5 backend/speedtester/errors.go
移入错误处理相关：
- `VlessTestError` 实现
- `AnalyzeError` - 错误分析
- `IsVlessProtocol` - 协议判断
- 错误常量定义

### 3. 实施步骤

#### 第一步：创建新文件
```bash
cd backend
touch types.go handlers.go
cd speedtester
touch types.go loader.go errors.go
```

#### 第二步：移动代码
1. 先移动类型定义到types.go
2. 再移动独立函数到对应文件
3. 确保import语句正确

#### 第三步：测试验证
```bash
# 格式化代码
go fmt ./...

# 构建测试
go build .

# 运行基本测试
./clash-speedtest -c config.yaml
```

### 4. 注意事项

1. **包名保持不变** - 新文件使用相同的package声明
2. **导入路径不变** - 不需要修改任何import
3. **逐步进行** - 可以一次只拆分一个文件
4. **保留注释** - 移动代码时保留原有注释

### 5. 拆分收益

1. **更容易定位代码**
   - 类型定义在types.go
   - HTTP处理在handlers.go
   - 配置加载在loader.go

2. **减少单文件大小**
   - main.go: 1,347行 → ~600行
   - speedtester.go: 1,276行 → ~500行

3. **提高开发效率**
   - 减少滚动查找
   - 相关代码集中
   - 修改时影响范围更小

4. **便于后续维护**
   - 新功能知道该放哪里
   - 代码审查更容易
   - 合并冲突更少

### 6. 后续可选优化

如果项目继续增长，可以考虑：
1. 将WebSocket相关代码单独分离
2. 将日志管理功能模块化
3. 添加单元测试文件

但对于当前规模，上述拆分已经足够。

## 总结

这个方案通过最小的改动实现了代码的模块化，让代码结构更清晰，同时保持了简单性。适合个人项目的维护需求，避免了过度工程化的问题。