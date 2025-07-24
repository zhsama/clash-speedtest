package tasks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/zhsama/clash-speedtest/logger"
	"github.com/zhsama/clash-speedtest/server/common"
	"github.com/zhsama/clash-speedtest/websocket"
)

// Manager 任务管理器
type Manager struct {
	tasks     map[string]Task
	tasksMutex sync.RWMutex
	wsHub     *websocket.Hub
	executor  *Executor
	
	// 事件处理
	eventHandlers map[string][]TaskEventHandler
	eventMutex    sync.RWMutex
}

// TaskEventHandler 任务事件处理器
type TaskEventHandler func(event *TaskEvent)

// NewManager 创建新的任务管理器
func NewManager(wsHub *websocket.Hub) *Manager {
	manager := &Manager{
		tasks:         make(map[string]Task),
		wsHub:         wsHub,
		eventHandlers: make(map[string][]TaskEventHandler),
	}
	
	// 创建任务执行器
	manager.executor = NewExecutor(manager)
	
	return manager
}

// generateTaskID 生成任务 ID
func (m *Manager) generateTaskID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("task-%s-%d", hex.EncodeToString(bytes), time.Now().Unix())
}

// CreateTask 创建新任务
func (m *Manager) CreateTask(req *common.TestRequest, options *TaskOptions) (Task, error) {
	if options == nil {
		options = DefaultTaskOptions()
	}
	
	// 生成任务 ID
	taskID := m.generateTaskID()
	
	// 创建任务上下文
	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	
	// 创建任务
	task := &SpeedTestTask{
		id:         taskID,
		config:     req,
		ctx:        ctx,
		cancelFunc: cancel,
		status:     TaskStatusPending,
		options:    options,
		startTime:  time.Now(),
		progress:   &TaskProgress{},
		manager:    m,
	}
	
	// 添加到任务列表
	m.tasksMutex.Lock()
	m.tasks[taskID] = task
	m.tasksMutex.Unlock()
	
	logger.Logger.InfoContext(ctx, "Task created",
		slog.String("task_id", taskID),
		slog.String("type", string(task.Type())),
		slog.String("test_mode", req.TestMode),
	)
	
	return task, nil
}

// GetTask 获取任务
func (m *Manager) GetTask(taskID string) (Task, bool) {
	m.tasksMutex.RLock()
	defer m.tasksMutex.RUnlock()
	
	task, exists := m.tasks[taskID]
	return task, exists
}

// GetAllTasks 获取所有任务
func (m *Manager) GetAllTasks() []Task {
	m.tasksMutex.RLock()
	defer m.tasksMutex.RUnlock()
	
	tasks := make([]Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	
	return tasks
}

// StartTask 启动任务
func (m *Manager) StartTask(taskID string) error {
	task, exists := m.GetTask(taskID)
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}
	
	// 通过执行器启动任务
	return m.executor.ExecuteTask(task)
}

// CancelTask 取消任务
func (m *Manager) CancelTask(taskID string) error {
	task, exists := m.GetTask(taskID)
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}
	
	return task.Cancel()
}

// RemoveTask 删除任务
func (m *Manager) RemoveTask(taskID string) error {
	m.tasksMutex.Lock()
	defer m.tasksMutex.Unlock()
	
	if _, exists := m.tasks[taskID]; !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}
	
	delete(m.tasks, taskID)
	
	logger.Logger.InfoContext(context.Background(), "Task removed",
		slog.String("task_id", taskID))
	
	return nil
}

// CleanupCompletedTasks 清理已完成的任务
func (m *Manager) CleanupCompletedTasks(maxAge time.Duration) {
	m.tasksMutex.Lock()
	defer m.tasksMutex.Unlock()
	
	now := time.Now()
	toDelete := make([]string, 0)
	
	for taskID, task := range m.tasks {
		if task.Status() == TaskStatusCompleted || task.Status() == TaskStatusFailed {
			// 检查任务是否超过最大保留时间
			if now.Sub(task.Progress().StartTime) > maxAge {
				toDelete = append(toDelete, taskID)
			}
		}
	}
	
	for _, taskID := range toDelete {
		delete(m.tasks, taskID)
		logger.Logger.DebugContext(context.Background(), "Cleaned up old task",
			slog.String("task_id", taskID))
	}
	
	if len(toDelete) > 0 {
		logger.Logger.InfoContext(context.Background(), "Cleaned up completed tasks",
			slog.Int("count", len(toDelete)))
	}
}

// AddEventHandler 添加事件处理器
func (m *Manager) AddEventHandler(eventType string, handler TaskEventHandler) {
	m.eventMutex.Lock()
	defer m.eventMutex.Unlock()
	
	if _, exists := m.eventHandlers[eventType]; !exists {
		m.eventHandlers[eventType] = make([]TaskEventHandler, 0)
	}
	
	m.eventHandlers[eventType] = append(m.eventHandlers[eventType], handler)
}

// EmitEvent 发送事件
func (m *Manager) EmitEvent(event *TaskEvent) {
	m.eventMutex.RLock()
	handlers := m.eventHandlers[event.Type]
	m.eventMutex.RUnlock()
	
	// 调用事件处理器
	for _, handler := range handlers {
		go handler(event)
	}
	
	// 通过 WebSocket 广播事件
	if m.wsHub != nil {
		m.wsHub.BroadcastMessage(websocket.MessageType(event.Type), event)
	}
}

// GetStats 获取任务统计信息
func (m *Manager) GetStats() map[string]interface{} {
	m.tasksMutex.RLock()
	defer m.tasksMutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_tasks":     len(m.tasks),
		"pending_tasks":   0,
		"running_tasks":   0,
		"completed_tasks": 0,
		"cancelled_tasks": 0,
		"failed_tasks":    0,
	}
	
	for _, task := range m.tasks {
		switch task.Status() {
		case TaskStatusPending:
			stats["pending_tasks"] = stats["pending_tasks"].(int) + 1
		case TaskStatusRunning:
			stats["running_tasks"] = stats["running_tasks"].(int) + 1
		case TaskStatusCompleted:
			stats["completed_tasks"] = stats["completed_tasks"].(int) + 1
		case TaskStatusCancelled:
			stats["cancelled_tasks"] = stats["cancelled_tasks"].(int) + 1
		case TaskStatusFailed:
			stats["failed_tasks"] = stats["failed_tasks"].(int) + 1
		}
	}
	
	return stats
}

// Start 启动任务管理器
func (m *Manager) Start() {
	// 启动定期清理任务
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				m.CleanupCompletedTasks(24 * time.Hour)
			}
		}
	}()
	
	logger.Logger.Info("Task manager started")
}

// Stop 停止任务管理器
func (m *Manager) Stop() {
	// 取消所有运行中的任务
	m.tasksMutex.RLock()
	tasks := make([]Task, 0)
	for _, task := range m.tasks {
		if task.Status() == TaskStatusRunning {
			tasks = append(tasks, task)
		}
	}
	m.tasksMutex.RUnlock()
	
	for _, task := range tasks {
		task.Cancel()
	}
	
	logger.Logger.Info("Task manager stopped")
}