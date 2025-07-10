import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Checkbox } from "@/components/ui/checkbox"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { Slider } from "@/components/ui/slider"
import { toast } from "sonner"
import {
  Play,
  Download,
  Filter,
  Globe,
  ServerCog,
  RefreshCw,
  AlertCircle,
  CheckCircle2,
  Loader2,
} from "lucide-react"
import ClientIcon from "./ClientIcon"
import RealTimeProgressTable from "./RealTimeProgressTable"
import { useWebSocket } from "../hooks/useWebSocket"
import { config } from "../lib/env"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

interface NodeInfo {
  name: string
  type: string
  server: string
  port: number
}

interface FilterConfig {
  includeNodes: string[]
  excludeNodes: string[]
  protocolFilter: string[]
  minDownloadSpeed: number
  minUploadSpeed: number
  maxLatency: number
  stashCompatible: boolean
}

interface TestConfig {
  configPaths: string
  serverUrl: string
  downloadSize: number
  uploadSize: number
  timeout: number
  concurrent: number
}

export default function SpeedTestPro() {
  // 状态管理
  const [configUrl, setConfigUrl] = useState("")
  const [nodes, setNodes] = useState<NodeInfo[]>([])
  const [filteredNodes, setFilteredNodes] = useState<NodeInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [testing, setTesting] = useState(false)
  const [taskId, setTaskId] = useState<string | null>(null)
  
  // 过滤配置
  const [filterConfig, setFilterConfig] = useState<FilterConfig>({
    includeNodes: [],
    excludeNodes: [],
    protocolFilter: [],
    minDownloadSpeed: 5,
    minUploadSpeed: 2,
    maxLatency: 3000,
    stashCompatible: false,
  })
  
  // 测试配置
  const [testConfig, setTestConfig] = useState<TestConfig>({
    configPaths: "",
    serverUrl: "https://speed.cloudflare.com",
    downloadSize: 50,
    uploadSize: 20,
    timeout: 10,
    concurrent: 4,
  })
  
  // UI状态
  const [includeNodesInput, setIncludeNodesInput] = useState("")
  const [excludeNodesInput, setExcludeNodesInput] = useState("")
  const [availableProtocols, setAvailableProtocols] = useState<string[]>([])
  
  // WebSocket
  const wsUrl = `${config.wsUrl}/ws`
  const {
    isConnected,
    connect,
    disconnect,
    sendMessage,
    testProgress,
    testResults,
    testCompleteData,
    testCancelledData,
    clearData
  } = useWebSocket(wsUrl)
  
  // 从localStorage加载配置
  useEffect(() => {
    const savedConfig = localStorage.getItem("clash-speedtest-config")
    if (savedConfig) {
      try {
        const parsed = JSON.parse(savedConfig)
        console.log(parsed);
        if (parsed.configUrl) setConfigUrl(parsed.configUrl)
        if (parsed.filterConfig) {
          setFilterConfig(parsed.filterConfig)
          handleIncludeNodesChange(parsed.filterConfig.includeNodes?.join(', ') || '')
          handleExcludeNodesChange(parsed.filterConfig.excludeNodes?.join(', ') || '')
        }
        if (parsed.testConfig) setTestConfig(prev => ({ ...prev, ...parsed.testConfig }))
      } catch (error) {
        console.error("Failed to load saved config:", error)
      }
    }
  }, [])
  
  // 保存配置到localStorage
  useEffect(() => {
    localStorage.setItem("clash-speedtest-config", JSON.stringify({
      configUrl,
      filterConfig,
      testConfig
    }))
  }, [configUrl, filterConfig, testConfig])
  
  // 连接WebSocket
  useEffect(() => {
    connect()
    return () => disconnect()
  }, [connect, disconnect])
  
  // 处理测试完成
  useEffect(() => {
    if (testCompleteData && testing) {
      setTesting(false)
      setTaskId(null)
      toast.success(
        `测试完成！成功: ${testCompleteData.successful_tests}, 失败: ${testCompleteData.failed_tests}`
      )
    }
  }, [testCompleteData, testing])
  
  // 处理测试取消
  useEffect(() => {
    if (testCancelledData && testing) {
      setTesting(false)
      setTaskId(null)
      toast.info(
        `测试已取消！已完成: ${testCancelledData.completed_tests}/${testCancelledData.total_tests}`
      )
    }
  }, [testCancelledData, testing])
  
  // 获取配置文件
  const fetchConfig = async () => {
    if (!configUrl.trim()) {
      toast.error("请输入配置文件路径或订阅链接")
      return
    }
    
    setLoading(true)
    setNodes([])
    setFilteredNodes([])
    
    try {
      const response = await fetch(`${config.apiUrl}/api/nodes`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          configPaths: configUrl,
          stashCompatible: filterConfig.stashCompatible,
        }),
      })
      
      const data = await response.json()
      
      if (data.success && data.nodes) {
        setNodes(data.nodes)
        
        // 提取可用协议
        const protocols = [...new Set(data.nodes.map((n: NodeInfo) => n.type))]
        setAvailableProtocols(protocols as string[])
        
        toast.success(`成功加载 ${data.nodes.length} 个节点`)
        
        // 自动应用过滤
        applyFilters(data.nodes)
      } else {
        toast.error(data.error || "加载配置失败")
      }
    } catch (error) {
      toast.error("请求失败：" + (error as Error).message)
    } finally {
      setLoading(false)
    }
  }
  
  // 应用过滤条件
  const applyFilters = (nodesToFilter: NodeInfo[] = nodes) => {
    // 这里只做客户端预览，实际过滤在后端进行
    let filtered = [...nodesToFilter]
    
    // 包含过滤
    if (filterConfig.includeNodes.length > 0) {
      filtered = filtered.filter(node =>
        filterConfig.includeNodes.some(include =>
          node.name.toLowerCase().includes(include.toLowerCase())
        )
      )
    }
    
    // 排除过滤
    if (filterConfig.excludeNodes.length > 0) {
      filtered = filtered.filter(node =>
        !filterConfig.excludeNodes.some(exclude =>
          node.name.toLowerCase().includes(exclude.toLowerCase())
        )
      )
    }
    
    // 协议过滤
    if (filterConfig.protocolFilter.length > 0) {
      filtered = filtered.filter(node =>
        filterConfig.protocolFilter.includes(node.type)
      )
    }
    
    setFilteredNodes(filtered)
  }
  
  // 监听过滤条件变化
  useEffect(() => {
    if (nodes.length > 0) {
      applyFilters()
    }
  }, [filterConfig, nodes])
  
  // 开始测试
  const startTest = async () => {
    if (!isConnected) {
      toast.error("WebSocket未连接，正在尝试重新连接...")
      connect()
      return
    }
    
    if (filteredNodes.length === 0) {
      toast.error("没有符合条件的节点可以测试")
      return
    }
    
    setTesting(true)
    clearData()
    
    try {
      // 发送异步测试请求
      const response = await fetch(`${config.apiUrl}/api/test/async`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          ...testConfig,
          configPaths: configUrl,
          ...filterConfig,
          filterRegex: ".+",
        }),
      })
      
      const data = await response.json()
      
      if (data.success && data.taskId) {
        setTaskId(data.taskId)
        toast.success(`测试任务已创建，任务ID: ${data.taskId}`)
      } else {
        toast.error(data.error || "创建测试任务失败")
        setTesting(false)
      }
    } catch (error) {
      toast.error("请求失败：" + (error as Error).message)
      setTesting(false)
    }
  }
  
  // 停止测试
  const stopTest = () => {
    if (isConnected && taskId) {
      sendMessage({
        type: 'stop_test',
        taskId: taskId,
        timestamp: new Date().toISOString()
      })
    }
    setTesting(false)
    toast.info("正在停止测试...")
  }
  
  // 处理包含节点输入
  const handleIncludeNodesChange = (value: string) => {
    setIncludeNodesInput(value)
    const nodes = value.split(',').map(s => s.trim()).filter(s => s.length > 0)
    setFilterConfig(prev => ({ ...prev, includeNodes: nodes }))
  }
  
  // 处理排除节点输入
  const handleExcludeNodesChange = (value: string) => {
    setExcludeNodesInput(value)
    const nodes = value.split(',').map(s => s.trim()).filter(s => s.length > 0)
    setFilterConfig(prev => ({ ...prev, excludeNodes: nodes }))
  }
  
  // 处理协议过滤
  const handleProtocolFilterChange = (protocol: string, checked: boolean) => {
    setFilterConfig(prev => ({
      ...prev,
      protocolFilter: checked
        ? [...prev.protocolFilter, protocol]
        : prev.protocolFilter.filter(p => p !== protocol)
    }))
  }
  
  return (
    <div className="min-h-screen p-8">
      <div className="max-w-7xl mx-auto">
        {/* 头部 */}
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold mb-3">
            <span className="text-gradient">Clash SpeedTest Pro</span>
          </h1>
          <p className="text-gray-400">专业的代理节点性能测试工具</p>
        </div>
        
        {/* 配置获取卡片 */}
        <Card className="glass-morphism border-gray-800 mb-6">
          <div className="p-6">
            <div className="flex items-center gap-2 mb-4">
              <ClientIcon icon={Globe} className="h-5 w-5 text-blue-400" />
              <h2 className="text-lg font-semibold text-white">配置获取</h2>
              <div className="ml-auto">
                {isConnected ? (
                  <Badge variant="outline" className="border-green-500 text-green-400">
                    <div className="w-2 h-2 bg-green-400 rounded-full mr-2 animate-pulse" />
                    WebSocket 已连接
                  </Badge>
                ) : (
                  <Badge variant="outline" className="border-red-500 text-red-400">
                    <div className="w-2 h-2 bg-red-400 rounded-full mr-2" />
                    WebSocket 未连接
                  </Badge>
                )}
              </div>
            </div>
            
            <div className="flex gap-3">
              <Input
                placeholder="输入配置文件路径或订阅链接..."
                value={configUrl}
                onChange={(e) => setConfigUrl(e.target.value)}
                className="flex-1 input-dark text-white placeholder:text-gray-500"
                disabled={loading || testing}
              />
              <Button
                onClick={fetchConfig}
                disabled={loading || testing}
                className="button-gradient min-w-[120px]"
              >
                {loading ? (
                  <>
                    <ClientIcon icon={Loader2} className="mr-2 h-4 w-4 animate-spin" />
                    获取中...
                  </>
                ) : (
                  <>
                    <ClientIcon icon={Download} className="mr-2 h-4 w-4" />
                    获取配置
                  </>
                )}
              </Button>
            </div>
            
            {nodes.length > 0 && (
              <div className="mt-4 flex items-center gap-4">
                <Badge variant="secondary" className="badge-dark">
                  总节点数: {nodes.length}
                </Badge>
                <Badge variant="secondary" className="badge-dark">
                  符合条件: {filteredNodes.length}
                </Badge>
              </div>
            )}
          </div>
        </Card>
        
        {/* 节点列表 */}
        {nodes.length > 0 && !testing && (
          <Card className="glass-morphism border-gray-800 mb-6">
            <div className="p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                  <ClientIcon icon={ServerCog} className="h-5 w-5 text-blue-400" />
                  节点列表
                </h2>
                <Button
                  onClick={() => applyFilters()}
                  variant="outline"
                  size="sm"
                  className="border-gray-700 text-gray-300 hover:text-white"
                >
                  <ClientIcon icon={RefreshCw} className="h-4 w-4 mr-1" />
                  刷新过滤
                </Button>
              </div>
              
              <div className="overflow-x-auto max-h-96">
                <Table className="table-dark">
                  <TableHeader>
                    <TableRow className="border-gray-800">
                      <TableHead className="text-gray-400">状态</TableHead>
                      <TableHead className="text-gray-400">节点名称</TableHead>
                      <TableHead className="text-gray-400">类型</TableHead>
                      <TableHead className="text-gray-400">服务器</TableHead>
                      <TableHead className="text-gray-400">端口</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredNodes.map((node, index) => (
                      <TableRow key={`${node.name}-${index}`} className="table-row-dark">
                        <TableCell>
                          {filterConfig.includeNodes.length > 0 || 
                           filterConfig.excludeNodes.length > 0 || 
                           filterConfig.protocolFilter.length > 0 ? (
                            <CheckCircle2 className="h-4 w-4 text-green-400" />
                          ) : (
                            <div className="w-4 h-4" />
                          )}
                        </TableCell>
                        <TableCell className="font-medium text-white">
                          <div className="truncate max-w-xs" title={node.name}>
                            {node.name}
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary" className="badge-dark text-xs">
                            {node.type}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-gray-400 font-mono text-sm">
                          {node.server}
                        </TableCell>
                        <TableCell className="text-gray-400">
                          {node.port}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          </Card>
        )}
        
        {/* 过滤条件卡片 */}
        <Card className="glass-morphism border-gray-800 mb-6">
          <div className="p-6">
            <div className="flex items-center gap-2 mb-6">
              <ClientIcon icon={Filter} className="h-5 w-5 text-purple-400" />
              <h2 className="text-lg font-semibold text-white">过滤条件</h2>
            </div>
            
            <div className="space-y-6">
              {/* 节点名称过滤 */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <Label className="text-gray-300 mb-2 block">
                    包含节点 (逗号分隔)
                  </Label>
                  <Textarea
                    placeholder="例如: 香港, HK, 新加坡..."
                    value={includeNodesInput}
                    onChange={(e) => handleIncludeNodesChange(e.target.value)}
                    className="input-dark text-white placeholder:text-gray-500 resize-none"
                    rows={2}
                  />
                </div>
                
                <div>
                  <Label className="text-gray-300 mb-2 block">
                    排除节点 (逗号分隔)
                  </Label>
                  <Textarea
                    placeholder="例如: 过期, 测试, 备用..."
                    value={excludeNodesInput}
                    onChange={(e) => handleExcludeNodesChange(e.target.value)}
                    className="input-dark text-white placeholder:text-gray-500 resize-none"
                    rows={2}
                  />
                </div>
              </div>
              
              {/* 协议过滤 */}
              {availableProtocols.length > 0 && (
                <div>
                  <Label className="text-gray-300 mb-3 block">
                    协议过滤
                  </Label>
                  <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-3">
                    {availableProtocols.map((protocol) => (
                      <div key={protocol} className="flex items-center space-x-2">
                        <Checkbox
                          id={`protocol-${protocol}`}
                          checked={filterConfig.protocolFilter.includes(protocol)}
                          onCheckedChange={(checked: boolean) => 
                            handleProtocolFilterChange(protocol, checked)
                          }
                          className="checkbox-dark"
                        />
                        <Label 
                          htmlFor={`protocol-${protocol}`} 
                          className="text-sm text-gray-300 cursor-pointer"
                        >
                          {protocol}
                        </Label>
                      </div>
                    ))}
                  </div>
                </div>
              )}
              
              {/* 性能过滤 */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div>
                  <Label className="text-gray-300 mb-2 block">
                    最低下载速度: {filterConfig.minDownloadSpeed} MB/s
                  </Label>
                  <Slider
                    value={[filterConfig.minDownloadSpeed]}
                    onValueChange={(v) => setFilterConfig(prev => ({ 
                      ...prev, 
                      minDownloadSpeed: v[0] 
                    }))}
                    max={100}
                    min={0}
                    step={5}
                    className="slider-dark"
                  />
                </div>
                
                <div>
                  <Label className="text-gray-300 mb-2 block">
                    最低上传速度: {filterConfig.minUploadSpeed} MB/s
                  </Label>
                  <Slider
                    value={[filterConfig.minUploadSpeed]}
                    onValueChange={(v) => setFilterConfig(prev => ({ 
                      ...prev, 
                      minUploadSpeed: v[0] 
                    }))}
                    max={50}
                    min={0}
                    step={1}
                    className="slider-dark"
                  />
                </div>
                
                <div>
                  <Label className="text-gray-300 mb-2 block">
                    最大延迟: {filterConfig.maxLatency} ms
                  </Label>
                  <Slider
                    value={[filterConfig.maxLatency]}
                    onValueChange={(v) => setFilterConfig(prev => ({ 
                      ...prev, 
                      maxLatency: v[0] 
                    }))}
                    max={5000}
                    min={100}
                    step={100}
                    className="slider-dark"
                  />
                </div>
              </div>
              
              {/* 其他选项 */}
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Switch
                    id="stashCompatible"
                    checked={filterConfig.stashCompatible}
                    onCheckedChange={(checked) => setFilterConfig(prev => ({ 
                      ...prev, 
                      stashCompatible: checked 
                    }))}
                    className="switch-dark"
                  />
                  <Label htmlFor="stashCompatible" className="text-gray-300">
                    Stash 兼容模式
                  </Label>
                </div>
                
                <Button
                  onClick={testing ? stopTest : startTest}
                  disabled={!isConnected || nodes.length === 0 || loading}
                  size="lg"
                  className={testing ? "bg-red-600 hover:bg-red-700" : "button-gradient"}
                >
                  {testing ? (
                    <>
                      <ClientIcon icon={Loader2} className="mr-2 h-4 w-4 animate-spin" />
                      停止测试
                    </>
                  ) : (
                    <>
                      <ClientIcon icon={Play} className="mr-2 h-4 w-4" />
                      开始测试
                    </>
                  )}
                </Button>
              </div>
            </div>
          </div>
        </Card>
        
        {/* 测试配置（折叠） */}
        <details className="mb-6">
          <summary className="cursor-pointer text-gray-400 hover:text-white transition-colors">
            高级测试配置
          </summary>
          <Card className="glass-morphism border-gray-800 mt-4">
            <div className="p-6">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <div>
                  <Label className="text-gray-300 mb-2 block">
                    测试服务器
                  </Label>
                  <Input
                    value={testConfig.serverUrl}
                    onChange={(e) => setTestConfig(prev => ({ 
                      ...prev, 
                      serverUrl: e.target.value 
                    }))}
                    className="input-dark text-white"
                  />
                </div>
                
                <div>
                  <Label className="text-gray-300 mb-2 block">
                    下载测试大小: {testConfig.downloadSize} MB
                  </Label>
                  <Slider
                    value={[testConfig.downloadSize]}
                    onValueChange={(v) => setTestConfig(prev => ({ 
                      ...prev, 
                      downloadSize: v[0] 
                    }))}
                    max={100}
                    min={10}
                    step={10}
                    className="slider-dark"
                  />
                </div>
                
                <div>
                  <Label className="text-gray-300 mb-2 block">
                    并发数: {testConfig.concurrent}
                  </Label>
                  <Slider
                    value={[testConfig.concurrent]}
                    onValueChange={(v) => setTestConfig(prev => ({ 
                      ...prev, 
                      concurrent: v[0] 
                    }))}
                    max={16}
                    min={1}
                    step={1}
                    className="slider-dark"
                  />
                </div>
                
                <div>
                  <Label className="text-gray-300 mb-2 block">
                    超时时间: {testConfig.timeout} 秒
                  </Label>
                  <Slider
                    value={[testConfig.timeout]}
                    onValueChange={(v) => setTestConfig(prev => ({ 
                      ...prev, 
                      timeout: v[0] 
                    }))}
                    max={30}
                    min={5}
                    step={5}
                    className="slider-dark"
                  />
                </div>
              </div>
            </div>
          </Card>
        </details>
        
        {/* 测试进度和结果 */}
        {testing && (
          <RealTimeProgressTable
            results={testResults}
            progress={testProgress}
            completeData={testCompleteData}
            cancelledData={testCancelledData}
            isConnected={isConnected}
          />
        )}
      </div>
    </div>
  )
}