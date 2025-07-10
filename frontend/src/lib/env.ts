import { z } from 'zod'

// 环境变量验证 schema
const envSchema = z.object({
  VITE_API_URL: z.string().url().min(1, "API URL is required"),
  VITE_WS_URL: z.string().min(1, "WebSocket URL is required"),
  NODE_ENV: z.enum(['development', 'production', 'test']).default('development'),
})

// 获取环境变量（支持客户端和服务端）
const getEnvVars = () => {
  // 在客户端，使用 import.meta.env
  if (typeof window !== 'undefined') {
    return import.meta.env
  }
  // 在服务端，也使用 import.meta.env（Astro 统一处理）
  return import.meta.env
}

// 验证环境变量
const validateEnv = () => {
  try {
    const envVars = getEnvVars()
    
    // 为缺失的环境变量提供默认值
    const processedEnv = {
      VITE_API_URL: envVars.VITE_API_URL || 'http://localhost:8080',
      VITE_WS_URL: envVars.VITE_WS_URL || 'ws://localhost:8080',
      NODE_ENV: envVars.NODE_ENV || 'development',
    }
    
    return envSchema.parse(processedEnv)
  } catch (error) {
    if (error instanceof z.ZodError) {
      console.error('Environment validation failed:', error.errors)
      // 在开发环境中，使用默认值而不是抛出错误
      return {
        VITE_API_URL: 'http://localhost:8080',
        VITE_WS_URL: 'ws://localhost:8080',
        NODE_ENV: 'development' as const,
      }
    }
    throw error
  }
}

// 导出验证后的环境变量
export const env = validateEnv()

// 导出常用的配置
export const config = {
  apiUrl: env.VITE_API_URL,
  wsUrl: env.VITE_WS_URL,
  isDev: env.NODE_ENV === 'development',
  isProd: env.NODE_ENV === 'production',
  isTest: env.NODE_ENV === 'test',
}

// 导出类型
export type Env = z.infer<typeof envSchema>