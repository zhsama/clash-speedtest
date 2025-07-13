import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { 
  AlertTriangle, 
  RefreshCw, 
  Network, 
  CheckCircle,
  Info,
  Loader2
} from "lucide-react"
import ClientIcon from "./ClientIcon"
import { config } from "@/lib/env"
import { toast } from "sonner"

interface TUNInterface {
  name: string
  type: string
  ip_addresses: string[]
  is_up: boolean
  mtu: number
  is_default: boolean
  associated_pid: number
}

interface ProxyProcess {
  name: string
  pid: number
  command: string
  process_type: string
}

interface RouteInfo {
  destination: string
  gateway: string
  interface: string
  metric: number
}

interface SystemInfo {
  os: string
  architecture: string
  hostname: string
}

interface TUNStatus {
  enabled: boolean
  interfaces: TUNInterface[]
  active_interface?: TUNInterface
  proxy_processes: ProxyProcess[]
  default_route?: RouteInfo
  detection_time: string
  system_info: SystemInfo
  additional_details: Record<string, any>
}

interface TUNCheckResponse {
  success: boolean
  tun_status: TUNStatus
  warning: string
}

interface TUNWarningProps {
  onTUNStatusChange?: (enabled: boolean) => void
  showDetails?: boolean
}

export default function TUNWarning({ onTUNStatusChange, showDetails = false }: TUNWarningProps) {
  const [tunStatus, setTunStatus] = useState<TUNStatus | null>(null)
  const [warning, setWarning] = useState("")
  const [loading, setLoading] = useState(false)
  const [showDetailedInfo, setShowDetailedInfo] = useState(showDetails)
  const [lastChecked, setLastChecked] = useState<Date | null>(null)

  const checkTUNMode = async () => {
    setLoading(true)
    try {
      const response = await fetch(`${config.apiUrl}/api/tun-check`)
      const data: TUNCheckResponse = await response.json()
      
      if (data.success) {
        setTunStatus(data.tun_status)
        setWarning(data.warning)
        setLastChecked(new Date())
        
        // 通知父组件TUN状态变化
        onTUNStatusChange?.(data.tun_status.enabled)
        
        if (data.tun_status.enabled) {
          toast.warning("检测到 TUN 模式", {
            description: "建议关闭 TUN 模式以获得更准确的测试结果"
          })
        } else {
          toast.success("TUN 模式检测", {
            description: "未检测到 TUN 模式，可以进行测试"
          })
        }
      } else {
        toast.error("检测失败", {
          description: "无法检测 TUN 模式状态"
        })
      }
    } catch (error) {
      console.error("TUN 模式检测失败:", error)
      toast.error("检测出错", {
        description: "TUN 模式检测请求失败"
      })
    } finally {
      setLoading(false)
    }
  }

  // 组件加载时自动检测
  useEffect(() => {
    checkTUNMode()
  }, [])

  const formatTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN')
  }

  if (!tunStatus) {
    return (
      <Card className="border-gray-700 bg-gray-800/50 p-4">
        <div className="flex items-center gap-3">
          <ClientIcon 
            icon={loading ? Loader2 : Network} 
            className={`h-5 w-5 text-blue-400 ${loading ? 'animate-spin' : ''}`} 
          />
          <span className="text-gray-300">检测 TUN 模式状态...</span>
          {!loading && (
            <Button
              variant="outline"
              size="sm"
              onClick={checkTUNMode}
              className="ml-auto border-gray-600 text-gray-300 hover:bg-gray-700"
            >
              <ClientIcon icon={RefreshCw} className="h-4 w-4 mr-1" />
              重新检测
            </Button>
          )}
        </div>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      {/* 主要警告信息 */}
      {tunStatus.enabled ? (
        <Card className="border-yellow-500 bg-yellow-500/10 p-4">
          <div className="flex items-start gap-3">
            <ClientIcon icon={AlertTriangle} className="h-5 w-5 text-yellow-500 mt-0.5" />
            <div className="flex-1">
              <div className="flex items-start justify-between">
                <div>
                  <div className="font-medium mb-1 text-yellow-200">检测到 TUN 模式已启用</div>
                  <div className="text-sm text-yellow-200">{warning}</div>
                </div>
                <Badge variant="outline" className="border-yellow-500 text-yellow-400 ml-4">
                  TUN 模式
                </Badge>
              </div>
            </div>
          </div>
        </Card>
      ) : (
        <Card className="border-green-500 bg-green-500/10 p-4">
          <div className="flex items-start gap-3">
            <ClientIcon icon={CheckCircle} className="h-5 w-5 text-green-500 mt-0.5" />
            <div className="flex-1">
              <div className="flex items-center justify-between">
                <div>
                  <div className="font-medium text-green-200">TUN 模式未启用</div>
                  <div className="text-sm text-green-200">系统网络配置正常，可以进行速度测试</div>
                </div>
                <Badge variant="outline" className="border-green-500 text-green-400">
                  正常
                </Badge>
              </div>
            </div>
          </div>
        </Card>
      )}

      {/* 操作按钮 */}
      <div className="flex items-center gap-3">
        <Button
          variant="outline"
          size="sm"
          onClick={checkTUNMode}
          disabled={loading}
          className="border-gray-600 text-gray-300 hover:bg-gray-700"
        >
          <ClientIcon 
            icon={loading ? Loader2 : RefreshCw} 
            className={`h-4 w-4 mr-1 ${loading ? 'animate-spin' : ''}`} 
          />
          重新检测
        </Button>
        
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setShowDetailedInfo(!showDetailedInfo)}
          className="text-gray-400 hover:text-gray-300"
        >
          <ClientIcon icon={Info} className="h-4 w-4 mr-1" />
          {showDetailedInfo ? '隐藏详情' : '显示详情'}
        </Button>
        
        {lastChecked && (
          <span className="text-xs text-gray-500 ml-auto">
            最后检测: {lastChecked.toLocaleTimeString('zh-CN')}
          </span>
        )}
      </div>

      {/* 详细信息 */}
      {showDetailedInfo && (
        <Card className="border-gray-700 bg-gray-800/30 p-4 space-y-4">
          <h4 className="text-gray-300 font-medium flex items-center gap-2">
            <ClientIcon icon={Network} className="h-4 w-4 text-blue-400" />
            TUN 模式详细信息
          </h4>

          {/* 系统信息 */}
          <div>
            <div className="text-sm text-gray-400 mb-2">系统信息</div>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-gray-500">操作系统:</span>
                <span className="text-gray-300 ml-2">{tunStatus.system_info.os}</span>
              </div>
              <div>
                <span className="text-gray-500">架构:</span>
                <span className="text-gray-300 ml-2">{tunStatus.system_info.architecture}</span>
              </div>
              <div>
                <span className="text-gray-500">主机名:</span>
                <span className="text-gray-300 ml-2">{tunStatus.system_info.hostname}</span>
              </div>
              <div>
                <span className="text-gray-500">检测时间:</span>
                <span className="text-gray-300 ml-2">{formatTime(tunStatus.detection_time)}</span>
              </div>
            </div>
          </div>

          {/* TUN 接口信息 */}
          {tunStatus.interfaces.length > 0 && (
            <div>
              <div className="text-sm text-gray-400 mb-2">TUN 接口</div>
              <div className="space-y-2">
                {tunStatus.interfaces.map((iface, index) => (
                  <div 
                    key={index} 
                    className={`p-3 rounded border ${
                      iface.is_up ? 'border-green-600 bg-green-900/20' : 'border-gray-600 bg-gray-800/50'
                    }`}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium text-gray-300">{iface.name}</span>
                      <div className="flex gap-2">
                        <Badge 
                          variant="outline" 
                          className={`text-xs ${
                            iface.is_up 
                              ? 'border-green-500 text-green-400' 
                              : 'border-gray-500 text-gray-400'
                          }`}
                        >
                          {iface.is_up ? '启用' : '禁用'}
                        </Badge>
                        {iface.is_default && (
                          <Badge variant="outline" className="border-blue-500 text-blue-400 text-xs">
                            默认
                          </Badge>
                        )}
                      </div>
                    </div>
                    <div className="text-xs text-gray-500 space-y-1">
                      <div>类型: {iface.type}</div>
                      <div>MTU: {iface.mtu}</div>
                      {iface.ip_addresses.length > 0 && (
                        <div>IP: {iface.ip_addresses.join(', ')}</div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* 代理进程信息 */}
          {tunStatus.proxy_processes.length > 0 && (
            <div>
              <div className="text-sm text-gray-400 mb-2">检测到的代理进程</div>
              <div className="space-y-2">
                {tunStatus.proxy_processes.map((process, index) => (
                  <div key={index} className="p-3 rounded border border-orange-600 bg-orange-900/20">
                    <div className="flex items-center justify-between mb-1">
                      <span className="font-medium text-gray-300">{process.name}</span>
                      <Badge variant="outline" className="border-orange-500 text-orange-400 text-xs">
                        {process.process_type}
                      </Badge>
                    </div>
                    {process.pid > 0 && (
                      <div className="text-xs text-gray-500">PID: {process.pid}</div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* 默认路由信息 */}
          {tunStatus.default_route && (
            <div>
              <div className="text-sm text-gray-400 mb-2">默认路由</div>
              <div className="p-3 rounded border border-gray-600 bg-gray-800/50">
                <div className="text-sm space-y-1">
                  <div>
                    <span className="text-gray-500">目标:</span>
                    <span className="text-gray-300 ml-2">{tunStatus.default_route.destination}</span>
                  </div>
                  <div>
                    <span className="text-gray-500">网关:</span>
                    <span className="text-gray-300 ml-2">{tunStatus.default_route.gateway}</span>
                  </div>
                  <div>
                    <span className="text-gray-500">接口:</span>
                    <span className="text-gray-300 ml-2">{tunStatus.default_route.interface}</span>
                  </div>
                  {tunStatus.default_route.metric > 0 && (
                    <div>
                      <span className="text-gray-500">优先级:</span>
                      <span className="text-gray-300 ml-2">{tunStatus.default_route.metric}</span>
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </Card>
      )}
    </div>
  )
}