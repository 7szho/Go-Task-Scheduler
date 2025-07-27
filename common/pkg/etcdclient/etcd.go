package etcdclient

import (
	"context"
	"crony/common/pkg/config"
	"crony/common/pkg/logger"
	"crony/common/pkg/utils/errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
)

var _defaultEtcd *Client

type Client struct {
	*clientv3.Client
	reqTimeout time.Duration // 管理每次请求的超时时间
}

// Init 函数负责初始化 etcd 客户端
func Init(endpoints []string, dialTimeout, reqTimeout int64) (*Client, error) {
	// 创建一个 etcd 客户端实例
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,                                // etcd 集群的地址列表
		DialTimeout: time.Duration(dialTimeout) * time.Second, // 建立连接的超时时间
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err: %v \n", err)
		return nil, err
	}
	// 如果连接成功, 则创建自定义的 Client 实例
	_defaultEtcd = &Client{
		Client:     cli,                                     // 将官方客户端实例内嵌进来
		reqTimeout: time.Duration(reqTimeout) * time.Second, // 设置请求超时时间
	}
	return _defaultEtcd, nil
}

// getter 函数, 返回一个 etcd 客户端实例
func GetEtcdClient() *Client {
	if _defaultEtcd == nil {
		logger.GetLogger().Error("etcd is not initialized")
		return nil
	}
	return _defaultEtcd
}

/* --- 下面是一系列对 etcd 常用操作的封装函数 --- */

// 对原始 Put 方法封装, 简化了 context 的创建, 并提供初始化检查
func Put(key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	if _defaultEtcd == nil {
		return nil, errors.ErrEtcdNotInit // 返回一个预定义的错误
	}
	// 使用 NewEtcdTimeoutContext 创建一个带超时和错误信息的 context
	ctx, cancel := NewEtcdTimeoutContext()
	defer cancel() // 确保在函数退出时取消 context, 释放资源
	// 调用官方客户端的 Put 方法
	return _defaultEtcd.Put(ctx, key, val, opts...)
}

// 将一个键值对与一个租约(Lease)绑定
// 当租约过期后, 这个简直会自动被 etcd 删除
// ttl: 租约的生命周期, 单位为秒
func PutWithTtl(key, val string, ttl int64) (*clientv3.PutResponse, error) {
	if _defaultEtcd == nil {
		return nil, errors.ErrEtcdNotInit
	}
	ctx, cancel := NewEtcdTimeoutContext()
	defer cancel()

	// 申请一个lease(租约)
	leaseRsp, err := Grant(ttl)
	if err != nil {
		return nil, err
	}
	// 在 Put 操作时, 使用 clientv3.WithLease 将键值对与租约 ID 关联起来
	return _defaultEtcd.Put(ctx, key, val, clientv3.WithLease(leaseRsp.ID))
}

// PutWithModRev 实现了基于版本号的 CAS(Compare-And-Swap) 更新
// 只有当 key 的当前 ModRevision 与传入的 rev 相同时, 更新才会成功
// 这样可以防止多个客户端同时修改同一个 key 时发生数据覆盖
func PutWithModRev(key, val string, rev int64) (*clientv3.PutResponse, error) {
	if _defaultEtcd == nil {
		return nil, errors.ErrEtcdNotInit
	}
	// 如果版本号为0, 进行普通的 Put 操作
	if rev == 0 {
		return Put(key, val)
	}

	ctx, cancel := NewEtcdTimeoutContext()
	// 使用 etcd 事务来实现 CAS
	tresp, err := _defaultEtcd.Txn(ctx).
		// If 条件: key 的 ModRevision 等于 rev
		If(clientv3.Compare(clientv3.ModRevision(key), "=", rev)).
		// Then 操作: 如果 If 成功, 则执行 Put 操作
		Then(clientv3.OpPut(key, val)).
		// 提交事务
		Commit()
	cancel()

	if err != nil {
		return nil, err
	}

	// 检查事务是否成功执行了 Then() 部分
	if !tresp.Succeeded {
		return nil, errors.ErrValueMayChanged // 返回一个表示值已改变的错误
	}

	// 事务成功后, 从响应中提取出 PutResponse 并返回
	resp := clientv3.PutResponse(*tresp.Responses[0].GetResponsePut())
	return &resp, nil
}

func Get(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	if _defaultEtcd == nil {
		return nil, errors.ErrEtcdNotInit
	}
	ctx, cancel := NewEtcdTimeoutContext()
	defer cancel()
	return _defaultEtcd.Get(ctx, key, opts...)
}

