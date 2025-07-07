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
import { 
  Activity, 
  Download, 
  Upload, 
  Zap,
  CheckCircle,
  XCircle,
  Loader2,
  TrendingUp
} from "lucide-react"
import ClientIcon from "./ClientIcon"
import type { TestResultData, TestProgressData, TestCompleteData, TestCancelledData } from "../hooks/useWebSocket"

interface RealTimeProgressTableProps {
  results: TestResultData[]
  progress: TestProgressData | null
  completeData: TestCompleteData | null
  cancelledData: TestCancelledData | null
  isConnected: boolean
}

export default function RealTimeProgressTable({ 
  results, 
  progress, 
  completeData,
  cancelledData,
  isConnected 
}: RealTimeProgressTableProps) {
  const formatLatency = (latencyMs: number) => {
    return latencyMs > 0 ? `${latencyMs}ms` : "N/A"
  }

  const formatSpeed = (speedMbps: number) => {
    return speedMbps > 0 ? `${speedMbps.toFixed(2)} MB/s` : "N/A"
  }

  const getLatencyColor = (latencyMs: number) => {
    if (latencyMs <= 0) return "text-red-400"
    if (latencyMs < 100) return "text-green-400"
    if (latencyMs < 300) return "text-yellow-400"
    return "text-red-400"
  }

  const getSpeedColor = (speedMbps: number) => {
    if (speedMbps <= 0) return "text-red-400"
    if (speedMbps >= 50) return "text-green-400"
    if (speedMbps >= 10) return "text-yellow-400"
    return "text-red-400"
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "success":
        return <CheckCircle className="h-4 w-4 text-green-400" />
      case "failed":
        return <XCircle className="h-4 w-4 text-red-400" />
      default:
        return <Loader2 className="h-4 w-4 text-blue-400 animate-spin" />
    }
  }

  const getStatusBadge = (status: string) => {
    const baseClasses = "text-xs"
    switch (status) {
      case "success":
        return <Badge variant="default" className={`${baseClasses} bg-green-600 hover:bg-green-700`}>成功</Badge>
      case "failed":
        return <Badge variant="destructive" className={baseClasses}>失败</Badge>
      default:
        return <Badge variant="secondary" className={baseClasses}>测试中</Badge>
    }
  }

  return (
    <div className="space-y-6">
      {/* Progress Summary */}
      {progress && (
        <Card className="glass-morphism border-gray-800 p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
              <ClientIcon icon={TrendingUp} className="h-5 w-5 text-blue-400" />
              测试进度
            </h3>
            <div className="flex items-center gap-2">
              <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-400 animate-pulse' : 'bg-red-400'}`} />
              <span className="text-sm text-gray-400">
                {isConnected ? '已连接' : '未连接'}
              </span>
            </div>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-white">{progress.completed_count}</div>
              <div className="text-sm text-gray-400">已完成</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-white">{progress.total_count}</div>
              <div className="text-sm text-gray-400">总数</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-400">{progress.progress_percent.toFixed(1)}%</div>
              <div className="text-sm text-gray-400">进度</div>
            </div>
          </div>

          <div className="w-full bg-gray-800 rounded-full h-3 overflow-hidden">
            <div
              className="h-full progress-bar transition-all duration-300 ease-out"
              style={{ width: `${progress.progress_percent}%` }}
            />
          </div>
          
          {progress.current_proxy && (
            <div className="mt-4 text-center">
              <span className="text-sm text-gray-400">当前测试: </span>
              <span className="text-sm text-white font-medium">{progress.current_proxy}</span>
            </div>
          )}
        </Card>
      )}

      {/* Completion Summary */}
      {completeData && (
        <Card className="glass-morphism border-gray-800 p-6">
          <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <ClientIcon icon={CheckCircle} className="h-5 w-5 text-green-400" />
            测试完成
          </h3>
          
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
            <div className="text-center">
              <div className="text-xl font-bold text-green-400">{completeData.successful_tests}</div>
              <div className="text-xs text-gray-400">成功</div>
            </div>
            <div className="text-center">
              <div className="text-xl font-bold text-red-400">{completeData.failed_tests}</div>
              <div className="text-xs text-gray-400">失败</div>
            </div>
            <div className="text-center">
              <div className="text-xl font-bold text-blue-400">{completeData.average_download_mbps.toFixed(1)}</div>
              <div className="text-xs text-gray-400">平均下载(MB/s)</div>
            </div>
            <div className="text-center">
              <div className="text-xl font-bold text-purple-400">{completeData.average_latency.toFixed(0)}</div>
              <div className="text-xs text-gray-400">平均延迟(ms)</div>
            </div>
          </div>

          {completeData.best_proxy && (
            <div className="text-center text-sm">
              <span className="text-gray-400">最佳节点: </span>
              <span className="text-white font-medium">{completeData.best_proxy}</span>
              <span className="text-green-400 ml-2">({completeData.best_download_speed_mbps.toFixed(2)} MB/s)</span>
            </div>
          )}
        </Card>
      )}

      {/* Cancellation Summary */}
      {cancelledData && (
        <Card className="glass-morphism border-gray-800 p-6">
          <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <ClientIcon icon={XCircle} className="h-5 w-5 text-orange-400" />
            测试已取消
          </h3>
          
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-4">
            <div className="text-center">
              <div className="text-xl font-bold text-orange-400">{cancelledData.completed_tests}</div>
              <div className="text-xs text-gray-400">已完成</div>
            </div>
            <div className="text-center">
              <div className="text-xl font-bold text-gray-400">{cancelledData.total_tests}</div>
              <div className="text-xs text-gray-400">总数</div>
            </div>
            <div className="text-center">
              <div className="text-xl font-bold text-blue-400">{cancelledData.partial_duration}</div>
              <div className="text-xs text-gray-400">用时</div>
            </div>
          </div>

          <div className="text-center text-sm">
            <span className="text-gray-400">取消原因: </span>
            <span className="text-orange-400 font-medium">{cancelledData.message}</span>
          </div>
        </Card>
      )}

      {/* Results Table */}
      {results.length > 0 && (
        <Card className="glass-morphism border-gray-800">
          <div className="p-6">
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-xl font-bold text-white">实时测试结果</h2>
              <Badge variant="outline" className="border-gray-700 text-gray-300">
                {results.length} 个结果
              </Badge>
            </div>

            <div className="overflow-x-auto">
              <Table className="table-dark">
                <TableHeader>
                  <TableRow className="border-gray-800">
                    <TableHead className="text-gray-400">状态</TableHead>
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
                    <TableHead className="text-gray-400">错误详情</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {results.map((result, index) => (
                    <TableRow 
                      key={`${result.proxy_name}-${index}`} 
                      className="table-row-dark animate-in slide-in-from-bottom-1 duration-300"
                    >
                      <TableCell>
                        <div className="flex items-center gap-2">
                          {getStatusIcon(result.status)}
                          {getStatusBadge(result.status)}
                        </div>
                      </TableCell>
                      <TableCell className="font-medium text-white max-w-xs">
                        <div className="truncate" title={result.proxy_name}>
                          {result.proxy_name}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant="secondary" className="badge-dark text-xs">
                          {result.proxy_type}
                        </Badge>
                      </TableCell>
                      <TableCell className={getLatencyColor(result.latency_ms)}>
                        {formatLatency(result.latency_ms)}
                      </TableCell>
                      <TableCell className={getSpeedColor(result.download_speed_mbps)}>
                        {formatSpeed(result.download_speed_mbps)}
                      </TableCell>
                      <TableCell className={getSpeedColor(result.upload_speed_mbps)}>
                        {formatSpeed(result.upload_speed_mbps)}
                      </TableCell>
                      <TableCell className="text-gray-400">
                        {result.packet_loss.toFixed(1)}%
                      </TableCell>
                      <TableCell className="text-gray-400 max-w-xs">
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
                            <div className="text-xs text-gray-500 truncate" title={result.error_message}>
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
      )}

      {/* Empty State */}
      {results.length === 0 && !progress && (
        <Card className="glass-morphism border-gray-800 p-12">
          <div className="text-center">
            <ClientIcon icon={Zap} className="h-12 w-12 text-gray-500 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-400 mb-2">等待测试开始</h3>
            <p className="text-sm text-gray-500">
              点击"开始测试"按钮开始代理速度测试，结果将在此处实时显示
            </p>
          </div>
        </Card>
      )}
    </div>
  )
}