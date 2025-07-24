import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card } from "@/components/ui/card"
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
  Loader2,
} from "lucide-react"
import ClientIcon from "./ClientIcon"
import RealTimeProgressTable from "./RealTimeProgressTable"
import TUNWarning from "./TUNWarning"
import { useWebSocket } from "../hooks/useWebSocket"
import type { TestProgressData } from "../hooks/useWebSocket"
import { config } from "@/lib/env"
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
  // 解锁检测相关配置
  testMode: string
  unlockPlatforms: string[]
  unlockConcurrent: number
  unlockTimeout: number
  unlockRetry: boolean
}

// 速度测试配置组件
const SpeedTestConfig = ({ testConfig, setTestConfig, filterConfig, setFilterConfig }: { 
  testConfig: TestConfig; 
  setTestConfig: React.Dispatch<React.SetStateAction<TestConfig>>;
  filterConfig: FilterConfig;
  setFilterConfig: React.Dispatch<React.SetStateAction<FilterConfig>>;
}) => (
  <div className="form-element">
    <h4 className="text-lg font-semibold text-lavender-50 flex items-center gap-2 mb-2">
      <ClientIcon icon={Download} className="h-5 w-5 text-lavender-400" />
      服务器测速配置
    </h4>
    <div className="space-y-2">
      <div>
        <label className="form-element-label">
          测试服务器
        </label>
        <Input
          value={testConfig.serverUrl}
          onChange={(e) => setTestConfig(prev => ({ 
            ...prev, 
            serverUrl: e.target.value 
          }))}
          className="input-standard"
        />
      </div>
      
      <div>
        <label className="form-element-label">
          测试包大小: {testConfig.downloadSize} MB
        </label>
        <Slider
          value={[testConfig.downloadSize]}
          onValueChange={(v) => setTestConfig(prev => ({ 
            ...prev, 
            downloadSize: v[0],
            uploadSize: v[0] // 同时更新上传大小
          }))}
          max={100}
          min={10}
          step={10}
          className="slider-dark"
        />
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="form-element-label">
            并发数: {testConfig.concurrent}
          </label>
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
          <label className="form-element-label">
            超时时间: {testConfig.timeout} 秒
          </label>
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

    {/* 速度过滤条件 */}
    <div className="border-t border-lavender-600 pt-4 mt-4">
      <h4 className="text-lg font-semibold text-lavender-50 flex items-center gap-2 mb-2">
        <ClientIcon icon={Filter} className="h-4 w-4 text-lavender-400" />
        速度过滤条件
      </h4>
      <div className="space-y-2">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="form-element-label">
              最低下载速度: {filterConfig.minDownloadSpeed} MB/s
            </label>
            <Slider
              value={[filterConfig.minDownloadSpeed]}
              onValueChange={(v) => setFilterConfig(prev => ({ 
                ...prev, 
                minDownloadSpeed: v[0] 
              }))}
              max={100}
              min={0}
              step={1}
              className="slider-dark"
            />
          </div>
          
          <div>
            <label className="form-element-label">
              最低上传速度: {filterConfig.minUploadSpeed} MB/s
            </label>
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
        </div>
        
        <div>
          <label className="form-element-label">
            最大延迟: {filterConfig.maxLatency} ms
          </label>
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
    </div>
  </div>
)

