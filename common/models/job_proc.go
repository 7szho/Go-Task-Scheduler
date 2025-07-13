package models

import (
	"encoding/json"
	"sync"
	"time"
)

// JobProcVal 存储作业进程的状态信息
type JobProcVal struct {
	Time   time.Time `json:"time"`    // 开启执行时间
	Killed bool      `json:"Kiddled"` // 是否强制杀死
}

// JobProc 表示一个作业进程，包括其基本信息和状态
type JobProc struct {
	ID         int    `json:"id"`        // 作业进程ID
	JobID      int    `json:"job_id"`    // 关联的作业ID
	NodeUUID   string `json:"node_uuid"` // 节点唯一标识
	JobProcVal        // 作业进程状态

	Running int32          // 运行状态标识
	Wg      sync.WaitGroup // 用于同步的等待组
}

// Val 返回 JobProcVal 的 JSON 字符串表示
func (p *JobProc) Val() (string, error) {
	b, err := json.Marshal(&p.JobProcVal)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
