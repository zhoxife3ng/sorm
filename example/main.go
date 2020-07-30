package main

import (
	"context"
	"fmt"
	"github.com/x554462/sorm"
	"github.com/x554462/sorm/db"
	"github.com/x554462/sorm/example/model"
	"time"
)

func setupDb() {
	db.Setup(db.Conf{
		Name:     "db",
		User:     "root",
		Password: "123456",
		Host:     "127.0.0.1",
		Port:     3306,
	})

	fmt.Println("建立连接成功")
}

func main() {

	setupDb()
	// 总查询时长限定60s
	//ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	//defer cancel()
	ctx := context.TODO()
	sess := sorm.NewSession(ctx)
	testD := sess.GetDao(new(model.Test))
	fmt.Println(testD.Insert(map[string]interface{}{
		"name": "test2",
		"time": time.Now(),
	}))
	a, err := testD.SelectById(66, sorm.Load())
	if err != nil {
		if err == model.TestNotFoundError {
			fmt.Println("记录未找到")
		} else {
			fmt.Println(err)
		}
	} else {
		fmt.Println(a.(*model.Test).Time.Value())
	}
}
