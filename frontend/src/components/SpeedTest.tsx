import { useEffect, useState } from "react"
import { FaHistory as History } from "react-icons/fa"
import { Button } from "@/components/ui/button"
import { config } from "@/lib/env"
import { useTestResultSaver } from "../hooks/useTestResultSaver"
import type { TestProgressData } from "../hooks/useWebSocket"
import { useWebSocket } from "../hooks/useWebSocket"
import ClientIcon from "./ClientIcon"
import ConfigManager from "./ConfigManager"
import FilterPanel from "./FilterPanel"
import NodeListTable from "./NodeListTable"
import RealTimeProgressTable from "./RealTimeProgressTable"
import ResultsExporter from "./ResultsExporter"
import TestConfigPanel from "./TestConfigPanel"
import TestController from "./TestController"
import TestHistoryModal from "./TestHistoryModal"
import TUNWarning from "./TUNWarning"

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
  testMode: string
  unlockPlatforms: string[]
  unlockConcurrent: number
  unlockTimeout: number
  unlockRetry: boolean
}

export default function SpeedTestPro() {
  // 基础状态
  const [configUrl, setConfigUrl] = useState("")
  const [nodes, setNodes] = useState<NodeInfo[]>([])
  const [filteredNodes, setFilteredNodes] = useState<NodeInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [testing, setTesting] = useState(false)
  const [taskId, setTaskId] = useState<string | null>(null)
  const [tunModeEnabled, setTunModeEnabled] = useState(false)
  const [showHistory, setShowHistory] = useState(false)

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
    testMode: "both",
    unlockPlatforms: [],
    unlockConcurrent: 5,
    unlockTimeout: 10,
    unlockRetry: true,
  })

  // 过滤器输入状态
  const [includeNodesInput, setIncludeNodesInput] = useState("")
  const [excludeNodesInput, setExcludeNodesInput] = useState("")
  const [availableProtocols, setAvailableProtocols] = useState<string[]>([])

  // WebSocket 和测试相关
  const { saveTestSession } = useTestResultSaver()
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
    testStartData,
    clearData,
    setTestProgress,
  } = useWebSocket(wsUrl)

  // 初始化配置
  useEffect(() => {
    const initializeConfig = async () => {
      const savedConfig = localStorage.getItem("clash-speedtest-config")
      let hasStoredUnlockPlatforms = false

      if (savedConfig) {
        try {
          const parsed = JSON.parse(savedConfig)
          if (parsed.configUrl) setConfigUrl(parsed.configUrl)
          if (parsed.filterConfig) {
            setFilterConfig((prev) => ({
              ...prev,
              ...parsed.filterConfig,
              protocolFilter: prev.protocolFilter,
            }))
            handleIncludeNodesChange(parsed.filterConfig.includeNodes?.join(", ") || "")
            handleExcludeNodesChange(parsed.filterConfig.excludeNodes?.join(", ") || "")
          }
          if (parsed.testConfig) {
            setTestConfig((prev) => ({ ...prev, ...parsed.testConfig }))
            hasStoredUnlockPlatforms =
              parsed.testConfig.unlockPlatforms && parsed.testConfig.unlockPlatforms.length > 0
          }
        } catch (error) {
          console.error("Failed to load saved config:", error)
        }
      }

      // 初始化默认解锁平台
      if (!hasStoredUnlockPlatforms) {
        try {
          const response = await fetch(`${config.apiUrl}/api/unlock/platforms`)
          const data = await response.json()

          if (data.success && data.data && data.data.platforms) {
            const sortedPlatforms = data.data.platforms
              .map((platform: any) => platform.display_name || platform.name)
              .sort((a: string, b: string) => a.localeCompare(b))

            const defaultPlatforms = sortedPlatforms.slice(0, 6)
            setTestConfig((prev) => ({
              ...prev,
              unlockPlatforms: defaultPlatforms,
            }))
          }
        } catch (error) {
          console.error("Error initializing unlock platforms:", error)
          const fallbackPlatforms = ["Netflix", "YouTube", "Disney+", "ChatGPT", "Spotify", "Bilibili"]
          setTestConfig((prev) => ({
            ...prev,
            unlockPlatforms: fallbackPlatforms.sort((a, b) => a.localeCompare(b)),
          }))
        }
      }
    }

    initializeConfig()
  }, [])

  // 保存配置到本地存储
  useEffect(() => {
    const { protocolFilter, ...filterConfigToSave } = filterConfig
    localStorage.setItem(
      "clash-speedtest-config",
      JSON.stringify({
        configUrl,
        filterConfig: filterConfigToSave,
        testConfig,
      })
    )
  }, [configUrl, filterConfig, testConfig])

  // WebSocket 连接管理
  useEffect(() => {
    connect()
    return () => disconnect()
  }, [connect, disconnect])

  // 监听测试模式变化，清理数据
  useEffect(() => {
    clearData()
  }, [testConfig.testMode, clearData])

  // 自动保存测试结果
  useEffect(() => {
    if (testCompleteData && testStartData && testResults.length > 0) {
      const testType: "speed" | "unlock" | "both" =
        testConfig.testMode === "speed_only"
          ? "speed"
          : testConfig.testMode === "unlock_only"
            ? "unlock"
            : "both"

      saveTestSession(testStartData, testResults, testCompleteData, testType).catch(console.error)
    }
  }, [testCompleteData, testStartData, testResults, saveTestSession, testConfig.testMode])

  // 过滤逻辑
  const applyFiltersWithConfig = (nodesToFilter: NodeInfo[], config: FilterConfig) => {
    let filtered = [...nodesToFilter]

    if (config.includeNodes.length > 0) {
      filtered = filtered.filter((node) =>
        config.includeNodes.some((include) =>
          node.name.toLowerCase().includes(include.toLowerCase())
        )
      )
    }

    if (config.excludeNodes.length > 0) {
      filtered = filtered.filter(
        (node) =>
          !config.excludeNodes.some((exclude) =>
            node.name.toLowerCase().includes(exclude.toLowerCase())
          )
      )
    }

    filtered = filtered.filter((node) => config.protocolFilter.includes(node.type))
    return filtered
  }

  const applyFilters = (nodesToFilter: NodeInfo[] = nodes) => {
    const filtered = applyFiltersWithConfig(nodesToFilter, filterConfig)
    setFilteredNodes(filtered)
  }

  useEffect(() => {
    if (nodes.length > 0) {
      applyFilters()
    }
  }, [filterConfig, nodes])

  // 处理过滤器输入变化
  const handleIncludeNodesChange = (value: string) => {
    setIncludeNodesInput(value)
    const nodes = value
      .split(/[,，]/)
      .map((s) => s.trim())
      .filter((s) => s.length > 0)
    setFilterConfig((prev) => ({ ...prev, includeNodes: nodes }))
  }

  const handleExcludeNodesChange = (value: string) => {
    setExcludeNodesInput(value)
    const nodes = value
      .split(/[,，]/)
      .map((s) => s.trim())
      .filter((s) => s.length > 0)
    setFilterConfig((prev) => ({ ...prev, excludeNodes: nodes }))
  }

  const isProtocolSelected = (protocol: string) => {
    return filterConfig.protocolFilter.includes(protocol)
  }

  const handleProtocolFilterChange = (protocol: string, checked: boolean) => {
    setFilterConfig((prev) => {
      let newProtocolFilter: string[]

      if (checked) {
        newProtocolFilter = [...prev.protocolFilter, protocol]
      } else {
        newProtocolFilter = prev.protocolFilter.filter((p) => p !== protocol)
      }

      return {
        ...prev,
        protocolFilter: newProtocolFilter,
      }
    })
  }

  // 测试控制器
  const testController = TestController({
    configUrl,
    filteredNodes,
    filterConfig,
    testConfig,
    testing,
    setTesting,
    taskId,
    setTaskId,
    tunModeEnabled,
    isConnected,
    connect,
    sendMessage,
    clearData,
    setTestProgress,
    testCompleteData,
    testCancelledData,
  })

  // 结果导出器
  const resultsExporter = ResultsExporter({
    testResults,
    testCompleteData,
    testConfig,
    configUrl,
    taskId,
  })

  return (
    <div className="min-h-screen bg-gradient-dark">
      <div className="max-w-7xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div className="text-center flex-1">
            <h1 className="text-4xl font-bold mb-3">
              <span className="text-gradient">Clash SpeedTest Pro</span>
            </h1>
            <p className="text-lavender-400">专业的代理节点性能测试工具</p>
          </div>

          <Button onClick={() => setShowHistory(true)} className="btn-outlined">
            <ClientIcon icon={History} className="h-4 w-4 mr-2" />
            历史记录
          </Button>
        </div>

        {/* TUN 模式检测警告 */}
        <TUNWarning onTUNStatusChange={setTunModeEnabled} showDetails={false} />

        {/* 配置管理组件 */}
        <ConfigManager
          configUrl={configUrl}
          setConfigUrl={setConfigUrl}
          nodes={nodes}
          setNodes={setNodes}
          filteredNodes={filteredNodes}
          setFilteredNodes={setFilteredNodes}
          filterConfig={filterConfig}
          setAvailableProtocols={setAvailableProtocols}
          setFilterConfig={setFilterConfig}
          applyFiltersWithConfig={applyFiltersWithConfig}
          loading={loading}
          setLoading={setLoading}
          testing={testing}
          isConnected={isConnected}
        />

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* 左侧面板 */}
          <div className="lg:col-span-2 space-y-6">
            {/* 过滤面板 */}
            <FilterPanel
              filterConfig={filterConfig}
              setFilterConfig={setFilterConfig}
              availableProtocols={availableProtocols}
              includeNodesInput={includeNodesInput}
              setIncludeNodesInput={setIncludeNodesInput}
              excludeNodesInput={excludeNodesInput}
              setExcludeNodesInput={setExcludeNodesInput}
              handleIncludeNodesChange={handleIncludeNodesChange}
              handleExcludeNodesChange={handleExcludeNodesChange}
              handleProtocolFilterChange={handleProtocolFilterChange}
              isProtocolSelected={isProtocolSelected}
              applyFilters={applyFilters}
              testing={testing}
            />

            {/* 节点列表表格 */}
            <NodeListTable nodes={nodes} filteredNodes={filteredNodes} testing={testing} />

            {/* 实时测试结果 */}
            {(testing || testResults.length > 0 || testCompleteData || testCancelledData) && (
              <RealTimeProgressTable
                results={testResults}
                progress={testProgress}
                completeData={testCompleteData}
                cancelledData={testCancelledData}
                isConnected={isConnected}
                testMode={testConfig.testMode}
                onExportMarkdown={resultsExporter.exportToMarkdown}
                onExportCSV={resultsExporter.exportToCSV}
                showExportButtons={Boolean(
                  (testCompleteData || testCancelledData) && testResults.length > 0
                )}
              />
            )}
          </div>

          {/* 右侧面板 - 测试配置和控制 */}
          <div className="space-y-4">
            <TestConfigPanel
              testConfig={testConfig}
              setTestConfig={setTestConfig}
              filterConfig={filterConfig}
              setFilterConfig={setFilterConfig}
              testing={testing}
              startTest={testController.startTest}
              stopTest={testController.stopTest}
              isConnected={isConnected}
              nodes={nodes}
              loading={loading}
            />
          </div>
        </div>
      </div>

      {/* 历史记录模态框 */}
      {showHistory && <TestHistoryModal onClose={() => setShowHistory(false)} />}
    </div>
  )
}