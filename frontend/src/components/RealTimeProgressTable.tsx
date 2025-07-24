import { Card } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { 
  TrendingUp,
  CheckCircle,
  XCircle,
  Zap,
  Download,
  Shield,
  FileText,
  TableIcon,
} from "lucide-react"
import ClientIcon from "./ClientIcon"
import SpeedTestTable from "./SpeedTestTable"
import UnlockTestTable from "./UnlockTestTable"
import type { TestResultData, TestProgressData, TestCompleteData, TestCancelledData } from "../hooks/useWebSocket"

interface RealTimeProgressTableProps {
  results: TestResultData[]
  progress: TestProgressData | null
  completeData: TestCompleteData | null
  cancelledData: TestCancelledData | null
  isConnected: boolean
  testMode?: string
  // 导出功能相关
  onExportMarkdown?: () => void
  onExportCSV?: () => void
  showExportButtons?: boolean
}

export default function RealTimeProgressTable({ 
  results, 
  progress, 
  completeData,
  cancelledData,
  isConnected,
  testMode = "both",
  onExportMarkdown,
  onExportCSV,
  showExportButtons = false
}: RealTimeProgressTableProps) {
  
  // 获取当前测试阶段的显示
  const getCurrentStageDisplay = (progress: TestProgressData | null) => {
    if (!progress) return null

    const { current_stage, unlock_platform } = progress

    if (current_stage === "speed_test") {
      return (
        <div className="flex items-center gap-2">
          <ClientIcon icon={Download} className="h-4 w-4 text-blue-400" />
          <span className="text-sm text-blue-400">速度测试</span>
        </div>
      )
    } else if (current_stage === "unlock_test") {
      return (
        <div className="flex items-center gap-2">
          <ClientIcon icon={Shield} className="h-4 w-4 text-green-400" />
          <span className="text-sm text-green-400">解锁检测</span>
          {unlock_platform && (
            <Badge variant="outline" className="text-xs border-green-500 text-green-400">
              {unlock_platform}
            </Badge>
          )}
        </div>
      )
    }

    return null
  }

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

    const unlockStats = [];
    if (data.unlock_stats) {
      unlockStats.push(
        {
          label: "解锁成功",
          value: data.unlock_stats.successful_unlock_tests,
          color: "text-green-400"
        },
        {
          label: "解锁总数",
          value: data.unlock_stats.total_unlock_tests,
          color: "text-cyan-400"
        }
      );
    }

    switch (mode) {
      case "speed_only":
        return [...baseStats, ...speedStats];
      case "unlock_only":
        return [...baseStats, ...unlockStats];
      case "both":
      default:
        return [...baseStats, ...speedStats, ...unlockStats];
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
        if (data.unlock_stats?.best_unlock_proxy) {
          return {
            name: data.unlock_stats.best_unlock_proxy,
            metric: `支持 ${data.unlock_stats.best_unlock_platforms?.join(', ') || '多个平台'}`
          };
        }
        return null;
      case "both":
      default:
        return {
          ...baseInfo,
          metric: `${data.best_download_speed_mbps.toFixed(2)} MB/s`
        };
    }
  };

  // 根据测试模式渲染相应的表格组件
  const renderTablesByMode = () => {
    if (results.length === 0) {
      return (
        <Card className="card-standard">
          <div className="text-center py-12">
            <ClientIcon icon={Zap} className="h-12 w-12 text-lavender-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-lavender-400 mb-2">等待测试开始</h3>
            <p className="text-sm text-lavender-500">
              {`点击"开始测试"按钮开始代理${testMode === "speed_only" ? "速度" : testMode === "unlock_only" ? "解锁" : "速度和解锁"}测试，结果将在此处实时显示`}
            </p>
          </div>
        </Card>
      )
    }

    switch (testMode) {
      case "speed_only":
        return (
          <SpeedTestTable 
            results={results} 
            onExportMarkdown={onExportMarkdown}
            onExportCSV={onExportCSV}
            showExportButtons={showExportButtons}
          />
        )
      
      case "unlock_only":
        return (
          <UnlockTestTable 
            results={results} 
            onExportMarkdown={onExportMarkdown}
            onExportCSV={onExportCSV}
            showExportButtons={showExportButtons}
          />
        )
      
      case "both":
      default:
        return (
          <div className="space-y-6">
            <SpeedTestTable 
              results={results} 
              onExportMarkdown={onExportMarkdown}
              onExportCSV={onExportCSV}
              showExportButtons={showExportButtons}
            />
            <UnlockTestTable 
              results={results} 
              onExportMarkdown={onExportMarkdown}
              onExportCSV={onExportCSV}
              showExportButtons={showExportButtons}
            />
          </div>
        )
    }
  }

  return (
    <div className="space-y-6">
      {/* Progress Summary */}
      {progress && (
        <Card className="card-standard">
          <div className="flex items-center justify-between form-element">
            <h3 className="text-lg font-semibold text-lavender-50 flex items-center gap-2">
              <ClientIcon icon={TrendingUp} className="h-5 w-5 text-lavender-400" />
              测试进度
            </h3>
            <div className="flex items-center gap-2">
              <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-400 animate-pulse' : 'bg-red-400'}`} />
              <span className="text-sm text-lavender-400">
                {isConnected ? '已连接' : '未连接'}
              </span>
            </div>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 form-element">
            <div className="text-center">
              <div className="text-2xl font-bold text-lavender-50">{progress.completed_count}</div>
              <div className="text-sm text-lavender-400">已完成</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-lavender-50">{progress.total_count}</div>
              <div className="text-sm text-lavender-400">总数</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-lavender-500">{progress.progress_percent.toFixed(1)}%</div>
              <div className="text-sm text-lavender-400">进度</div>
            </div>
          </div>

          <div className="w-full bg-lavender-800 rounded-full h-3 overflow-hidden form-element">
            <div
              className="h-full bg-lavender-500 transition-all duration-300 ease-out"
              style={{ width: `${progress.progress_percent}%` }}
            />
          </div>
          
          <div className="space-y-2">
            {progress.current_proxy && (
              <div className="text-center">
                <span className="text-sm text-lavender-400">当前测试: </span>
                <span className="text-sm text-lavender-50 font-medium">{progress.current_proxy}</span>
              </div>
            )}
            {getCurrentStageDisplay(progress) && (
              <div className="flex justify-center">
                {getCurrentStageDisplay(progress)}
              </div>
            )}
          </div>
        </Card>
      )}

      {/* Completion Summary */}
      {completeData && (
        <Card className="card-standard">
          <h3 className="text-lg font-semibold text-lavender-50 form-element flex items-center gap-2">
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
          
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 form-element">
            {getCompletionSummary(completeData, testMode).map((stat, index) => (
              <div key={index} className="text-center">
                <div className={`text-xl font-bold ${stat.color}`}>{stat.value}</div>
                <div className="text-xs text-lavender-400">{stat.label}</div>
              </div>
            ))}
          </div>

          {(() => {
            const bestNode = getBestNodeInfo(completeData, testMode);
            return bestNode && (
              <div className="text-center text-sm">
                <span className="text-lavender-400">最佳节点: </span>
                <span className="text-lavender-50 font-medium">{bestNode.name}</span>
                <span className="text-green-400 ml-2">({bestNode.metric})</span>
              </div>
            );
          })()}
        </Card>
      )}

      {/* Cancellation Summary */}
      {cancelledData && (
        <Card className="card-standard">
          <h3 className="text-lg font-semibold text-lavender-50 form-element flex items-center gap-2">
            <ClientIcon icon={XCircle} className="h-5 w-5 text-orange-400" />
            测试已取消
          </h3>
          
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4 form-element">
            <div className="text-center">
              <div className="text-xl font-bold text-orange-400">{cancelledData.completed_tests}</div>
              <div className="text-xs text-lavender-400">已完成</div>
            </div>
            <div className="text-center">
              <div className="text-xl font-bold text-lavender-400">{cancelledData.total_tests}</div>
              <div className="text-xs text-lavender-400">总数</div>
            </div>
            <div className="text-center">
              <div className="text-xl font-bold text-lavender-500">{cancelledData.partial_duration}</div>
              <div className="text-xs text-lavender-400">用时</div>
            </div>
          </div>

          <div className="text-center text-sm">
            <span className="text-lavender-400">取消原因: </span>
            <span className="text-orange-400 font-medium">{cancelledData.message}</span>
          </div>
        </Card>
      )}

      {/* Results Tables */}
      <div className="space-y-4">
        {renderTablesByMode()}
      </div>
    </div>
  )
}