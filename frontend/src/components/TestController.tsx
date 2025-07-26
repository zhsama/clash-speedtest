import { useEffect } from "react"
import { toast } from "sonner"
import { config } from "@/lib/env"
import type { TestProgressData } from "../hooks/useWebSocket"

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

interface TestControllerProps {
  configUrl: string
  filteredNodes: NodeInfo[]
  filterConfig: FilterConfig
  testConfig: TestConfig
  testing: boolean
  setTesting: (testing: boolean) => void
  taskId: string | null
  setTaskId: (taskId: string | null) => void
  tunModeEnabled: boolean
  isConnected: boolean
  connect: () => void
  sendMessage: (message: any) => void
  clearData: () => void
  setTestProgress: (progress: TestProgressData) => void
  testCompleteData: any
  testCancelledData: any
}

export default function TestController({
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
}: TestControllerProps) {
  // 监听测试完成
  useEffect(() => {
    if (testCompleteData && testing) {
      setTesting(false)
      setTaskId(null)
      toast.success(
        `测试完成！成功: ${testCompleteData.successful_tests}, 失败: ${testCompleteData.failed_tests}`
      )
    }
  }, [testCompleteData, testing, setTesting, setTaskId])

  // 监听测试取消
  useEffect(() => {
    if (testCancelledData && testing) {
      setTesting(false)
      setTaskId(null)
      toast.info(
        `测试已取消！已完成: ${testCancelledData.completed_tests}/${testCancelledData.total_tests}`
      )
    }
  }, [testCancelledData, testing, setTesting, setTaskId])

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
      current_stage: testConfig.testMode === "unlock_only" ? "unlock_test" : "speed_test",
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
      toast.error(`请求失败：${(error as Error).message}`)
      setTesting(false)
    }
  }

  const stopTest = () => {
    if (isConnected && taskId) {
      sendMessage({
        type: "stop_test",
        taskId: taskId,
        timestamp: new Date().toISOString(),
      })
    }
    setTesting(false)
    toast.info("正在停止测试...")
  }

  return {
    startTest,
    stopTest,
  }
}