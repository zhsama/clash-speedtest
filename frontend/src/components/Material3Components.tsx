import React from "react"
import { cn } from "@/lib/utils"

/* ============================================================================
   Material 3 组件库
   基于 Material Design 3.0 的 React 组件集合
   ============================================================================ */

// ============================================================================
// 按钮组件
// ============================================================================

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "filled" | "outlined" | "text"
  size?: "sm" | "md" | "lg"
  children: React.ReactNode
}

export const MaterialButton = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "filled", size = "md", children, ...props }, ref) => {
    const baseStyles =
      "inline-flex items-center justify-center gap-2 font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none"

    const variants = {
      filled: "btn-filled",
      outlined: "btn-outlined",
      text: "btn-text",
    }

    const sizes = {
      sm: "px-3 py-1.5 text-sm",
      md: "px-6 py-2",
      lg: "px-8 py-3 text-lg",
    }

    return (
      <button
        ref={ref}
        className={cn(baseStyles, variants[variant], sizes[size], className)}
        {...props}
      >
        {children}
      </button>
    )
  }
)
MaterialButton.displayName = "MaterialButton"

// ============================================================================
// 卡片组件
// ============================================================================

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "elevated" | "filled" | "outlined"
  children: React.ReactNode
}

export const MaterialCard = React.forwardRef<HTMLDivElement, CardProps>(
  ({ className, variant = "elevated", children, ...props }, ref) => {
    const variants = {
      elevated: "card-elevated",
      filled: "card-filled",
      outlined: "card-outlined",
    }

    return (
      <div ref={ref} className={cn(variants[variant], className)} {...props}>
        {children}
      </div>
    )
  }
)
MaterialCard.displayName = "MaterialCard"

// ============================================================================
// 输入组件
// ============================================================================

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  variant?: "filled" | "outlined"
}

export const MaterialInput = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, variant = "outlined", ...props }, ref) => {
    const variants = {
      filled: "input-filled",
      outlined: "input-outlined",
    }

    return <input ref={ref} className={cn(variants[variant], className)} {...props} />
  }
)
MaterialInput.displayName = "MaterialInput"

// ============================================================================
// 徽章组件
// ============================================================================

interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: "filled" | "outlined"
  children: React.ReactNode
}

export const MaterialBadge = React.forwardRef<HTMLSpanElement, BadgeProps>(
  ({ className, variant = "filled", children, ...props }, ref) => {
    const variants = {
      filled: "badge-filled",
      outlined: "badge-outlined",
    }

    return (
      <span ref={ref} className={cn(variants[variant], className)} {...props}>
        {children}
      </span>
    )
  }
)
MaterialBadge.displayName = "MaterialBadge"

// ============================================================================
// 状态指示器组件
// ============================================================================

interface StatusIndicatorProps extends React.HTMLAttributes<HTMLDivElement> {
  status: "success" | "warning" | "error" | "info"
  children: React.ReactNode
}

export const MaterialStatusIndicator = React.forwardRef<HTMLDivElement, StatusIndicatorProps>(
  ({ className, status, children, ...props }, ref) => {
    const statusStyles = {
      success: "status-success",
      warning: "status-warning",
      error: "status-error",
      info: "status-info",
    }

    return (
      <div ref={ref} className={cn("status-indicator", statusStyles[status], className)} {...props}>
        <div className="status-dot" />
        {children}
      </div>
    )
  }
)
MaterialStatusIndicator.displayName = "MaterialStatusIndicator"

// ============================================================================
// 进度指示器组件
// ============================================================================

interface ProgressProps extends React.HTMLAttributes<HTMLDivElement> {
  value: number
  max?: number
}

export const MaterialProgress = React.forwardRef<HTMLDivElement, ProgressProps>(
  ({ className, value, max = 100, ...props }, ref) => {
    const percentage = (Math.min(Math.max(value, 0), max) / max) * 100

    return (
      <div ref={ref} className={cn("progress-linear", className)} {...props}>
        <div className="progress-linear-indicator" style={{ width: `${percentage}%` }} />
      </div>
    )
  }
)
MaterialProgress.displayName = "MaterialProgress"

