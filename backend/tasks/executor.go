package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/faceair/clash-speedtest/logger"
	"github.com/faceair/clash-speedtest/server/common"
	"github.com/faceair/clash-speedtest/speedtester"
	"github.com/faceair/clash-speedtest/unlock"
)

// Executor 任务执行器
type Executor struct {
	manager       *Manager
	workerPool    chan struct{}
	maxWorkers    int
	activeWorkers int
	workersMutex  sync.RWMutex
}

// NewExecutor 创建新的任务执行器
func NewExecutor(manager *Manager) *Executor {
	maxWorkers := 10 // 可以配置化
	return &Executor{
		manager:    manager,
		workerPool: make(chan struct{}, maxWorkers),
		maxWorkers: maxWorkers,
	}
}

// ExecuteTask 执行任务
func (e *Executor) ExecuteTask(task Task) error {
	// 检查任务状态
	if task.Status() != TaskStatusPending {
		logger.Logger.Warn("Task is not in pending status",
			slog.String("task_id", task.ID()),
			slog.String("current_status", string(task.Status())),
		)
		return fmt.Errorf("task is not in pending status: %s", task.Status())
	}
	
	// 获取工作池信号量
	select {
	case e.workerPool <- struct{}{}:
		logger.Logger.Debug("Task accepted by worker pool",
			slog.String("task_id", task.ID()),
			slog.Int("active_workers", e.activeWorkers),
		)
	default:
		logger.Logger.Warn("Worker pool is full, task rejected",
			slog.String("task_id", task.ID()),
			slog.Int("max_workers", e.maxWorkers),
		)
		return fmt.Errorf("worker pool is full")
	}
	
	// 异步执行任务
	go func() {
		defer func() {
			<-e.workerPool // 释放工作池槽位
		}()
		
		e.executeTaskInternal(task)
	}()
	
	return nil
}

// executeTaskInternal 内部任务执行逻辑
func (e *Executor) executeTaskInternal(task Task) {
	ctx := task.Context()
	
	e.workersMutex.Lock()
	e.activeWorkers++
	e.workersMutex.Unlock()
	
	defer func() {
		e.workersMutex.Lock()
		e.activeWorkers--
		e.workersMutex.Unlock()
	}()
	
	// 启动任务
	if err := task.Start(); err != nil {
		logger.Logger.ErrorContext(ctx, "Failed to start task",
			slog.String("task_id", task.ID()),
			slog.String("error", err.Error()))
		
		e.manager.EmitEvent(&TaskEvent{
			TaskID:    task.ID(),
			Type:      TaskEventTypeFailed,
			Data:      map[string]interface{}{"error": err.Error()},
			Timestamp: time.Now(),
		})
		return
	}
	
	// 发送任务开始事件
	e.manager.EmitEvent(&TaskEvent{
		TaskID:    task.ID(),
		Type:      TaskEventTypeStarted,
		Data:      task.Progress(),
		Timestamp: time.Now(),
	})
	
	// 等待任务完成或取消
	e.waitForTaskCompletion(task)
}

