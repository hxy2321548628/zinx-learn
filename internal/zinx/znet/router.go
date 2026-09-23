package znet

import "zinx-learn/internal/zinx/zitface"

// BaseRouter 为 IRouter 的三个阶段提供空实现。
// 自定义路由可以嵌入 BaseRouter，只实现当前业务真正需要的处理阶段。
type BaseRouter struct{}

// PreHandle 默认不执行任何预处理。
func (br *BaseRouter) PreHandle(req zitface.IRequest) {}

// Handle 默认不执行任何核心业务。
func (br *BaseRouter) Handle(req zitface.IRequest) {}

// PostHandle 默认不执行任何收尾处理。
func (br *BaseRouter) PostHandle(req zitface.IRequest) {}
