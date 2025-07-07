import React, { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card } from "@/components/ui/card"
import { Slider } from "@/components/ui/slider"
import { Switch } from "@/components/ui/switch"
import { Badge } from "@/components/ui/badge"
import { toast } from "sonner"
import {
  Play,
  Pause,
  Download,
  Upload,
  Settings,
  ChevronDown,
  ChevronUp,
  Wifi,
  WifiOff,
  RotateCcw,
  Globe,
} from "lucide-react"
import ClientIcon from "./ClientIcon"
import RealTimeProgressTable from "./RealTimeProgressTable"
import { useWebSocket } from "../hooks/useWebSocket"

interface TestRequest {
  configPaths: string
  filterRegex: string
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

  // LocalStorage key for config persistence
  const CONFIG_STORAGE_KEY = "clash-speedtest-config"

  // WebSocket hook
  const wsUrl = `ws://localhost:8080/ws`
  const {
    isConnected,
    connect,
    disconnect,
    testStartData,
    testProgress,
    testResults,
    testCompleteData,
    error: wsError,
    clearData
  } = useWebSocket(wsUrl)

  // Helper function to update config and save to localStorage
  const updateConfig = (newConfig: Partial<TestRequest>) => {
    setConfig(prev => ({ ...prev, ...newConfig }))
  }

  // Load config from localStorage on component mount
  useEffect(() => {
    const savedConfig = localStorage.getItem(CONFIG_STORAGE_KEY)
    if (savedConfig) {
      try {
        const parsedConfig = JSON.parse(savedConfig)
        setConfig(parsedConfig)
      } catch (error) {
        console.error("Failed to parse saved config:", error)
      }
    }
  }, [])

  // Save config to localStorage whenever config changes
  useEffect(() => {
    localStorage.setItem(CONFIG_STORAGE_KEY, JSON.stringify(config))
  }, [config])

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
      })

      const data: TestResponse = await response.json()

      if (data.success && data.results) {
        setResults(data.results)
      } else {
        toast.error(data.error || "测试失败")
        setTesting(false)
      }
    } catch (error) {
      toast.error("请求失败：" + (error as Error).message)
      setTesting(false)
    }
  }

  const stopTest = () => {
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

    const exportData = dataToExport.map((r) => ({
      name: 'proxy_name' in r ? r.proxy_name : (r as Result).proxy_name,
      type: 'proxy_type' in r ? r.proxy_type : (r as Result).proxy_type,
      latency: 'latency_ms' in r ? `${r.latency_ms}ms` : `${Math.round((r as Result).latency / 1000000)}ms`,
      download: 'download_speed_mbps' in r ? `${r.download_speed_mbps.toFixed(2)} MB/s` : `${((r as Result).download_speed / (1024 * 1024)).toFixed(2)} MB/s`,
      upload: 'upload_speed_mbps' in r ? `${r.upload_speed_mbps.toFixed(2)} MB/s` : `${((r as Result).upload_speed / (1024 * 1024)).toFixed(2)} MB/s`,
      packetLoss: `${r.packet_loss.toFixed(1)}%`,
      status: 'status' in r ? r.status : 'unknown'
    }))

    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: "application/json" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `speedtest-results-${Date.now()}.json`
    a.click()
    URL.revokeObjectURL(url)
    toast.success("结果已导出")
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
                  className="flex-1 bg-gray-900/50 border-gray-700 text-white placeholder:text-gray-500"
                />
                <Button
                  onClick={testing ? stopTest : handleTest}
                  disabled={!isConnected && !testing}
                  size="lg"
                  className={`min-w-[140px] ${
                    testing
                      ? "bg-orange-600 hover:bg-orange-700"
                      : "bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
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
                        className="[&_[role=slider]]:bg-blue-600"
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
                        className="[&_[role=slider]]:bg-purple-600"
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
                        className="[&_[role=slider]]:bg-green-600"
                      />
                    </div>
                  </div>

                  <div className="flex items-center gap-6">
                    <div className="flex items-center space-x-2">
                      <Switch
                        id="stashCompatible"
                        checked={config.stashCompatible}
                        onCheckedChange={(v) => updateConfig({ stashCompatible: v })}
                      />
                      <Label htmlFor="stashCompatible" className="text-gray-300">
                        Stash 兼容模式
                      </Label>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Switch
                        id="renameNodes"
                        checked={config.renameNodes}
                        onCheckedChange={(v) => updateConfig({ renameNodes: v })}
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
          isConnected={isConnected}
        />
      </div>
    </div>
  )
}