// 解锁检测配置组件
const UnlockTestConfig = ({ testConfig, setTestConfig, hasSpeedConfig }: { 
  testConfig: TestConfig; 
  setTestConfig: React.Dispatch<React.SetStateAction<TestConfig>>;
  hasSpeedConfig: boolean;
}) => {
  const [availablePlatforms, setAvailablePlatforms] = useState<string[]>([])
  const [platformsLoading, setPlatformsLoading] = useState(false)

  // 获取支持的解锁检测平台
  useEffect(() => {
    const fetchUnlockPlatforms = async () => {
      setPlatformsLoading(true)
      let platforms = ["Netflix", "YouTube", "Disney+", "ChatGPT", "Spotify", "Bilibili"]
      try {
        const response = await fetch(`${config.apiUrl}/api/unlock/platforms`)
        const data = await response.json()
        
        if (data.success && data.data && data.data.platforms) {
          platforms = data.data.platforms
            .map((platform: any) => platform.display_name || platform.name)
            .sort((a: string, b: string) => a.localeCompare(b))
          setAvailablePlatforms(platforms)
        } else {
          setAvailablePlatforms(platforms.sort((a, b) => a.localeCompare(b)))
          console.warn("Failed to fetch unlock platforms, using defaults")
        }
      } catch (error) {
        console.error("Error fetching unlock platforms:", error)
        setAvailablePlatforms(platforms.sort((a, b) => a.localeCompare(b)))
      } finally {
        setPlatformsLoading(false)
      }
    }

    fetchUnlockPlatforms()
  }, [])

  return (
    <div className={`form-element ${hasSpeedConfig ? "border-t border-lavender-700 pt-4" : ""}`}>
      <h4 className="text-lg font-semibold text-lavender-50 flex items-center gap-2 mb-2">
        <ClientIcon icon={Globe} className="h-5 w-5 text-lavender-400" />
        流媒体解锁检测
      </h4>
      
      <div className="space-y-2">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="form-element-label">
              解锁检测并发数: {testConfig.unlockConcurrent}
            </label>
            <Slider
              value={[testConfig.unlockConcurrent]}
              onValueChange={(v) => setTestConfig(prev => ({ 
                ...prev, 
                unlockConcurrent: v[0] 
              }))}
              max={10}
              min={1}
              step={1}
              className="slider-dark"
            />
          </div>
          
          <div>
            <label className="form-element-label">
              解锁检测超时: {testConfig.unlockTimeout} 秒
            </label>
            <Slider
              value={[testConfig.unlockTimeout]}
              onValueChange={(v) => setTestConfig(prev => ({ 
                ...prev, 
                unlockTimeout: v[0] 
              }))}
              max={30}
              min={5}
              step={5}
              className="slider-dark"
            />
          </div>
        </div>
      </div>
      
      <div className="form-element">
        <label className="form-element-label">
          检测平台 {platformsLoading && <span className="text-xs text-lavender-400">(加载中...)</span>}
        </label>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 component-gap">
          {availablePlatforms.map((platform) => (
            <label key={platform} className="flex items-center gap-2 cursor-pointer min-w-0">
              <Checkbox
                checked={testConfig.unlockPlatforms.includes(platform)}
                onCheckedChange={(checked) => {
                  setTestConfig(prev => ({
                    ...prev,
                    unlockPlatforms: checked
                      ? [...prev.unlockPlatforms, platform]
                      : prev.unlockPlatforms.filter(p => p !== platform)
                  }))
                }}
                className="checkbox-dark"
              />
              <span className="text-lavender-100 text-sm truncate">{platform}</span>
            </label>
          ))}
        </div>
        {availablePlatforms.length === 0 && !platformsLoading && (
          <div className="text-center text-lavender-400 py-4">
            无可用的解锁检测平台
          </div>
        )}
      </div>
    </div>
  )
}

