package main

import (
	"context"
	"fmt"
	"github.com/x554462/sorm"
	"github.com/x554462/sorm/db"
	"github.com/x554462/sorm/example/model"
	"reflect"
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

	a := sess.GetDao(&model.Test{}).Select(false, 2).Load(false)
	fmt.Println(a)
	fmt.Println(a.(*model.Test).Name.Value())

}
