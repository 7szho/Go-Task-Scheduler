package handler

import (
	"bytes"
	"context"
	"crony/common/models"
	"crony/common/pkg/logger"
	"fmt"
	"os/exec"
	"time"
)

// CMDHandler 结构体用于处理命令执行
type CMDHandler struct{}

// Run 方法负责执行一个job
func (c *CMDHandler) Run(job *Job) (result string, err error) {
	var (
		cmd  *exec.Cmd // 用于表示一个外部命令
		proc *JobProc  // 用于追踪正在运行的工作进程
	)
	// 如果设定了超时时间，则创建一个带有超时的context
	if job.Timeout > 0 {
		// 建立一个带有超时的context
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(job.Timeout)*time.Second)
		defer cancel() // 确保在函数结束时取消context，释放资源
		// 使用带有context的CommandContext来创建命令，如果context被取消，命令也被终止
		cmd = exec.CommandContext(ctx, job.Cmd[0], job.Cmd[1:]...)
	} else {
		// 如果没有设置超时，则创建一个常规的命令
		cmd = exec.Command(job.Cmd[0], job.Cmd[1:]...)
	}
	// 创建一个bytes.Buffer，用于捕获命令的标准输出和标准错误
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	// 异步启动命令
	err = cmd.Start()
	// 即使 Start() 出错，缓冲区中也可能已经有输出
	result = b.String()
	if err != nil {
		// 若果启动命令时出错，记录错误并返回
		logger.GetLogger().Error(fmt.Sprintf("%s\n%s", b.String(), err.Error()))
		return
	}

	// 创建一个新的JobProc实例来追踪这个进程
	proc = &JobProc{
		JobProc: &models.JobProc{
			ID:       cmd.Process.Pid, // 获取进程的ID
			JobID:    job.ID,          // 关联的Job ID
			NodeUUID: job.RunOn,       // 运行该作业的节点UUID
			JobProcVal: models.JobProcVal{
				Time:   time.Now(), // 记录当前时间
				Killed: false,      // 初始化为为被杀死状态
			},
		},
	}
	// 启动进程追踪
	err = proc.Start()
	if err != nil {
		return // 如果追踪失败，则返回错误
	}
	defer proc.Stop() // 确保在函数退出时停止进程追踪
	if err = cmd.Wait(); err != nil {
		// 如果命令执行出错，记录错误
		logger.GetLogger().Error(fmt.Sprintf("%s%s", b.String(), err.Error()))
		// 返回输出和错误
		return b.String(), err
	}
	return b.String(), nil
}

// RunPresetScript 函数用于执行一个预设的脚本
func RunPresetScript(script *models.Script) (result string, err error) {
	var cmd *exec.Cmd
	// 从script模型来创建命令
	cmd = exec.Command(script.Cmd[0], script.Cmd[1:]...)
	// 创建一个bytes.Buffer，用于捕获命令的标准输出和标准错误
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	// 异步启动命令
	err = cmd.Start()
	// 即使 Start() 出错，缓冲区中也可能已经有输出
	result = b.String()
	if err != nil {
		// 如果启动时出错，记录错误并返回
		logger.GetLogger().Error(fmt.Sprintf("run preset sctipt:%s\n%s", b.String(), err.Error()))
		return
	}
	// 等待命令执行完成
	if err = cmd.Wait(); err != nil {
		// 如果命令执行出错，记录错误
		logger.GetLogger().Error(fmt.Sprintf("run preset script:%s%s", b.String(), err.Error()))
		// 返回输出和错误
		return b.String(), err
	}
	// 返回命令的输出和nil错误
	return b.String(), nil
}
