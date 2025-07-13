package models

import (
	"crony/common/pkg/dbclient"
	"fmt"
)

// JobLog 结构体表示作业日志表
type JobLog struct {
	ID       int    `json:"id" gorm:"column:id;primary_key;auto_increment"`                             // 主键，自增
	Name     string `json:"name" gorm:"size:64;column:name;index:idx_job_log_name;not null"`            // 作业名称
	JobId    int    `json:"job_id" gorm:"column:job_id;index:idx_job_log_id; not null"`                 // 作业ID
	Command  string `json:"command" gorm:"size:512;column:command"`                                     // 执行命令
	IP       string `json:"ip" gorm:"size:32;column:ip"`                                                // 节点IP
	Hostname string `json:"hostname" gorm:"size:32;column:hostname"`                                    // 节点主机名
	NodeUUID string `json:"node_uuid" gorm:"size:128;column:node_uuid;not null;index:idx_job_log_node"` // 节点唯一标识
	Success  bool   `json:"success" gorm:"size:1;column:success;not null"`                              // 是否成功

	Output string `json:"output" gorm:"size:512;column:output;"`     // 执行输出
	Spec   string `json:"spec" gorm:"size:64;column:spec;not null" ` // 定时表达式

	RetryTimes int   `json:"retry_times" gorm:"size:4;column:retry_times;default:0"` // 重试次数
	StartTime  int64 `json:"start_time" gorm:"column:start_time;not null;"`          // 开始时间
	EndTime    int64 `json:"end_time" gorm:"column:end_time;default:0;"`             // 结束时间
}

// Update 更新当前作业日志
func (jb *JobLog) Update() error {
	return dbclient.GetMysqlDB().Table(CronyJobLogTableName).Updates(jb).Error
}

// Delete 删除当前作业日志
func (jb *JobLog) Delete() error {
	return dbclient.GetMysqlDB().Exec(fmt.Sprintf("delete from %s where id = ?", CronyJobLogTableName), jb.ID).Error
}

// Insert 插入新的作业日志
func (jb *JobLog) Insert() (insertId int, err error) {
	err = dbclient.GetMysqlDB().Table(CronyJobLogTableName).Create(jb).Error
	if err == nil {
		insertId = jb.ID
	}
	return
}

// TableName 返回作业日志表名
func (jb *JobLog) TableName() string {
	return CronyJobLogTableName
}
