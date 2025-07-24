package tasks

import (
	"context"
	"time"

	"github.com/zhsama/clash-speedtest/server/common"
	"github.com/zhsama/clash-speedtest/speedtester"
)

// TaskStatus 任务状态枚举
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusCancelled TaskStatus = "cancelled"
	TaskStatusFailed    TaskStatus = "failed"
)

// TaskType 任务类型枚举
type TaskType string

const (
	TaskTypeSpeedTest TaskType = "speed_test"
	TaskTypeUnlock    TaskType = "unlock_test"
	TaskTypeBoth      TaskType = "both"
)

// Task 任务接口
type Task interface {
	ID() string
	Type() TaskType
	Status() TaskStatus
	Config() *common.TestRequest
	Context() context.Context
	Start() error
	Cancel() error
	Results() []*speedtester.Result
	Progress() *TaskProgress
	Error() error
}

// TaskProgress 任务进度信息
type TaskProgress struct {
	CompletedCount  int     `json:"completed_count"`
	TotalCount      int     `json:"total_count"`
	ProgressPercent float64 `json:"progress_percent"`
	CurrentProxy    string  `json:"current_proxy"`
	Status          string  `json:"status"`
	StartTime       time.Time `json:"start_time"`
	Duration        time.Duration `json:"duration"`
}

// TaskResult 任务结果
type TaskResult struct {
	TaskID        string                `json:"task_id"`
	Status        TaskStatus            `json:"status"`
	Results       []*speedtester.Result `json:"results"`
	Error         error                 `json:"error,omitempty"`
	Progress      *TaskProgress         `json:"progress"`
	StartTime     time.Time             `json:"start_time"`
	CompleteTime  time.Time             `json:"complete_time"`
	Duration      time.Duration         `json:"duration"`
	SuccessCount  int                   `json:"success_count"`
	FailedCount   int                   `json:"failed_count"`
}

// TaskEvent 任务事件
type TaskEvent struct {
	TaskID    string      `json:"task_id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// TaskEventType 任务事件类型
const (
	TaskEventTypeStarted    = "started"
	TaskEventTypeProgress   = "progress"
	TaskEventTypeCompleted  = "completed"
	TaskEventTypeCancelled  = "cancelled"
	TaskEventTypeFailed     = "failed"
	TaskEventTypeResult     = "result"
)

// TaskOptions 任务选项
type TaskOptions struct {
	Timeout    time.Duration
	RetryCount int
	Async      bool
}

// DefaultTaskOptions 默认任务选项
func DefaultTaskOptions() *TaskOptions {
	return &TaskOptions{
		Timeout:    10 * time.Minute,
		RetryCount: 0,
		Async:      true,
	}
}