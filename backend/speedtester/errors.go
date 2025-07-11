package speedtester

import (
	"fmt"
	"strings"

	"github.com/metacubex/mihomo/constant"
)

// VlessTestError represents vless test error details
type VlessTestError struct {
	Stage     string `json:"stage"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	ProxyName string `json:"proxy_name"`
}

// Error stage constants
const (
	StageValidation = "validation"
	StageDNS        = "dns"
	StageConnect    = "connect"
	StageHandshake  = "handshake"
	StageTransfer   = "transfer"
)

// Error code constants
const (
	ErrorInvalidConfig     = "INVALID_CONFIG"
	ErrorDNSResolution     = "DNS_RESOLUTION_FAILED"
	ErrorConnectionRefused = "CONNECTION_REFUSED"
	ErrorConnectionTimeout = "CONNECTION_TIMEOUT"
	ErrorHandshakeTimeout  = "HANDSHAKE_TIMEOUT"
	ErrorProtocolError     = "PROTOCOL_ERROR"
	ErrorAuthFailed        = "AUTHENTICATION_FAILED"
	ErrorTransferTimeout   = "TRANSFER_TIMEOUT"
	ErrorUnknown           = "UNKNOWN_ERROR"
)

// NewVlessTestError creates a new vless test error
func NewVlessTestError(stage, code, message, proxyName string) *VlessTestError {
	return &VlessTestError{
		Stage:     stage,
		Code:      code,
		Message:   message,
		ProxyName: proxyName,
	}
}

// Error implements error interface
func (e *VlessTestError) Error() string {
	return fmt.Sprintf("[%s:%s] %s - %s", e.Stage, e.Code, e.ProxyName, e.Message)
}

// AnalyzeError analyzes error and returns detailed VlessTestError
func AnalyzeError(err error, proxyName string, defaultStage string) *VlessTestError {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	errLower := strings.ToLower(errStr)

	switch {
	case strings.Contains(errLower, "dns"):
		return NewVlessTestError(StageDNS, ErrorDNSResolution, errStr, proxyName)
	case strings.Contains(errLower, "connection refused"):
		return NewVlessTestError(StageConnect, ErrorConnectionRefused, errStr, proxyName)
	case strings.Contains(errLower, "timeout") && strings.Contains(errLower, "connect"):
		return NewVlessTestError(StageConnect, ErrorConnectionTimeout, errStr, proxyName)
	case strings.Contains(errLower, "handshake"):
		return NewVlessTestError(StageHandshake, ErrorHandshakeTimeout, errStr, proxyName)
	case strings.Contains(errLower, "protocol"):
		return NewVlessTestError(StageHandshake, ErrorProtocolError, errStr, proxyName)
	case strings.Contains(errLower, "auth"):
		return NewVlessTestError(StageHandshake, ErrorAuthFailed, errStr, proxyName)
	case strings.Contains(errLower, "timeout"):
		return NewVlessTestError(StageTransfer, ErrorTransferTimeout, errStr, proxyName)
	default:
		return NewVlessTestError(defaultStage, ErrorUnknown, errStr, proxyName)
	}
}

// IsVlessProtocol checks if proxy type is VLESS
func IsVlessProtocol(proxyType constant.AdapterType) bool {
	return proxyType == constant.Vless
}
