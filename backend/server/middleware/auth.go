package middleware

import (
	"net/http"
	"strings"

	"github.com/faceair/clash-speedtest/server/response"
)

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled bool
	APIKey  string
	Tokens  []string
}

// Auth 认证中间件 (预留功能)
func Auth(config *AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 如果未启用认证，直接通过
			if config == nil || !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}
			
			// 从请求头获取认证信息
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.SendError(r.Context(), w, http.StatusUnauthorized, "Missing authorization header")
				return
			}
			
			// 检查 Bearer token
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if isValidToken(token, config.Tokens) {
					next.ServeHTTP(w, r)
					return
				}
			}
			
			// 检查 API Key
			if strings.HasPrefix(authHeader, "ApiKey ") {
				apiKey := strings.TrimPrefix(authHeader, "ApiKey ")
				if apiKey == config.APIKey {
					next.ServeHTTP(w, r)
					return
				}
			}
			
			// 也可以从查询参数获取 API Key
			queryAPIKey := r.URL.Query().Get("api_key")
			if queryAPIKey != "" && queryAPIKey == config.APIKey {
				next.ServeHTTP(w, r)
				return
			}
			
			// 认证失败
			response.SendError(r.Context(), w, http.StatusUnauthorized, "Invalid authentication credentials")
		})
	}
}

// isValidToken 检查 token 是否有效
func isValidToken(token string, validTokens []string) bool {
	for _, validToken := range validTokens {
		if token == validToken {
			return true
		}
	}
	return false
}

// RequireAuth 需要认证的处理器装饰器
func RequireAuth(config *AuthConfig, handler http.HandlerFunc) http.HandlerFunc {
	authMiddleware := Auth(config)
	return authMiddleware(handler).ServeHTTP
}