// waitForTaskCompletion 等待任务完成
func (e *Executor) waitForTaskCompletion(task Task) {
	ctx := task.Context()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	startTime := time.Now()
	
	for {
		select {
		case <-ctx.Done():
			// 任务被取消
			duration := time.Since(startTime)
			logger.Logger.InfoContext(ctx, "Task cancelled",
				slog.String("task_id", task.ID()),
				slog.Duration("duration", duration),
				slog.String("reason", ctx.Err().Error()))
			
			e.manager.EmitEvent(&TaskEvent{
				TaskID:    task.ID(),
				Type:      TaskEventTypeCancelled,
				Data:      task.Progress(),
				Timestamp: time.Now(),
			})
			return
			
		case <-ticker.C:
			// 检查任务状态
			status := task.Status()
			duration := time.Since(startTime)
			
			// 发送进度更新
			if status == TaskStatusRunning {
				e.manager.EmitEvent(&TaskEvent{
					TaskID:    task.ID(),
					Type:      TaskEventTypeProgress,
					Data:      task.Progress(),
					Timestamp: time.Now(),
				})
			}
			
			// 检查是否完成
			if status == TaskStatusCompleted {
				logger.Logger.InfoContext(ctx, "Task completed",
					slog.String("task_id", task.ID()),
					slog.Duration("duration", duration))
				
				e.manager.EmitEvent(&TaskEvent{
					TaskID:    task.ID(),
					Type:      TaskEventTypeCompleted,
					Data:      task.Progress(),
					Timestamp: time.Now(),
				})
				return
			}
			
			// 检查是否失败
			if status == TaskStatusFailed {
				logger.Logger.ErrorContext(ctx, "Task failed",
					slog.String("task_id", task.ID()),
					slog.Duration("duration", duration),
					slog.String("error", task.Error().Error()))
				
				e.manager.EmitEvent(&TaskEvent{
					TaskID:    task.ID(),
					Type:      TaskEventTypeFailed,
					Data:      map[string]interface{}{
						"error":    task.Error().Error(),
						"progress": task.Progress(),
					},
					Timestamp: time.Now(),
				})
				return
			}
		}
	}
}

// SpeedTestTask 速度测试任务实现
type SpeedTestTask struct {
	id         string
	config     *common.TestRequest
	ctx        context.Context
	cancelFunc context.CancelFunc
	status     TaskStatus
	options    *TaskOptions
	startTime  time.Time
	progress   *TaskProgress
	results    []*speedtester.Result
	err        error
	manager    *Manager
	
	// 同步锁
	mutex sync.RWMutex
}

// ID 返回任务 ID
func (t *SpeedTestTask) ID() string {
	return t.id
}

// Type 返回任务类型
func (t *SpeedTestTask) Type() TaskType {
	switch t.config.TestMode {
	case "speed_only":
		return TaskTypeSpeedTest
	case "unlock_only":
		return TaskTypeUnlock
	case "both":
		return TaskTypeBoth
	default:
		return TaskTypeSpeedTest
	}
}

// Status 返回任务状态
func (t *SpeedTestTask) Status() TaskStatus {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.status
}

// Config 返回任务配置
func (t *SpeedTestTask) Config() *common.TestRequest {
	return t.config
}

// Context 返回任务上下文
func (t *SpeedTestTask) Context() context.Context {
	return t.ctx
}

// Start 启动任务
func (t *SpeedTestTask) Start() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	if t.status != TaskStatusPending {
		return fmt.Errorf("task is not in pending status")
	}
	
	t.status = TaskStatusRunning
	t.startTime = time.Now()
	
	// 异步执行实际的测试逻辑
	go t.executeTest()
	
	return nil
}

// Cancel 取消任务
func (t *SpeedTestTask) Cancel() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	if t.status != TaskStatusRunning {
		return fmt.Errorf("task is not running")
	}
	
	t.status = TaskStatusCancelled
	t.cancelFunc()
	
	return nil
}

// Results 返回任务结果
func (t *SpeedTestTask) Results() []*speedtester.Result {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.results
}

// Progress 返回任务进度
func (t *SpeedTestTask) Progress() *TaskProgress {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	
	progress := *t.progress
	progress.Duration = time.Since(t.startTime)
	return &progress
}

// Error 返回任务错误
func (t *SpeedTestTask) Error() error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.err
}

