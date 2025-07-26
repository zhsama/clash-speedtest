import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Card } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import {
  FaChartLine as Activity,
  FaDownload as Download,
  FaUpload as Upload,
  FaCheckCircle as CheckCircle,
  FaTimesCircle as XCircle,
  FaSpinner as Loader2,
  FaGlobe as Globe,
  FaFile as FileText,
  FaTable as TableIcon,
} from "react-icons/fa"
import ClientIcon from "./ClientIcon"
import type { TestResultData } from "../hooks/useWebSocket"

interface SpeedTestTableProps {
  results: TestResultData[]
  title?: string
  // 导出功能相关
  onExportMarkdown?: () => void
  onExportCSV?: () => void
  showExportButtons?: boolean
}

export default function SpeedTestTable({
  results,
  title = "速度测试结果",
  onExportMarkdown,
  onExportCSV,
  showExportButtons = false,
}: SpeedTestTableProps) {
  // 工具函数
  const formatLatency = (latencyMs: number | null | undefined) => {
    return latencyMs && latencyMs > 0 ? `${latencyMs}ms` : "N/A"
  }

  const formatSpeed = (speedMbps: number | null | undefined) => {
    return speedMbps && speedMbps > 0 ? `${speedMbps.toFixed(2)} MB/s` : "N/A"
  }

  const getLatencyColor = (latencyMs: number | null | undefined) => {
    if (!latencyMs || latencyMs <= 0) return "text-red-400"
    if (latencyMs < 100) return "text-green-400"
    if (latencyMs < 300) return "text-yellow-400"
    return "text-red-400"
  }

  const getSpeedColor = (speedMbps: number | null | undefined) => {
    if (!speedMbps || speedMbps <= 0) return "text-red-400"
    if (speedMbps >= 50) return "text-green-400"
    if (speedMbps >= 10) return "text-yellow-400"
    return "text-red-400"
  }

  const getSpeedIndicator = (speedMbps: number | null | undefined, maxSpeed: number = 100) => {
    if (!speedMbps || speedMbps <= 0) return null

    const percentage = Math.min((speedMbps / maxSpeed) * 100, 100)
    let colorClass = "bg-red-500"

    if (speedMbps >= 50) colorClass = "bg-green-500"
    else if (speedMbps >= 20) colorClass = "bg-yellow-500"
    else if (speedMbps >= 5) colorClass = "bg-orange-500"

    return (
      <div className="w-full mt-1 bg-lavender-800 rounded-full h-2">
        <div
          className={`h-full rounded-full transition-all duration-300 ${colorClass}`}
          style={{ width: `${percentage}%` }}
        />
      </div>
    )
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "success":
        return <CheckCircle className="h-4 w-4 text-green-400" />
      case "failed":
        return <XCircle className="h-4 w-4 text-red-400" />
      default:
        return <Loader2 className="h-4 w-4 text-lavender-400 animate-spin" />
    }
  }

  const getStatusBadge = (status: string) => {
    const baseClasses = "text-xs"
    switch (status) {
      case "success":
        return (
          <Badge variant="default" className={`${baseClasses} bg-green-600 hover:bg-green-700`}>
            成功
          </Badge>
        )
      case "failed":
        return (
          <Badge variant="destructive" className={baseClasses}>
            失败
          </Badge>
        )
      default:
        return (
          <Badge variant="secondary" className={baseClasses}>
            测试中
          </Badge>
        )
    }
  }

  // 显示所有结果，不进行过滤
  const speedResults = results

  if (speedResults.length === 0) {
    return (
      <Card className="card-standard">
        <div className="text-center py-8">
          <ClientIcon icon={Download} className="h-12 w-12 text-lavender-600 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-lavender-400 mb-2">暂无速度测试数据</h3>
          <p className="text-sm text-lavender-500">速度测试结果将在此处显示</p>
        </div>
      </Card>
    )
  }

  return (
    <Card className="card-standard">
      <div className="form-element">
        <div className="flex justify-between items-center form-element">
          <h2 className="text-xl font-bold text-lavender-50 flex items-center gap-2">
            <ClientIcon icon={Download} className="h-5 w-5 text-blue-400" />
            {title}
          </h2>
          <div className="flex items-center gap-3">
            {showExportButtons && (onExportMarkdown || onExportCSV) && (
              <div className="flex gap-2">
                {onExportMarkdown && (
                  <Button
                    onClick={onExportMarkdown}
                    variant="outline"
                    size="sm"
                    className="button-dark border-lavender-600 hover:border-lavender-500 text-lavender-200 hover:text-white"
                  >
                    <ClientIcon icon={FileText} className="h-4 w-4 mr-2" />
                    导出 Markdown
                  </Button>
                )}
                {onExportCSV && (
                  <Button
                    onClick={onExportCSV}
                    variant="outline"
                    size="sm"
                    className="button-dark border-lavender-600 hover:border-lavender-500 text-lavender-200 hover:text-white"
                  >
                    <ClientIcon icon={TableIcon} className="h-4 w-4 mr-2" />
                    导出 CSV
                  </Button>
                )}
              </div>
            )}
            <Badge variant="outline" className="badge-standard">
              {speedResults.length} 个结果
            </Badge>
          </div>
        </div>

        <div className="table-wrapper">
          <Table className="table-standard table-content">
            <TableHeader>
              <TableRow>
                <TableHead className="text-lavender-400">
                  <div className="flex items-center gap-1">状态</div>
                </TableHead>
                <TableHead className="text-lavender-400">
                  <div className="flex items-center gap-1">节点名称</div>
                </TableHead>
                <TableHead className="text-lavender-400">
                  <div className="flex items-center gap-1">类型</div>
                </TableHead>
                <TableHead className="text-lavender-400">
                  <div className="flex items-center gap-1">
                    <ClientIcon icon={Globe} className="h-4 w-4" />
                    IP地址
                  </div>
                </TableHead>
                <TableHead className="text-lavender-400">
                  <div className="flex items-center gap-1">
                    <ClientIcon icon={Activity} className="h-4 w-4" />
                    延迟
                  </div>
                </TableHead>
                <TableHead className="text-lavender-400">
                  <div className="flex items-center gap-1">
                    <ClientIcon icon={Download} className="h-4 w-4" />
                    下载
                  </div>
                </TableHead>
                <TableHead className="text-lavender-400">
                  <div className="flex items-center gap-1">
                    <ClientIcon icon={Upload} className="h-4 w-4" />
                    上传
                  </div>
                </TableHead>
                <TableHead className="text-lavender-400">
                  <div className="flex items-center gap-1">丢包率</div>
                </TableHead>
                <TableHead className="text-lavender-400">
                  <div className="flex items-center gap-1">错误详情</div>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {speedResults.map((result, index) => (
                <TableRow key={`${result.proxy_name}-${index}`} className="table-row-dark">
                  <TableCell className="w-24">
                    <div className="flex items-center gap-2">
                      {getStatusIcon(result.status)}
                      {getStatusBadge(result.status)}
                    </div>
                  </TableCell>
                  <TableCell className="min-w-48">
                    <div
                      className="truncate max-w-xs font-medium text-lavender-50"
                      title={result.proxy_name}
                    >
                      {result.proxy_name}
                    </div>
                  </TableCell>
                  <TableCell className="w-20">
                    <Badge variant="secondary" className="badge-standard text-xs">
                      {result.proxy_type}
                    </Badge>
                  </TableCell>
                  <TableCell className="w-32">
                    <span className="text-lavender-400 font-mono text-xs">
                      {result.proxy_ip || "-"}
                    </span>
                  </TableCell>
                  <TableCell className="w-20">
                    <span className={getLatencyColor(result.latency_ms)}>
                      {formatLatency(result.latency_ms)}
                    </span>
                  </TableCell>
                  <TableCell className="w-32">
                    <div>
                      <span className={getSpeedColor(result.download_speed_mbps)}>
                        {formatSpeed(result.download_speed_mbps)}
                      </span>
                      {getSpeedIndicator(result.download_speed_mbps)}
                    </div>
                  </TableCell>
                  <TableCell className="w-32">
                    <div>
                      <span className={getSpeedColor(result.upload_speed_mbps)}>
                        {formatSpeed(result.upload_speed_mbps)}
                      </span>
                      {getSpeedIndicator(result.upload_speed_mbps, 50)}
                    </div>
                  </TableCell>
                  <TableCell className="w-20">
                    <span className="text-lavender-400">
                      {result.packet_loss != null ? `${result.packet_loss.toFixed(1)}%` : "N/A"}
                    </span>
                  </TableCell>
                  <TableCell className="min-w-40">
                    {result.error_message ? (
                      <div className="space-y-1">
                        <div className="text-xs text-red-400">
                          {result.error_stage && (
                            <Badge variant="outline" className="border-red-500 text-red-400 mr-1">
                              {result.error_stage}
                            </Badge>
                          )}
                          {result.error_code && (
                            <Badge variant="outline" className="border-orange-500 text-orange-400">
                              {result.error_code}
                            </Badge>
                          )}
                        </div>
                        <div
                          className="text-xs text-lavender-500 truncate"
                          title={result.error_message}
                        >
                          {result.error_message}
                        </div>
                      </div>
                    ) : (
                      <span className="text-green-400 text-xs">-</span>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </Card>
  )
}
