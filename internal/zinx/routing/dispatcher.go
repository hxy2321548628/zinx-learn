package routing

import (
	"fmt"
	"log/slog"
)

// Dispatcher 按消息 ID 分发请求，并管理工作队列。
type Dispatcher struct {
	handlers         map[uint32]Handler
	workerPoolSize   uint32
	maxWorkerTaskLen uint32
	taskQueues       []chan *Request
}

// NewDispatcher 创建请求分发器。
func NewDispatcher(workerPoolSize, maxWorkerTaskLen uint32) *Dispatcher {
	return &Dispatcher{
		handlers:         make(map[uint32]Handler),
		workerPoolSize:   workerPoolSize,
		maxWorkerTaskLen: maxWorkerTaskLen,
		taskQueues:       make([]chan *Request, workerPoolSize),
	}
}

func (d *Dispatcher) startWorker(workerID int, taskQueue <-chan *Request) {
	slog.Info("工作协程已启动", "worker_id", workerID)
	for request := range taskQueue {
		d.dispatch(request)
	}
}

// Start 启动工作池。
func (d *Dispatcher) Start() {
	for i := range d.taskQueues {
		d.taskQueues[i] = make(chan *Request, d.maxWorkerTaskLen)
		go d.startWorker(i, d.taskQueues[i])
	}
}

// Submit 将请求提交给工作池或独立协程。
func (d *Dispatcher) Submit(request *Request) {
	if d.workerPoolSize == 0 {
		go d.dispatch(request)
		return
	}

	workerID := request.ConnectionID() % d.workerPoolSize
	slog.Debug("请求已入队",
		"connection_id", request.ConnectionID(),
		"message_id", request.MessageID(),
		"worker_id", workerID,
	)
	d.taskQueues[workerID] <- request
}

func (d *Dispatcher) dispatch(request *Request) {
	handler, ok := d.handlers[request.MessageID()]
	if !ok {
		slog.Warn("未找到消息处理器", "message_id", request.MessageID())
		return
	}
	handler.Handle(request)
}

// Register 为消息 ID 注册处理器。
func (d *Dispatcher) Register(messageID uint32, handler Handler) error {
	if handler == nil {
		return fmt.Errorf("消息 %d 的处理器不能为 nil", messageID)
	}
	if _, exists := d.handlers[messageID]; exists {
		return fmt.Errorf("消息 %d 已注册处理器", messageID)
	}
	d.handlers[messageID] = handler
	slog.Info("消息处理器已注册", "message_id", messageID)
	return nil
}
