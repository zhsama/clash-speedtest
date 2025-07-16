package common

// TestRequest 表示测试请求的结构
type TestRequest struct {
	ConfigPaths      string   `json:"configPaths"`
	FilterRegex      string   `json:"filterRegex"`
	IncludeNodes     []string `json:"includeNodes"`
	ExcludeNodes     []string `json:"excludeNodes"`
	ProtocolFilter   []string `json:"protocolFilter"`
	ServerURL        string   `json:"serverUrl"`
	DownloadSize     int      `json:"downloadSize"`
	UploadSize       int      `json:"uploadSize"`
	Timeout          int      `json:"timeout"`
	Concurrent       int      `json:"concurrent"`
	MaxLatency       int      `json:"maxLatency"`
	MinDownloadSpeed float64  `json:"minDownloadSpeed"`
	MinUploadSpeed   float64  `json:"minUploadSpeed"`
	StashCompatible  bool     `json:"stashCompatible"`
	// 新增字段
	FastMode     bool   `json:"fastMode"`     // 快速模式：只测试延迟
	RenameNodes  bool   `json:"renameNodes"`  // 节点重命名：添加地理位置信息
	ExportFormat string `json:"exportFormat"` // 导出格式：json, csv, yaml, clash
	ExportPath   string `json:"exportPath"`   // 导出路径
	// 解锁检测相关字段
	TestMode         string   `json:"testMode"`         // 测试模式：speed_only, unlock_only, both
	UnlockEnabled    bool     `json:"unlockEnabled"`    // 是否启用解锁检测
	UnlockPlatforms  []string `json:"unlockPlatforms"`  // 要检测的平台列表
	UnlockConcurrent int      `json:"unlockConcurrent"` // 解锁检测并发数
	UnlockTimeout    int      `json:"unlockTimeout"`    // 解锁检测超时时间
	UnlockRetry      bool     `json:"unlockRetry"`      // 解锁检测失败时是否重试
}

// SetRequestDefaults 设置请求默认值
func SetRequestDefaults(req *TestRequest) {
	if req.FilterRegex == "" {
		req.FilterRegex = ".+"
	}
	if req.ServerURL == "" {
		req.ServerURL = "https://speed.cloudflare.com"
	}
	if req.DownloadSize == 0 {
		req.DownloadSize = 50
	}
	if req.UploadSize == 0 {
		req.UploadSize = 20
	}
	if req.Timeout == 0 {
		req.Timeout = 5
	}
	if req.Concurrent == 0 {
		req.Concurrent = 4
	}
	if req.MaxLatency == 0 {
		req.MaxLatency = 800
	}
	if req.TestMode == "" {
		req.TestMode = "speed_only"
	}
	if req.UnlockConcurrent == 0 {
		req.UnlockConcurrent = 5
	}
	if req.UnlockTimeout == 0 {
		req.UnlockTimeout = 10
	}
	if len(req.UnlockPlatforms) == 0 {
		req.UnlockPlatforms = []string{"Netflix", "YouTube", "Disney+", "ChatGPT", "Spotify", "Bilibili"}
	}
}

// ValidateRequest 验证请求参数
func ValidateRequest(req *TestRequest) error {
	if req.ConfigPaths == "" {
		return NewValidationError("config paths cannot be empty")
	}
	if req.Concurrent < 1 || req.Concurrent > 100 {
		return NewValidationError("concurrent must be between 1 and 100")
	}
	if req.Timeout < 1 || req.Timeout > 300 {
		return NewValidationError("timeout must be between 1 and 300 seconds")
	}
	if req.DownloadSize < 1 || req.DownloadSize > 1000 {
		return NewValidationError("download size must be between 1 and 1000 MB")
	}
	if req.UploadSize < 1 || req.UploadSize > 1000 {
		return NewValidationError("upload size must be between 1 and 1000 MB")
	}
	if req.MaxLatency < 10 || req.MaxLatency > 10000 {
		return NewValidationError("max latency must be between 10 and 10000 ms")
	}
	
	validTestModes := []string{"speed_only", "unlock_only", "both"}
	validMode := false
	for _, mode := range validTestModes {
		if req.TestMode == mode {
			validMode = true
			break
		}
	}
	if !validMode {
		return NewValidationError("test mode must be one of: speed_only, unlock_only, both")
	}
	
	return nil
}

// ValidationError 验证错误类型
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError 创建验证错误
func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}