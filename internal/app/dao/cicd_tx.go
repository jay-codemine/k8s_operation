package dao

import (
	"context"
	"gorm.io/gorm"
)

//func (d *Dao) WithTx(ctx context.Context, fn func(tx *Dao) error) error {
//	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
//		return fn(&Dao{db: tx})
//	})
//}

func (d *Dao) WithTx(ctx context.Context, fn func(tx *Dao) error) error {
	// 1) 先拿到带 ctx 的 db（不是事务，只是把 context 绑定进去）
	dbWithCtx := d.db.WithContext(ctx)

	// 2) 定义一个“包装函数”：把 *gorm.DB → *dao.Dao
	wrapper := func(gormTx *gorm.DB) error {
		daoTx := &Dao{db: gormTx} // 用事务 tx 包一层 Dao
		return fn(daoTx)          // 调用你传进来的业务回调
	}

	// 3) 交给 gorm 开事务并执行 wrapper
	err := dbWithCtx.Transaction(wrapper)
	return err
}
