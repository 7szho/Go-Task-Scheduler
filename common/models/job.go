package models

import (
	"crony/common/pkg/dbclient"
	"crony/common/pkg/utils"
	"crony/common/pkg/utils/errors"
	"encoding/json"
	"fmt"
	"strings"
)

// 任务类型
type JobType int

const (
	JobTypeCmd  = JobType(1) // 命令任务
	JobTypeHttp = JobType(2) // HTTP任务

	HttpMethodGet  = 1 // GET请求
	HttpMethodPost = 2 // POST请求

	JobExcSuccess = 1 // 执行成功
	JobExcFail    = 0 // 执行失败

	JobStatusNotAssigned = 0 // 未分配
	JobStatusAssigned    = 1 // 已分配

	ManualAllocation = 1 // 手动分配
	AutoAllocation   = 2 // 自动分配
)

// 注册到 /crony/job/<node_uuid>/<job_id>
type Job struct {
	ID      int    `json:"id" gorm:"column:id;primary_key;auto_increment"`                                 // 任务ID
	Name    string `json:"name" gorm:"size:64;column:name;not null;index:idx_job_name" binding:"required"` // 任务名称
	Command string `json:"command" gorm:"type:text;column:command;not null" binding:"required"`            // 执行命令
	// 预设脚本ID
	ScriptID      []byte `josn:"-" gorm:"size:256;column:script_id;default:null"` // 脚本ID（字节数组）
	ScriptIDArray []int  `json:"script_id" gorm:"-"`                              // 脚本ID数组
	// 任务执行超时时间设置，大于0时生效
	Timeout int64 `json:"timeout" gorm:"size:13;column:timeout;default:0"` // 超时时间
	// 任务执行失败重试次数，默认0
	RetryTimes int `json:"retry_times" gorm:"size:4,column:retry_times;default:0"` // 重试次数
	// 任务执行失败重试间隔，单位秒，小于0时立即重试
	RetryInterval int64   `json:"retry_interval" gorm:"size:10;column:retry_interval;default:0"`   // 重试间隔
	Type          JobType `json:"job_type" gorm:"size:1;column:type;not null;" binding:"required"` // 任务类型
	HttpMethod    int     `json:"http_method" gorm:"size:1;column:http_method"`                    // HTTP方法
	NotifyType    int     `json:"notify_type" gorm:"size:1;column:notify_type;not null"`           // 通知类型
	// 是否分配节点
	Status        int    `json:"status" gorm:"size:1;column:status;not null;default:0;index:idx_job_status"` // 状态
	NotifyTo      []byte `json:"-" gorm:"size:256;column:notify_to;default:null"`                            // 通知对象（字节数组）
	NotifyToArray []int  `json:"notify_to" gorm:"-"`                                                         // 通知对象数组
	Spec          string `json:"spec" gorm:"size:64;column:spec;not null"`                                   // 定时表达式
	RunOn         string `json:"run_on" gorm:"size:128;column:run_on;index:idx_job_run_on;"`                 // 运行节点
	Note          string `json:"note" gorm:"size:512;column:note;default:''"`                                // 备注
	Created       int64  `json:"created" gorm:"column:created;not null"`                                     // 创建时间
	Updated       int64  `json:"updated" gorm:"column:upddated;default:0"`                                   // 更新时间

	Hostname string   `json:"host_name" gorm:"-"` // 主机名
	Ip       string   `json:"ip" gorm:"-"`        // IP地址
	Cmd      []string `json:"cmd" gorm:"-"`       // 命令参数数组
}

// 初始化节点信息
func (j *Job) InitNodeInfo(status int, nodeUUID, hostname, ip string) {
	j.Status = status
	j.RunOn = nodeUUID
	j.Hostname = hostname
	j.Ip = ip
}

// 更新任务
func (j *Job) Update() error {
	return dbclient.GetMysqlDB().Table(CronyJobTableName).Updates(j).Error
}

// 删除任务
func (j *Job) Delete() error {
	return dbclient.GetMysqlDB().Exec(fmt.Sprintf("delete from %s where id = ?", CronyJobTableName), j.ID).Error
}

// 根据ID查找任务
func (j *Job) FindById() error {
	return dbclient.GetMysqlDB().Table(CronyJobTableName).Where("id = ?", j.ID).First(j).Error
}

// 校验任务参数
func (j *Job) Check() error {
	j.Name = strings.TrimSpace(j.Name)
	if len(j.Name) == 0 {
		return errors.ErrEmptyJobName
	}
	if j.RetryInterval == 0 {
		j.RetryTimes = 1
	}
	if len(strings.TrimSpace(j.Command)) == 0 {
		return errors.ErrEmptyJobCommand
	}
	if len(j.Cmd) == 0 && j.Type == JobTypeCmd {
		j.SplitCmd()
	}
	return nil
}

// 拆分命令字符串为命令和参数
func (j *Job) SplitCmd() {
	ps := strings.SplitN(j.Command, " ", 2)
	if len(ps) == 1 {
		j.Cmd = ps
		return
	}
	j.Cmd = make([]string, 0, 2)
	j.Cmd = append(j.Cmd, ps[0])
	j.Cmd = append(j.Cmd, utils.ParseCmdArguments(ps[1])...)
}

// 返回任务的JSON字符串
func (j *Job) Val() string {
	data, err := json.Marshal(j)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

// 返回表名
func (j *Job) TableName() string {
	return CronyJobTableName
}

// 反序列化通知对象和脚本ID
func (j *Job) Unmarshal() (err error) {
	if err = json.Unmarshal(j.NotifyTo, &j.NotifyToArray); err != nil {
		return
	}
	if err = json.Unmarshal(j.ScriptID, &j.ScriptIDArray); err != nil {
		return
	}
	return
}
