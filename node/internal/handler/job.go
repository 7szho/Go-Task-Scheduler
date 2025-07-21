package handler

import (
	"crony/common/models"
	"crony/common/pkg/config"
	"crony/common/pkg/etcdclient"
	"crony/common/pkg/logger"
	"crony/common/pkg/notify"
	"crony/common/pkg/utils"
	"crony/common/pkg/utils/errors"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/jakecoffman/cron"
)

// Job 结构体用于封装models.Job
type Job struct {
	*models.Job
}

// Jobs 是一个map，用于存储一组Job，其中键是作业的ID，值是指向Job实例的指针
type Jobs map[int]*Job

// JobKey 是一个辅助函数，用于生成一个etcd中唯一标识某个节点上特定作业的key
// nodeUUID：作业被分配执行的节点的唯一标识
// jobID：作业自身的唯一ID
func JobKey(nodeUUID string, jobId int) string {
	return fmt.Sprintf(etcdclient.KeyEtcdJob, nodeUUID, jobId)
}

// GetJobAndRev 函数用于从etcd中获取指定的任务以其元数据
func GetJobAndRev(nodeUUID string, jobId int) (job *Job, rev int64, err error) {
	// 调用etcd客户端的Get方法获取数据
	resp, err := etcdclient.Get(JobKey(nodeUUID, jobId))
	if err != nil {
		return
	}

	// 如果etcd中没有对应的键
	if resp.Count == 0 {
		err = errors.ErrNotFound
		return
	}

	// rev是etcd中该键的最后修改版本号
	rev = resp.Kvs[0].ModRevision
	// 将从etcd获取的JSON格式的value反序列化到结构体中
	if err = json.Unmarshal(resp.Kvs[0].Value, &job); err != nil {
		return
	}

	// 解析任务的命令字符串
	job.SplitCmd()
	return
}

// GetJobs 函数用于从etcd中获取制定节点上的所有任务
func GetJobs(nodeUUID string) (jobs Jobs, err error) {
	// 使用前缀查询获取该节点下的所有任务
	resp, err := etcdclient.Get(fmt.Sprintf(etcdclient.KeyEtcdJobProfile, nodeUUID), clientv3.WithPrefix())
	if err != nil {
		return
	}

	count := len(resp.Kvs)
	jobs = make(Jobs, count)
	if count == 0 {
		return
	}

	// 遍历所有获取到的键值对
	for _, j := range resp.Kvs {
		job := new(Job)
		// 反序列化任务数据
		if e := json.Unmarshal(j.Value, job); e != nil {
			logger.GetLogger().Warn(fmt.Sprintf("job[%s] unmarshal err: %s", string(j.Key), e.Error()))
			continue // 如果解析失败，记录警告并跳过此任务
		}
		// 检查任务数据的有效性
		if err := job.Check(); err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("job[%s] is invalid: %s", string(j.Key), err.Error()))
			continue // 如果数据无效，记录警告并跳过
		}
		// 将有效的任务存入jobs映射中
		jobs[job.ID] = job
	}
	return
}

// RunWithRecovery 一个安全执行任务的方法，包含panic恢复机制
func (j *Job) RunWithRecovery() {
	defer func() {
		// 使用defer和recover来捕获任务执行过程中的panic
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)] // 获取panic发生时的堆栈信息
			logger.GetLogger().Warn(fmt.Sprintf("panic running job: %v\n%s", r, buf))
		}
	}()
	t := time.Now()
	// 为执行创建一条日志记录
	jobLogId, err := j.CreateJobLog()
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("Failed to write to job log with jobID:%d nodeUUID: %s", j.ID, j.RunOn, err.Error()))
	}
	// 根据任务类型创建对应的执行处理器
	h := CreateHandler(j)
	if h == nil {
		return
	}
	// 执行任务
	result, runErr := h.Run(j)
	if runErr != nil {
		// 如果任务执行失败
		// 1. 更新任务日志为失败状态
		err = j.Fail(jobLogId, t, runErr.Error(), 0)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed to write to job log with jobID:%d nodeUUID: %s", j.ID, j.RunOn, err.Error()))
		}
		// 2. 准备发送失败通知
		node := &models.Node{UUID: j.RunOn}
		err = node.FindByUUID()
		var to []string
		// 遍历需要通知的用户ID，查询用户的联系方式
		for _, userId := range j.NotifyToArray {
			userModel := &models.User{ID: userId}
			err = userModel.FindById()
			if err != nil {
				continue
			}
			if j.NotifyType == notify.NotifyTypeMail {
				to = append(to, userModel.Email)
			} else if j.NotifyType == notify.NotifyTypeWebHook && config.GetConfigModels().WebHook.Kind == "feishu" {
				to = append(to, userModel.UserName)
			}
		}
		// 3. 构建通知消息体
		msg := &notify.Message{
			Type:      j.NotifyType,
			IP:        fmt.Sprintf("%s:%s", node.IP, node.PID),
			Subject:   fmt.Sprintf("任务[%s]立即执行失败", j.Name),
			Body:      fmt.Sprintf("job[%d] run on node[%s] oince execute failed, output: %s, eror:%s", j.ID, j.RunOn, result, runErr.Error()),
			To:        to,
			OccurTime: time.Now().Format(utils.TimeFormatSecond),
		}
		// 4. 异步发送通知
		go notify.Send(msg)
	} else {
		// 如果任务执行成功，更新日志为成功状态
		err = j.Success(jobLogId, t, result, 0)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed to write to job log with jobID:%d nodeUUID"))
		}
	}
}

