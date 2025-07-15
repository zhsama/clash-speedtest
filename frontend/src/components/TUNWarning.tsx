import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
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
      <Card className="card-standard my-4">
        <div className="flex items-center gap-2">
          <ClientIcon 
            icon={loading ? Loader2 : Network} 
            className={`h-5 w-5 text-shamrock-400 ${loading ? 'animate-spin' : ''}`} 
          />
          <span className="text-shamrock-100">检测 TUN 模式状态...</span>
          {!loading && (
            <Button
              variant="outline"
              size="sm"
              onClick={checkTUNMode}
              className="button-standard ml-auto"
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
    <div className="form-element">
      {/* 主要警告信息 */}
      {tunStatus.enabled ? (
        <Card className="card-standard border-yellow-500 bg-yellow-500/10">
          <div className="flex items-start gap-2">
            <ClientIcon icon={AlertTriangle} className="h-5 w-5 text-yellow-500 mt-0.5" />
            <div className="flex-1">
              <div className="flex items-start justify-between">
                <div>
                  <div className="font-medium form-element-label text-yellow-200">检测到 TUN 模式已启用</div>
                  <div className="text-sm text-yellow-200">{warning}</div>
                </div>
                <span className="badge-standard border-yellow-500 text-yellow-400 ml-4">
                  TUN 模式
                </span>
              </div>
            </div>
          </div>
        </Card>
      ) : (
        <Card className="card-standard border-green-500 bg-green-500/10">
          <div className="flex items-start gap-2">
            <ClientIcon icon={CheckCircle} className="h-5 w-5 text-green-500 mt-0.5" />
            <div className="flex-1">
              <div className="flex items-center justify-between">
                <div>
                  <div className="font-medium text-green-200">TUN 模式未启用</div>
                  <div className="text-sm text-green-200">系统网络配置正常，可以进行速度测试</div>
                </div>
                <span className="badge-standard border-green-500 text-green-400">
                  正常
                </span>
              </div>
            </div>
          </div>
        </Card>
      )}

      {/* 操作按钮 */}
      <div className="flex items-center component-gap my-4">
        <Button
          variant="outline"
          size="sm"
          onClick={checkTUNMode}
          disabled={loading}
          className="button-standard"
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
          className="button-standard"
        >
          <ClientIcon icon={Info} className="h-4 w-4 mr-1" />
          {showDetailedInfo ? '隐藏详情' : '显示详情'}
        </Button>
        
        {lastChecked && (
          <span className="text-xs text-shamrock-400 ml-auto">
            最后检测: {lastChecked.toLocaleTimeString('zh-CN')}
          </span>
        )}
      </div>

      {/* 详细信息 */}
      {showDetailedInfo && (
        <Card className="card-standard">
          <h4 className="form-element-label flex items-center gap-2">
            <ClientIcon icon={Network} className="h-4 w-4 text-shamrock-400" />
            TUN 模式详细信息
          </h4>

          {/* 系统信息 */}
          <div className="form-element">
            <div className="form-element-label">系统信息</div>
            <div className="grid grid-cols-2 component-gap text-sm">
              <div>
                <span className="text-shamrock-400">操作系统:</span>
                <span className="text-shamrock-100 ml-2">{tunStatus.system_info.os}</span>
              </div>
              <div>
                <span className="text-shamrock-400">架构:</span>
                <span className="text-shamrock-100 ml-2">{tunStatus.system_info.architecture}</span>
              </div>
              <div>
                <span className="text-shamrock-400">主机名:</span>
                <span className="text-shamrock-100 ml-2">{tunStatus.system_info.hostname}</span>
              </div>
              <div>
                <span className="text-shamrock-400">检测时间:</span>
                <span className="text-shamrock-100 ml-2">{formatTime(tunStatus.detection_time)}</span>
              </div>
            </div>
          </div>

          {/* TUN 接口信息 */}
          {tunStatus.interfaces.length > 0 && (
            <div className="form-element">
              <div className="form-element-label">TUN 接口</div>
              <div className="space-y-2">
                {tunStatus.interfaces.map((iface, index) => (
                  <div 
                    key={index} 
                    className={`card-standard ${
                      iface.is_up ? 'border-green-600 bg-green-500/10' : 'border-shamrock-600'
                    }`}
                  >
                    <div className="flex items-center justify-between form-element">
                      <span className="font-medium text-shamrock-100">{iface.name}</span>
                      <div className="flex component-gap">
                        <span className={`badge-standard ${
                          iface.is_up 
                            ? 'border-green-500 text-green-400' 
                            : 'border-shamrock-500 text-shamrock-400'
                        }`}>
                          {iface.is_up ? '启用' : '禁用'}
                        </span>
                        {iface.is_default && (
                          <span className="badge-standard border-shamrock-500 text-shamrock-400">
                            默认
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="text-xs text-shamrock-400 space-y-1">
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
            <div className="form-element">
              <div className="form-element-label">检测到的代理进程</div>
              <div className="space-y-2">
                {tunStatus.proxy_processes.map((process, index) => (
                  <div key={index} className="card-standard border-orange-600 bg-orange-500/10">
                    <div className="flex items-center justify-between form-element">
                      <span className="font-medium text-shamrock-100">{process.name}</span>
                      <span className="badge-standard border-orange-500 text-orange-400">
                        {process.process_type}
                      </span>
                    </div>
                    {process.pid > 0 && (
                      <div className="text-xs text-shamrock-400">PID: {process.pid}</div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* 默认路由信息 */}
          {tunStatus.default_route && (
            <div className="form-element">
              <div className="form-element-label">默认路由</div>
              <div className="card-standard">
                <div className="text-sm space-y-1">
                  <div>
                    <span className="text-shamrock-400">目标:</span>
                    <span className="text-shamrock-100 ml-2">{tunStatus.default_route.destination}</span>
                  </div>
                  <div>
                    <span className="text-shamrock-400">网关:</span>
                    <span className="text-shamrock-100 ml-2">{tunStatus.default_route.gateway}</span>
                  </div>
                  <div>
                    <span className="text-shamrock-400">接口:</span>
                    <span className="text-shamrock-100 ml-2">{tunStatus.default_route.interface}</span>
                  </div>
                  {tunStatus.default_route.metric > 0 && (
                    <div>
                      <span className="text-shamrock-400">优先级:</span>
                      <span className="text-shamrock-100 ml-2">{tunStatus.default_route.metric}</span>
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