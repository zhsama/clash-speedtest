package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/faceair/clash-speedtest/logger"
	"github.com/faceair/clash-speedtest/server/response"
	"github.com/faceair/clash-speedtest/speedtester"
	"github.com/faceair/clash-speedtest/utils/export"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	*Handler
}

// NewConfigHandler 创建新的配置处理器
func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{
		Handler: NewHandler(),
	}
}

// HandleGetProtocols 处理获取协议列表请求
func (h *ConfigHandler) HandleGetProtocols(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if r.Method != http.MethodPost {
		h.handleMethodNotAllowed(ctx, w, r, "POST")
		return
	}
	
	logger.Logger.InfoContext(ctx, "Get protocols request received")
	
	var req struct {
		ConfigPaths string `json:"configPaths"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.ErrorContext(ctx, "Failed to decode request body", 
			slog.String("error", err.Error()))
		response.SendError(ctx, w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	
	if req.ConfigPaths == "" {
		response.SendError(ctx, w, http.StatusBadRequest, "Config paths are required")
		return
	}
	
	// 创建速度测试器
	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths: req.ConfigPaths,
		FilterRegex: ".+",
	})
	
	logger.Logger.InfoContext(ctx, "Loading proxies for protocol discovery", 
		slog.String("config_paths", req.ConfigPaths))
	
	allProxies, err := speedTester.LoadProxies(false)
	if err != nil {
		logger.Logger.ErrorContext(ctx, "Failed to load proxies", 
			slog.String("error", err.Error()),
			slog.String("config_paths", req.ConfigPaths))
		response.SendError(ctx, w, http.StatusBadRequest, "Failed to load proxies: "+err.Error())
		return
	}
	
	protocols := speedTester.GetAvailableProtocols(allProxies)
	logger.Logger.InfoContext(ctx, "Protocols discovered", 
		slog.Int("protocol_count", len(protocols)))
	
	response.SendJSON(ctx, w, http.StatusOK, response.ProtocolsResponse{
		Success:   true,
		Protocols: protocols,
	})
}

// HandleGetNodes 处理获取节点列表请求
func (h *ConfigHandler) HandleGetNodes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if r.Method != http.MethodPost {
		h.handleMethodNotAllowed(ctx, w, r, "POST")
		return
	}
	
	logger.Logger.InfoContext(ctx, "Get nodes request received")
	
	var req struct {
		ConfigPaths     string   `json:"configPaths"`
		IncludeNodes    []string `json:"includeNodes"`
		ExcludeNodes    []string `json:"excludeNodes"`
		ProtocolFilter  []string `json:"protocolFilter"`
		StashCompatible bool     `json:"stashCompatible"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.ErrorContext(ctx, "Failed to decode request body", 
			slog.String("error", err.Error()))
		response.SendError(ctx, w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	
	// 创建速度测试器
	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths:    req.ConfigPaths,
		FilterRegex:    ".+",
		IncludeNodes:   req.IncludeNodes,
		ExcludeNodes:   req.ExcludeNodes,
		ProtocolFilter: req.ProtocolFilter,
	})
	
	logger.Logger.InfoContext(ctx, "Loading nodes", 
		slog.String("config_paths", req.ConfigPaths))
	
	allProxies, err := speedTester.LoadProxies(req.StashCompatible)
	if err != nil {
		logger.Logger.ErrorContext(ctx, "Failed to load proxies", 
			slog.String("error", err.Error()),
			slog.String("config_paths", req.ConfigPaths))
		response.SendError(ctx, w, http.StatusBadRequest, "Failed to load proxies: "+err.Error())
		return
	}
	
	// 转换代理信息为节点信息
	nodes := make([]response.NodeInfo, 0, len(allProxies))
	for name, proxy := range allProxies {
		nodeInfo := response.NodeInfo{
			Name: name,
			Type: proxy.Type().String(),
		}
		
		// 从配置中提取服务器和端口信息
		if server, ok := proxy.Config["server"]; ok {
			nodeInfo.Server = server.(string)
		}
		if port, ok := proxy.Config["port"]; ok {
			switch p := port.(type) {
			case int:
				nodeInfo.Port = p
			case float64:
				nodeInfo.Port = int(p)
			}
		}
		
		nodes = append(nodes, nodeInfo)
	}
	
	logger.Logger.InfoContext(ctx, "Nodes loaded successfully", 
		slog.Int("node_count", len(nodes)))
	
	response.SendJSON(ctx, w, http.StatusOK, response.NodesResponse{
		Success: true,
		Nodes:   nodes,
	})
}

// HandleExportResults 处理导出结果请求
func (h *ConfigHandler) HandleExportResults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if r.Method != http.MethodPost {
		h.handleMethodNotAllowed(ctx, w, r, "POST")
		return
	}
	
	var exportReq struct {
		TaskID  string                 `json:"taskId"`
		Options export.ExportOptions `json:"options"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&exportReq); err != nil {
		response.SendError(ctx, w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	
	// 验证导出选项
	if err := export.ValidateExportOptions(exportReq.Options); err != nil {
		response.SendError(ctx, w, http.StatusBadRequest, "Invalid export options: "+err.Error())
		return
	}
	
	// TODO: 实现实际的导出逻辑
	// 这里需要从任务管理器中获取测试结果，然后进行导出
	
	response.SendJSON(ctx, w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Export functionality will be implemented",
		"format":  exportReq.Options.Format,
		"path":    exportReq.Options.OutputPath,
	})
}