package handler

import (
	"crony/common/pkg/etcdclient"

	"github.com/coreos/etcd/clientv3"
)

// WatchOnce 函数用于创建一个etcd的watch通道，专门用于监听一次性任务
func WatchOnce() clientv3.WatchChan {
	// 调用etcd客户端的Watch方法，监听预定义的“一次性任务”的key前缀
	// etcdclient.KeyEtcdOnceProfile 是用于从此一次性任务的etcd key
	// clientv3.WithPrefix() 用于监听所有以此key为前缀的键值对的变化
	return etcdclient.Watch(etcdclient.KeyEtcdOnceProfile, clientv3.WithPrefix())
}
