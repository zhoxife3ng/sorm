# sorm

***sorm***是一个封装了简易orm模型的Go包，使用了内置`builder`库辅助生成`Sql`语句。封装了`dao`抽象方法，可直接通过API进行数据库查询，查询成功后统一构造对应的`model`实例返回。

## 特性

* 支持Struct和数据库表之间的灵活映射

* 支持事务，主从库配置读写分离

* 同时支持原始SQL语句和ORM操作的混合执行

* 使用连写来简化调用

* 支持使用Where, Limit, Join, Having, Table, Columns等函数和结构体等方式作为条件

* 内置SQL Builder支持

* 上下文缓存支持

* 抽象Dao方法可直接构造对象，无需事先声明变量

* 支持懒查询机制，当访问到非主键字段时自动进行数据库查询

## 安装

	go get github.com/xkisas/sorm
	
# 快速开始

* 创建数据库连接
```go
db.Setup(
    db.Conf{
		Name:     "dbName",
		User:     "root",
		Password: "123456",
		Host:     "127.0.0.1",
		Port:     3306,
	},
	db.Param("charset", "utf8"),
)
```

* 创建一个和数据库表同步的结构体
```go
type Test struct {
	sorm.BaseModel
	Id   int          `db:"id,pk"` //pk标记主键
	Name _type.String `db:"name"`
	Time _type.Time   `db:"time"`
}
```

* 创建上下文并操作数据库
```go
sess := sorm.NewSession(context.TODO())
defer sess.Close()

testD := sess.GetDao(new(Test)) //获取操作该对象的dao

//插入一条记录，并返回对象
model, err := testD.Insert(map[string]interface{}{
	"name": "test",
	"time": time.Now(),
})
fmt.Println(model, err) //&Test{...}    nil
id := model.GetId() //得到刚刚插入的自增id

//删除刚插入的记录
model.Remove()

//查询刚刚删除的记录(立即查询数据库)
model, err = testD.SelectById(id, sorm.Load())
fmt.Println(model, err) //nil    model not found

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
	"name":"test3",
	"time":time.Now(),
})
fmt.Println(testModel.Id, testModel.Name.MustValue(), testModel.Time.MustValue())

//查询刚插入的记录(懒查询)
model, _ = testD.SelectById(testModel.Id)
testModel = model.(*Test)
//...
v, err := testModel.Name.Value() //访问到name字段时，底层自动进行数据库查询，查询记录不存在，则err不为空
fmt.Println(v, err)
sorm.TryCatch(func() {
    //强制读取字段时，如该条记录不存在，则会引发panic，可通过TryCatch捕获
    v = model.(*Test).Name.MustValue()
	fmt.Println(v)
}, func(err error) {
    fmt.Println(err)
}, sorm.ModelNotFoundError)
```