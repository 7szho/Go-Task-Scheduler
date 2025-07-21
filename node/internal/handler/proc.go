package handler

import (
	"crony/common/models"
	"crony/common/pkg/config"
	"crony/common/pkg/etcdclient"
	"crony/common/pkg/logger"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/coreos/etcd/clientv3"
)

// JobProc 结构体代表一个正在执行的任务的信息
type JobProc struct {
	*models.JobProc
}

// GetProcFromKey 函数从一个etcd的key字符串中解析出JobProc的信息
func GetProcFromKey(key string) (proc *JobProc, err error) {
	// 通过”/“分割etcd key字符串
	// 预期的 key 格式应为： /crony/proc/{nodeUUID}/{jobId}/{procId}
	ss := strings.Split(key, "/")
	var sslen = len(ss)
	if sslen < 5 { // 根据与其的key格式，分割后的长度至少为5
		err = fmt.Errorf("invalid proc key [%s]", key)
		return
	}
	// 从分割后的切片末尾解析出进程ID
	id, err := strconv.Atoi(ss[sslen-1])
	if err != nil {
		return
	}
	// 从倒数第二个位置解析出作业ID
	jobId, err := strconv.Atoi(ss[sslen-2])
	if err != nil {
		return
	}
	// 创建JobProc实例并填充解析出的值
	proc = &JobProc{
		JobProc: &models.JobProc{
			ID:       id,
			JobID:    jobId,
			NodeUUID: ss[sslen-3], // 使用 ss[sslen-3] 来获取 {nodeUUID}
		},
	}
	return
}

// Key 方法为JobProc实例生成一个唯一的etcd key
// 这个key用于在 etcd 中存储该进程的运行信息
func (p *JobProc) Key() string {
	// 使用预定义的etcd key格式模板来生成 key 字符串
	return fmt.Sprintf(etcdclient.KeyEtcdProc, p.NodeUUID, p.JobID, p.ID)
}

// del 是一个内部方法，用于从etcd中删除该进程的key
func (p *JobProc) del() error {
	_, err := etcdclient.Delete(p.Key())
	return err
}

// Start 方法负责在etcd中注册一个正在运行的进程
// 这是一个线程安全的方法
func (p *JobProc) Start() error {
	// 使用原子操作确保 Start 逻辑只被执行一次。
	// 它原子地检查 p.Running 是否为 0，如果是，则将其置为 1，并返回 true
	if !atomic.CompareAndSwapInt32(&p.Running, 0, 1) {
		return nil // 如果已经启动，则直接返回nil，不做任何操作
	}

	// 为WaitGroup增加一个计数，表示有一个长时间运行的操作（etcd put）开始了
	p.Wg.Add(1)
	// 在函数退出时，调用`Done()`来减少WaitGroup的计数
	defer p.Wg.Done()
	// 将进程的动态值（如启动时间、是否被杀死等）序列化为JSON字符串
	b, err := json.Marshal(p.JobProcVal)
	if err != nil {
		// 如果写入etcd失败，返回错误
		return err
	}
	// 将进程信息写入etcd，并设置一个租约TTL
	// 这是一种心跳机制：如果节点崩溃，无法续约，该key会在TTL到期后被etcd自动删除
	// 这可以有效防止etcd中出现僵尸进程记录
	_, err = etcdclient.PutWithTtl(p.Key(), string(b), config.GetConfigModels().System.JobProcTtl)
	if err != nil {
		return err
	}
	return nil
}

// Stop 方法负责停止进程的追踪并清理etcd中的记录
// 这是一个线程安全的方法
func (p *JobProc) Stop() {
	if p == nil {
		return
	}
	// 使用原子操作 CompareAndSwapInt32 来确保 Stop 逻辑只被执行一次
	// 它会原子地检查 p.Running 是否为1，如果是，则将其置为0，并返回true
	// 如果p.Running不为1，则操作失败，返回false，函数直接退出
	if !atomic.CompareAndSwapInt32(&p.Running, 1, 0) {
		return
	}
	// 等待所有相关的goroutine完成，这里主要是等待Start方法中的etcd操作完成
	// 这样可以防止在etcd的put操作完成前就执行删除操作
	p.Wg.Wait()

	// 从etcd中删除该进程的记录
	if err := p.del(); err != nil {
		// 如果删除失败，记录一条警告日志
		logger.GetLogger().Warn(fmt.Sprintf("proc del[%s] err: %s", p.Key(), err.Error()))
	}
}

// WatchProc 函数创建一个etcd watch通道，用于监听指定节点上所有进程的变化
func WatchProc(nodeUUID string) clientv3.WatchChan {
	// 监听的key是该节点下所有进程的公共前缀
	keyPrefix := fmt.Sprintf(etcdclient.KeyEtcdNodeProcProfile, nodeUUID)
	// clientv3.WithPrefix() 表示监听所有以此key为前缀的键值对的变化
	return etcdclient.Watch(keyPrefix, clientv3.WithPrefix())
}
