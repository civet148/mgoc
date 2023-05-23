<h1 align="center"> MongoDB ORM client - mgoc </h1>

 

## 部署MongoDB

- 启动容器

```shell

# start mongodb by docker
$ docker run -p 27017:27017 --restart always -v /data/mongodb/db:/data/db --name mongodb -d mongo:4.4.10
```

- 创建登录账号

账户: root 密码: 123456

```shell
$ docker exec -it mongodb mongo admin
> use admin
> db.createUser({user:"root", pwd: "123456", roles: ["root"]})
> exit
```

## 导入测试数据

- 进入容器终端

```shell
$ docker exec -it mongodb bash
# 在容器内安装wget
root@072bedc2e6c5:/# apt-get update && apt-get install wget
root@072bedc2e6c5:/# cd /tmp
root@072bedc2e6c5:/tmp# 
```

- **基于经纬度的测试数据**

restaurants.json是餐馆所在经纬度数据

https://raw.githubusercontent.com/mongodb/docs-assets/geospatial/restaurants.json

```sh
# 下载保存到/tmp目录
root@072bedc2e6c5:/tmp# wget https://raw.githubusercontent.com/mongodb/docs-assets/geospatial/restaurants.json
```



- **基于GeoJSON的测试数据**

neighborhoods.json数据是一些多边形范围数据(由N个点围成一圈)

https://raw.githubusercontent.com/mongodb/docs-assets/geospatial/neighborhoods.json

```sh
# 下载保存到/tmp目录
root@072bedc2e6c5:/tmp# wget https://raw.githubusercontent.com/mongodb/docs-assets/geospatial/neighborhoods.json
```
- 导入测试数据到test库

```sh
root@072bedc2e6c5:/tmp# ls *.json
neighborhoods.json  restaurants.json

root@072bedc2e6c5:/tmp# mongoimport restaurants.json -c restaurants 
2023-05-10T09:20:32.821+0000    connected to: mongodb://localhost/
2023-05-10T09:20:35.517+0000    25359 document(s) imported successfully. 0 document(s) failed to import.

root@072bedc2e6c5:/tmp# mongoimport neighborhoods.json -c neighborhoods
2023-05-10T09:21:05.400+0000    connected to: mongodb://localhost/
2023-05-10T09:21:07.305+0000    195 document(s) imported successfully. 0 document(s) failed to import.
```

- 登录mongo终端创建索引

```sh
root@072bedc2e6c5:/tmp# mongo

# 切换到test库
> use test
# 查询一条数据看看表结构
> db.restaurants.find().limit(1).pretty()
{
        "_id" : ObjectId("55cba2476c522cafdb053ae2"),
        "location" : {
                "coordinates" : [
                        -73.98513559999999,
                        40.7676919
                ],
                "type" : "Point"
        },
        "name" : "Dj Reynolds Pub And Restaurant"
}
# 为restaurants表location字段创建2D球面索引
> db.restaurants.createIndex({"location":"2dsphere"})
{
        "createdCollectionAutomatically" : false,
        "numIndexesBefore" : 1,
        "numIndexesAfter" : 2,
        "ok" : 1
}
# 为neighborhoods表geometry字段建立2D球面索引
> db.neighborhoods.createIndex({ geometry: "2dsphere" })
{
        "createdCollectionAutomatically" : false,
        "numIndexesBefore" : 1,
        "numIndexesAfter" : 2,
        "ok" : 1
}
```


## 快速开始

- 所有的ORM操作必须是以Model方法开始,参数除执行delete操作之外都是必填
- TableName方法通常情况下也是必须调用的方法（除了数据库层面的聚合操作）

```go
package main
import (
    "time"
    "github.com/civet148/log"
    "github.com/civet148/mgoc"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Student struct {
	Id          string    `bson:"_id,omitempty"`
	Name        string    `bson:"name"`
	Sex         string    `bson:"sex"`
	Age         int       `bson:"age"`
	ClassNo     string    `bson:"class_no"`
	CreatedTime time.Time `bson:"created_time"`
}

func main() {
	e, err := mgoc.NewEngine("mongodb://root:123456@127.0.0.1:27017/test?authSource=admin")
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	var students []*Student
	err = e.Model(&students).
            Table("student_info").
            Options(&options.FindOptions{}).
            Desc("created_time").
            Limit(10).
            Query()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	for _, s := range students {
		log.Infof("student %+v", s)
	}
}
```

## 选项
- Options方法 
  根据实际操作类型不同可输入不同的Option类型，比如查询时选填options.FindOptions，更新时可选填options.UpdateOptions
  插入单条记录时可选填options.InsertOneOptions，插入多条则是options.InsertManyOptions等等

## 插入操作

- 单条插入

```go
var student = &docStudent{
		Name:        "john",
		Sex:         "male",
		Age:         13,
		ClassNo:     "CLASS-3",
		Balance:     mgoc.NewDecimal("532.324"),
		CreatedTime: time.Now(),
	}
ids, err := e.Model(student).
		Table("student_info").
		Options(&options.InsertOneOptions{}).
		Insert()
if err != nil {
    log.Errorf(err.Error())
    return
}
log.Infof("[Single] insert ids %+v", ids)
```

- 多条插入(非事务)

```go
var students = []*docStudent{
		{
			Name:        "lory",
			Sex:         "male",
			Age:         14,
			ClassNo:     "CLASS-1",
			CreatedTime: time.Now(),
		},
		{
			Name:        "katy",
			Sex:         "female",
			Age:         15,
			ClassNo:     "CLASS-2",
			CreatedTime: time.Now(),
		},
	}
ids, err = e.Model(students).
		Options(&options.InsertManyOptions{}).
		Table("student_info").
		Insert()
if err != nil {
    log.Errorf(err.Error())
    return 
}
log.Infof("[Many] insert ids %+v", ids)
```



## 普通查询

## 更新操作

## 聚合查询

## 地理位置查询

## 

