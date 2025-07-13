package models

import (
	"crony/common/pkg/dbclient"
	"fmt"
)

const (
	RoleNormal = 1 // 普通用户角色
	RoleAdmin  = 2 // 管理员角色
)

// User 用户结构体，映射数据库表
type User struct {
	ID       int    `json:"id" gorm:"column:id;primary_key;auto_increment"`    // 用户ID，主键，自增
	UserName string `json:"username" gorm:"size:128;column:username;not null"` // 用户名
	Password string `json:"password" gorm:"size:128;column:password;not null"` // 密码
	Email    string `json:"email" gorm:"size:64;column:email;default:''"`      // 邮箱
	Role     int    `json:"role" gorm:"size:1;column:role;default:1"`          // 角色

	Created int64 `json:"created" gorm:"column:created;not null"`  // 创建时间
	Updated int64 `json:"updated" gorm:"column:updated;default:0"` // 更新时间
}

// Update 更新用户信息
func (u *User) Update() error {
	return dbclient.GetMysqlDB().Table(CronyUserTableName).Updates(u).Error
}

// Delete 删除用户
func (u *User) Delete() error {
	return dbclient.GetMysqlDB().Exec(fmt.Sprintf("delete from %s where id = ?", CronyUserTableName), u.ID).Error
}

// Insert 插入新用户
func (u *User) Insert() (insertId int, err error) {
	err = dbclient.GetMysqlDB().Table(CronyUserTableName).Create(u).Error
	if err == nil {
		insertId = u.ID
	}
	return
}

// FindById 根据ID查找用户
func (u *User) FindById() error {
	return dbclient.GetMysqlDB().Table(CronyUserTableName).Select("id", "username", "email", "role", "created", "updated").Where("id = ?", u.ID).First(u).Error
}

// TableName 返回用户表名
func (u *User) TableName() string {
	return CronyUserTableName
}
