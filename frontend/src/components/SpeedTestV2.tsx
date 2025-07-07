import React, { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card } from "@/components/ui/card"
import { Slider } from "@/components/ui/slider"
import { Switch } from "@/components/ui/switch"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { toast } from "sonner"
import {
  Play,
  Pause,
  Download,
  Upload,
  Activity,
  Clock,
  Settings,
  ChevronDown,
  ChevronUp,
  Zap,
  Wifi,
  Globe,
} from "lucide-react"
import ClientIcon from "./ClientIcon"

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

export default function SpeedTest() {
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
  const [progress, setProgress] = useState(0)
  const [currentNode, setCurrentNode] = useState("")

  // LocalStorage key for config persistence
  const CONFIG_STORAGE_KEY = "clash-speedtest-config"

  // Load config from localStorage on component mount
  useEffect(() => {
    const savedConfig = localStorage.getItem(CONFIG_STORAGE_KEY)
    if (savedConfig) {
      try {
        const parsedConfig = JSON.parse(savedConfig)
        setConfig(parsedConfig)
        // toast.success("已加载上次的配置参数")
      } catch (error) {
        console.error("Failed to parse saved config:", error)
        // toast.error("配置参数解析失败，使用默认配置")
      }
    }
  }, [])

  // Save config to localStorage whenever config changes
  useEffect(() => {
    localStorage.setItem(CONFIG_STORAGE_KEY, JSON.stringify(config))
  }, [config])

  // Helper function to update config and save to localStorage
  const updateConfig = (newConfig: Partial<TestRequest>) => {
    setConfig(prev => ({ ...prev, ...newConfig }))
  }

  // 模拟测试进度
  useEffect(() => {
    if (testing && progress < 100) {
      const timer = setTimeout(() => {
        setProgress((prev) => Math.min(prev + 2, 100))
      }, 100)
      return () => clearTimeout(timer)
    }
  }, [testing, progress])

  const handleTest = async () => {
    if (!config.configPaths) {
      toast.error("请输入配置文件路径")
      return
    }

    setTesting(true)
    setResults([])
    setProgress(0)
    setCurrentNode("正在加载配置...")

    try {
      const response = await fetch("http://localhost:8090/api/test", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          ...config,
          downloadSize: config.downloadSize * 1024 * 1024,
          uploadSize: config.uploadSize * 1024 * 1024,
        }),
      })

      const data: TestResponse = await response.json()

      if (data.success && data.results) {
        setResults(data.results)
        toast.success(`测试完成，共 ${data.results.length} 个节点`)
        setProgress(100)
      } else {
        toast.error(data.error || "测试失败")
      }
    } catch (error) {
      toast.error("请求失败：" + (error as Error).message)
    } finally {
      setTesting(false)
      setCurrentNode("")
    }
  }

  const formatSpeed = (bytesPerSecond: number) => {
    const mbps = bytesPerSecond / (1024 * 1024)
    return mbps.toFixed(2)
  }

  const formatLatency = (latency: number) => {
    return Math.round(latency / 1000000)
  }

  const getMetricColor = (value: number, type: "speed" | "latency") => {
    if (type === "speed") {
      if (value >= 100) return "text-green-400"
      if (value >= 50) return "text-yellow-400"
      return "text-red-400"
    } else {
      if (value < 50) return "text-green-400"
      if (value < 150) return "text-yellow-400"
      return "text-red-400"
    }
  }

  const exportResults = () => {
    const data = results.map((r) => ({
      name: r.proxy_name,
      type: r.proxy_type,
      latency: `${formatLatency(r.latency)}ms`,
      download: `${formatSpeed(r.download_speed)} MB/s`,
      upload: `${formatSpeed(r.upload_speed)} MB/s`,
      packetLoss: `${r.packet_loss.toFixed(1)}%`,
    }))

    const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `speedtest-results-${Date.now()}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  // 计算平均值
  const avgDownload = results.length
    ? results.reduce((sum, r) => sum + r.download_speed, 0) / results.length / (1024 * 1024)
    : 0
  const avgUpload = results.length
    ? results.reduce((sum, r) => sum + r.upload_speed, 0) / results.length / (1024 * 1024)
    : 0
  const avgLatency = results.length
    ? results.reduce((sum, r) => sum + r.latency, 0) / results.length / 1000000
    : 0

  return (
    <div className="min-h-screen p-8">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-5xl font-bold mb-4">
            <span className="text-gradient">Clash SpeedTest</span>
          </h1>
          <p className="text-gray-400 text-lg">测试您的代理节点性能</p>
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
                  onClick={handleTest}
                  disabled={testing}
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
                      测试中
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

            {/* Progress Section */}
            {testing && (
              <div className="mb-8">
                <div className="flex justify-between text-sm text-gray-400 mb-2">
                  <span>{currentNode}</span>
                  <span>{progress}%</span>
                </div>
                <div className="w-full bg-gray-800 rounded-full h-2 overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-blue-600 to-purple-600 transition-all duration-300"
                    style={{ width: `${progress}%` }}
                  />
                </div>
              </div>
            )}

            {/* Metrics Cards */}
            {results.length > 0 && (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                <div className="metric-card">
                  <div className="flex items-center justify-between mb-4">
                    <ClientIcon icon={Download} className="h-8 w-8 text-blue-400" />
                    <span className="text-sm text-gray-400">下载</span>
                  </div>
                  <div className="text-3xl font-bold text-white">
                    {avgDownload.toFixed(2)}
                  </div>
                  <div className="text-sm text-gray-400 mt-1">MB/s 平均</div>
                </div>

                <div className="metric-card">
                  <div className="flex items-center justify-between mb-4">
                    <ClientIcon icon={Upload} className="h-8 w-8 text-purple-400" />
                    <span className="text-sm text-gray-400">上传</span>
                  </div>
                  <div className="text-3xl font-bold text-white">
                    {avgUpload.toFixed(2)}
                  </div>
                  <div className="text-sm text-gray-400 mt-1">MB/s 平均</div>
                </div>

                <div className="metric-card">
                  <div className="flex items-center justify-between mb-4">
                    <ClientIcon icon={Activity} className="h-8 w-8 text-green-400" />
                    <span className="text-sm text-gray-400">延迟</span>
                  </div>
                  <div className="text-3xl font-bold text-white">
                    {avgLatency.toFixed(0)}
                  </div>
                  <div className="text-sm text-gray-400 mt-1">ms 平均</div>
                </div>
              </div>
            )}

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
                        onCheckedChange={(v) => setConfig({ ...config, stashCompatible: v })}
                      />
                      <Label htmlFor="stashCompatible" className="text-gray-300">
                        Stash 兼容模式
                      </Label>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Switch
                        id="renameNodes"
                        checked={config.renameNodes}
                        onCheckedChange={(v) => setConfig({ ...config, renameNodes: v })}
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

        {/* Results Table */}
        {results.length > 0 && (
          <Card className="glass-morphism border-gray-800">
            <div className="p-6">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold text-white">测试结果</h2>
                <Button
                  onClick={exportResults}
                  variant="outline"
                  className="border-gray-700 text-gray-300 hover:text-white"
                >
                  <ClientIcon icon={Download} className="mr-2 h-4 w-4" />
                  导出结果
                </Button>
              </div>

              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="border-gray-800">
                      <TableHead className="text-gray-400">节点名称</TableHead>
                      <TableHead className="text-gray-400">类型</TableHead>
                      <TableHead className="text-gray-400">
                        <div className="flex items-center gap-1">
                          <ClientIcon icon={Activity} className="h-4 w-4" />
                          延迟
                        </div>
                      </TableHead>
                      <TableHead className="text-gray-400">
                        <div className="flex items-center gap-1">
                          <ClientIcon icon={Download} className="h-4 w-4" />
                          下载
                        </div>
                      </TableHead>
                      <TableHead className="text-gray-400">
                        <div className="flex items-center gap-1">
                          <ClientIcon icon={Upload} className="h-4 w-4" />
                          上传
                        </div>
                      </TableHead>
                      <TableHead className="text-gray-400">丢包率</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {results.map((result, index) => (
                      <TableRow key={index} className="border-gray-800">
                        <TableCell className="font-medium text-white">
                          {result.proxy_name}
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary" className="bg-gray-800 text-gray-300">
                            {result.proxy_type}
                          </Badge>
                        </TableCell>
                        <TableCell
                          className={getMetricColor(
                            formatLatency(result.latency),
                            "latency"
                          )}
                        >
                          {formatLatency(result.latency)} ms
                        </TableCell>
                        <TableCell
                          className={getMetricColor(
                            parseFloat(formatSpeed(result.download_speed)),
                            "speed"
                          )}
                        >
                          {formatSpeed(result.download_speed)} MB/s
                        </TableCell>
                        <TableCell
                          className={getMetricColor(
                            parseFloat(formatSpeed(result.upload_speed)),
                            "speed"
                          )}
                        >
                          {formatSpeed(result.upload_speed)} MB/s
                        </TableCell>
                        <TableCell className="text-gray-400">
                          {result.packet_loss.toFixed(1)}%
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          </Card>
        )}
      </div>
    </div>
  )
}