// CreateJob 函数用于将一个Job对象包装成一个cron库可以执行的`cron.FuncJob`函数
func CreateJob(j *Job) cron.FuncJob {
	h := CreateHandler(j)
	if h == nil {
		return nil
	}
	// 返回一个闭包函数，这个函数就是cron调度器实际执行的内容
	jobFunc := func() {
		logger.GetLogger().Info(fmt.Sprintf("start the job#%s#command-%s", j.Name, j.Command))
		var execTimes int = 1
		if j.RetryTimes > 0 {
			// 计算总执行次数 = 1次正常执行 + N次重试
			execTimes += j.RetryTimes
		}
		var i = 0
		var output string
		var runErr error
		var err error
		var jobLogId int
		t := time.Now()
		// 创建初始的任务日志
		jobLogId, err = j.CreateJobLog()
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed to write to job with jobId:%s nodeUUID: %s error:%s", j.ID, j.RunOn, err.Error()))
		}
		// 循环执行，直到成功或达到最大次数
		for i < execTimes {
			output, runErr = h.Run(j)
			if runErr == nil {
				// 执行成功，更新日志并直接返回
				err = j.Success(jobLogId, t, output, i)
				if err != nil {
					logger.GetLogger().Warn(fmt.Sprintf("Failed to write to job log with jobID: %d nodeUUID: %s error: %s", j.ID, j.RunOn, err.Error()))
				}
			}
			i++
			if i < execTimes {
				// 如果还未达到最大次数，准备重试
				logger.GetLogger().Warn(fmt.Sprintf("job execution failure%jobId-%d %retry %d times #output-%s#error-%v", j.ID, i, output, runErr))
				if j.RetryInterval > 0 {
					time.Sleep(time.Duration(j.RetryInterval) * time.Second)
				} else {
					// 默认的重试是递增的，每次增加1分钟
					time.Sleep(time.Duration(i) * time.Minute)
				}
			}
		}
		// 所有尝试都失败后，更新日志为失败状态
		err = j.Fail(jobLogId, t, runErr.Error(), execTimes-1)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed to write to job with jobID:%d nodeID: %s error: %s", j.ID, j.RunOn, err.Error()))
		}
		// 发送最终失败的通知
		node := &models.Node{UUID: j.RunOn}
		err = node.FindByUUID()
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed to find node with jobID: %d nodeUUID: %s error:%s", j.ID, j.RunOn, err.Error()))
		}
		var to []string
		for _, userId := range j.NotifyToArray {
			userModel := &models.User{ID: userId}
			err = userModel.FindById()
			if err != nil {
				continue
			}
			if j.NotifyType == notify.NotifyTypeMail {
				to = append(to, userModel.Email)
			} else if j.NotifyType == notify.NotifyTypeWebHook && config.GetConfigModels().WebHook.Kind == "feishu" {
				to = append(to, userModel.UserName)
			}
		}
		msg := &notify.Message{
			Type:      j.NotifyType,
			IP:        fmt.Sprintf("%s:%s", node.IP, node.PID),
			Subject:   fmt.Sprintf("任务[%s]执行失败", j.Name),
			Body:      fmt.Sprintf("job[%d] run on node[%s] execute failed ,retry %d times ,output :%s, error:%v", j.ID, j.RunOn, j.RetryTimes, output, runErr),
			To:        to,
			OccurTime: time.Now().Format(utils.TimeFormatSecond),
		}
		go notify.Send(msg)
	}
	return jobFunc
}

// WatchJobs 函数用于在etcd上为指定节点的任务创建一个监视器
func WatchJobs(nodeUUID string) clientv3.WatchChan {
	// 监视指定前缀下的所有键值变化
	return etcdclient.Watch(fmt.Sprintf(etcdclient.KeyEtcdJobProfile, nodeUUID), clientv3.WithPrefix())
}

// GetJobIDFromKey 是一个工具函数，用于从etcd的key中解析出任务ID
func GetJobIDFromKey(key string) int {
	// 任务ID通常是key路径的最后一部分
	index := strings.LastIndex(key, "/")
	if index < 0 {
		return 0
	}
	// 将字符串ID转换为整数
	jobId, err := strconv.Atoi(key[index+1:])
	if err != nil {
		return 0
	}
	return jobId
}

// CreateJobLog 方法用于为任务的一次执行创建一个日志条目
func (j *Job) CreateJobLog() (int, error) {
	start := time.Now()
	jobLog := &models.JobLog{
		Name:      j.Name,
		JobId:     j.ID,
		Command:   j.Command,
		IP:        j.Ip,
		Hostname:  j.Hostname,
		NodeUUID:  j.RunOn,
		Spec:      j.Spec,
		StartTime: start.Unix(),
	}
	// 将日志插入数据库并返回新日志的ID
	return jobLog.Insert()
}

// UpdateJobLog 函数用于更新制定的任务日志条目
func UpdateJobLog(jobLogId int, start time.Time, output string, retry int, success bool) error {
	end := time.Now()
	jobLog := &models.JobLog{
		ID:         jobLogId,
		StartTime:  start.Unix(),
		RetryTimes: retry,      // 记录重试次数
		Success:    success,    // 记录成功或失败
		Output:     output,     // 记录输出或错误信息
		EndTime:    end.Unix(), // 记录结束时间
	}
	// 更新数据库中的日志记录
	return jobLog.Update()
}

// Success 是一个辅助方法，用于将任务日志标记为成功
func (j *Job) Success(jobLogId int, start time.Time, output string, retry int) error {
	return UpdateJobLog(jobLogId, start, output, retry, true)
}

// Fail 是一个辅助方法，用于将任务日志标记为失败
func (j *Job) Fail(jobLogId int, start time.Time, errMsg string, retry int) error {
	return UpdateJobLog(jobLogId, start, errMsg, retry, false)
}
