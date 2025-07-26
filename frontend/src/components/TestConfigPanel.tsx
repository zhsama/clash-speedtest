import { useEffect, useState } from "react"
import {
  FaDownload as Download,
  FaFilter as Filter,
  FaGlobe as Globe,
  FaPlay as Play,
  FaSpinner as Loader2,
} from "react-icons/fa"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Slider } from "@/components/ui/slider"
import { config } from "@/lib/env"
import ClientIcon from "./ClientIcon"

interface FilterConfig {
  includeNodes: string[]
  excludeNodes: string[]
  protocolFilter: string[]
  minDownloadSpeed: number
  minUploadSpeed: number
  maxLatency: number
  stashCompatible: boolean
}

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

interface TestConfigPanelProps {
  testConfig: TestConfig
  setTestConfig: React.Dispatch<React.SetStateAction<TestConfig>>
  filterConfig: FilterConfig
  setFilterConfig: React.Dispatch<React.SetStateAction<FilterConfig>>
  testing: boolean
  startTest: () => void
  stopTest: () => void
  isConnected: boolean
  nodes: any[]
  loading: boolean
}

// 速度测试配置子组件
const SpeedTestConfig = ({
  testConfig,
  setTestConfig,
  filterConfig,
  setFilterConfig,
}: {
  testConfig: TestConfig
  setTestConfig: React.Dispatch<React.SetStateAction<TestConfig>>
  filterConfig: FilterConfig
  setFilterConfig: React.Dispatch<React.SetStateAction<FilterConfig>>
}) => (
  <div className="form-element">
    <h4 className="text-lg font-semibold text-lavender-50 flex items-center gap-2 mb-2">
      <ClientIcon icon={Download} className="h-5 w-5 text-lavender-400" />
      服务器测速配置
    </h4>
    <div className="space-y-2">
      <div>
        <label htmlFor="server-url" className="form-element-label">
          测试服务器
        </label>
        <Input
          id="server-url"
          value={testConfig.serverUrl}
          onChange={(e) =>
            setTestConfig((prev) => ({
              ...prev,
              serverUrl: e.target.value,
            }))
          }
          className="input-outlined"
        />
      </div>

      <div>
        <div className="form-element-label">测试包大小: {testConfig.downloadSize} MB</div>
        <Slider
          value={[testConfig.downloadSize]}
          onValueChange={(v) =>
            setTestConfig((prev) => ({
              ...prev,
              downloadSize: v[0],
              uploadSize: v[0],
            }))
          }
          max={100}
          min={10}
          step={10}
          className="slider-dark"
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <div className="form-element-label">并发数: {testConfig.concurrent}</div>
          <Slider
            value={[testConfig.concurrent]}
            onValueChange={(v) =>
              setTestConfig((prev) => ({
                ...prev,
                concurrent: v[0],
              }))
            }
            max={16}
            min={1}
            step={1}
            className="slider-dark"
          />
        </div>

        <div>
          <div className="form-element-label">超时时间: {testConfig.timeout} 秒</div>
          <Slider
            value={[testConfig.timeout]}
            onValueChange={(v) =>
              setTestConfig((prev) => ({
                ...prev,
                timeout: v[0],
              }))
            }
            max={30}
            min={5}
            step={5}
            className="slider-dark"
          />
        </div>
      </div>
    </div>

    {/* 速度过滤条件 */}
    <div className="border-t border-lavender-600 pt-4 mt-4">
      <h4 className="text-lg font-semibold text-lavender-50 flex items-center gap-2 mb-2">
        <ClientIcon icon={Filter} className="h-4 w-4 text-lavender-400" />
        速度过滤条件
      </h4>
      <div className="space-y-2">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <div className="form-element-label">
              最低下载速度: {filterConfig.minDownloadSpeed} MB/s
            </div>
            <Slider
              value={[filterConfig.minDownloadSpeed]}
              onValueChange={(v) =>
                setFilterConfig((prev) => ({
                  ...prev,
                  minDownloadSpeed: v[0],
                }))
              }
              max={100}
              min={0}
              step={1}
              className="slider-dark"
            />
          </div>

          <div>
            <div className="form-element-label">
              最低上传速度: {filterConfig.minUploadSpeed} MB/s
            </div>
            <Slider
              value={[filterConfig.minUploadSpeed]}
              onValueChange={(v) =>
                setFilterConfig((prev) => ({
                  ...prev,
                  minUploadSpeed: v[0],
                }))
              }
              max={50}
              min={0}
              step={1}
              className="slider-dark"
            />
          </div>
        </div>

        <div>
          <div className="form-element-label">最大延迟: {filterConfig.maxLatency} ms</div>
          <Slider
            value={[filterConfig.maxLatency]}
            onValueChange={(v) =>
              setFilterConfig((prev) => ({
                ...prev,
                maxLatency: v[0],
              }))
            }
            max={5000}
            min={100}
            step={100}
            className="slider-dark"
          />
        </div>
      </div>
    </div>
  </div>
)

