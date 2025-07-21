package handler

import (
	"crony/common/pkg/etcdclient"
	"fmt"

	"github.com/coreos/etcd/clientv3"
)

// WatchSystem 函数用于创建一个etcd watch通道，用于监听特定节点的系统级事件或开关
// nodeUUID参数指定了要监听的目标节点的唯一标识符
func WatchSystem(nodeUUID string) clientv3.WatchChan {
	// 使用 fmt.Sprintf 将节点的 UUID 格式化到预定义的 etcd key 模板中，从而生成一个节点专属的 key
	// etcdclient.KeyEtcdSystemSwitch 是一个类似 "/crony/system/switch/%s" 的字符串模板
	key := fmt.Sprintf(etcdclient.KeyEtcdSystemSwitch, nodeUUID)

	// 调用 etcd 客户端的 Watch 方法，监听这个为特定节点生成的 key
	// clientv3.WithPrefix() 确保也会监听到该 key 下的所有子 key 的变化
	return etcdclient.Watch(key, clientv3.WithPrefix())
}
