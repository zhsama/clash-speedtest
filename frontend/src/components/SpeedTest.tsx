import React, { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
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
import { Loader2, Play, Download } from "lucide-react"

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

  const handleTest = async () => {
    if (!config.configPaths) {
      toast.error("请输入配置文件路径")
      return
    }

    setTesting(true)
    setResults([])

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
      } else {
        toast.error(data.error || "测试失败")
      }
    } catch (error) {
      toast.error("请求失败：" + (error as Error).message)
    } finally {
      setTesting(false)
    }
  }

  const formatSpeed = (bytesPerSecond: number) => {
    const mbps = bytesPerSecond / (1024 * 1024)
    return `${mbps.toFixed(2)} MB/s`
  }

  const formatLatency = (latency: number) => {
    return `${Math.round(latency / 1000000)} ms`
  }

  const getLatencyColor = (latency: number) => {
    const ms = latency / 1000000
    if (ms < 100) return "text-green-600"
    if (ms < 300) return "text-yellow-600"
    return "text-red-600"
  }

  const getSpeedColor = (speed: number) => {
    const mbps = speed / (1024 * 1024)
    if (mbps >= 10) return "text-green-600"
    if (mbps >= 5) return "text-yellow-600"
    return "text-red-600"
  }

  const exportResults = () => {
    const data = results.map((r) => ({
      name: r.proxy_name,
      type: r.proxy_type,
      latency: formatLatency(r.latency),
      download: formatSpeed(r.download_speed),
      upload: formatSpeed(r.upload_speed),
      packetLoss: `${r.packet_loss.toFixed(1)}%`,
    }))

    const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `speedtest-results-${new Date().toISOString()}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="container mx-auto p-6 max-w-7xl">
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Clash SpeedTest</CardTitle>
          <CardDescription>测试 Clash 配置文件中的代理节点速度</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* 基础配置 */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="configPaths">配置文件路径</Label>
              <Input
                id="configPaths"
                placeholder="config.yaml 或 https://example.com/sub"
                value={config.configPaths}
                onChange={(e) => setConfig({ ...config, configPaths: e.target.value })}
              />
              <p className="text-sm text-muted-foreground">支持本地文件和订阅链接，多个用逗号分隔</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="filterRegex">节点过滤正则</Label>
              <Input
                id="filterRegex"
                placeholder=".+"
                value={config.filterRegex}
                onChange={(e) => setConfig({ ...config, filterRegex: e.target.value })}
              />
            </div>
          </div>

          {/* 测试参数 */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label>下载测试大小: {config.downloadSize} MB</Label>
              <Slider
                value={[config.downloadSize]}
                onValueChange={(v) => setConfig({ ...config, downloadSize: v[0] })}
                max={100}
                min={10}
                step={10}
              />
            </div>

            <div className="space-y-2">
              <Label>上传测试大小: {config.uploadSize} MB</Label>
              <Slider
                value={[config.uploadSize]}
                onValueChange={(v) => setConfig({ ...config, uploadSize: v[0] })}
                max={50}
                min={5}
                step={5}
              />
            </div>

            <div className="space-y-2">
              <Label>并发数: {config.concurrent}</Label>
              <Slider
                value={[config.concurrent]}
                onValueChange={(v) => setConfig({ ...config, concurrent: v[0] })}
                max={16}
                min={1}
                step={1}
              />
            </div>

            <div className="space-y-2">
              <Label>超时时间: {config.timeout} 秒</Label>
              <Slider
                value={[config.timeout]}
                onValueChange={(v) => setConfig({ ...config, timeout: v[0] })}
                max={30}
                min={3}
                step={1}
              />
            </div>

            <div className="space-y-2">
              <Label>最大延迟: {config.maxLatency} ms</Label>
              <Slider
                value={[config.maxLatency]}
                onValueChange={(v) => setConfig({ ...config, maxLatency: v[0] })}
                max={2000}
                min={100}
                step={100}
              />
            </div>

            <div className="space-y-2">
              <Label>最小下载速度: {config.minDownloadSpeed} MB/s</Label>
              <Slider
                value={[config.minDownloadSpeed]}
                onValueChange={(v) => setConfig({ ...config, minDownloadSpeed: v[0] })}
                max={50}
                min={0}
                step={1}
              />
            </div>
          </div>

          {/* 高级选项 */}
          <div className="flex items-center gap-6">
            <div className="flex items-center space-x-2">
              <Switch
                id="stashCompatible"
                checked={config.stashCompatible}
                onCheckedChange={(v) => setConfig({ ...config, stashCompatible: v })}
              />
              <Label htmlFor="stashCompatible">Stash 兼容模式</Label>
            </div>

            <div className="flex items-center space-x-2">
              <Switch
                id="renameNodes"
                checked={config.renameNodes}
                onCheckedChange={(v) => setConfig({ ...config, renameNodes: v })}
              />
              <Label htmlFor="renameNodes">重命名节点</Label>
            </div>
          </div>

          {/* 操作按钮 */}
          <div className="flex gap-4">
            <Button onClick={handleTest} disabled={testing} size="lg">
              {testing ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  测试中...
                </>
              ) : (
                <>
                  <Play className="mr-2 h-4 w-4" />
                  开始测试
                </>
              )}
            </Button>

            {results.length > 0 && (
              <Button onClick={exportResults} variant="outline" size="lg">
                <Download className="mr-2 h-4 w-4" />
                导出结果
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* 结果表格 */}
      {results.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>测试结果</CardTitle>
            <CardDescription>共 {results.length} 个节点通过测试</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>节点名称</TableHead>
                  <TableHead>类型</TableHead>
                  <TableHead>延迟</TableHead>
                  <TableHead>抖动</TableHead>
                  <TableHead>丢包率</TableHead>
                  <TableHead>下载速度</TableHead>
                  <TableHead>上传速度</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {results.map((result, index) => (
                  <TableRow key={index}>
                    <TableCell className="font-medium">{result.proxy_name}</TableCell>
                    <TableCell>
                      <Badge variant="secondary">{result.proxy_type}</Badge>
                    </TableCell>
                    <TableCell className={getLatencyColor(result.latency)}>
                      {formatLatency(result.latency)}
                    </TableCell>
                    <TableCell>{formatLatency(result.jitter)}</TableCell>
                    <TableCell>{result.packet_loss.toFixed(1)}%</TableCell>
                    <TableCell className={getSpeedColor(result.download_speed)}>
                      {formatSpeed(result.download_speed)}
                    </TableCell>
                    <TableCell className={getSpeedColor(result.upload_speed)}>
                      {formatSpeed(result.upload_speed)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  )
}