func Delete(key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	if _defaultEtcd == nil {
		return nil, errors.ErrEtcdNotInit
	}
	ctx, cancel := NewEtcdTimeoutContext()
	defer cancel()
	return _defaultEtcd.Delete(ctx, key, opts...)
}

func Watch(key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return _defaultEtcd.Watch(context.Background(), key, opts...)
}

// Grant 封装了申请租约的操作
func Grant(ttl int64) (*clientv3.LeaseGrantResponse, error) {
	if _defaultEtcd == nil {
		return nil, errors.ErrEtcdNotInit
	}
	ctx, cancel := NewEtcdTimeoutContext()
	defer cancel()
	return _defaultEtcd.Grant(ctx, ttl)
}

// Revoke 封装了撤销租约的操作
// 撤销租约后, 与该租约关联的所有键值对都会被删除
func Revoke(id clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error) {
	if _defaultEtcd == nil {
		return nil, errors.ErrEtcdNotInit
	}
	ctx, cancel := context.WithTimeout(context.Background(), _defaultEtcd.reqTimeout)
	defer cancel()
	return _defaultEtcd.Revoke(ctx, id)
}

// GetLock 尝试获取一个分布式锁
// 它利用 etcd 的事务 和 "CreateRevision" 实现了一个 "Create-If-Not-Exists" 的原子操作
// key: 锁的名称
// id: 租约 ID. 锁会与这个租约绑定, 如果持有锁的客户端崩溃, 租约到期后锁会自动释放
func GetLock(key string, id clientv3.LeaseID) (bool, error) {
	if _defaultEtcd == nil {
		return false, errors.ErrEtcdNotInit
	}
	// 使用预定义的锁前缀来构造完整的锁key
	key = fmt.Sprintf(KeyEtcdLock, key)
	ctx, cancel := NewEtcdTimeoutContext()
	defer cancel()

	// 启动一个事务
	resp, err := _defaultEtcd.Txn(ctx).
		// If 条件: key 的 CreateRevision 等于 0, 即 key 不存在
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		// Then 操作: 如果 key 不存在, 就创建它, 并绑定租约 id
		Then(clientv3.OpPut(key, "", clientv3.WithLease(id))).
		// 提交事务
		Commit()

	if err != nil {
		return false, err
	}

	// resp.Succeeded 为 true 意味着 If 条件满足, Then 操作被执行, 即成功获取了锁
	return resp.Succeeded, nil
}

// DelLock 用于删除一个锁
func DelLock(key string) error {
	_, err := Delete(fmt.Sprintf(KeyEtcdLock, key))
	return err
}

// IsValidAsKeyPath 检查一个字符串是否包含路径分隔符
func IsValidAsKeyPath(s string) bool {
	return strings.IndexAny(s, "/\\") == -1
}

/* --- 下面是自定义 context 的实现 --- */

type etcdTimeoutContext struct {
	context.Context
	etcdEndpoints []string
}

func (c *etcdTimeoutContext) Err() error {
	// 调用 context.Context 的原始 Err() 方法获取错误
	err := c.Context.Err()
	// 检查返回错误是否为超时错误
	if err == context.DeadlineExceeded {
		err = fmt.Errorf("%s: etcd(%v) maybe lost",
			err, c.etcdEndpoints)
	}
	return err
}

// 工厂函数, 用于创建自定义的 etcdTimoutContext
// 封装了创建和初始化的细节, 可以像使用标准 context.WithTimeout 一样方便地使用
func NewEtcdTimeoutContext() (context.Context, context.CancelFunc) {
	// 1. 从配置中获取 etcd 的请求超时时间
	ctx, cancel := context.WithTimeout(context.Background(), _defaultEtcd.reqTimeout)
	// 2. 创建自定义的 etcdTimeoutContext 实例
	etcdCtx := &etcdTimeoutContext{}
	// 3. 将刚刚创建的标准 context 赋值给 etcdCtx 的内嵌字段
	etcdCtx.Context = ctx
	// 4. 从全局配置获取 etcd 的端点地址, 并存储到 etcdCtx 的自定义字段中
	etcdCtx.etcdEndpoints = config.GetConfigModels().Etcd.Endpoints
	// 5. 返回打包后的 etcdCtx 和原始的 cancel 函数
	return etcdCtx, cancel
}