export default function SpeedTestPro() {
  const [configUrl, setConfigUrl] = useState("")
  const [nodes, setNodes] = useState<NodeInfo[]>([])
  const [filteredNodes, setFilteredNodes] = useState<NodeInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [testing, setTesting] = useState(false)
  const [taskId, setTaskId] = useState<string | null>(null)
  const [tunModeEnabled, setTunModeEnabled] = useState(false)
  
  const [filterConfig, setFilterConfig] = useState<FilterConfig>({
    includeNodes: [],
    excludeNodes: [],
    protocolFilter: [],
    minDownloadSpeed: 5,
    minUploadSpeed: 2,
    maxLatency: 3000,
    stashCompatible: false,
  })
  
  const [testConfig, setTestConfig] = useState<TestConfig>({
    configPaths: "",
    serverUrl: "https://speed.cloudflare.com",
    downloadSize: 50,
    uploadSize: 20,
    timeout: 10,
    concurrent: 4,
    // 解锁检测配置 - 初始为空，等待API返回后动态填充
    testMode: "both", // both, speed_only, unlock_only
    unlockPlatforms: [], // 将由API动态填充
    unlockConcurrent: 5,
    unlockTimeout: 10,
    unlockRetry: true,
  })
  
  const [includeNodesInput, setIncludeNodesInput] = useState("")
  const [excludeNodesInput, setExcludeNodesInput] = useState("")
  const [availableProtocols, setAvailableProtocols] = useState<string[]>([])
  
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
    clearData,
    setTestProgress
  } = useWebSocket(wsUrl)
  
  useEffect(() => {
    const initializeConfig = async () => {
      // 首先尝试从localStorage加载配置
      const savedConfig = localStorage.getItem("clash-speedtest-config")
      let hasStoredUnlockPlatforms = false
      
      if (savedConfig) {
        try {
          const parsed = JSON.parse(savedConfig)
          if (parsed.configUrl) setConfigUrl(parsed.configUrl)
          if (parsed.filterConfig) {
            setFilterConfig(prev => ({
              ...prev,
              ...parsed.filterConfig,
              protocolFilter: prev.protocolFilter
            }))
            handleIncludeNodesChange(parsed.filterConfig.includeNodes?.join(', ') || '')
            handleExcludeNodesChange(parsed.filterConfig.excludeNodes?.join(', ') || '')
          }
          if (parsed.testConfig) {
            setTestConfig(prev => ({ ...prev, ...parsed.testConfig }))
            // 检查是否有保存的解锁平台配置
            hasStoredUnlockPlatforms = parsed.testConfig.unlockPlatforms && 
                                      parsed.testConfig.unlockPlatforms.length > 0
          }
        } catch (error) {
          console.error("Failed to load saved config:", error)
        }
      }
      
      // 只有在没有保存的解锁平台配置时，才初始化默认的解锁平台
      if (!hasStoredUnlockPlatforms) {
        try {
          const response = await fetch(`${config.apiUrl}/api/unlock/platforms`)
          const data = await response.json()
          
          if (data.success && data.data && data.data.platforms) {
            // 选择前6个平台作为默认选中的平台（按首字母排序后选择）
            const sortedPlatforms = data.data.platforms
              .map((platform: any) => platform.display_name || platform.name)
              .sort((a: string, b: string) => a.localeCompare(b))
            
            const defaultPlatforms = sortedPlatforms.slice(0, 6)
            
            setTestConfig(prev => ({
              ...prev,
              unlockPlatforms: defaultPlatforms
            }))
          }
        } catch (error) {
          console.error("Error initializing unlock platforms:", error)
          // 使用硬编码的默认值作为fallback
          const fallbackPlatforms = ["Netflix", "YouTube", "Disney+", "ChatGPT", "Spotify", "Bilibili"]
          setTestConfig(prev => ({
            ...prev,
            unlockPlatforms: fallbackPlatforms.sort((a, b) => a.localeCompare(b))
          }))
        }
      }
    }

    initializeConfig()
  }, [])
  
  useEffect(() => {
    const { protocolFilter, ...filterConfigToSave } = filterConfig
    localStorage.setItem("clash-speedtest-config", JSON.stringify({
      configUrl,
      filterConfig: filterConfigToSave,
      testConfig
    }))
  }, [configUrl, filterConfig, testConfig])
  
  useEffect(() => {
    connect()
    return () => disconnect()
  }, [connect, disconnect])
  
  useEffect(() => {
    if (testCompleteData && testing) {
      setTesting(false)
      setTaskId(null)
      toast.success(
        `测试完成！成功: ${testCompleteData.successful_tests}, 失败: ${testCompleteData.failed_tests}`
      )
    }
  }, [testCompleteData, testing])
  
  // 监听testMode变化，清理旧的测试数据
  useEffect(() => {
    // 只在测试模式改变时清理数据，不要在测试完成时清理
    clearData()
  }, [testConfig.testMode, clearData])
  
  useEffect(() => {
    if (testCancelledData && testing) {
      setTesting(false)
      setTaskId(null)
      toast.info(
        `测试已取消！已完成: ${testCancelledData.completed_tests}/${testCancelledData.total_tests}`
      )
    }
  }, [testCancelledData, testing])
  
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
        
        const protocols = [...new Set(data.nodes.map((n: NodeInfo) => n.type))]
        setAvailableProtocols(protocols as string[])
        
        // 更新过滤配置
        const newFilterConfig = {
          ...filterConfig,
          protocolFilter: protocols as string[]
        }
        setFilterConfig(newFilterConfig)
        
        // 使用新的过滤配置来应用过滤和计算统计
        const filtered = applyFiltersWithConfig(data.nodes, newFilterConfig)
        setFilteredNodes(filtered)
        
        const filteredCount = filtered.length
        const hasFilters = newFilterConfig.includeNodes.length > 0 || 
                          newFilterConfig.excludeNodes.length > 0
        
        if (hasFilters && filteredCount < data.nodes.length) {
          const filteredOutCount = data.nodes.length - filteredCount
          toast.success(`成功加载 ${data.nodes.length} 个节点，已过滤 ${filteredOutCount} 个节点，符合条件 ${filteredCount} 个节点`)
        } else {
          toast.success(`成功加载 ${data.nodes.length} 个节点`)
        }
      } else {
        toast.error(data.error || "加载配置失败")
      }
    } catch (error) {
      toast.error("请求失败：" + (error as Error).message)
    } finally {
      setLoading(false)
    }
  }
  
  const applyFilters = (nodesToFilter: NodeInfo[] = nodes) => {
    const filtered = applyFiltersWithConfig(nodesToFilter, filterConfig)
    setFilteredNodes(filtered)
  }

  const applyFiltersWithConfig = (nodesToFilter: NodeInfo[], config: FilterConfig) => {
    let filtered = [...nodesToFilter]
    
    if (config.includeNodes.length > 0) {
      filtered = filtered.filter(node =>
        config.includeNodes.some(include =>
          node.name.toLowerCase().includes(include.toLowerCase())
        )
      )
    }
    
    if (config.excludeNodes.length > 0) {
      filtered = filtered.filter(node =>
        !config.excludeNodes.some(exclude =>
          node.name.toLowerCase().includes(exclude.toLowerCase())
        )
      )
    }
    
    filtered = filtered.filter(node =>
      config.protocolFilter.includes(node.type)
    )
    
    return filtered
  }
  
  useEffect(() => {
    if (nodes.length > 0) {
      applyFilters()
    }
  }, [filterConfig, nodes])
  
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
    
    // 检查TUN模式状态
    if (tunModeEnabled) {
      toast.warning("检测到 TUN 模式已启用", {
        description: "建议先关闭 TUN 模式以获得更准确的测试结果",
        duration: 5000,
      })
      
      // 可以选择是否继续测试
      const confirmed = window.confirm(
        "检测到系统已启用 TUN 模式，这可能影响测试结果的准确性。\n\n是否仍要继续测试？"
      )
      
      if (!confirmed) {
        return
      }
    }
    
    setTesting(true)
    clearData()
    
    const initialProgress: TestProgressData = {
      current_proxy: "",
      completed_count: 0,
      total_count: filteredNodes.length,
      progress_percent: 0,
      status: "starting",
      current_stage: testConfig.testMode === "unlock_only" ? "unlock_test" : "speed_test"
    }
    
    setTestProgress(initialProgress)
    
    try {
      const getFilteredParams = () => {
        const baseParams = {
          configPaths: configUrl,
          testMode: testConfig.testMode,
          timeout: testConfig.timeout,
          concurrent: testConfig.concurrent,
          ...filterConfig,
          filterRegex: ".+",
        }
        
        switch (testConfig.testMode) {
          case "speed_only":
            return {
              ...baseParams,
              serverUrl: testConfig.serverUrl,
              downloadSize: testConfig.downloadSize,
              uploadSize: testConfig.uploadSize,
            }
          case "unlock_only":
            return {
              ...baseParams,
              unlockPlatforms: testConfig.unlockPlatforms,
              unlockConcurrent: testConfig.unlockConcurrent,
              unlockTimeout: testConfig.unlockTimeout,
              unlockRetry: testConfig.unlockRetry,
            }
          case "both":
          default:
            return {
              ...baseParams,
              // 速度测试参数
              serverUrl: testConfig.serverUrl,
              downloadSize: testConfig.downloadSize,
              uploadSize: testConfig.uploadSize,
              // 解锁检测参数
              unlockPlatforms: testConfig.unlockPlatforms,
              unlockConcurrent: testConfig.unlockConcurrent,
              unlockTimeout: testConfig.unlockTimeout,
              unlockRetry: testConfig.unlockRetry,
            }
        }
      }
      
      const response = await fetch(`${config.apiUrl}/api/test/async`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(getFilteredParams()),
      })
      
      const data = await response.json()
      
      if (data.success && data.data && data.data.taskId) {
        setTaskId(data.data.taskId)
        toast.success(`测试任务已创建，任务ID: ${data.data.taskId}`)
      } else {
        toast.error(data.error || "创建测试任务失败")
        setTesting(false)
      }
    } catch (error) {
      toast.error("请求失败：" + (error as Error).message)
      setTesting(false)
    }
  }
  
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
  
  const handleIncludeNodesChange = (value: string) => {
    setIncludeNodesInput(value)
    const nodes = value.split(/[,，]/).map(s => s.trim()).filter(s => s.length > 0)
    setFilterConfig(prev => ({ ...prev, includeNodes: nodes }))
  }
  
  const handleExcludeNodesChange = (value: string) => {
    setExcludeNodesInput(value)
    const nodes = value.split(/[,，]/).map(s => s.trim()).filter(s => s.length > 0)
    setFilterConfig(prev => ({ ...prev, excludeNodes: nodes }))
  }
  
  const isProtocolSelected = (protocol: string) => {
    return filterConfig.protocolFilter.includes(protocol)
  }

  // 导出功能
  const generateFileName = (format: string) => {
    const testType = testConfig.testMode === 'both' ? '速度+解锁测试' : 
                     testConfig.testMode === 'speed_only' ? '速度测试' : '解锁测试'
    const timestamp = new Date().toISOString().slice(0, 19).replace(/:/g, '-')
    const testId = taskId || 'unknown'
    return `${testType}_${timestamp}_${testId}.${format}`
  }

  const exportToMarkdown = () => {
    if (!testResults.length && !testCompleteData) {
      toast.error("没有测试结果可导出")
      return
    }

    let markdown = `# Clash SpeedTest 测试结果\n\n`
    markdown += `**测试时间**: ${new Date().toLocaleString('zh-CN')}\n`
    markdown += `**测试模式**: ${testConfig.testMode === 'both' ? '速度+解锁测试' : 
                                testConfig.testMode === 'speed_only' ? '速度测试' : '解锁测试'}\n`
    markdown += `**配置来源**: ${configUrl || testConfig.configPaths || 'N/A'}\n`
    
    if (testCompleteData) {
      markdown += `**测试统计**: 总计 ${testCompleteData.total_tested} 个节点，成功 ${testCompleteData.successful_tests} 个，失败 ${testCompleteData.failed_tests} 个\n`
      markdown += `**测试耗时**: ${testCompleteData.total_duration}\n`
      if (testCompleteData.average_latency > 0) {
        markdown += `**平均延迟**: ${testCompleteData.average_latency.toFixed(2)} ms\n`
      }
      if (testCompleteData.average_download_mbps > 0) {
        markdown += `**平均下载速度**: ${testCompleteData.average_download_mbps.toFixed(2)} Mbps\n`
      }
      if (testCompleteData.average_upload_mbps > 0) {
        markdown += `**平均上传速度**: ${testCompleteData.average_upload_mbps.toFixed(2)} Mbps\n`
      }
    }
    
    markdown += `\n## 详细测试结果\n\n`

    // 根据测试模式生成不同的表格
    if (testConfig.testMode === 'speed_only' || testConfig.testMode === 'both') {
      markdown += `| 节点名称 | 类型 | IP地址 | 延迟(ms) | 抖动(ms) | 丢包率(%) | 下载(Mbps) | 上传(Mbps) | 状态 |\n`
      markdown += `|----------|------|--------|----------|----------|-----------|------------|------------|------|\n`
      
      testResults.forEach(result => {
        const ip = result.proxy_ip || 'N/A'
        const latency = result.latency_ms > 0 ? result.latency_ms.toFixed(2) : 'N/A'
        const jitter = result.jitter_ms > 0 ? result.jitter_ms.toFixed(2) : 'N/A'
        const loss = result.packet_loss > 0 ? result.packet_loss.toFixed(2) : '0'
        const download = result.download_speed_mbps > 0 ? result.download_speed_mbps.toFixed(2) : 'N/A'
        const upload = result.upload_speed_mbps > 0 ? result.upload_speed_mbps.toFixed(2) : 'N/A'
        
        markdown += `| ${result.proxy_name} | ${result.proxy_type} | ${ip} | ${latency} | ${jitter} | ${loss} | ${download} | ${upload} | ${result.status} |\n`
      })
    }

    if (testConfig.testMode === 'unlock_only' || testConfig.testMode === 'both') {
      if (testConfig.testMode === 'both') {
        markdown += `\n## 解锁测试结果\n\n`
      }
      
      markdown += `| 节点名称 | 类型 | IP地址 | 解锁平台 | 状态 |\n`
      markdown += `|----------|------|--------|----------|------|\n`
      
      testResults.forEach(result => {
        if (result.unlock_results && result.unlock_results.length > 0) {
          const ip = result.proxy_ip || 'N/A'
          result.unlock_results.forEach(unlock => {
            const status = unlock.supported ? `✅ ${unlock.region || '支持'}` : '❌ 不支持'
            markdown += `| ${result.proxy_name} | ${result.proxy_type} | ${ip} | ${unlock.platform} | ${status} |\n`
          })
        } else {
          const ip = result.proxy_ip || 'N/A'
          markdown += `| ${result.proxy_name} | ${result.proxy_type} | ${ip} | 无数据 | ❌ 测试失败 |\n`
        }
      })
    }

    const blob = new Blob([markdown], { type: 'text/markdown;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = generateFileName('md')
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    toast.success("Markdown 文件已导出")
  }

  const exportToCSV = () => {
    if (!testResults.length && !testCompleteData) {
      toast.error("没有测试结果可导出")
      return
    }

    let csvContent = ''
    
    // 添加配置信息头部
    csvContent += `# Clash SpeedTest 测试结果\n`
    csvContent += `# 测试时间: ${new Date().toLocaleString('zh-CN')}\n`
    csvContent += `# 测试模式: ${testConfig.testMode === 'both' ? '速度+解锁测试' : 
                                testConfig.testMode === 'speed_only' ? '速度测试' : '解锁测试'}\n`
    csvContent += `# 配置来源: ${configUrl || testConfig.configPaths || 'N/A'}\n`
    
    if (testCompleteData) {
      csvContent += `# 测试统计: 总计 ${testCompleteData.total_tested} 个节点，成功 ${testCompleteData.successful_tests} 个，失败 ${testCompleteData.failed_tests} 个\n`
      csvContent += `# 测试耗时: ${testCompleteData.total_duration}\n`
      if (testCompleteData.average_latency > 0) {
        csvContent += `# 平均延迟: ${testCompleteData.average_latency.toFixed(2)} ms\n`
      }
      if (testCompleteData.average_download_mbps > 0) {
        csvContent += `# 平均下载速度: ${testCompleteData.average_download_mbps.toFixed(2)} Mbps\n`
      }
      if (testCompleteData.average_upload_mbps > 0) {
        csvContent += `# 平均上传速度: ${testCompleteData.average_upload_mbps.toFixed(2)} Mbps\n`
      }
    }
    csvContent += `\n`
    
    // 根据测试模式生成不同的CSV
    if (testConfig.testMode === 'speed_only' || testConfig.testMode === 'both') {
      csvContent += '节点名称,节点类型,IP地址,延迟(ms),抖动(ms),丢包率(%),下载速度(Mbps),上传速度(Mbps),状态\n'
      
      testResults.forEach(result => {
        const ip = result.proxy_ip || 'N/A'
        const latency = result.latency_ms > 0 ? result.latency_ms.toFixed(2) : 'N/A'
        const jitter = result.jitter_ms > 0 ? result.jitter_ms.toFixed(2) : 'N/A'
        const loss = result.packet_loss > 0 ? result.packet_loss.toFixed(2) : '0'
        const download = result.download_speed_mbps > 0 ? result.download_speed_mbps.toFixed(2) : 'N/A'
        const upload = result.upload_speed_mbps > 0 ? result.upload_speed_mbps.toFixed(2) : 'N/A'
        
        csvContent += `"${result.proxy_name}","${result.proxy_type}","${ip}",${latency},${jitter},${loss},${download},${upload},"${result.status}"\n`
      })
    }

    if (testConfig.testMode === 'unlock_only' || testConfig.testMode === 'both') {
      if (testConfig.testMode === 'both') {
        csvContent += '\n解锁测试结果\n'
      }
      
      csvContent += '节点名称,节点类型,IP地址,解锁平台,支持状态,区域\n'
      
      testResults.forEach(result => {
        if (result.unlock_results && result.unlock_results.length > 0) {
          const ip = result.proxy_ip || 'N/A'
          result.unlock_results.forEach(unlock => {
            const status = unlock.supported ? '支持' : '不支持'
            const region = unlock.region || ''
            csvContent += `"${result.proxy_name}","${result.proxy_type}","${ip}","${unlock.platform}","${status}","${region}"\n`
          })
        } else {
          const ip = result.proxy_ip || 'N/A'
          csvContent += `"${result.proxy_name}","${result.proxy_type}","${ip}","无数据","测试失败",""\n`
        }
      })
    }

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = generateFileName('csv')
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    toast.success("CSV 文件已导出")
  }

  const handleProtocolFilterChange = (protocol: string, checked: boolean) => {
    console.log(protocol, checked);
    setFilterConfig(prev => {
      let newProtocolFilter: string[]
      
      if (checked) {
        newProtocolFilter = [...prev.protocolFilter, protocol]
      } else {
        newProtocolFilter = prev.protocolFilter.filter(p => p !== protocol)
      }
      
      console.log('protocolFilter changed from', prev.protocolFilter, 'to', newProtocolFilter);
      
      return {
        ...prev,
        protocolFilter: newProtocolFilter
      }
    })
  }
  
  return (
    <div className="min-h-screen bg-gradient-dark">
      <div className="max-w-7xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="text-center">
          <h1 className="text-4xl font-bold mb-3">
            <span className="text-gradient">Clash SpeedTest Pro</span>
          </h1>
          <p className="text-lavender-400">专业的代理节点性能测试工具</p>
        </div>
        
        {/* TUN 模式检测警告 */}
        <TUNWarning 
          onTUNStatusChange={setTunModeEnabled}
          showDetails={false}
        />
        
        {/* 配置获取 */}
        <Card className="card-standard">
          <div className="flex items-center gap-2 form-element">
            <ClientIcon icon={Globe} className="h-5 w-5 text-lavender-400" />
            <h2 className="text-lg font-semibold text-lavender-50">配置获取</h2>
            <div className="ml-auto">
              {isConnected ? (
                <div className="status-indicator">
                  <div className="status-dot success animate-pulse" />
                  <span className="text-lavender-300 text-sm">WebSocket 已连接</span>
                </div>
              ) : (
                <div className="status-indicator">
                  <div className="status-dot error" />
                  <span className="text-lavender-300 text-sm">WebSocket 未连接</span>
                </div>
              )}
            </div>
          </div>
          
          <div className="flex component-gap">
            <Input
              placeholder="输入配置文件路径或订阅链接..."
              value={configUrl}
              onChange={(e) => setConfigUrl(e.target.value)}
              className="flex-1 input-standard"
              disabled={loading || testing}
            />
            <Button
              onClick={fetchConfig}
              disabled={loading || testing}
              className="button-standard min-w-[120px]"
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
            <div className="flex items-center component-gap">
              <span className="badge-standard">
                总节点数: {nodes.length}
              </span>
              <span className="badge-standard">
                符合条件: {filteredNodes.length}
              </span>
              {testing && (
                <span className="badge-standard bg-lavender-600 text-lavender-50">
                  测试中...
                </span>
              )}
            </div>
          )}
        </Card>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* 左侧面板 */}
          <div className="lg:col-span-2 space-y-6">
            {/* 节点过滤条件 */}
            <Card className="card-standard">
              <h4 className="text-lg font-semibold text-lavender-50 flex items-center gap-2 mb-2">
                <ClientIcon icon={Filter} className="h-4 w-4 text-lavender-400" />
                过滤条件
              </h4>
            
              <div className="space-y-2">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="form-element-label">
                      包含节点 (逗号分隔)
                    </label>
                    <Input
                      placeholder="例如: 香港, HK, 新加坡..."
                      value={includeNodesInput}
                      onChange={(e) => handleIncludeNodesChange(e.target.value)}
                      className="input-standard"
                    />
                  </div>
                  
                  <div>
                    <label className="form-element-label">
                      排除节点 (逗号分隔)
                    </label>
                    <Input
                      placeholder="例如: 过期, 测试, 备用..."
                      value={excludeNodesInput}
                      onChange={(e) => handleExcludeNodesChange(e.target.value)}
                      className="input-standard"
                    />
                  </div>
                </div>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {availableProtocols.length > 0 && (
                    <div>
                      <label className="form-element-label">
                        协议过滤
                      </label>
                      <div className="grid grid-cols-1 sm:grid-cols-2 component-gap">
                        {availableProtocols.map((protocol) => (
                          <div key={protocol} className="flex items-center gap-2 min-w-0">
                            <Checkbox
                              id={`protocol-${protocol}`}
                              checked={isProtocolSelected(protocol)}
                              onCheckedChange={(checked: boolean) => 
                                handleProtocolFilterChange(protocol, checked)
                              }
                              className="checkbox-dark"
                            />
                            <label 
                              htmlFor={`protocol-${protocol}`} 
                              className="text-sm text-lavender-100 cursor-pointer truncate"
                            >
                              {protocol}
                            </label>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                  
                  <div>
                    <label className="form-element-label">
                      其他选项
                    </label>
                    <div className="flex items-center gap-2">
                      <Switch
                        id="stashCompatible"
                        checked={filterConfig.stashCompatible}
                        onCheckedChange={(checked) => setFilterConfig(prev => ({ 
                          ...prev, 
                          stashCompatible: checked 
                        }))}
                        className="switch-dark"
                      />
                      <label htmlFor="stashCompatible" className="text-lavender-100">
                        Stash 兼容模式
                      </label>
                    </div>
                  </div>
                </div>
              </div>
            </Card>
            {/* 节点列表预览 */}
            <Card className="card-standard">
              <div className="flex items-center justify-between form-element">
                <h2 className="text-lg font-semibold text-lavender-50 flex items-center gap-2">
                  <ClientIcon icon={ServerCog} className="h-5 w-5 text-lavender-400" />
                  节点列表 {testing ? '(测试中)' : '(预览)'}
                </h2>
                <Button
                  onClick={() => applyFilters()}
                  variant="outline"
                  size="sm"
                  className="button-standard"
                  disabled={testing}
                >
                  <ClientIcon icon={RefreshCw} className="h-4 w-4 mr-1" />
                  刷新过滤
                </Button>
              </div>
              
              <div className="table-scroll-container">
                <div className="table-scroll-content">
                  <Table>
                    <TableHeader className="table-header-fixed">
                      <TableRow>
                        <TableHead>名称</TableHead>
                        <TableHead>协议</TableHead>
                        <TableHead>IP / 域名</TableHead>
                        <TableHead>端口</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredNodes.length > 0 ? (
                        filteredNodes.map((node, index) => (
                          <TableRow key={`${node.name}-${index}`}>
                            <TableCell className="font-medium text-lavender-50">
                              <div className="truncate max-w-xs" title={node.name}>
                                {node.name}
                              </div>
                            </TableCell>
                            <TableCell>
                              <span className="badge-standard">
                                {node.type}
                              </span>
                            </TableCell>
                            <TableCell className="text-lavender-300 font-mono text-sm">
                              {node.server}
                            </TableCell>
                            <TableCell className="text-lavender-300">
                              {node.port}
                            </TableCell>
                          </TableRow>
                        ))
                      ) : (
                        <TableRow>
                          <TableCell colSpan={4} className="text-center text-lavender-400 py-8">
                            {nodes.length === 0 ? "暂无节点信息" : "没有符合条件的节点"}
                          </TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>
                </div>
              </div>
            </Card>

            {/* 实时测试结果 */}
            {(testing || testResults.length > 0 || testCompleteData || testCancelledData) && (
              <RealTimeProgressTable
                results={testResults}
                progress={testProgress}
                completeData={testCompleteData}
                cancelledData={testCancelledData}
                isConnected={isConnected}
                testMode={testConfig.testMode}
                onExportMarkdown={exportToMarkdown}
                onExportCSV={exportToCSV}
                showExportButtons={Boolean((testCompleteData || testCancelledData) && testResults.length > 0)}
              />
            )}
          </div>

          {/* 右侧面板 - 测试配置和控制 */}
          <div className="space-y-4">
            {/* 测试配置 */}
            <Card className="card-standard">
              <div className="flex items-center gap-2 mb-2">
                <ClientIcon icon={Filter} className="h-5 w-5 text-lavender-400" />
                <h2 className="text-lg font-semibold text-lavender-50">测试配置</h2>
              </div>
              
              {/* 测试模式选择器 */}
              <div className="space-y-2">
                <label className="form-element-label">
                  测试模式
                </label>
                <select
                  value={testConfig.testMode}
                  onChange={(e) => setTestConfig(prev => ({ 
                    ...prev, 
                    testMode: e.target.value 
                  }))}
                  className="test-mode-selector w-full"
                >
                  <option value="both">全面测试（测速+解锁）</option>
                  <option value="speed_only">仅测速</option>
                  <option value="unlock_only">仅解锁检测</option>
                </select>
                <p className="text-sm text-lavender-400 mt-2">
                  {testConfig.testMode === "both" && "同时进行速度测试和流媒体解锁检测"}
                  {testConfig.testMode === "speed_only" && "只进行网络速度测试，跳过解锁检测"}
                  {testConfig.testMode === "unlock_only" && "只进行流媒体解锁检测，跳过速度测试"}
                </p>
              </div>

              {/* 启动测试按钮 */}
              <Button
                onClick={testing ? stopTest : startTest}
                disabled={!isConnected || nodes.length === 0 || loading}
                size="lg"
                className={`w-full ${testing ? "bg-red-600 hover:bg-red-700 text-white" : "button-standard"}`}
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
            </Card>

            {/* 高级配置 */}
            <details>
              <summary className="cursor-pointer text-lavender-300 hover:text-lavender-100 transition-colors">
                高级测试配置
              </summary>
              <Card className="card-standard mt-4">
                {/* 速度测试配置 - 条件显示 */}
                {(testConfig.testMode === "both" || testConfig.testMode === "speed_only") && (
                  <SpeedTestConfig 
                    testConfig={testConfig} 
                    setTestConfig={setTestConfig}
                    filterConfig={filterConfig}
                    setFilterConfig={setFilterConfig}
                  />
                )}
                
                {/* 解锁检测配置 - 条件显示 */}
                {(testConfig.testMode === "both" || testConfig.testMode === "unlock_only") && (
                  <UnlockTestConfig 
                    testConfig={testConfig} 
                    setTestConfig={setTestConfig}
                    hasSpeedConfig={testConfig.testMode === "both"}
                  />
                )}
              </Card>
            </details>
          </div>
        </div>
      </div>
    </div>
  )
}