// ============================================================================
// 表格组件
// ============================================================================

interface TableProps extends React.TableHTMLAttributes<HTMLTableElement> {
  children: React.ReactNode
}

export const MaterialTable = React.forwardRef<HTMLTableElement, TableProps>(
  ({ className, children, ...props }, ref) => {
    return (
      <div className="table-container scrollbar-modern">
        <table ref={ref} className={cn("table-modern", className)} {...props}>
          {children}
        </table>
      </div>
    )
  }
)
MaterialTable.displayName = "MaterialTable"

// ============================================================================
// 速度测试专用组件
// ============================================================================

interface SpeedBadgeProps {
  speed: number
  unit?: "Mbps" | "MB/s"
  className?: string
}

export const SpeedBadge: React.FC<SpeedBadgeProps> = ({ speed, unit = "Mbps", className }) => {
  const getSpeedClass = (speed: number) => {
    if (speed >= 100) return "speed-excellent"
    if (speed >= 50) return "speed-good"
    if (speed >= 20) return "speed-average"
    if (speed >= 5) return "speed-poor"
    return "speed-bad"
  }

  return (
    <MaterialBadge className={cn(getSpeedClass(speed), className)}>
      {speed.toFixed(1)} {unit}
    </MaterialBadge>
  )
}

interface LatencyBadgeProps {
  latency: number
  className?: string
}

export const LatencyBadge: React.FC<LatencyBadgeProps> = ({ latency, className }) => {
  const getLatencyClass = (latency: number) => {
    if (latency <= 50) return "latency-low"
    if (latency <= 200) return "latency-medium"
    return "latency-high"
  }

  return (
    <MaterialBadge className={cn(getLatencyClass(latency), className)}>{latency}ms</MaterialBadge>
  )
}

interface ProtocolBadgeProps {
  protocol: string
  className?: string
}

export const ProtocolBadge: React.FC<ProtocolBadgeProps> = ({ protocol, className }) => {
  const protocolClass = `protocol-${protocol.toLowerCase()}`

  return (
    <MaterialBadge className={cn(protocolClass, className)}>{protocol.toUpperCase()}</MaterialBadge>
  )
}

// ============================================================================
// 动画包装器组件
// ============================================================================

interface AnimatedContainerProps extends React.HTMLAttributes<HTMLDivElement> {
  animation?: "fade-in" | "slide-up" | "pulse-gentle"
  children: React.ReactNode
}

export const AnimatedContainer = React.forwardRef<HTMLDivElement, AnimatedContainerProps>(
  ({ className, animation = "fade-in", children, ...props }, ref) => {
    const animations = {
      "fade-in": "animate-fade-in",
      "slide-up": "animate-slide-up",
      "pulse-gentle": "animate-pulse-gentle",
    }

    return (
      <div ref={ref} className={cn(animations[animation], className)} {...props}>
        {children}
      </div>
    )
  }
)
AnimatedContainer.displayName = "AnimatedContainer"

// ============================================================================
// Surface 组件
// ============================================================================

interface SurfaceProps extends React.HTMLAttributes<HTMLDivElement> {
  level?: "lowest" | "low" | "container" | "high" | "highest"
  children: React.ReactNode
}

export const MaterialSurface = React.forwardRef<HTMLDivElement, SurfaceProps>(
  ({ className, level = "container", children, ...props }, ref) => {
    const levels = {
      lowest: "surface-container-lowest",
      low: "surface-container-low",
      container: "surface-container",
      high: "surface-container-high",
      highest: "surface-container-highest",
    }

    return (
      <div ref={ref} className={cn(levels[level], className)} {...props}>
        {children}
      </div>
    )
  }
)
MaterialSurface.displayName = "MaterialSurface"

// ============================================================================
// 类型导出
// ============================================================================

export type {
  ButtonProps,
  CardProps,
  InputProps,
  BadgeProps,
  StatusIndicatorProps,
  ProgressProps,
  TableProps,
  SpeedBadgeProps,
  LatencyBadgeProps,
  ProtocolBadgeProps,
  AnimatedContainerProps,
  SurfaceProps,
}