// 解锁检测配置子组件
const UnlockTestConfig = ({
  testConfig,
  setTestConfig,
  hasSpeedConfig,
}: {
  testConfig: TestConfig
  setTestConfig: React.Dispatch<React.SetStateAction<TestConfig>>
  hasSpeedConfig: boolean
}) => {
  const [availablePlatforms, setAvailablePlatforms] = useState<string[]>([])
  const [platformsLoading, setPlatformsLoading] = useState(false)

  useEffect(() => {
    const fetchUnlockPlatforms = async () => {
      setPlatformsLoading(true)
      let platforms = ["Netflix", "YouTube", "Disney+", "ChatGPT", "Spotify", "Bilibili"]
      try {
        const response = await fetch(`${config.apiUrl}/api/unlock/platforms`)
        const data = await response.json()

        if (data.success && data.data && data.data.platforms) {
          platforms = data.data.platforms
            .map((platform: any) => platform.display_name || platform.name)
            .sort((a: string, b: string) => a.localeCompare(b))
          setAvailablePlatforms(platforms)
        } else {
          setAvailablePlatforms(platforms.sort((a, b) => a.localeCompare(b)))
          console.warn("Failed to fetch unlock platforms, using defaults")
        }
      } catch (error) {
        console.error("Error fetching unlock platforms:", error)
        setAvailablePlatforms(platforms.sort((a, b) => a.localeCompare(b)))
      } finally {
        setPlatformsLoading(false)
      }
    }

    fetchUnlockPlatforms()
  }, [])

  return (
    <div className={`form-element ${hasSpeedConfig ? "border-t border-lavender-700 pt-4" : ""}`}>
      <h4 className="text-lg font-semibold text-lavender-50 flex items-center gap-2 mb-2">
        <ClientIcon icon={Globe} className="h-5 w-5 text-lavender-400" />
        流媒体解锁检测
      </h4>

      <div className="space-y-2">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <div className="form-element-label">解锁检测并发数: {testConfig.unlockConcurrent}</div>
            <Slider
              value={[testConfig.unlockConcurrent]}
              onValueChange={(v) =>
                setTestConfig((prev) => ({
                  ...prev,
                  unlockConcurrent: v[0],
                }))
              }
              max={10}
              min={1}
              step={1}
              className="slider-dark"
            />
          </div>

          <div>
            <div className="form-element-label">解锁检测超时: {testConfig.unlockTimeout} 秒</div>
            <Slider
              value={[testConfig.unlockTimeout]}
              onValueChange={(v) =>
                setTestConfig((prev) => ({
                  ...prev,
                  unlockTimeout: v[0],
                }))
              }
              max={30}
              min={5}
              step={5}
              className="slider-dark"
            />
          </div>
        </div>
      </div>

      <div className="form-element">
        <div className="form-element-label">
          检测平台{" "}
          {platformsLoading && <span className="text-xs text-lavender-400">(加载中...)</span>}
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 component-gap">
          {availablePlatforms.map((platform) => (
            <label
              key={platform}
              htmlFor={`platform-${platform}`}
              className="flex items-center gap-2 cursor-pointer min-w-0"
            >
              <Checkbox
                id={`platform-${platform}`}
                checked={testConfig.unlockPlatforms.includes(platform)}
                onCheckedChange={(checked) => {
                  setTestConfig((prev) => ({
                    ...prev,
                    unlockPlatforms: checked
                      ? [...prev.unlockPlatforms, platform]
                      : prev.unlockPlatforms.filter((p) => p !== platform),
                  }))
                }}
                className="checkbox-dark"
              />
              <span className="text-lavender-100 text-sm truncate">{platform}</span>
            </label>
          ))}
        </div>
        {availablePlatforms.length === 0 && !platformsLoading && (
          <div className="text-center text-lavender-400 py-4">无可用的解锁检测平台</div>
        )}
      </div>
    </div>
  )
}

