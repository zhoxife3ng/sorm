package main

import (
	"context"
	"fmt"
	"github.com/x554462/go-exception"
	"github.com/x554462/sorm"
	"github.com/x554462/sorm/db"
	"github.com/x554462/sorm/example/model"
)

func main() {

	fmt.Println(reflect.TypeOf(model.Test{}).Name())

	db.Setup(db.Conf{
		Name:     "name",
		User:     "root",
		Password: "123456",
		Host:     "127.0.0.1",
		Port:     3306,
	})

	fmt.Println("建立连接成功")

	// 总查询时长限定60s
	//ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	//defer cancel()
	ctx := context.TODO()
	sess := sorm.NewSession(ctx)

	exception.TryCatch(func() {

		testModel := new(model.Test)

		// 通过selectById查询主键id为1的记录
		a := sess.GetDao(testModel).SelectById(1, sorm.Load())
		fmt.Println(a, a.(*model.Test).Name.Value())

		sess.BeginTransaction()
		// 在事务里面查询id为2的记录并加forupdate锁，调用Select方法会进行缓存，下次查询时直接返回缓存对象
		b := sess.GetDao(testModel).Select(false, 2).Load(sorm.ForUpdate())
		fmt.Println(b, b.(*model.Test).Name.Value())
		sess.SubmitTransaction()

		testDao := sess.GetDao(testModel)

		// 通过SelectOne查询单条记录
		c := testDao.SelectOne(map[string]interface{}{
			"id": 3,
		})
		fmt.Println(c, c.(*model.Test).Name.Value())

		// 更新一条记录
		c.Update(map[string]interface{}{
			"name": "test update",
		})

		// 删除一条记录
		c.Remove()

		// 通过SelectMulti查询多条记录
		for _, m := range testDao.SelectMulti(map[string]interface{}{
			"name": "test",
		}) {
			fmt.Println(m, m.(*model.Test).Id)
		}

	}, func(err exception.ErrorWrapper) {
		fmt.Printf("%d  %s  %s\n", err.Code(), err.Error(), err.Position())
		// 捕获未找到异常
	}, model.TestNotFoundError)
}
