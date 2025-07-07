import React, { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Checkbox } from "@/components/ui/checkbox"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { toast } from "sonner"
import {
  Play,
  Pause,
  Download,
  Settings,
  ChevronDown,
  ChevronUp,
  RotateCcw,
  Filter,
  TestTubes,
  Globe,
  ServerCog,
} from "lucide-react"
import ClientIcon from "./ClientIcon"
import RealTimeProgressTable from "./RealTimeProgressTable"
import { useWebSocket } from "../hooks/useWebSocket"
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
  selected?: boolean
}

interface TestRequest {
  configPaths: string
  filterRegex: string
  includeNodes: string[]
  excludeNodes: string[]
  protocolFilter: string[]
  serverUrl: string
  downloadSize: number
  uploadSize: number
  timeout: number
  concurrent: number
  maxLatency: number
  minDownloadSpeed: number
  minUploadSpeed: number
  stashCompatible: boolean
  renameNodes: boolean
}

export default function SpeedTestV2() {
  const [config, setConfig] = useState<TestRequest>({
    configPaths: "",
    filterRegex: ".+",
    includeNodes: [],
    excludeNodes: [],
    protocolFilter: [],
    serverUrl: "https://speed.cloudflare.com",
    downloadSize: 50,
    uploadSize: 20,
    timeout: 10,
    concurrent: 4,
    maxLatency: 3000,
    minDownloadSpeed: 5,
    minUploadSpeed: 2,
    stashCompatible: false,
    renameNodes: false,
  })

  const [phase, setPhase] = useState<"idle" | "loading" | "filtering" | "testing">("idle")
  const [nodes, setNodes] = useState<NodeInfo[]>([])
  const [filteredNodes, setFilteredNodes] = useState<NodeInfo[]>([])
  const [selectedNodes, setSelectedNodes] = useState<Set<string>>(new Set())
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [includeNodesInput, setIncludeNodesInput] = useState("")
  const [excludeNodesInput, setExcludeNodesInput] = useState("")
  const [availableProtocols, setAvailableProtocols] = useState<string[]>([])
  const [abortController, setAbortController] = useState<AbortController | null>(null)

  // WebSocket hook
  const wsUrl = `ws://localhost:8080/ws`
  const {
    isConnected,
    connect,
    disconnect,
    sendMessage,
    testStartData,
    testProgress,
    testResults,
    testCompleteData,
    testCancelledData,
    error: wsError,
    clearData
  } = useWebSocket(wsUrl)

  // Load config from localStorage on mount
  useEffect(() => {
    const savedConfig = localStorage.getItem("clash-speedtest-config-v2")
    if (savedConfig) {
      try {
        const parsedConfig = JSON.parse(savedConfig)
        setConfig({ ...config, ...parsedConfig })
        setIncludeNodesInput(parsedConfig.includeNodes?.join(', ') || '')
        setExcludeNodesInput(parsedConfig.excludeNodes?.join(', ') || '')
      } catch (error) {
        console.error("Failed to parse saved config:", error)
      }
    }
  }, [])

  // Save config to localStorage whenever it changes
  useEffect(() => {
    localStorage.setItem("clash-speedtest-config-v2", JSON.stringify(config))
  }, [config])

  // Connect to WebSocket on mount
  useEffect(() => {
    connect()
    return () => {
      disconnect()
    }
  }, [connect, disconnect])

  // Handle include/exclude nodes input changes
  const handleIncludeNodesChange = (value: string) => {
    setIncludeNodesInput(value)
    const nodes = value.split(',').map(s => s.trim()).filter(s => s.length > 0)
    setConfig(prev => ({ ...prev, includeNodes: nodes }))
  }

  const handleExcludeNodesChange = (value: string) => {
    setExcludeNodesInput(value)
    const nodes = value.split(',').map(s => s.trim()).filter(s => s.length > 0)
    setConfig(prev => ({ ...prev, excludeNodes: nodes }))
  }

  // Handle protocol filter changes
  const handleProtocolFilterChange = (protocol: string, checked: boolean) => {
    const currentProtocolFilter = config.protocolFilter || []
    const newProtocolFilter = checked
      ? [...currentProtocolFilter, protocol]
      : currentProtocolFilter.filter(p => p !== protocol)
    setConfig(prev => ({ ...prev, protocolFilter: newProtocolFilter }))
  }

  // Phase 1: Load nodes from config
  const loadNodes = async () => {
    if (!config.configPaths) {
      toast.error("请输入配置文件路径")
      return
    }

    setPhase("loading")
    setNodes([])
    setFilteredNodes([])
    setSelectedNodes(new Set())

    try {
      const response = await fetch("http://localhost:8080/api/nodes", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          configPaths: config.configPaths,
          includeNodes: config.includeNodes,
          excludeNodes: config.excludeNodes,
          protocolFilter: config.protocolFilter,
          stashCompatible: config.stashCompatible,
        }),
      })

      const data = await response.json()

      if (data.success && data.nodes) {
        setNodes(data.nodes)
        setFilteredNodes(data.nodes)
        
        // Extract available protocols
        const protocols = [...new Set(data.nodes.map((n: NodeInfo) => n.type))]
        setAvailableProtocols(protocols as string[])
        
        // Select all nodes by default
        setSelectedNodes(new Set(data.nodes.map((n: NodeInfo) => n.name)))
        
        setPhase("filtering")
        toast.success(`成功加载 ${data.nodes.length} 个节点`)
      } else {
        toast.error(data.error || "加载节点失败")
        setPhase("idle")
      }
    } catch (error) {
      toast.error("请求失败：" + (error as Error).message)
      setPhase("idle")
    }
  }

  // Apply filters to nodes
  const applyFilters = () => {
    let filtered = [...nodes]

    // Apply include filter
    if (config.includeNodes.length > 0) {
      filtered = filtered.filter(node => 
        config.includeNodes.some(include => 
          node.name.toLowerCase().includes(include.toLowerCase())
        )
      )
    }

    // Apply exclude filter
    if (config.excludeNodes.length > 0) {
      filtered = filtered.filter(node => 
        !config.excludeNodes.some(exclude => 
          node.name.toLowerCase().includes(exclude.toLowerCase())
        )
      )
    }

    // Apply protocol filter
    if (config.protocolFilter.length > 0) {
      filtered = filtered.filter(node => 
        config.protocolFilter.includes(node.type)
      )
    }

    setFilteredNodes(filtered)
    
    // Update selected nodes to only include filtered ones
    const filteredNames = new Set(filtered.map(n => n.name))
    setSelectedNodes(prev => {
      const newSelected = new Set<string>()
      prev.forEach(name => {
        if (filteredNames.has(name)) {
          newSelected.add(name)
        }
      })
      return newSelected
    })
  }

  // Apply filters when they change
  useEffect(() => {
    if (phase === "filtering" && nodes.length > 0) {
      applyFilters()
    }
  }, [config.includeNodes, config.excludeNodes, config.protocolFilter, nodes, phase])

  // Phase 2: Test selected nodes
  const testSelectedNodes = async () => {
    if (selectedNodes.size === 0) {
      toast.error("请至少选择一个节点进行测试")
      return
    }

    if (!isConnected) {
      toast.error("WebSocket未连接，正在尝试重新连接...")
      connect()
      return
    }

    // Create new AbortController for this test
    const controller = new AbortController()
    setAbortController(controller)
    
    setPhase("testing")
    clearData()

    // Create a filtered config that only includes selected nodes
    const selectedNodeNames = Array.from(selectedNodes)
    const testConfig = {
      ...config,
      includeNodes: selectedNodeNames,
      excludeNodes: [],
    }

    try {
      const response = await fetch("http://localhost:8080/api/test", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(testConfig),
        signal: controller.signal,
      })

      const data = await response.json()

      if (!data.success) {
        toast.error(data.error || "测试失败")
        setPhase("filtering")
      }
    } catch (error) {
      if (error instanceof Error && error.name === 'AbortError') {
        toast.info("测试已被取消")
      } else {
        toast.error("请求失败：" + (error as Error).message)
      }
      setPhase("filtering")
    } finally {
      setAbortController(null)
    }
  }

  // Stop current operation
  const stopOperation = () => {
    if (abortController) {
      abortController.abort()
    }
    
    if (phase === "testing" && isConnected) {
      sendMessage({
        type: 'stop_test',
        timestamp: new Date().toISOString()
      })
    }
    
    setPhase(nodes.length > 0 ? "filtering" : "idle")
    toast.info("操作已停止")
  }

  // Handle test completion
  useEffect(() => {
    if (testCompleteData && phase === "testing") {
      setPhase("filtering")
      toast.success(
        `测试完成！成功: ${testCompleteData.successful_tests}, 失败: ${testCompleteData.failed_tests}`
      )
    }
  }, [testCompleteData, phase])

  // Handle test cancellation
  useEffect(() => {
    if (testCancelledData && phase === "testing") {
      setPhase("filtering")
      toast.info(
        `测试已取消！已完成: ${testCancelledData.completed_tests}/${testCancelledData.total_tests}`
      )
    }
  }, [testCancelledData, phase])

  // Toggle node selection
  const toggleNodeSelection = (nodeName: string) => {
    setSelectedNodes(prev => {
      const newSet = new Set(prev)
      if (newSet.has(nodeName)) {
        newSet.delete(nodeName)
      } else {
        newSet.add(nodeName)
      }
      return newSet
    })
  }

  // Select/deselect all filtered nodes
  const toggleAllNodes = () => {
    if (selectedNodes.size === filteredNodes.length) {
      setSelectedNodes(new Set())
    } else {
      setSelectedNodes(new Set(filteredNodes.map(n => n.name)))
    }
  }

  return (
    <div className="min-h-screen p-8">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-5xl font-bold mb-4">
            <span className="text-gradient">Clash SpeedTest V2</span>
          </h1>
          <p className="text-gray-400 text-lg">两阶段测试：先加载节点，再选择测试</p>
        </div>

        {/* Main Control Card */}
        <Card className="glass-morphism border-gray-800 mb-8">
          <div className="p-8">
            {/* Config Input */}
            <div className="mb-8">
              <Label className="text-gray-300 mb-2 block">配置文件路径</Label>
              <div className="flex gap-4">
                <Input
                  placeholder="输入配置文件路径或订阅链接..."
                  value={config.configPaths}
                  onChange={(e) => setConfig(prev => ({ ...prev, configPaths: e.target.value }))}
                  className="flex-1 input-dark text-white placeholder:text-gray-500"
                  disabled={phase !== "idle"}
                />
                <Button
                  onClick={phase === "idle" ? loadNodes : phase === "filtering" ? testSelectedNodes : stopOperation}
                  disabled={(!isConnected && phase === "filtering") || (phase === "loading" && !abortController)}
                  size="lg"
                  className={`min-w-[140px] ${
                    phase === "loading" || phase === "testing"
                      ? "bg-orange-600 hover:bg-orange-700"
                      : phase === "filtering"
                      ? "bg-green-600 hover:bg-green-700"
                      : "button-gradient"
                  }`}
                >
                  {phase === "idle" ? (
                    <>
                      <ClientIcon icon={Globe} className="mr-2 h-4 w-4" />
                      加载节点
                    </>
                  ) : phase === "loading" ? (
                    <>
                      <ClientIcon icon={Pause} className="mr-2 h-4 w-4" />
                      停止加载
                    </>
                  ) : phase === "filtering" ? (
                    <>
                      <ClientIcon icon={TestTubes} className="mr-2 h-4 w-4" />
                      测试选中节点 ({selectedNodes.size})
                    </>
                  ) : (
                    <>
                      <ClientIcon icon={Pause} className="mr-2 h-4 w-4" />
                      停止测试
                    </>
                  )}
                </Button>
              </div>
            </div>

            {/* Status Bar */}
            <div className="mb-6 flex items-center justify-between">
              <div className="flex items-center gap-4">
                <Badge variant="outline" className={`border-gray-700 ${
                  phase === "idle" ? "text-gray-400" :
                  phase === "loading" ? "text-blue-400 animate-pulse" :
                  phase === "filtering" ? "text-green-400" :
                  "text-orange-400 animate-pulse"
                }`}>
                  {phase === "idle" ? "待机" :
                   phase === "loading" ? "加载中..." :
                   phase === "filtering" ? `已加载 ${nodes.length} 个节点` :
                   "测试中..."}
                </Badge>
                <div className="flex items-center gap-2">
                  <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-400 animate-pulse' : 'bg-red-400'}`} />
                  <span className="text-sm text-gray-400">
                    WebSocket: {isConnected ? '已连接' : '未连接'}
                  </span>
                </div>
              </div>
            </div>

            {/* Advanced Settings */}
            <div className="border-t border-gray-800 pt-6">
              <button
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="flex items-center gap-2 text-gray-400 hover:text-white transition-colors"
              >
                <ClientIcon icon={Settings} className="h-4 w-4" />
                高级设置
                {showAdvanced ? (
                  <ClientIcon icon={ChevronUp} className="h-4 w-4" />
                ) : (
                  <ClientIcon icon={ChevronDown} className="h-4 w-4" />
                )}
              </button>

              {showAdvanced && (
                <div className="mt-6 space-y-6">
                  {/* Filtering Options */}
                  <div className="space-y-4">
                    <h3 className="text-gray-300 font-medium">节点过滤选项</h3>
                    
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div>
                        <Label className="text-gray-300 mb-2 block">
                          包含节点 (用逗号分隔)
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
                          排除节点 (用逗号分隔)
                        </Label>
                        <Textarea
                          placeholder="例如: 过期, 到期, 测试..."
                          value={excludeNodesInput}
                          onChange={(e) => handleExcludeNodesChange(e.target.value)}
                          className="input-dark text-white placeholder:text-gray-500 resize-none"
                          rows={2}
                        />
                      </div>
                    </div>

                    {/* Protocol Filter */}
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
                                checked={config.protocolFilter.length === 0 || config.protocolFilter.includes(protocol)}
                                onCheckedChange={(checked: boolean) => handleProtocolFilterChange(protocol, checked)}
                                className="checkbox-dark"
                              />
                              <Label htmlFor={`protocol-${protocol}`} className="text-sm text-gray-300 cursor-pointer">
                                {protocol}
                              </Label>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>

                  <div className="flex items-center gap-6">
                    <div className="flex items-center space-x-2">
                      <Switch
                        id="stashCompatible"
                        checked={config.stashCompatible}
                        onCheckedChange={(checked) => setConfig(prev => ({ ...prev, stashCompatible: checked }))}
                        className="switch-dark"
                      />
                      <Label htmlFor="stashCompatible" className="text-gray-300">
                        Stash 兼容模式
                      </Label>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </Card>

        {/* Node Selection Table (Phase 1) */}
        {phase === "filtering" && filteredNodes.length > 0 && (
          <Card className="glass-morphism border-gray-800 mb-8">
            <div className="p-6">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-xl font-bold text-white flex items-center gap-2">
                  <ClientIcon icon={Filter} className="h-5 w-5 text-blue-400" />
                  选择要测试的节点
                </h2>
                <div className="flex items-center gap-4">
                  <Badge variant="outline" className="border-gray-700 text-gray-300">
                    已选择 {selectedNodes.size} / {filteredNodes.length}
                  </Badge>
                  <Button
                    onClick={toggleAllNodes}
                    variant="outline"
                    size="sm"
                    className="border-gray-700 text-gray-300 hover:text-white"
                  >
                    {selectedNodes.size === filteredNodes.length ? "取消全选" : "全选"}
                  </Button>
                </div>
              </div>

              <div className="overflow-x-auto max-h-96">
                <Table className="table-dark">
                  <TableHeader>
                    <TableRow className="border-gray-800">
                      <TableHead className="text-gray-400 w-12">选择</TableHead>
                      <TableHead className="text-gray-400">节点名称</TableHead>
                      <TableHead className="text-gray-400">类型</TableHead>
                      <TableHead className="text-gray-400">服务器</TableHead>
                      <TableHead className="text-gray-400">端口</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredNodes.map((node) => (
                      <TableRow 
                        key={node.name} 
                        className="table-row-dark cursor-pointer"
                        onClick={() => toggleNodeSelection(node.name)}
                      >
                        <TableCell className="text-center">
                          <Checkbox
                            checked={selectedNodes.has(node.name)}
                            onCheckedChange={() => toggleNodeSelection(node.name)}
                            onClick={(e) => e.stopPropagation()}
                            className="checkbox-dark"
                          />
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
                        <TableCell className="text-gray-400">
                          <div className="flex items-center gap-2">
                            <ClientIcon icon={ServerCog} className="h-4 w-4 text-gray-500" />
                            {node.server}
                          </div>
                        </TableCell>
                        <TableCell className="text-gray-400">{node.port}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          </Card>
        )}

        {/* Real-time Progress Table (Phase 2) */}
        {phase === "testing" && (
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