export default function TestConfigPanel({
  testConfig,
  setTestConfig,
  filterConfig,
  setFilterConfig,
  testing,
  startTest,
  stopTest,
  isConnected,
  nodes,
  loading,
}: TestConfigPanelProps) {
  return (
    <div className="space-y-4">
      {/* 基础测试配置 */}
      <Card className="card-elevated">
        <div className="flex items-center gap-2 mb-2">
          <ClientIcon icon={Filter} className="h-5 w-5 text-lavender-400" />
          <h2 className="text-lg font-semibold text-lavender-50">测试配置</h2>
        </div>

        {/* 测试模式选择器 */}
        <div className="space-y-2">
          <label htmlFor="test-mode" className="form-element-label">
            测试模式
          </label>
          <Select
            value={testConfig.testMode}
            onValueChange={(value) =>
              setTestConfig((prev) => ({
                ...prev,
                testMode: value,
              }))
            }
          >
            <SelectTrigger id="test-mode" className="w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="both">全面测试（测速+解锁）</SelectItem>
              <SelectItem value="speed_only">仅测速</SelectItem>
              <SelectItem value="unlock_only">仅解锁检测</SelectItem>
            </SelectContent>
          </Select>
          <p className="text-sm text-lavender-400 mt-2">
            {testConfig.testMode === "both" && "同时进行速度测试和流媒体解锁检测"}
            {testConfig.testMode === "speed_only" && "只进行网络速度测试，跳过解锁检测"}
            {testConfig.testMode === "unlock_only" && "只进行流媒体解锁检测，跳过速度测试"}
          </p>
        </div>

        {/* 启动测试按钮 */}
        <Button
          onClick={testing ? stopTest : startTest}
          disabled={!isConnected || nodes.length === 0 || loading}
          size="lg"
          className={`w-full ${testing ? "bg-red-500 hover:bg-red-600 text-white font-medium transition-colors duration-200" : "btn-filled"}`}
        >
          {testing ? (
            <>
              <ClientIcon icon={Loader2} className="mr-2 h-4 w-4 animate-spin" />
              停止测试
            </>
          ) : (
            <>
              <ClientIcon icon={Play} className="mr-2 h-4 w-4" />
              开始测试
            </>
          )}
        </Button>
      </Card>

      {/* 高级配置 */}
      <details>
        <summary className="cursor-pointer text-lavender-300 hover:text-lavender-100 transition-colors">
          高级测试配置
        </summary>
        <Card className="card-elevated mt-4">
          {/* 速度测试配置 - 条件显示 */}
          {(testConfig.testMode === "both" || testConfig.testMode === "speed_only") && (
            <SpeedTestConfig
              testConfig={testConfig}
              setTestConfig={setTestConfig}
              filterConfig={filterConfig}
              setFilterConfig={setFilterConfig}
            />
          )}

          {/* 解锁检测配置 - 条件显示 */}
          {(testConfig.testMode === "both" || testConfig.testMode === "unlock_only") && (
            <UnlockTestConfig
              testConfig={testConfig}
              setTestConfig={setTestConfig}
              hasSpeedConfig={testConfig.testMode === "both"}
            />
          )}
        </Card>
      </details>
    </div>
  )
}