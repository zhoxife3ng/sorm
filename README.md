# sorm

***sorm***是一个封装了简易orm模型的Go包，使用了`didi/gendry`库辅助生成`Sql`语句。封装了`dao`抽象方法，可直接通过API进行数据库查询，查询成功后统一构造对应的`model`实例返回。