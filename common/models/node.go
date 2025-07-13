package models

import (
	"crony/common/pkg/dbclient"
	"fmt"
)

const (
	NodeConnSuccess      = 1       // 节点连接成功
	NodeConnFail         = 2       // 节点连接失败
	NodeSystemInfoSwitch = "alive" // 节点系统信息开关
)

// 注册到 /crony/node/<node_uuid>
type Node struct {
	ID       int    `json:"id" gorm:"column:id;primary_key;auto_increment"`                // 主键ID
	PID      string `json:"pid" gorm:"size:16;column:pid;not null"`                        // 进程ID
	IP       string `json:"ip" gorm:"size:32;column:ip;default:''"`                        // IP地址
	Hostname string `json:"hostname" gorm:"size64;column:hostname;default:''"`             // 主机名
	UUID     string `json:"uuid" gorm:"size:128;column:uuid;not null;index:idx_node_uuid"` // 节点唯一标识
	Version  string `json:"version" gorm:"size:64;column:version;default:''"`              // 版本号
	Status   int    `json:"statur" gorm:"size:1;column:status"`                            // 状态

	UpTime   int64 `json:"up" gorm:"column:up;not null"`      // 上线时间
	DownTime int64 `json:"down" gorm:"column:down;default:0"` // 下线时间
}

// 返回节点和PID的字符串表示
func (n *Node) String() string {
	return "node[" + n.UUID + "] pid[" + n.PID + "]"
}

// 插入节点数据
func (n *Node) Insert() (insertId int, err error) {
	err = dbclient.GetMysqlDB().Table(CronyNodeTableName).Create(n).Error
	if err == nil {
		insertId = n.ID
	}
	return
}

// 更新节点数据
func (n *Node) Update() error {
	return dbclient.GetMysqlDB().Table(CronyNodeTableName).Updates(n).Error
}

// 删除节点数据
func (n *Node) Delete() error {
	return dbclient.GetMysqlDB().Exec(fmt.Sprintf("delete from %s where uuid = ?", CronyNodeTableName), n.UUID).Error
}

// 根据UUID查找节点
func (n *Node) FindByUUID() error {
	return dbclient.GetMysqlDB().Table(CronyNodeTableName).Where("uuid = ?", n.UUID).First(n).Error
}

// 返回表名
func (n *Node) TableName() string {
	return CronyNodeTableName
}
