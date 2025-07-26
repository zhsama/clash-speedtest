import { FaDownload as Download, FaGlobe as Globe, FaSpinner as Loader2 } from "react-icons/fa"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { config } from "@/lib/env"
import ClientIcon from "./ClientIcon"

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

interface ConfigManagerProps {
  configUrl: string
  setConfigUrl: (url: string) => void
  nodes: NodeInfo[]
  setNodes: (nodes: NodeInfo[]) => void
  filteredNodes: NodeInfo[]
  setFilteredNodes: (nodes: NodeInfo[]) => void
  filterConfig: FilterConfig
  setAvailableProtocols: (protocols: string[]) => void
  setFilterConfig: (config: FilterConfig | ((prev: FilterConfig) => FilterConfig)) => void
  applyFiltersWithConfig: (nodes: NodeInfo[], config: FilterConfig) => NodeInfo[]
  loading: boolean
  setLoading: (loading: boolean) => void
  testing: boolean
  isConnected: boolean
}

export default function ConfigManager({
  configUrl,
  setConfigUrl,
  nodes,
  setNodes,
  filteredNodes,
  setFilteredNodes,
  filterConfig,
  setAvailableProtocols,
  setFilterConfig,
  applyFiltersWithConfig,
  loading,
  setLoading,
  testing,
  isConnected,
}: ConfigManagerProps) {
  const fetchConfig = async () => {
    if (!configUrl.trim()) {
      toast.error("请输入配置文件路径或订阅链接")
      return
    }

    setLoading(true)
    setNodes([])
    setFilteredNodes([])

    try {
      const response = await fetch(`${config.apiUrl}/api/nodes`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          configPaths: configUrl,
          stashCompatible: filterConfig.stashCompatible,
        }),
      })

      const data = await response.json()

      if (data.success && data.nodes) {
        setNodes(data.nodes)

        const protocols = [...new Set(data.nodes.map((n: NodeInfo) => n.type))]
        setAvailableProtocols(protocols as string[])

        // 更新过滤配置
        const newFilterConfig = {
          ...filterConfig,
          protocolFilter: protocols as string[],
        }
        setFilterConfig(newFilterConfig)

        // 使用新的过滤配置来应用过滤和计算统计
        const filtered = applyFiltersWithConfig(data.nodes, newFilterConfig)
        setFilteredNodes(filtered)

        const filteredCount = filtered.length
        const hasFilters =
          newFilterConfig.includeNodes.length > 0 || newFilterConfig.excludeNodes.length > 0

        if (hasFilters && filteredCount < data.nodes.length) {
          const filteredOutCount = data.nodes.length - filteredCount
          toast.success(
            `成功加载 ${data.nodes.length} 个节点，已过滤 ${filteredOutCount} 个节点，符合条件 ${filteredCount} 个节点`
          )
        } else {
          toast.success(`成功加载 ${data.nodes.length} 个节点`)
        }
      } else {
        toast.error(data.error || "加载配置失败")
      }
    } catch (error) {
      toast.error(`请求失败：${(error as Error).message}`)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card className="card-elevated">
      <div className="flex items-center gap-2 form-element">
        <ClientIcon icon={Globe} className="h-5 w-5 text-lavender-400" />
        <h2 className="text-lg font-semibold text-lavender-50">配置获取</h2>
        <div className="ml-auto">
          {isConnected ? (
            <div className="status-indicator">
              <div className="status-dot success animate-pulse" />
              <span className="text-lavender-300 text-sm">WebSocket 已连接</span>
            </div>
          ) : (
            <div className="status-indicator">
              <div className="status-dot error" />
              <span className="text-lavender-300 text-sm">WebSocket 未连接</span>
            </div>
          )}
        </div>
      </div>

      <div className="flex component-gap">
        <Input
          placeholder="输入配置文件路径或订阅链接..."
          value={configUrl}
          onChange={(e) => setConfigUrl(e.target.value)}
          className="flex-1 input-outlined"
          disabled={loading || testing}
        />
        <Button
          onClick={fetchConfig}
          disabled={loading || testing}
          className="btn-filled min-w-[120px]"
        >
          {loading ? (
            <>
              <ClientIcon icon={Loader2} className="mr-2 h-4 w-4 animate-spin" />
              获取中...
            </>
          ) : (
            <>
              <ClientIcon icon={Download} className="mr-2 h-4 w-4" />
              获取配置
            </>
          )}
        </Button>
      </div>

      {nodes.length > 0 && (
        <div className="flex items-center component-gap">
          <span className="badge-standard">总节点数: {nodes.length}</span>
          <span className="badge-standard">符合条件: {filteredNodes.length}</span>
          {testing && (
            <span className="badge-standard bg-lavender-600 text-lavender-50">测试中...</span>
          )}
        </div>
      )}
    </Card>
  )
}