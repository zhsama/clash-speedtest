import React, { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card } from "@/components/ui/card"
import { Slider } from "@/components/ui/slider"
import { Switch } from "@/components/ui/switch"
import { Checkbox } from "@/components/ui/checkbox"
import { Textarea } from "@/components/ui/textarea"
import { toast } from "sonner"
import {
  Play,
  Pause,
  Download,
  Settings,
  ChevronDown,
  ChevronUp,
  RotateCcw,
} from "lucide-react"
import ClientIcon from "./ClientIcon"
import RealTimeProgressTable from "./RealTimeProgressTable"
import { useWebSocket } from "../hooks/useWebSocket"

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

interface Result {
  proxy_name: string
  proxy_type: string
  latency: number
  jitter: number
  packet_loss: number
  download_speed: number
  upload_speed: number
}

interface TestResponse {
  success: boolean
  error?: string
  results?: Result[]
}

export default function SpeedTestWithWebSocket() {
  const [config, setConfig] = useState<TestRequest>({
    configPaths: "",
    filterRegex: ".+",
    includeNodes: [],
    excludeNodes: [],
    protocolFilter: [],
    serverUrl: "https://speed.cloudflare.com",
    downloadSize: 50,
    uploadSize: 20,
    timeout: 5,
    concurrent: 4,
    maxLatency: 800,
    minDownloadSpeed: 5,
    minUploadSpeed: 2,
    stashCompatible: false,
    renameNodes: false,
  })

  const [results, setResults] = useState<Result[]>([])
  const [testing, setTesting] = useState(false)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [availableProtocols, setAvailableProtocols] = useState<string[]>([])
  const [includeNodesInput, setIncludeNodesInput] = useState("")
  const [excludeNodesInput, setExcludeNodesInput] = useState("")
  const [abortController, setAbortController] = useState<AbortController | null>(null)

  // LocalStorage key for config persistence
  const CONFIG_STORAGE_KEY = "clash-speedtest-config"

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

  // Helper function to update config and save to localStorage
  const updateConfig = (newConfig: Partial<TestRequest>) => {
    setConfig(prev => ({ ...prev, ...newConfig }))
  }

  // Fetch available protocols from backend
  const fetchAvailableProtocols = async (configPaths: string) => {
    if (!configPaths.trim()) return
    
    try {
      const response = await fetch("http://localhost:8080/api/protocols", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ configPaths }),
      })

      const data = await response.json()
      if (data.success && data.protocols) {
        setAvailableProtocols(data.protocols)
      }
    } catch (error) {
      console.error("Failed to fetch protocols:", error)
    }
  }

  // Handle include/exclude nodes input changes
  const handleIncludeNodesChange = (value: string) => {
    setIncludeNodesInput(value)
    const nodes = value.split(',').map(s => s.trim()).filter(s => s.length > 0)
    updateConfig({ includeNodes: nodes })
  }

  const handleExcludeNodesChange = (value: string) => {
    setExcludeNodesInput(value)
    const nodes = value.split(',').map(s => s.trim()).filter(s => s.length > 0)
    updateConfig({ excludeNodes: nodes })
  }

  // Handle protocol filter changes
  const handleProtocolFilterChange = (protocol: string, checked: boolean) => {
    const currentProtocolFilter = config.protocolFilter || []
    const newProtocolFilter = checked
      ? [...currentProtocolFilter, protocol]
      : currentProtocolFilter.filter(p => p !== protocol)
    updateConfig({ protocolFilter: newProtocolFilter })
  }

  // Load config from localStorage on component mount
  useEffect(() => {
    const savedConfig = localStorage.getItem(CONFIG_STORAGE_KEY)
    if (savedConfig) {
      try {
        const parsedConfig = JSON.parse(savedConfig)
        // Ensure new fields have default values if missing
        const configWithDefaults = {
          ...parsedConfig,
          includeNodes: parsedConfig.includeNodes || [],
          excludeNodes: parsedConfig.excludeNodes || [],
          protocolFilter: parsedConfig.protocolFilter || [],
        }
        setConfig(configWithDefaults)
        setIncludeNodesInput(configWithDefaults.includeNodes?.join(', ') || '')
        setExcludeNodesInput(configWithDefaults.excludeNodes?.join(', ') || '')
      } catch (error) {
        console.error("Failed to parse saved config:", error)
      }
    }
  }, [])

  // Save config to localStorage whenever config changes
  useEffect(() => {
    localStorage.setItem(CONFIG_STORAGE_KEY, JSON.stringify(config))
  }, [config])

  // Fetch protocols when config path changes
  useEffect(() => {
    if (config.configPaths) {
      fetchAvailableProtocols(config.configPaths)
    }
  }, [config.configPaths])

  // Connect to WebSocket on component mount
  useEffect(() => {
    connect()
    return () => {
      disconnect()
    }
  }, [connect, disconnect])

  // Handle WebSocket errors
  useEffect(() => {
    if (wsError) {
      toast.error(`WebSocket错误: ${wsError.message}`)
    }
  }, [wsError])

  // Handle test completion
  useEffect(() => {
    if (testCompleteData && testing) {
      setTesting(false)
      toast.success(
        `测试完成！成功: ${testCompleteData.successful_tests}, 失败: ${testCompleteData.failed_tests}`
      )
    }
  }, [testCompleteData, testing])

  // Handle test cancellation
  useEffect(() => {
    if (testCancelledData && testing) {
      setTesting(false)
      toast.info(
        `测试已取消！已完成: ${testCancelledData.completed_tests}/${testCancelledData.total_tests}, 用时: ${testCancelledData.partial_duration}`
      )
    }
  }, [testCancelledData, testing])

  const handleTest = async () => {
    if (!config.configPaths) {
      toast.error("请输入配置文件路径")
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
    
    setTesting(true)
    setResults([])
    clearData()

    try {
      const response = await fetch("http://localhost:8080/api/test", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(config),
        signal: controller.signal, // Add abort signal
      })

      const data: TestResponse = await response.json()

      if (data.success && data.results) {
        setResults(data.results)
      } else {
        toast.error(data.error || "测试失败")
        setTesting(false)
      }
    } catch (error) {
      if (error instanceof Error && error.name === 'AbortError') {
        toast.info("测试已被取消")
      } else {
        toast.error("请求失败：" + (error as Error).message)
      }
      setTesting(false)
    } finally {
      setAbortController(null)
    }
  }

  const stopTest = () => {
    // Cancel the ongoing fetch request
    if (abortController) {
      abortController.abort()
    }
    
    // Send stop signal via WebSocket
    if (isConnected) {
      sendMessage({
        type: 'stop_test',
        timestamp: new Date().toISOString()
      })
    }
    
    setTesting(false)
    toast.info("测试已停止")
  }

  const reconnectWebSocket = () => {
    disconnect()
    setTimeout(connect, 1000)
    toast.info("正在重新连接WebSocket...")
  }

  const exportResults = () => {
    const dataToExport = testResults.length > 0 ? testResults : results
    if (dataToExport.length === 0) {
      toast.error("没有结果可导出")
      return
    }

    // 准备CSV数据
    const csvHeaders = [
      "节点名称",
      "代理类型", 
      "延迟(ms)",
      "下载速度(MB/s)",
      "上传速度(MB/s)",
      "丢包率(%)",
      "状态",
      "错误阶段",
      "错误代码",
      "错误信息"
    ]

    const csvRows = dataToExport.map((r) => {
      const name = 'proxy_name' in r ? r.proxy_name : (r as Result).proxy_name
      const type = 'proxy_type' in r ? r.proxy_type : (r as Result).proxy_type
      const latency = 'latency_ms' in r ? r.latency_ms : Math.round((r as Result).latency / 1000000)
      const download = 'download_speed_mbps' in r ? r.download_speed_mbps.toFixed(2) : ((r as Result).download_speed / (1024 * 1024)).toFixed(2)
      const upload = 'upload_speed_mbps' in r ? r.upload_speed_mbps.toFixed(2) : ((r as Result).upload_speed / (1024 * 1024)).toFixed(2)
      const packetLoss = r.packet_loss.toFixed(1)
      const status = 'status' in r ? r.status : 'unknown'
      const errorStage = 'error_stage' in r ? (r.error_stage || '') : ''
      const errorCode = 'error_code' in r ? (r.error_code || '') : ''
      const errorMessage = 'error_message' in r ? (r.error_message || '') : ''

      return [
        `"${name.replace(/"/g, '""')}"`,
        `"${type}"`,
        latency,
        download,
        upload,
        packetLoss,
        `"${status}"`,
        `"${errorStage.replace(/"/g, '""')}"`,
        `"${errorCode.replace(/"/g, '""')}"`,
        `"${errorMessage.replace(/"/g, '""')}"`
      ].join(',')
    })

    // 生成CSV内容
    const csvContent = [
      csvHeaders.map(h => `"${h}"`).join(','),
      ...csvRows
    ].join('\n')

    // 添加BOM以支持中文显示
    const BOM = '\uFEFF'
    const blob = new Blob([BOM + csvContent], { type: "text/csv;charset=utf-8" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `clash-speedtest-results-${new Date().toISOString().slice(0, 19).replace(/[T:]/g, '-')}.csv`
    a.click()
    URL.revokeObjectURL(url)
    toast.success("CSV结果已导出")
  }

  return (
    <div className="min-h-screen p-8">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-5xl font-bold mb-4">
            <span className="text-gradient">Clash SpeedTest</span>
          </h1>
          <p className="text-gray-400 text-lg">实时测试您的代理节点性能</p>
        </div>

        {/* Main Test Card */}
        <Card className="glass-morphism border-gray-800 mb-8">
          <div className="p-8">
            {/* Config Input */}
            <div className="mb-8">
              <Label className="text-gray-300 mb-2 block">配置文件路径</Label>
              <div className="flex gap-4">
                <Input
                  placeholder="输入配置文件路径或订阅链接..."
                  value={config.configPaths}
                  onChange={(e) => updateConfig({ configPaths: e.target.value })}
                  className="flex-1 input-dark text-white placeholder:text-gray-500"
                />
                <Button
                  onClick={testing ? stopTest : handleTest}
                  disabled={!isConnected && !testing}
                  size="lg"
                  className={`min-w-[140px] ${
                    testing
                      ? "bg-orange-600 hover:bg-orange-700"
                      : "button-gradient"
                  }`}
                >
                  {testing ? (
                    <>
                      <ClientIcon icon={Pause} className="mr-2 h-4 w-4" />
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

            {/* WebSocket Status */}
            <div className="mb-6 flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="flex items-center gap-2">
                  <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-400 animate-pulse' : 'bg-red-400'}`} />
                  <span className="text-sm text-gray-400">
                    WebSocket: {isConnected ? '已连接' : '未连接'}
                  </span>
                  {!isConnected && (
                    <Button
                      onClick={reconnectWebSocket}
                      size="sm"
                      variant="outline"
                      className="ml-2 border-gray-700 text-gray-300 hover:text-white"
                    >
                      <ClientIcon icon={RotateCcw} className="h-3 w-3 mr-1" />
                      重连
                    </Button>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-2">
                {(testResults.length > 0 || results.length > 0) && (
                  <Button
                    onClick={exportResults}
                    variant="outline"
                    size="sm"
                    className="border-gray-700 text-gray-300 hover:text-white"
                  >
                    <ClientIcon icon={Download} className="mr-2 h-4 w-4" />
                    导出结果
                  </Button>
                )}
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
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    <div>
                      <Label className="text-gray-300 mb-2 block">
                        下载测试大小: {config.downloadSize} MB
                      </Label>
                      <Slider
                        value={[config.downloadSize]}
                        onValueChange={(v) => updateConfig({ downloadSize: v[0] })}
                        max={100}
                        min={10}
                        step={10}
                        className="slider-dark"
                      />
                    </div>

                    <div>
                      <Label className="text-gray-300 mb-2 block">
                        并发数: {config.concurrent}
                      </Label>
                      <Slider
                        value={[config.concurrent]}
                        onValueChange={(v) => updateConfig({ concurrent: v[0] })}
                        max={16}
                        min={1}
                        step={1}
                        className="slider-dark"
                      />
                    </div>

                    <div>
                      <Label className="text-gray-300 mb-2 block">
                        最大延迟: {config.maxLatency} ms
                      </Label>
                      <Slider
                        value={[config.maxLatency]}
                        onValueChange={(v) => updateConfig({ maxLatency: v[0] })}
                        max={2000}
                        min={100}
                        step={100}
                        className="slider-dark"
                      />
                    </div>
                  </div>

                  {/* Filtering Options */}
                  <div className="border-t border-gray-800 pt-6">
                    <h3 className="text-gray-300 mb-4 font-medium">节点过滤选项</h3>
                    
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
                      {/* Include Nodes Filter */}
                      <div>
                        <Label className="text-gray-300 mb-2 block">
                          包含节点 (用逗号分隔)
                        </Label>
                        <Textarea
                          placeholder="例如: 香港, HK, 新加坡..."
                          value={includeNodesInput}
                          onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => handleIncludeNodesChange(e.target.value)}
                          className="input-dark text-white placeholder:text-gray-500 resize-none"
                          rows={2}
                        />
                        <p className="text-xs text-gray-500 mt-1">只测试包含这些关键词的节点（模糊匹配）</p>
                      </div>

                      {/* Exclude Nodes Filter */}
                      <div>
                        <Label className="text-gray-300 mb-2 block">
                          排除节点 (用逗号分隔)
                        </Label>
                        <Textarea
                          placeholder="例如: 过期, 到期, 测试..."
                          value={excludeNodesInput}
                          onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => handleExcludeNodesChange(e.target.value)}
                          className="input-dark text-white placeholder:text-gray-500 resize-none"
                          rows={2}
                        />
                        <p className="text-xs text-gray-500 mt-1">排除包含这些关键词的节点（模糊匹配）</p>
                      </div>
                    </div>

                    {/* Protocol Filter */}
                    {availableProtocols.length > 0 && (
                      <div>
                        <Label className="text-gray-300 mb-3 block">
                          协议过滤 ({(config.protocolFilter?.length || 0) > 0 ? `已选择 ${config.protocolFilter.length} 个` : '全选'})
                        </Label>
                        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-3">
                          {availableProtocols.map((protocol) => (
                            <div key={protocol} className="flex items-center space-x-2">
                              <Checkbox
                                id={`protocol-${protocol}`}
                                checked={(config.protocolFilter?.length || 0) === 0 || config.protocolFilter?.includes(protocol) || false}
                                onCheckedChange={(checked: boolean) => handleProtocolFilterChange(protocol, checked)}
                                className="checkbox-dark"
                              />
                              <Label htmlFor={`protocol-${protocol}`} className="text-sm text-gray-300 cursor-pointer">
                                {protocol}
                              </Label>
                            </div>
                          ))}
                        </div>
                        <p className="text-xs text-gray-500 mt-2">留空表示选择所有协议</p>
                      </div>
                    )}
                  </div>

                  <div className="flex items-center gap-6">
                    <div className="flex items-center space-x-2">
                      <Switch
                        id="stashCompatible"
                        checked={config.stashCompatible}
                        onCheckedChange={(checked) => updateConfig({ stashCompatible: checked })}
                        className="switch-dark"
                      />
                      <Label htmlFor="stashCompatible" className="text-gray-300">
                        Stash 兼容模式
                      </Label>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Switch
                        id="renameNodes"
                        checked={config.renameNodes}
                        onCheckedChange={(checked) => updateConfig({ renameNodes: checked })}
                        className="switch-dark"
                      />
                      <Label htmlFor="renameNodes" className="text-gray-300">
                        重命名节点
                      </Label>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </Card>

        {/* Real-time Progress Table */}
        <RealTimeProgressTable
          results={testResults}
          progress={testProgress}
          completeData={testCompleteData}
          cancelledData={testCancelledData}
          isConnected={isConnected}
        />
      </div>
    </div>
  )
}