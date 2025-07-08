package etcdclient

import "github.com/coreos/etcd/clientv3"

// ServerReg 结构体用于管理一个服务在 etcd 上的注册信息和生命周期
type ServerReg struct {
	Client        *Client                                 // 封装了 etcd 客户端连接的自定义结构体
	stop          chan error                              // 用于goroutine停止工作的通道
	leaseId       clientv3.LeaseID                        // etcd 生成的租约ID
	cancelFunc    func()                                  // 与 context 关联的取消函数
	KeepAliveChan <-chan *clientv3.LeaseKeepAliveResponse // 接收 etcd 服务端对租约续约请求响应的只读通道
	// time-to-live
	Ttl int64 // 租约的有效期
}

// NewServerReg 是 ServerReg 的构造函数
func NewServerReg(ttl int64) *ServerReg {
	return &ServerReg{
		// 单例模式, 避免重复创建 etcd 连接
		Client: _defaultEtcd,
		Ttl:    ttl,              // 设置租约的 TTL
		stop:   make(chan error), // 初始化用于停止信号的通道
	}
}
