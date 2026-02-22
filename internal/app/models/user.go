package models

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"k8soperation/global"
	"k8soperation/internal/errorcode"
	"strconv"
	"time"
)

type User struct {
	Username string `json:"username" description:"用户名"`
	Password string `json:"-" description:"密码"`
	*Base
}

func NewUser() *User {
	return &User{}
}

func (u *User) TableName() string {
	return "user"
}

// 注册用户
func (u *User) Create(db *gorm.DB) error {
	return db.Create(u).Error
}

// Delete 删除用户
// Delete 方法用于删除用户（实际上是标记为已删除）
// 参数:
//   - db: 数据库连接对象
//
// 返回值:
//   - error: 操作过程中出现的错误
//
// Delete 方法用于删除用户，实际上是软删除，通过标记删除字段实现
// Delete 方法用于删除用户，实际上是软删除操作
// 参数:
//
//	db: *gorm.DB 类型的数据库连接对象
//
// 返回值:
//
//	error: 操作过程中发生的错误
func (u *User) Delete(db *gorm.DB) error {
	// 判断用户是否被删除
	// 声明一个User结构体变量user
	var user User
	// 从数据库中查询记录
	// 条件：id字段等于u.ID且is_del字段等于0
	// First方法用于查询第一条匹配的记录并将结果赋值给user变量
	// 如果查询过程中发生错误，将错误返回
	if err := db.Where("id=? and is_del=?", u.ID, 0).First(&user).Error; err != nil {
		return err
	}

	// 获取当前时间戳，转换为uint32类型
	nowTime := uint32(time.Now().Unix())

	// 标记删除：设置用户为删除状态
	user.IsDel = 1            // 设置删除标记为1，表示已删除
	user.DeletedAt = nowTime  // 记录删除时间
	user.ModifiedAt = nowTime // 更新最后修改时间

	// 执行数据库更新操作
	if err := db.Updates(&user).Error; err != nil {
		return err // 如果更新失败，返回错误
	}

	// 返回 nil，表示操作成功或无错误
	// 在 Go 语言中，nil 通常用于表示接口、函数、指针、映射、切片和通道的零值
	// 当函数返回 nil 时，通常表示操作成功完成或没有错误发生
	return nil
}

// Update 方法用于更新用户信息
// 参数:
//   - db: 数据库连接对象，使用 GORM 进行数据库操作
//   - values: 包含要更新字段的接口类型参数，通常是一个 map 或结构体
//
// 返回值:
//   - error: 操作成功返回 nil，失败返回错误信息
func (u *User) Update(db *gorm.DB, values interface{}) error {
	// 创建一个数据库事务 tx，用于更新用户模型 u 的数据
	// 使用 Where 子句指定更新条件：用户 ID 匹配且未被删除 (is_del=0)
	// Updates 方法用于执行实际的更新操作，传入要更新的字段值 values
	tx := db.Model(u).
		Where("id=? AND is_del=?", u.ID, 0).
		Updates(values)

	// 检查 SQL 执行是否出错
	// 如果 tx.Error 不为 nil，表示 SQL 执行失败，返回错误信息
	if tx.Error != nil {
		return tx.Error // SQL 执行失败
	}
	// 检查是否有记录被更新
	// 如果 tx.RowsAffected 为 0，表示没有找到符合条件的记录，返回更新失败错误
	if tx.RowsAffected == 0 {
		return errorcode.ErrorUserUpdateFail // 没有找到要更新的记录
	}
	// 如果一切正常，返回 nil 表示更新成功
	return nil // 更新成功
}

// List 是User结构体的方法，用于查询用户列表
// 参数:
//
//	db: 数据库连接对象
//	page: 页码，从1开始
//	limit: 每页记录数
//
// 返回值:
//
//	[]*User: 用户列表
//	error: 错误信息
func (u *User) List(db *gorm.DB, page, limit int) ([]*User, error) {
	// 校验页码，最小为1
	if page < 1 {
		page = 1
	}
	// 校验每页记录数，限制在1-1000之间，默认为20
	if limit <= 0 || limit > 1000 {
		limit = 20
	} // 上限防止一次拉太多

	var users []*User

	// 1) 统一构建查询条件
	q := db.Model(&User{}).Where("is_del = 0") // 只查询未删除的用户
	if u.Username != "" {
		q = q.Where("username = ?", u.Username) // 如果指定了用户名，则添加用户名条件
	}

	// 2) 排序 + 分页 + 一次性查询
	// 计算分页偏移量，(当前页码-1)每页条数
	offset := (page - 1) * limit // 计算偏移量
	// 执行数据库查询：
	// 1. 按id降序排序
	// 2. 设置查询偏移量
	// 3. 限制返回记录数
	// 4. 将结果存入users变量
	// 如果查询过程中发生错误，则返回错误信息

	global.Logger.Info("分页参数",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	if err := q.Order("id DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, err // 返回空结果和错误信息
	}
	// 查询成功，返回用户列表和nil错误
	return users, nil
}

func (u *User) GetByName(db *gorm.DB) (*User, error) {
	var user User
	if u.Username == "" {
		return nil, nil
	}
	err := db.Where("username = ?", u.Username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 通过ID查找用户
func (u *User) GetUserByID(id string) User {
	var user = NewUser()
	global.DB.Where("id", id).First(&user)
	return *user
}

// GetStringID 是 User 结构体的一个方法，用于将用户ID转换为字符串格式
// 接收者 u 指向 User 结构体的指针
// 返回值类型为 string
func (u *User) GetStringID() string {
	// 使用 strconv 包的 FormatUint 函数将 uint64 类型的 ID 转换为十进制字符串表示
	// 首先将 u.ID 转换为 uint64 类型，然后以10为基数进行格式化
	return strconv.FormatUint(uint64(u.ID), 10)
}

func (u *User) UpdatePasswordByName(db *gorm.DB, newPassword string) error {
	return db.Model(&User{}).
		Where("username = ?", u.Username).
		Update("password", newPassword).
		Error
}
