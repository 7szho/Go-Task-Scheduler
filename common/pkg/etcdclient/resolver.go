package etcdclient

type Watcher interface {
	// 启动监视
	Watch() error

	// 停止监视
	Close() error
}
