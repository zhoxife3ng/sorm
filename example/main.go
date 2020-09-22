package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xkisas/sorm"
	"github.com/xkisas/sorm/db"
	_type "github.com/xkisas/sorm/type"
)

var TestNotFoundError = sorm.NewError(sorm.ModelNotFoundError, "Test记录未找到")

type Test struct {
	sorm.BaseModel
	Id   int          `db:"id,pk"`
	Name _type.String `db:"name"`
	Time _type.Time   `db:"time"`
}

func (t *Test) GetNotFoundError() error {
	return TestNotFoundError
}

func (t *Test) CustomDao() sorm.DaoIfe {
	return &TestDao{}
}

func (t *Test) GetDao() *TestDao {
	return t.GetDaoIfe().(*TestDao)
}

type TestDao struct {
	sorm.Dao
}

func GetTestDao(sess *sorm.Session) *TestDao {
	return sess.GetDao(new(Test)).(*TestDao)
}

func (td *TestDao) Count() (int, error) {
	return td.GetCount("*", map[string]interface{}{})
}

func setupDb() {
	db.Setup(
		db.Conf{
			Name:     "db",
			User:     "root",
			Password: "123456",
			Host:     "127.0.0.1",
			Port:     3306,
		},
		db.Loc(time.Local),
		db.ParseTime(true),
		db.AllowCleartextPasswords(true),
		db.InterpolateParams(true),
		db.Param("charset", "utf8"),
	)

	fmt.Println("建立连接成功")
}

func main() {

	setupDb()

	sess := sorm.NewSession(context.TODO())
	defer sess.Close()

	testD := GetTestDao(sess) //获取操作该对象的dao

	fmt.Println(testD.Count())
	fmt.Println(testD.GetSum("id", map[string]interface{}{}))

	//插入一条记录，并返回对象
	model, err := testD.Insert(map[string]interface{}{
		"name": "test",
		"time": time.Now(),
	})
	fmt.Println(model, err) //&Test{...}    nil
	fmt.Println(model.GetDaoIfe().(*TestDao).Count())
	id := model.GetId() //得到刚刚插入的自增id

	//删除刚插入的记录
	model.Remove()

	//查询刚刚删除的记录(立即查询数据库)
	model, err = testD.SelectById(id, sorm.Load())
	fmt.Println(model, err) //nil    Test记录未找到

	//插入一条新记录
	model, _ = testD.Insert(map[string]interface{}{
		"name": "test2",
		"time": time.Now(),
	})
	//获取结构体值
	testModel := model.(*Test)
	fmt.Println(testModel.Id, testModel.Name.MustValue(), testModel.Time.MustValue())

	//更新记录
	testModel.Update(map[string]interface{}{
		"name": "test3",
		"time": time.Now(),
	})
	fmt.Println(testModel.Id, testModel.Name.MustValue(), testModel.Time.MustValue())

	//查询刚插入的记录(懒查询)
	model, _ = testD.SelectById(testModel.Id)
	testModel = model.(*Test)
	//...
	v, err := testModel.Name.Value() //访问到name字段时，进行数据库查询，查询记录不存在，则err不为空
	fmt.Println(v, err)
	sorm.TryCatch(func() {
		//强制读取字段时，如该条记录不存在，则会引发panic，可通过TryCatch捕获
		v = testModel.Name.MustValue()
		fmt.Println(v)
	}, func(err error) {
		fmt.Println(err)
	}, TestNotFoundError)
}
