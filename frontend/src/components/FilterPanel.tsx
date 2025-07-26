import { FaFilter as Filter, FaSync as RefreshCw } from "react-icons/fa"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Switch } from "@/components/ui/switch"
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

interface FilterPanelProps {
  filterConfig: FilterConfig
  setFilterConfig: (config: FilterConfig | ((prev: FilterConfig) => FilterConfig)) => void
  availableProtocols: string[]
  includeNodesInput: string
  setIncludeNodesInput: (value: string) => void
  excludeNodesInput: string
  setExcludeNodesInput: (value: string) => void
  handleIncludeNodesChange: (value: string) => void
  handleExcludeNodesChange: (value: string) => void
  handleProtocolFilterChange: (protocol: string, checked: boolean) => void
  isProtocolSelected: (protocol: string) => boolean
  applyFilters: () => void
  testing: boolean
}

export default function FilterPanel({
  filterConfig,
  setFilterConfig,
  availableProtocols,
  includeNodesInput,
  setIncludeNodesInput,
  excludeNodesInput,
  setExcludeNodesInput,
  handleIncludeNodesChange,
  handleExcludeNodesChange,
  handleProtocolFilterChange,
  isProtocolSelected,
  applyFilters,
  testing,
}: FilterPanelProps) {
  return (
    <Card className="card-elevated">
      <div className="flex items-center justify-between form-element">
        <h4 className="text-lg font-semibold text-lavender-50 flex items-center gap-2">
          <ClientIcon icon={Filter} className="h-4 w-4 text-lavender-400" />
          过滤条件
        </h4>
        <Button
          onClick={applyFilters}
          variant="outline"
          size="sm"
          className="btn-outlined"
          disabled={testing}
        >
          <ClientIcon icon={RefreshCw} className="h-4 w-4 mr-1" />
          刷新过滤
        </Button>
      </div>

      <div className="space-y-2">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label htmlFor="include-nodes" className="form-element-label">
              包含节点 (逗号分隔)
            </label>
            <Input
              id="include-nodes"
              placeholder="例如: 香港, HK, 新加坡..."
              value={includeNodesInput}
              onChange={(e) => {
                setIncludeNodesInput(e.target.value)
                handleIncludeNodesChange(e.target.value)
              }}
              className="input-outlined"
            />
          </div>

          <div>
            <label htmlFor="exclude-nodes" className="form-element-label">
              排除节点 (逗号分隔)
            </label>
            <Input
              id="exclude-nodes"
              placeholder="例如: 过期, 测试, 备用..."
              value={excludeNodesInput}
              onChange={(e) => {
                setExcludeNodesInput(e.target.value)
                handleExcludeNodesChange(e.target.value)
              }}
              className="input-outlined"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {availableProtocols.length > 0 && (
            <div>
              <div className="form-element-label">协议过滤</div>
              <div className="grid grid-cols-1 sm:grid-cols-2 component-gap">
                {availableProtocols.map((protocol) => (
                  <div key={protocol} className="flex items-center gap-2 min-w-0">
                    <Checkbox
                      id={`protocol-${protocol}`}
                      checked={isProtocolSelected(protocol)}
                      onCheckedChange={(checked: boolean) =>
                        handleProtocolFilterChange(protocol, checked)
                      }
                      className="checkbox-dark"
                    />
                    <label
                      htmlFor={`protocol-${protocol}`}
                      className="text-sm text-lavender-100 cursor-pointer truncate"
                    >
                      {protocol}
                    </label>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div>
            <div className="form-element-label">其他选项</div>
            <div className="flex items-center gap-2">
              <Switch
                id="stashCompatible"
                checked={filterConfig.stashCompatible}
                onCheckedChange={(checked) =>
                  setFilterConfig((prev) => ({
                    ...prev,
                    stashCompatible: checked,
                  }))
                }
                className="switch-dark"
              />
              <label htmlFor="stashCompatible" className="text-lavender-100">
                Stash 兼容模式
              </label>
            </div>
          </div>
        </div>
      </div>
    </Card>
  )
}