package middleware

import (
	"net/http"
)

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// DefaultCORSConfig 默认 CORS 配置
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With"},
	}
}

// CORS 创建 CORS 中间件
func CORS(config *CORSConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultCORSConfig()
	}
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// 检查是否允许该来源
			allowed := false
			for _, allowedOrigin := range config.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			
			// 设置允许的方法
			allowedMethods := ""
			for i, method := range config.AllowedMethods {
				if i > 0 {
					allowedMethods += ", "
				}
				allowedMethods += method
			}
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			
			// 设置允许的头部
			allowedHeaders := ""
			for i, header := range config.AllowedHeaders {
				if i > 0 {
					allowedHeaders += ", "
				}
				allowedHeaders += header
			}
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			
			// 处理 OPTIONS 请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}