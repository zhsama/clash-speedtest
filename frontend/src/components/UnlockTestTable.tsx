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
  CheckCircle,
  XCircle,
  Loader2,
  Globe,
  Shield,
  Lock,
  Unlock,
} from "lucide-react"
import ClientIcon from "./ClientIcon"
import type { TestResultData, UnlockResult } from "../hooks/useWebSocket"

interface UnlockTestTableProps {
  results: TestResultData[]
  title?: string
}

export default function UnlockTestTable({ results, title = "解锁检测结果" }: UnlockTestTableProps) {
  
  const getStatusIcon = (status: string) => {
    switch (status) {
      case "success":
        return <CheckCircle className="h-4 w-4 text-green-400" />;
      case "failed":
        return <XCircle className="h-4 w-4 text-red-400" />;
      default:
        return <Loader2 className="h-4 w-4 text-lavender-400 animate-spin" />;
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

  // 获取解锁结果的格式化显示
  const formatUnlockResults = (unlockResults: UnlockResult[]) => {
    if (!unlockResults || unlockResults.length === 0) {
      return <span className="text-lavender-500 text-xs">-</span>
    }

    const supported = unlockResults.filter(r => r.supported)
    const unsupported = unlockResults.filter(r => !r.supported)

    return (
      <div className="space-y-1">
        {supported.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {supported.map((result) => (
              <Badge
                key={result.platform}
                variant="outline"
                className="text-xs border-green-500 text-green-400"
              >
                <ClientIcon icon={Unlock} className="h-3 w-3 mr-1" />
                {result.platform}
                {result.region && ` (${result.region})`}
              </Badge>
            ))}
          </div>
        )}
        {unsupported.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {unsupported.map((result) => (
              <Badge
                key={result.platform}
                variant="outline"
                className="text-xs border-red-500 text-red-400"
              >
                <ClientIcon icon={Lock} className="h-3 w-3 mr-1" />
                {result.platform}
              </Badge>
            ))}
          </div>
        )}
      </div>
    )
  }

  // 获取解锁摘要的格式化显示
  const formatUnlockSummary = (unlockSummary: any) => {
    if (!unlockSummary) {
      return <span className="text-lavender-500 text-xs">-</span>
    }

    const { total_tested = 0, total_supported = 0 } = unlockSummary
    const supportedPlatforms = unlockSummary.supported_platforms || []
    const supportRate = total_tested > 0 ? (total_supported / total_tested * 100).toFixed(0) : 0

    return (
      <div className="space-y-1">
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="text-xs border-green-500 text-green-400">
            {total_supported}/{total_tested} ({supportRate}%)
          </Badge>
        </div>
        {supportedPlatforms.length > 0 && (
          <div className="text-xs text-green-400">
            {supportedPlatforms.join(', ')}
          </div>
        )}
      </div>
    )
  }

  // 显示所有结果，不进行过滤
  const unlockResults = results

  if (unlockResults.length === 0) {
    return (
      <Card className="card-standard">
        <div className="text-center py-8">
          <ClientIcon icon={Shield} className="h-12 w-12 text-lavender-600 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-lavender-400 mb-2">暂无解锁检测数据</h3>
          <p className="text-sm text-lavender-500">解锁检测结果将在此处显示</p>
        </div>
      </Card>
    )
  }

  return (
    <Card className="card-standard">
      <div className="form-element">
        <div className="flex justify-between items-center form-element">
          <h2 className="text-xl font-bold text-lavender-50 flex items-center gap-2">
            <ClientIcon icon={Shield} className="h-5 w-5 text-green-400" />
            {title}
          </h2>
          <Badge variant="outline" className="badge-standard">
            {unlockResults.length} 个结果
          </Badge>
        </div>

        <div className="table-scroll-container">
          <div className="overflow-x-auto table-horizontal-scroll">
            <Table className="table-standard table-unlock-mode">
            <TableHeader>
              <TableRow>
                <TableHead className="text-lavender-400 w-24">
                  <div className="flex items-center gap-1">
                    状态
                  </div>
                </TableHead>
                <TableHead className="text-lavender-400 min-w-48">
                  <div className="flex items-center gap-1">
                    节点名称
                  </div>
                </TableHead>
                <TableHead className="text-lavender-400 w-20">
                  <div className="flex items-center gap-1">
                    类型
                  </div>
                </TableHead>
                <TableHead className="text-lavender-400 w-32">
                  <div className="flex items-center gap-1">
                    <ClientIcon icon={Globe} className="h-4 w-4" />
                    IP地址
                  </div>
                </TableHead>
                <TableHead className="text-lavender-400 w-32">
                  <div className="flex items-center gap-1">
                    <ClientIcon icon={Shield} className="h-4 w-4" />
                    解锁摘要
                  </div>
                </TableHead>
                <TableHead className="text-lavender-400 min-w-48">
                  <div className="flex items-center gap-1">
                    <ClientIcon icon={Globe} className="h-4 w-4" />
                    平台详情
                  </div>
                </TableHead>
                <TableHead className="text-lavender-400 min-w-40">
                  <div className="flex items-center gap-1">
                    错误详情
                  </div>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {unlockResults.map((result, index) => (
                <TableRow key={`${result.proxy_name}-${index}`} className="table-row-dark">
                  <TableCell className="w-24">
                    <div className="flex items-center gap-2">
                      {getStatusIcon(result.status)}
                      {getStatusBadge(result.status)}
                    </div>
                  </TableCell>
                  <TableCell className="min-w-48">
                    <div className="truncate max-w-xs font-medium text-lavender-50" title={result.proxy_name}>
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
                      {result.proxy_ip || '-'}
                    </span>
                  </TableCell>
                  <TableCell className="w-32">
                    {formatUnlockSummary(result.unlock_summary)}
                  </TableCell>
                  <TableCell className="min-w-48">
                    {formatUnlockResults(result.unlock_results || [])}
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
                        <div className="text-xs text-lavender-500 truncate" title={result.error_message}>
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
      </div>
    </Card>
  )
}