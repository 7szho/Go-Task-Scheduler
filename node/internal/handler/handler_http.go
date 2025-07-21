package handler

import (
	"crony/common/models"
	"crony/common/pkg/httpclient"
	"strings"
	"time"
)

// HTTPHandler 是一个用于执行HTTP类型任务的处理器
type HTTPHandler struct{}

// HttpExecTimeout 定义了HTTP请求的最大允许超时时间
const HttpExecTimeout = 300

// Run 方法实现了Handler接口，负责执行一个HTTP类型的作业
func (h *HTTPHandler) Run(job *Job) (result string, err error) {
	var proc *JobProc
	// 初始化一个JobProc来追踪此次HTTP任务的执行状态
	proc = &JobProc{
		JobProc: &models.JobProc{
			ID:       0, // HTTP任务没有操作系统进程ID
			JobID:    job.ID,
			NodeUUID: job.RunOn,
			JobProcVal: models.JobProcVal{
				Time: time.Now(), // 记录任务开始执行的时间
			},
		},
	}

	// 记录任务开始执行
	err = proc.Start()
	if err != nil {
		return // 如果记录失败，直接返回错误
	}
	// 使用defer确保函数退出时，会调用proc.Stop()来清理执行记录
	defer proc.Stop()
	// 检查并修正任务的超时设置
	if job.Timeout <= 0 || job.Timeout > HttpExecTimeout {
		job.Timeout = HttpExecTimeout
	}
	// 根据job中定义的HTTP方法执行不同的逻辑
	if job.HttpMethod == models.HttpMethodGet {
		// 如果是GET请求，直接调用httpclient的Get方法
		// job.Command字段此时应包含完整的URL（包括查询参数）
		result, err = httpclient.Get(job.Command, job.Timeout)
	} else {
		// 否则，默认为POST请求
		// 在Command字段中，使用'?'来分割URL和POST的body数据
		urlFields := strings.Split(job.Command, "?")
		url := urlFields[0]
		var body string
		if len(urlFields) >= 2 {
			body = urlFields[1] // '?'之后的部分被用作POST请求的body

		}
		// 调用httpclient的PostJson方法发送请求
		result, err = httpclient.PostJson(url, body, job.Timeout)
	}
	// 返回result和err
	return
}