// executeTest 执行测试逻辑
func (t *SpeedTestTask) executeTest() {
	ctx := t.ctx
	
	// 创建解锁配置
	unlockConfig := t.createUnlockConfig()
	
	// 创建速度测试器
	speedTester := speedtester.New(&speedtester.Config{
		ConfigPaths:      t.config.ConfigPaths,
		FilterRegex:      t.config.FilterRegex,
		IncludeNodes:     t.config.IncludeNodes,
		ExcludeNodes:     t.config.ExcludeNodes,
		ProtocolFilter:   t.config.ProtocolFilter,
		ServerURL:        t.config.ServerURL,
		DownloadSize:     t.config.DownloadSize * 1024 * 1024,
		UploadSize:       t.config.UploadSize * 1024 * 1024,
		Timeout:          time.Duration(t.config.Timeout) * time.Second,
		Concurrent:       t.config.Concurrent,
		MaxLatency:       time.Duration(t.config.MaxLatency) * time.Millisecond,
		MinDownloadSpeed: t.config.MinDownloadSpeed * 1024 * 1024,
		MinUploadSpeed:   t.config.MinUploadSpeed * 1024 * 1024,
		FastMode:         t.config.FastMode,
		RenameNodes:      t.config.RenameNodes,
		TestMode:         t.config.TestMode,
		UnlockConfig:     unlockConfig,
	})
	
	// 加载代理
	allProxies, err := speedTester.LoadProxies(t.config.StashCompatible)
	if err != nil {
		t.setError(fmt.Errorf("failed to load proxies: %w", err))
		return
	}
	
	if len(allProxies) == 0 {
		t.setError(fmt.Errorf("no proxies found"))
		return
	}
	
	// 更新进度
	t.updateProgress(0, len(allProxies), "")
	
	// 执行测试
	results := make([]*speedtester.Result, 0)
	completed := 0
	
	err = speedTester.TestProxiesWithContext(ctx, allProxies, func(result *speedtester.Result) {
		results = append(results, result)
		completed++
		
		// 更新进度
		t.updateProgress(completed, len(allProxies), result.ProxyName)
		
		// 发送结果事件
		t.manager.EmitEvent(&TaskEvent{
			TaskID:    t.id,
			Type:      TaskEventTypeResult,
			Data:      result,
			Timestamp: time.Now(),
		})
	})
	
	if err != nil && err != context.Canceled {
		t.setError(fmt.Errorf("test execution failed: %w", err))
		return
	}
	
	// 设置结果
	t.mutex.Lock()
	t.results = results
	if err == context.Canceled {
		t.status = TaskStatusCancelled
	} else {
		t.status = TaskStatusCompleted
	}
	t.mutex.Unlock()
}

// createUnlockConfig 创建解锁配置
func (t *SpeedTestTask) createUnlockConfig() *unlock.UnlockTestConfig {
	needsUnlock := t.config.TestMode == "unlock_only" || t.config.TestMode == "both"
	
	if !t.config.UnlockEnabled && !needsUnlock {
		return &unlock.UnlockTestConfig{
			Enabled: false,
		}
	}
	
	platforms := t.config.UnlockPlatforms
	if len(platforms) == 0 {
		platforms = []string{"Netflix", "YouTube", "Disney+", "ChatGPT", "Spotify", "Bilibili"}
	}
	
	concurrent := t.config.UnlockConcurrent
	if concurrent <= 0 {
		concurrent = 5
	}
	
	timeout := t.config.UnlockTimeout
	if timeout <= 0 {
		timeout = 10
	}
	
	return &unlock.UnlockTestConfig{
		Enabled:       true,
		Platforms:     platforms,
		Concurrent:    concurrent,
		Timeout:       timeout,
		RetryOnError:  t.config.UnlockRetry,
		IncludeIPInfo: true,
	}
}

// updateProgress 更新进度
func (t *SpeedTestTask) updateProgress(completed, total int, currentProxy string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	t.progress.CompletedCount = completed
	t.progress.TotalCount = total
	t.progress.CurrentProxy = currentProxy
	
	if total > 0 {
		t.progress.ProgressPercent = float64(completed) / float64(total) * 100
	}
}

// setError 设置错误
func (t *SpeedTestTask) setError(err error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	t.err = err
	t.status = TaskStatusFailed
}