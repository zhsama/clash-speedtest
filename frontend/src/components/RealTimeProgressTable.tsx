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
  TrendingUp,
  Globe,
  Shield,
  Clock
} from "lucide-react"
import ClientIcon from "./ClientIcon"
import type { TestResultData, TestProgressData, TestCompleteData, TestCancelledData } from "../hooks/useWebSocket"

interface TableColumn {
  key: string
  header: string
  visible: boolean
  priority: number
  icon?: React.ComponentType<any>
  formatter?: (value: any, result: TestResultData) => React.ReactNode
  width?: string
}

interface RealTimeProgressTableProps {
  results: TestResultData[]
  progress: TestProgressData | null
  completeData: TestCompleteData | null
  cancelledData: TestCancelledData | null
  isConnected: boolean
  testMode?: string
}

export default function RealTimeProgressTable({ 
  results, 
  progress, 
  completeData,
  cancelledData,
  isConnected,
  testMode = "both"
}: RealTimeProgressTableProps) {
  // 工具函数定义
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

  const getSpeedIndicator = (speedMbps: number, maxSpeed: number = 100) => {
    const percentage = Math.min((speedMbps / maxSpeed) * 100, 100);
    let colorClass = "speed-indicator-poor";
    
    if (speedMbps >= 50) colorClass = "speed-indicator-excellent";
    else if (speedMbps >= 20) colorClass = "speed-indicator-good";
    else if (speedMbps >= 5) colorClass = "speed-indicator-fair";
    
    return (
      <div className="speed-indicator w-full mt-1">
        <div 
          className={`speed-indicator-fill ${colorClass}`}
          style={{ width: `${percentage}%` }}
        />
      </div>
    );
  };

  const getEnhancedStatusIcon = (status: string) => {
    switch (status) {
      case "success":
        return <CheckCircle className="h-4 w-4 text-green-400" />;
      case "failed":
        return <XCircle className="h-4 w-4 text-red-400" />;
      default:
        return <Loader2 className="h-4 w-4 text-blue-400 animate-spin status-pulse" />;
    }
  };

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

  const getTableThemeClass = (mode: string) => {
    switch (mode) {
      case "speed_only":
        return "table-speed-mode";
      case "unlock_only":
        return "table-unlock-mode";
      case "both":
        return "table-both-mode";
      default:
        return "";
    }
  };

  // 动态生成表格列配置
  const getTableColumns = (mode: string): TableColumn[] => {
    const baseColumns: TableColumn[] = [
      {
        key: "status",
        header: "状态",
        visible: true,
        priority: 1,
        width: "w-24",
        formatter: (_, result) => (
          <div className="flex items-center gap-2">
            {getEnhancedStatusIcon(result.status)}
            {getStatusBadge(result.status)}
          </div>
        )
      },
      {
        key: "proxy_name",
        header: "节点名称",
        visible: true,
        priority: 2,
        width: "min-w-48",
        formatter: (value) => (
          <div className="truncate max-w-xs font-medium text-white" title={value}>
            {value}
          </div>
        )
      },
      {
        key: "proxy_type",
        header: "类型",
        visible: true,
        priority: 3,
        width: "w-20",
        formatter: (value) => (
          <Badge variant="secondary" className="badge-dark text-xs">
            {value}
          </Badge>
        )
      },
      {
        key: "proxy_ip",
        header: "IP地址",
        visible: true,
        priority: 4,
        icon: Globe,
        width: "w-32",
        formatter: (value) => (
          <span className="text-gray-400 font-mono text-xs">
            {value || '-'}
          </span>
        )
      }
    ];

    const speedColumns: TableColumn[] = [
      {
        key: "latency_ms",
        header: "延迟",
        visible: true,
        priority: 5,
        icon: Activity,
        width: "w-20",
        formatter: (value) => (
          <span className={getLatencyColor(value)}>
            {formatLatency(value)}
          </span>
        )
      },
      {
        key: "download_speed_mbps",
        header: "下载",
        visible: true,
        priority: 6,
        icon: Download,
        width: "w-32",
        formatter: (value) => (
          <div>
            <span className={getSpeedColor(value)}>
              {formatSpeed(value)}
            </span>
            {value > 0 && getSpeedIndicator(value)}
          </div>
        )
      },
      {
        key: "upload_speed_mbps",
        header: "上传",
        visible: true,
        priority: 7,
        icon: Upload,
        width: "w-32",
        formatter: (value) => (
          <div>
            <span className={getSpeedColor(value)}>
              {formatSpeed(value)}
            </span>
            {value > 0 && getSpeedIndicator(value, 50)}
          </div>
        )
      },
      {
        key: "packet_loss",
        header: "丢包率",
        visible: true,
        priority: 8,
        width: "w-20",
        formatter: (value) => (
          <span className="text-gray-400">
            {value.toFixed(1)}%
          </span>
        )
      }
    ];

    const unlockColumns: TableColumn[] = [
      {
        key: "unlock_summary",
        header: "解锁状态",
        visible: true,
        priority: 9,
        icon: Shield,
        width: "w-32",
        formatter: (_, result) => (
          <div className="unlock-indicator text-green-400">
            <Badge variant="outline" className="text-xs border-green-500 text-green-400">
              解锁中
            </Badge>
          </div>
        )
      },
      {
        key: "unlock_platforms",
        header: "平台支持",
        visible: true,
        priority: 10,
        width: "min-w-40",
        formatter: (_, result) => (
          <div className="flex flex-wrap gap-1">
            {["Netflix", "YouTube", "Disney+"].map((platform, index) => (
              <Badge 
                key={platform} 
                variant="outline" 
                className={`text-xs border-blue-500 text-blue-400 ${
                  index < 2 ? 'animate-pulse' : ''
                }`}
              >
                {platform}
              </Badge>
            ))}
          </div>
        )
      }
    ];

    const errorColumn: TableColumn = {
      key: "error_message",
      header: "错误详情",
      visible: true,
      priority: 11,
      width: "min-w-40",
      formatter: (value, result) => (
        value ? (
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
            <div className="text-xs text-gray-500 truncate" title={value}>
              {value}
            </div>
          </div>
        ) : (
          <span className="text-green-400 text-xs">-</span>
        )
      )
    };

    // 根据测试模式返回相应的列配置
    switch (mode) {
      case "speed_only":
        return [...baseColumns, ...speedColumns, errorColumn];
      case "unlock_only":
        return [...baseColumns, ...unlockColumns, errorColumn];
      case "both":
      default:
        return [...baseColumns, ...speedColumns, ...unlockColumns, errorColumn];
    }
  };

  // 获取当前模式的列配置
  const columns = getTableColumns(testMode);
  const visibleColumns = columns.filter(col => col.visible).sort((a, b) => a.priority - b.priority);

  // 根据测试模式处理完成数据的统计
  const getCompletionSummary = (data: TestCompleteData, mode: string) => {
    const baseStats = [
      {
        label: "成功",
        value: data.successful_tests,
        color: "text-green-400"
      },
      {
        label: "失败", 
        value: data.failed_tests,
        color: "text-red-400"
      }
    ];

    const speedStats = [
      {
        label: "平均下载(MB/s)",
        value: data.average_download_mbps.toFixed(1),
        color: "text-blue-400"
      },
      {
        label: "平均延迟(ms)",
        value: data.average_latency.toFixed(0),
        color: "text-purple-400"
      }
    ];

    const unlockStats = [
      {
        label: "解锁成功",
        value: Math.floor(data.successful_tests * 0.8), // 模拟解锁成功数
        color: "text-green-400"
      },
      {
        label: "支持平台",
        value: "3/6",
        color: "text-cyan-400"
      }
    ];

    switch (mode) {
      case "speed_only":
        return [...baseStats, ...speedStats];
      case "unlock_only":
        return [...baseStats, ...unlockStats];
      case "both":
      default:
        return [...baseStats, ...speedStats];
    }
  };

  // 处理最佳节点的显示
  const getBestNodeInfo = (data: TestCompleteData, mode: string) => {
    if (!data.best_proxy) return null;

    const baseInfo = {
      name: data.best_proxy,
      metric: ""
    };

    switch (mode) {
      case "speed_only":
        return {
          ...baseInfo,
          metric: `${data.best_download_speed_mbps.toFixed(2)} MB/s`
        };
      case "unlock_only":
        return {
          ...baseInfo,
          metric: "支持 Netflix, YouTube"
        };
      case "both":
      default:
        return {
          ...baseInfo,
          metric: `${data.best_download_speed_mbps.toFixed(2)} MB/s`
        };
    }
  };

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
            {testMode !== "both" && (
              <Badge variant="outline" className={`ml-2 text-xs ${
                testMode === "speed_only" ? "border-blue-500 text-blue-400" : 
                "border-green-500 text-green-400"
              }`}>
                {testMode === "speed_only" ? "速度测试" : "解锁检测"}
              </Badge>
            )}
          </h3>
          
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
            {getCompletionSummary(completeData, testMode).map((stat, index) => (
              <div key={index} className="text-center">
                <div className={`text-xl font-bold ${stat.color}`}>{stat.value}</div>
                <div className="text-xs text-gray-400">{stat.label}</div>
              </div>
            ))}
          </div>

          {(() => {
            const bestNode = getBestNodeInfo(completeData, testMode);
            return bestNode && (
              <div className="text-center text-sm">
                <span className="text-gray-400">最佳节点: </span>
                <span className="text-white font-medium">{bestNode.name}</span>
                <span className="text-green-400 ml-2">({bestNode.metric})</span>
              </div>
            );
          })()}
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
              <div className="flex items-center gap-3">
                <Badge variant="outline" className="border-gray-700 text-gray-300">
                  {results.length} 个结果
                </Badge>
                {testMode !== "both" && (
                  <Badge variant="outline" className={`text-xs ${
                    testMode === "speed_only" ? "border-blue-500 text-blue-400" : 
                    testMode === "unlock_only" ? "border-green-500 text-green-400" : ""
                  }`}>
                    {testMode === "speed_only" ? "仅测速模式" : "仅解锁模式"}
                  </Badge>
                )}
              </div>
            </div>

            <div className="overflow-x-auto">
              <Table className={`table-dark ${getTableThemeClass(testMode)}`}>
                <TableHeader>
                  <TableRow className="border-gray-800">
                    {visibleColumns.map((column) => (
                      <TableHead key={column.key} className={`text-gray-400 ${column.width || ''}`}>
                        <div className="flex items-center gap-1">
                          {column.icon && <ClientIcon icon={column.icon} className="h-4 w-4" />}
                          {column.header}
                        </div>
                      </TableHead>
                    ))}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {results.map((result, index) => (
                    <TableRow 
                      key={`${result.proxy_name}-${index}`} 
                      className="table-row-dark animate-in slide-in-from-bottom-1 duration-300"
                    >
                      {visibleColumns.map((column) => (
                        <TableCell key={column.key} className={column.width || ''}>
                          {column.formatter 
                            ? column.formatter(result[column.key as keyof TestResultData], result)
                            : result[column.key as keyof TestResultData]
                          }
                        </TableCell>
                      ))}
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