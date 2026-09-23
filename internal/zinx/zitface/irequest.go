package zitface

// IRequest 将客户端连接和本次读取的数据封装为一个请求。
// 路由只依赖该接口即可同时取得请求数据和响应所需的连接。
type IRequest interface {
	GetConnection() IConnection // GetConnection 返回产生该请求的客户端连接。
	GetData() []byte            // GetData 返回本次从客户端读取到的有效数据。
}
