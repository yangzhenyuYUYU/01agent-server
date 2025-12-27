package service

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type ConcurrentService struct{}

// NewConcurrentService 创建并发服务
func NewConcurrentService() *ConcurrentService {
	return &ConcurrentService{}
}

// TaskResult 任务结果
type TaskResult struct {
	TaskID    int           `json:"task_id"`
	TaskName  string        `json:"task_name"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Delay     time.Duration `json:"delay"`
	Result    interface{}   `json:"result"`
}

// TaskSummary 任务摘要
type TaskSummary struct {
	MinDuration    time.Duration `json:"min_duration"`
	MaxDuration    time.Duration `json:"max_duration"`
	AvgDuration    time.Duration `json:"avg_duration"`
	TotalTaskTime  time.Duration `json:"total_task_time"`
	EfficiencyRate float64       `json:"efficiency_rate"` // 总任务时间 / 实际执行时间
}

// ConcurrentTestResponse 并发测试响应
type ConcurrentTestResponse struct {
	Method         string        `json:"method"`
	TotalTasks     int           `json:"total_tasks"`
	TotalStartTime time.Time     `json:"total_start_time"`
	TotalEndTime   time.Time     `json:"total_end_time"`
	TotalDuration  time.Duration `json:"total_duration"`
	Tasks          []TaskResult  `json:"tasks"`
	Summary        TaskSummary   `json:"summary"`
}

// simulateTask 模拟一个有延迟的任务
func (s *ConcurrentService) simulateTask(taskID int, taskName string, delayMs int) TaskResult {
	startTime := time.Now()
	delay := time.Duration(delayMs) * time.Millisecond

	// 模拟任务处理（随机延迟）
	time.Sleep(delay)

	// 模拟一些计算工作
	result := fmt.Sprintf("Task %d completed with %dms delay", taskID, delayMs)

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	return TaskResult{
		TaskID:    taskID,
		TaskName:  taskName,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  duration,
		Delay:     delay,
		Result:    result,
	}
}

// calculateSummary 计算任务摘要
func (s *ConcurrentService) calculateSummary(tasks []TaskResult, totalDuration time.Duration) TaskSummary {
	if len(tasks) == 0 {
		return TaskSummary{}
	}

	var minDuration, maxDuration, totalTaskTime time.Duration
	minDuration = tasks[0].Duration
	maxDuration = tasks[0].Duration

	for _, task := range tasks {
		if task.Duration < minDuration {
			minDuration = task.Duration
		}
		if task.Duration > maxDuration {
			maxDuration = task.Duration
		}
		totalTaskTime += task.Duration
	}

	avgDuration := totalTaskTime / time.Duration(len(tasks))
	efficiencyRate := float64(totalTaskTime) / float64(totalDuration)

	return TaskSummary{
		MinDuration:    minDuration,
		MaxDuration:    maxDuration,
		AvgDuration:    avgDuration,
		TotalTaskTime:  totalTaskTime,
		EfficiencyRate: efficiencyRate,
	}
}

// SerialExecution 串行执行任务
func (s *ConcurrentService) SerialExecution(taskCount int) *ConcurrentTestResponse {
	if taskCount > 20 {
		taskCount = 20 // 限制最大任务数
	}

	totalStartTime := time.Now()
	var tasks []TaskResult

	// 串行执行任务
	for i := 1; i <= taskCount; i++ {
		delayMs := rand.Intn(1000) + 500 // 500-1500ms 随机延迟
		taskName := fmt.Sprintf("Serial Task %d", i)
		task := s.simulateTask(i, taskName, delayMs)
		tasks = append(tasks, task)
	}

	totalEndTime := time.Now()
	totalDuration := totalEndTime.Sub(totalStartTime)

	summary := s.calculateSummary(tasks, totalDuration)

	return &ConcurrentTestResponse{
		Method:         "Serial",
		TotalTasks:     taskCount,
		TotalStartTime: totalStartTime,
		TotalEndTime:   totalEndTime,
		TotalDuration:  totalDuration,
		Tasks:          tasks,
		Summary:        summary,
	}
}

// ConcurrentExecution 并发执行任务
func (s *ConcurrentService) ConcurrentExecution(taskCount int) *ConcurrentTestResponse {
	if taskCount > 20 {
		taskCount = 20 // 限制最大任务数
	}

	totalStartTime := time.Now()
	var tasks []TaskResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 并发执行任务
	for i := 1; i <= taskCount; i++ {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()

			delayMs := rand.Intn(1000) + 500 // 500-1500ms 随机延迟
			taskName := fmt.Sprintf("Concurrent Task %d", taskID)
			task := s.simulateTask(taskID, taskName, delayMs)

			// 安全地添加到结果切片
			mu.Lock()
			tasks = append(tasks, task)
			mu.Unlock()
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()

	totalEndTime := time.Now()
	totalDuration := totalEndTime.Sub(totalStartTime)

	summary := s.calculateSummary(tasks, totalDuration)

	return &ConcurrentTestResponse{
		Method:         "Concurrent",
		TotalTasks:     taskCount,
		TotalStartTime: totalStartTime,
		TotalEndTime:   totalEndTime,
		TotalDuration:  totalDuration,
		Tasks:          tasks,
		Summary:        summary,
	}
}

// CompareExecution 对比串行和并发执行
func (s *ConcurrentService) CompareExecution(taskCount int) map[string]interface{} {
	if taskCount > 20 {
		taskCount = 20 // 限制最大任务数
	}

	// 准备相同的任务延迟列表，确保公平对比
	taskDelays := make([]int, taskCount)
	for i := 0; i < taskCount; i++ {
		taskDelays[i] = rand.Intn(1000) + 500 // 500-1500ms
	}

	// 串行执行
	serialStartTime := time.Now()
	var serialTasks []TaskResult
	for i := 0; i < taskCount; i++ {
		taskName := fmt.Sprintf("Serial Task %d", i+1)
		task := s.simulateTask(i+1, taskName, taskDelays[i])
		serialTasks = append(serialTasks, task)
	}
	serialEndTime := time.Now()
	serialDuration := serialEndTime.Sub(serialStartTime)

	// 并发执行
	concurrentStartTime := time.Now()
	var concurrentTasks []TaskResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func(taskID int, delay int) {
			defer wg.Done()

			taskName := fmt.Sprintf("Concurrent Task %d", taskID+1)
			task := s.simulateTask(taskID+1, taskName, delay)

			mu.Lock()
			concurrentTasks = append(concurrentTasks, task)
			mu.Unlock()
		}(i, taskDelays[i])
	}

	wg.Wait()
	concurrentEndTime := time.Now()
	concurrentDuration := concurrentEndTime.Sub(concurrentStartTime)

	// 计算性能提升
	speedup := float64(serialDuration) / float64(concurrentDuration)

	// 构建对比结果
	return map[string]interface{}{
		"task_count":  taskCount,
		"task_delays": taskDelays,
		"serial": map[string]interface{}{
			"duration":    serialDuration,
			"duration_ms": serialDuration.Milliseconds(),
			"tasks":       serialTasks,
			"summary":     s.calculateSummary(serialTasks, serialDuration),
		},
		"concurrent": map[string]interface{}{
			"duration":    concurrentDuration,
			"duration_ms": concurrentDuration.Milliseconds(),
			"tasks":       concurrentTasks,
			"summary":     s.calculateSummary(concurrentTasks, concurrentDuration),
		},
		"performance": map[string]interface{}{
			"speedup":         speedup,
			"time_saved":      serialDuration - concurrentDuration,
			"time_saved_ms":   serialDuration.Milliseconds() - concurrentDuration.Milliseconds(),
			"efficiency_gain": fmt.Sprintf("%.2f%%", (speedup-1)*100),
		},
	}
}

// StressTest 压力测试
func (s *ConcurrentService) StressTest(goroutineCount int) map[string]interface{} {
	if goroutineCount > 1000 {
		goroutineCount = 1000 // 限制最大协程数
	}

	totalStartTime := time.Now()
	var completed int32
	var wg sync.WaitGroup

	for i := 1; i <= goroutineCount; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// 模拟随机工作负载
			workTime := time.Duration(rand.Intn(100)+50) * time.Millisecond
			time.Sleep(workTime)

			// 计数
			completed++
		}(i)
	}

	wg.Wait()
	totalEndTime := time.Now()
	totalDuration := totalEndTime.Sub(totalStartTime)

	return map[string]interface{}{
		"goroutines_count":              goroutineCount,
		"completed_count":               completed,
		"total_duration":                totalDuration,
		"total_duration_ms":             totalDuration.Milliseconds(),
		"goroutines_per_ms":             float64(goroutineCount) / float64(totalDuration.Milliseconds()),
		"average_time_per_goroutine":    totalDuration / time.Duration(goroutineCount),
	}
} 