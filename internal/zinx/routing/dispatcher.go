package routing

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

// Dispatcher 按消息 ID 分发请求，并管理固定大小的工作池。
//
// 同一连接的请求始终进入同一个队列，从而保持处理顺序；不同连接可以由不同 worker
// 并行处理。队列是有界的，队列满时 Submit 会阻塞，形成背压而不是无限占用内存。
type Dispatcher struct {
	handlers       map[uint32]Handler
	workerPoolSize uint32
	taskQueues     []chan task
	handlersMu     sync.RWMutex
}

type task struct {
	// ctx 属于产生该任务的连接。服务器关闭后，已入队但尚未处理的任务可据此跳过。
	ctx     context.Context
	request *Request
}

// NewDispatcher 创建请求分发器。
func NewDispatcher(workerPoolSize, maxWorkerTaskLen uint32) *Dispatcher {
	return &Dispatcher{
		handlers:       make(map[uint32]Handler),
		workerPoolSize: workerPoolSize,
		taskQueues:     makeTaskQueues(workerPoolSize, maxWorkerTaskLen),
	}
}

func makeTaskQueues(workerPoolSize, maxWorkerTaskLen uint32) []chan task {
	queues := make([]chan task, workerPoolSize)
	for i := range queues {
		queues[i] = make(chan task, maxWorkerTaskLen)
	}
	return queues
}

func (d *Dispatcher) runWorker(ctx context.Context, workerID int, taskQueue <-chan task) {
	slog.Info("工作协程已启动", "worker_id", workerID)
	defer slog.Info("工作协程已退出", "worker_id", workerID)

	for {
		select {
		case <-ctx.Done():
			return
		case task := <-taskQueue:
			// 优先结束关闭流程，不在停机时继续消费排队中的任务。
			if ctx.Err() != nil {
				return
			}
			if task.ctx.Err() != nil {
				continue
			}
			d.dispatch(task.ctx, task.request)
		}
	}
}

// Run 启动工作池，并阻塞到 ctx 取消且所有 worker 退出。
// 当前采用“立即停止”语义：ctx 取消后不再排空队列。
func (d *Dispatcher) Run(ctx context.Context) {
	if len(d.taskQueues) == 0 {
		<-ctx.Done()
		return
	}

	var workers sync.WaitGroup
	workers.Add(len(d.taskQueues))
	for i := range d.taskQueues {
		go func(workerID int) {
			defer workers.Done()
			d.runWorker(ctx, workerID, d.taskQueues[workerID])
		}(i)
	}
	workers.Wait()
}

// Submit 将请求提交给工作池。
// 按连接 ID 取模可以让同一连接稳定地落到同一个 worker；代价是取模相同的连接也会串行。
func (d *Dispatcher) Submit(ctx context.Context, request *Request) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d.workerPoolSize == 0 {
		return errors.New("工作池未配置")
	}

	workerID := request.ConnectionID() % d.workerPoolSize
	slog.Debug("请求已入队",
		"connection_id", request.ConnectionID(),
		"message_id", request.MessageID(),
		"worker_id", workerID,
	)
	select {
	case d.taskQueues[workerID] <- task{ctx: ctx, request: request}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *Dispatcher) dispatch(ctx context.Context, request *Request) {
	// 注册和查询可能并发发生，只在访问 map 时持锁，避免执行用户 Handler 时占用锁。
	d.handlersMu.RLock()
	handler, ok := d.handlers[request.MessageID()]
	d.handlersMu.RUnlock()
	if !ok {
		slog.Warn("未找到消息处理器", "message_id", request.MessageID())
		return
	}
	// Handler 在 worker 协程内同步执行；慢 Handler 会阻塞该 worker 负责的连接。
	if err := handler.Handle(ctx, request); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("处理请求失败", "message_id", request.MessageID(), "error", err)
	}
}

// Register 为消息 ID 注册处理器。
func (d *Dispatcher) Register(messageID uint32, handler Handler) error {
	if handler == nil {
		return fmt.Errorf("消息 %d 的处理器不能为 nil", messageID)
	}
	d.handlersMu.Lock()
	defer d.handlersMu.Unlock()
	if _, exists := d.handlers[messageID]; exists {
		return fmt.Errorf("消息 %d 已注册处理器", messageID)
	}
	d.handlers[messageID] = handler
	slog.Info("消息处理器已注册", "message_id", messageID)
	return nil
}
