package models

import (
	"crony/common/pkg/dbclient"
	"crony/common/pkg/utils"
	"crony/common/pkg/utils/errors"
	"fmt"
	"strings"
)

// Preset Script
// Script 结构体定义了预设脚本的数据模型
type Script struct {
	ID      int    `json:"id" gorm:"column:id;primary_key:auto_increment"`                                     // 脚本ID
	Name    string `json:"name" gorm:"size:256;column:name;not null;index:idx_script_name" binding:"required"` // 脚本名称
	Command string `json:"command" gorm:"type:text;column:command:not null" binding:"required"`                // 脚本执行的命令
	Created int64  `json:"created" gorm:"column:created;not null"`                                             // 传建时间的时间戳
	Updated int64  `json:"updated" gorm:"column: updated;default:0"`                                           // 更新时间的时间戳

	Cmd []string `json:"cmd" gorm:"-"` // 用于存储分割后的命令和参数
}

// Insert 向数据库中插入一条新的脚本记录
func (s *Script) Insert() (insertId int, err error) {
	// 使用GORM的Create方法传建记录
	err = dbclient.GetMysqlDB().Table(CronyScriptTableName).Create(s).Error
	if err == nil {
		// 返回插入记录的ID
		insertId = s.ID
	}
	return
}

// Update 更新数据库中已有的脚本记录
func (s *Script) Update() error {
	return dbclient.GetMysqlDB().Table(CronyScriptTableName).Updates(s).Error
}

// Delete 从数据库中删除一条脚本记录
func (s *Script) Delete() error {
	// 使用GORM的Updates方法更新记录
	return dbclient.GetMysqlDB().Exec(fmt.Sprintf("delete from %s where id =?", CronyScriptTableName), s.ID).Error
}

// FindById 根据ID查找一条脚本记录
func (s *Script) FindById() error {
	// 执行Sql删除语句
	return dbclient.GetMysqlDB().Table(CronyScriptTableName).Where("id = ?", s.ID).First(s).Error
}

// Table 返回此模型对应的数据库表名
func (s *Script) TableName() string {
	// 使用GORM的First方法根据ID查找记录，并将结果填充到s结构体中
	return CronyScriptTableName
}

// Check 检查脚本字段的有效性
func (s *Script) Check() error {
	// 去除命令名称前后的空格
	s.Name = strings.TrimSpace(s.Name)
	if len(s.Name) == 0 {
		// 如果命令名称为空，返回错误
		return errors.ErrEmptyScriptName
	}
	// 去除命令前后的空格
	if len(strings.TrimSpace(s.Command)) == 0 {
		return errors.ErrEmptyScriptCommand
	}
	// 如果Cmd字段为空，则调用SplitCmd来分割命令
	if len(s.Cmd) == 0 {
		s.SplitCmd()
	}
	return nil
}

// SplitCmd 将Command字符串分割成命令和参数的切片
func (s *Script) SplitCmd() {
	// 使用SplitN按空格最多分割成两部分
	ps := strings.SplitN(s.Command, " ", 2)
	// 如果分割后只有一部分，说明没有参数
	if len(ps) == 1 {
		s.Cmd = ps
		return
	}
	// 初始化Cmd切片
	s.Cmd = make([]string, 0, 2)
	// 第一条元素是命令
	s.Cmd = append(s.Cmd, ps[0])
	// 解析并追加其余的参数
	s.Cmd = append(s.Cmd, utils.ParseCmdArguments(ps[1])...)
}
