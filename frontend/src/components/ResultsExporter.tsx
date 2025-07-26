import { toast } from "sonner"

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

interface TestResult {
  proxy_name: string
  proxy_type: string
  proxy_ip?: string
  latency_ms: number
  jitter_ms: number
  packet_loss: number
  download_speed_mbps: number
  upload_speed_mbps: number
  status: string
  unlock_results?: Array<{
    platform: string
    supported: boolean
    region?: string
  }>
}

interface TestCompleteData {
  total_tested: number
  successful_tests: number
  failed_tests: number
  total_duration: string
  average_latency: number
  average_download_mbps: number
  average_upload_mbps: number
}

interface ResultsExporterProps {
  testResults: TestResult[]
  testCompleteData: TestCompleteData | null
  testConfig: TestConfig
  configUrl: string
  taskId: string | null
}

export default function ResultsExporter({
  testResults,
  testCompleteData,
  testConfig,
  configUrl,
  taskId,
}: ResultsExporterProps) {
  // 生成文件名
  const generateFileName = (format: string) => {
    const testType =
      testConfig.testMode === "both"
        ? "速度+解锁测试"
        : testConfig.testMode === "speed_only"
          ? "速度测试"
          : "解锁测试"
    const timestamp = new Date().toISOString().slice(0, 19).replace(/:/g, "-")
    const testId = taskId || "unknown"
    return `${testType}_${timestamp}_${testId}.${format}`
  }

  const exportToMarkdown = () => {
    if (!testResults.length && !testCompleteData) {
      toast.error("没有测试结果可导出")
      return
    }

    let markdown = "# Clash SpeedTest 测试结果\n\n"
    markdown += `**测试时间**: ${new Date().toLocaleString("zh-CN")}\n`
    markdown += `**测试模式**: ${
      testConfig.testMode === "both"
        ? "速度+解锁测试"
        : testConfig.testMode === "speed_only"
          ? "速度测试"
          : "解锁测试"
    }\n`
    markdown += `**配置来源**: ${configUrl || testConfig.configPaths || "N/A"}\n`

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

    markdown += "\n## 详细测试结果\n\n"

    // 根据测试模式生成不同的表格
    if (testConfig.testMode === "speed_only" || testConfig.testMode === "both") {
      markdown +=
        "| 节点名称 | 类型 | IP地址 | 延迟(ms) | 抖动(ms) | 丢包率(%) | 下载(Mbps) | 上传(Mbps) | 状态 |\n"
      markdown +=
        "|----------|------|--------|----------|----------|-----------|------------|------------|------|\n"

      testResults.forEach((result) => {
        const ip = result.proxy_ip || "N/A"
        const latency = result.latency_ms > 0 ? result.latency_ms.toFixed(2) : "N/A"
        const jitter = result.jitter_ms > 0 ? result.jitter_ms.toFixed(2) : "N/A"
        const loss = result.packet_loss > 0 ? result.packet_loss.toFixed(2) : "0"
        const download =
          result.download_speed_mbps > 0 ? result.download_speed_mbps.toFixed(2) : "N/A"
        const upload = result.upload_speed_mbps > 0 ? result.upload_speed_mbps.toFixed(2) : "N/A"

        markdown += `| ${result.proxy_name} | ${result.proxy_type} | ${ip} | ${latency} | ${jitter} | ${loss} | ${download} | ${upload} | ${result.status} |\n`
      })
    }

    if (testConfig.testMode === "unlock_only" || testConfig.testMode === "both") {
      if (testConfig.testMode === "both") {
        markdown += "\n## 解锁测试结果\n\n"
      }

      markdown += "| 节点名称 | 类型 | IP地址 | 解锁平台 | 状态 |\n"
      markdown += "|----------|------|--------|----------|------|\n"

      testResults.forEach((result) => {
        if (result.unlock_results && result.unlock_results.length > 0) {
          const ip = result.proxy_ip || "N/A"
          result.unlock_results.forEach((unlock) => {
            const status = unlock.supported ? `✅ ${unlock.region || "支持"}` : "❌ 不支持"
            markdown += `| ${result.proxy_name} | ${result.proxy_type} | ${ip} | ${unlock.platform} | ${status} |\n`
          })
        } else {
          const ip = result.proxy_ip || "N/A"
          markdown += `| ${result.proxy_name} | ${result.proxy_type} | ${ip} | 无数据 | ❌ 测试失败 |\n`
        }
      })
    }

    const blob = new Blob([markdown], { type: "text/markdown;charset=utf-8" })
    const url = URL.createObjectURL(blob)
    const link = document.createElement("a")
    link.href = url
    link.download = generateFileName("md")
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

    let csvContent = ""

    // 添加配置信息头部
    csvContent += "# Clash SpeedTest 测试结果\n"
    csvContent += `# 测试时间: ${new Date().toLocaleString("zh-CN")}\n`
    csvContent += `# 测试模式: ${
      testConfig.testMode === "both"
        ? "速度+解锁测试"
        : testConfig.testMode === "speed_only"
          ? "速度测试"
          : "解锁测试"
    }\n`
    csvContent += `# 配置来源: ${configUrl || testConfig.configPaths || "N/A"}\n`

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
    csvContent += "\n"

    // 根据测试模式生成不同的CSV
    if (testConfig.testMode === "speed_only" || testConfig.testMode === "both") {
      csvContent +=
        "节点名称,节点类型,IP地址,延迟(ms),抖动(ms),丢包率(%),下载速度(Mbps),上传速度(Mbps),状态\n"

      testResults.forEach((result) => {
        const ip = result.proxy_ip || "N/A"
        const latency = result.latency_ms > 0 ? result.latency_ms.toFixed(2) : "N/A"
        const jitter = result.jitter_ms > 0 ? result.jitter_ms.toFixed(2) : "N/A"
        const loss = result.packet_loss > 0 ? result.packet_loss.toFixed(2) : "0"
        const download =
          result.download_speed_mbps > 0 ? result.download_speed_mbps.toFixed(2) : "N/A"
        const upload = result.upload_speed_mbps > 0 ? result.upload_speed_mbps.toFixed(2) : "N/A"

        csvContent += `"${result.proxy_name}","${result.proxy_type}","${ip}",${latency},${jitter},${loss},${download},${upload},"${result.status}"\n`
      })
    }

    if (testConfig.testMode === "unlock_only" || testConfig.testMode === "both") {
      if (testConfig.testMode === "both") {
        csvContent += "\n解锁测试结果\n"
      }

      csvContent += "节点名称,节点类型,IP地址,解锁平台,支持状态,区域\n"

      testResults.forEach((result) => {
        if (result.unlock_results && result.unlock_results.length > 0) {
          const ip = result.proxy_ip || "N/A"
          result.unlock_results.forEach((unlock) => {
            const status = unlock.supported ? "支持" : "不支持"
            const region = unlock.region || ""
            csvContent += `"${result.proxy_name}","${result.proxy_type}","${ip}","${unlock.platform}","${status}","${region}"\n`
          })
        } else {
          const ip = result.proxy_ip || "N/A"
          csvContent += `"${result.proxy_name}","${result.proxy_type}","${ip}","无数据","测试失败",""\n`
        }
      })
    }

    const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8" })
    const url = URL.createObjectURL(blob)
    const link = document.createElement("a")
    link.href = url
    link.download = generateFileName("csv")
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)

    toast.success("CSV 文件已导出")
  }

  return {
    exportToMarkdown,
    exportToCSV,
  }
}