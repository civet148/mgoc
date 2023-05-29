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

- 所有的ORM操作必须是以Model方法开始,参数除执行delete/update操作之外都是必填
- Table方法通常情况下也是必须调用的方法（除了数据库层面的聚合操作）

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
      Id          string            `bson:"_id,omitempty"`
      Name        string            `bson:"name"`
      Sex         string            `bson:"sex"`
      Age         int               `bson:"age"`
      Balance     mgoc.Decimal      `bson:"balance"`
      ClassNo     string            `bson:"class_no"`
      CreatedTime time.Time         `bson:"created_time"`
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
- Options方法 [optional]

  根据实际操作类型不同可输入不同的Option类型，比如查询时选填options.FindOptions，更新时可选填options.UpdateOptions
  插入单条记录时可选填options.InsertOneOptions，插入多条则是options.InsertManyOptions等等(选填)

## 插入操作

- 单条插入

```go
var student = Student{
		Name:        "john",
		Sex:         "male",
		Age:         13,
		ClassNo:     "CLASS-3",
		Balance:     mgoc.NewDecimal("532.324"),
		CreatedTime: time.Now(),
	}
ids, err := e.Model(&student).
		Table("student_info").
		Options(&options.InsertOneOptions{}).
		Insert()
if err != nil {
    log.Errorf(err.Error())
    return
}
log.Infof("[Single] insert id %+v", ids)
```

- 多条插入(非事务)

```go
var students = []*Student{
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
ids, err := e.Model(&students).
		Options(&options.InsertManyOptions{}).
		Table("student_info").
		Insert()
if err != nil {
    log.Errorf(err.Error())
    return 
}
log.Infof("[Many] insert ids %+v", ids)
```



## 查询操作

- 单条查询

SELECT *  FROM student_info LIMIT 1

```go
var err error
var student *Student
err = e.Model(&student).
        Table("student_info").
        Limit(1).
        Query()
if err != nil {
    log.Errorf(err.Error())
    return
}
log.Infof("single student %+v", student)
```

- 多条查询

SELECT _id, name, age, sex FROM student_info ORDER BY created_time DESC LIMIT 10

```go
var err error
var students []*Student
err = e.Model(&students).
        Table("student_info").
        Select("_id", "name", "age", "sex").
        Options(&options.FindOptions{}).
        Desc("created_time").
        Limit(10).
        Query()
if err != nil {
    log.Errorf(err.Error())
    return
}
for _, student := range students {
    log.Infof("student %+v", student)
}
```

- 分页查询

SELECT _id, name, age, sex FROM student_info ORDER BY created_time DESC LIMIT 0,10

```go
var students []*Student
total, err := e.Model(&students).
        Table("student_info").
        Select("_id", "name", "age", "sex").
        Options(&options.FindOptions{}).
        Desc("created_time").
        Page(0, 10). //Page(2, 10) == LIMIT 2*10, 10
        QueryEx()
if err != nil {
    log.Errorf(err.Error())
    return
}
log.Infof("student total %+v", total)
for _, student := range students {
    log.Infof("student %+v", student)
}
```

- 条件查询

SELECT _id, name, age, sex FROM student_info WHERE class_no='CLASS-2' AND age >= 11 and age <=16 ORDER BY created_time DESC

```go
var err error
var students []*Student
err = e.Model(&students).
        Table("student_info").
        Select("_id", "name", "age", "sex").
        Options(&options.FindOptions{}).
        Eq("class_no", "CLASS-2").
        Gte("age", 11).
        Lte("age", 16).
        Desc("created_time").
        Query()
if err != nil {
    log.Errorf(err.Error())
    return
}
for _, student := range students {
    log.Infof("student %+v", student)
}
```



- 自定义查询

SELECT _id, name, age, sex FROM student_info WHERE class_no='CLASS-2' AND age >= 11 and age <=16 ORDER BY created_time DESC

```go
var err error
var students []*Student
err = e.Model(&students).
        Table("student_info").
        Select("_id", "name", "age", "sex").
        Options(&options.FindOptions{}).
        Filter(bson.M{
            "class_no":"CLASS-2",
            "age": bson.M{"$gte":11},
            "age": bson.M{"$lte":16},
        }).
        Desc("created_time").
        Query()
if err != nil {
    log.Errorf(err.Error())
    return
}
for _, student := range students {
    log.Infof("student %+v", student)
}
```



## 更新操作

- **通过数据模型更新字段**

UPDATE student_info SET name='kary', sex='female', age=39, balance='123.456', created_time=NOW() WHERE _id='6438f32fd71fc42e601558aa'

```go
// 更新Id值6438f32fd71fc42e601558aa对应的数据记录
var student = &Student{
		Id:          mgoc.ToObjectID("6438f32fd71fc42e601558aa").(mgoc.ObjectID),
		Name:        "kary",
		Sex:         "female",
		Age:         39,
		Balance:     mgoc.NewDecimal("123.456"),
		CreatedTime: time.Now(),
	}
_, err := e.Model(&student).
            Table("student_info").
            Options(&options.UpdateOptions{}).
            Select("name", "sex", "age", "balance", "created_time").
            Update()
if err != nil {
    log.Errorf(err.Error())
    return
}
```

- **通过Set方式更新字段**

```go
_, err := e.Model().
            Table("student_info").
            Options(&options.UpdateOptions{}).
            Id("6438f32fd71fc42e601558aa").
            Set("name", "mason").
            Set("sex", "male").
            Set("balance", mgoc.NewDecimal("123.456")).
            Update()
if err != nil {
    log.Errorf(err.Error())
    return
}
```

- **结构嵌套更新**

```go
type ExtraData struct {
    IdCard      string   `bson:"id_card"`
    Address     string   `bson:"address"`
}

type Student struct {
    Id          string          `bson:"_id,omitempty"`
    Name        string          `bson:"name"`
    Sex         string          `bson:"sex"`
    Age         int             `bson:"age"`
    ClassNo     string          `bson:"class_no"`
    Balance     mgoc.Decimal    `bson:"balance"`
    CreatedTime time.Time       `bson:"created_time"`
    ExtraData   ExtraData       `bson:"extra_data"`
}
oid, err := mgoc.NewObjectIDFromString("6438f32fd71fc42e601558aa")
if err != nil {
    log.Errorf(err.Error())
    return
}
var student = &Student{
        Id:          oid,
        ClassNo:     "CLASS-3",
        ExtraData:   ExtraData {
            IdCard: "6553239109322",
        },
    }
// UPDATE student_info 
// SET class_no='CLASS-3', extra_data.id_card='6553239109322'
// WHERE _id='6438f32fd71fc42e601558aa'
_, err = e.Model(&student).
            Table("student_info").
            Options(&options.UpdateOptions{}).
            Select("class_no", "extra_data.id_card").
            Update()
if err != nil {
    log.Errorf(err.Error())
    return
}

//等同于下面的方式
_, err = e.Model().
            Table("student_info").
            Id("6438f32fd71fc42e601558aa").
            Options(&options.UpdateOptions{}).
            Set("class_no", "CLASS-3").
            Set("extra_data.id_card", "6553239109322").
            Update()
if err != nil {
    log.Errorf(err.Error())
    return
}
```



## 更新或插入



```go
var err error
_, err = e.Model().
        Table("student_info").
        Id("6438f32fd71fc42e601558aa").
        Set("name", "rose").
        Set("sex", "female").
        Set("age", 18).
        Set("created_time", time.Now()).
        Set("balance", mgoc.NewDecimal("520.1314")).
        Upsert()
if err != nil {
    log.Errorf(err.Error())
    return
}
```





## 删除操作

```go
rows, err := e.Model().
                Table("student_info").
                Options(&options.DeleteOptions{}).
                Id("6438f32fd71fc42e601558aa").
                Delete()
if err != nil {
    log.Errorf(err.Error())
    return
}
log.Infof("rows %d deleted", rows)
```



## 聚合查询

- ORM聚合查询

SELECT AVG(age) AS age, SUM(1) AS total, SUM(balance) as balance FROM  student_info WHERE sex='female' GROUP BY name, age

```go
/*
	db.getCollection("student_info").aggregate([
        {
            "$match":{
                "sex":"female"
            },
        },
        {
            "$group":{
                        "_id":{"name":"$name", "age":"$age"},
                        "age":{ "$avg":"$age"},
                        "balance":{ "$sum":"$balance"},
                        "total":{ "$sum":1}
            }
        }
    ])
----------------------------------------------------------------------
{
    "_id": {
        "name": "katy3",
        "age": NumberInt("28")
    },
    "age": 28,
    "balance": NumberDecimal("24149.3374"),
    "total": 8
}
// 2
{
    "_id": {
        "name": "katy3",
        "age": NumberInt("27")
    },
    "age": 27,
    "balance": NumberDecimal("234"),
    "total": 1
}
// 3
{
    "_id": {
        "name": "rose",
        "age": NumberInt("18")
    },
    "age": 18,
    "balance": NumberDecimal("520.1314"),
    "total": 1
}
*/
  type AggID struct {
      Name string `bson:"name"`
  }
  type StudentAgg struct {
    ID    AggID             `bson:"_id"`
    Age   float64           `bson:"age"`
    Total int               `bson:"total"`
    Balance mgoc.Decimal    `bson:"balance"`
  }
  var agg []*StudentAgg
  err := e.Model(&agg).
        Table("student_info").
        Avg("age").
        Sum("total", 1).
        Sum("balance").
        Eq("sex", "female").
        GroupBy("name", "age").
        Aggregate()
  if err != nil {
    log.Errorf(err.Error())
    return
  }
  log.Infof("aggregate rows %d", len(agg))
  for _, a := range agg {
    log.Infof("%+v", a)
  }
```

- 自定义聚合查询

SELECT AVG(age) AS age, COUNT(1) AS total FROM  student_info WHERE sex='female' 

```go
/*
		db.getCollection("student_info").aggregate([
		   {
		     "$match":{
				    "sex":"female"
			   },
			 },
			 {
			   "$group":{
			      		"_id":null,
						"age":{ "$avg":"$age"},
						"total":{ "$sum":1}
					}
		   },
			 {
			   "$project":{
			         "_id":0,
					 "age":1,
					 "total":1
					}
			 }
		]
		)
		----------
		{
		    "age": 18,
		    "total": 14
		}
*/

type StudentAgg struct {
  Age   float64 `bson:"age"`
  Total int     `bson:"total"`
}
var agg []*StudentAgg
// create match stage
match := bson.D{
    {
        "$match", bson.D{
            {"sex", "female"},
        },
    },
}
// create group stage
group := bson.D{
    {"$group", bson.D{
        {"_id", nil},
        {"age", bson.D{{"$avg", "$age"}}},
        {"total", bson.D{{"$sum", 1}}},
    }}}
// create projection stage
project := bson.D{
    {"$project", bson.D{
        {"_id", 0},
        {"age", 1},
        {"total", 1},
    }}}
err := e.Model(&agg).
        Table("student_info").
        Options(&options.AggregateOptions{}).
        Pipeline(match, group, project).
        Aggregate()
if err != nil {
    log.Errorf(err.Error())
    return
}
log.Infof("aggregate rows %d", len(agg))
for _, a := range agg {
    log.Infof("%+v", a)
}
```





## 地理位置查询

- 查询某一点1000米范围内所有的餐馆

```go
type Restaurant struct {
	Id       string `json:"_id" bson:"_id,omitempty"`
	Location struct {
		Type        string    `json:"type" bson:"type"`
		Coordinates []float64 `json:"coordinates" bson:"coordinates"`
	} `json:"location" bson:"location"`
	Name     string  `json:"name" bson:"name"`
	Distance float64 `json:"distance" bson:"distance"`
}
	
const maxMeters = 1000 //meters
var pos = mgoc.Coordinate{X: -73.93414657, Y: 40.82302903}
//query restaurants near by distance 1000 meters
var restaurants []*Restaurant
err := e.Model(&restaurants).
        Table("restaurants").
        GeoCenterSphere("location", pos, maxMeters).
        Query()
if err != nil {
    log.Errorf(err.Error())
    return
}
log.Infof("center sphere restaurants total [%d]", len(restaurants))
```



- 查询某个社区范围内的所有餐馆

```go
type Neighborhood struct {
	Id       string   `json:"_id" bson:"_id,omitempty"`
	Geometry mgoc.Geometry `json:"geometry" bson:"geometry"`
	Name     string   `json:"name" bson:"name"`
}
type Restaurant struct {
	Id       string `json:"_id" bson:"_id,omitempty"`
	Location struct {
		Type        string    `json:"type" bson:"type"`
		Coordinates []float64 `json:"coordinates" bson:"coordinates"`
	} `json:"location" bson:"location"`
	Name     string  `json:"name" bson:"name"`
	Distance float64 `json:"distance" bson:"distance"`
}

var neighbor *Neighborhood
var pos = Coordinate{X: -73.93414657, Y: 40.82302903}
//查询 -73.93414657, 40.82302903 所在社区的社区范围信息
err = e.Model(&neighbor).
        Table("neighborhoods").
        Filter(bson.M{
            "geometry": bson.M{
                mgoc.KeyGeoIntersects: bson.M{
                    mgoc.KeyGeoMetry: mgoc.NewGeoMetry(mgoc.GeoTypePoint, mgoc.FloatArray{pos.X, pos.Y}),
                },
            },
        }).
        Limit(1).
        Query()
if err != nil {
    log.Errorf(err.Error())
    return
}
log.Infof("neighborhood [%+v]", neighbor)
//查询社区范围内的餐馆
var restaurants []*Restaurant
err = e.Model(&restaurants).
        Table("restaurants").
        Geometry("location", &neighbor.Geometry).
        Query()
if err != nil {
    log.Errorf(err.Error())
    return
}
log.Infof("neighborhood restaurants total [%d]", len(restaurants))
```

- 查询某一点附近1000米内的所有餐馆数据并附带距离

```go
type Restaurant struct {
	Id       string `json:"_id" bson:"_id,omitempty"`
	Location struct {
		Type        string    `json:"type" bson:"type"`
		Coordinates []float64 `json:"coordinates" bson:"coordinates"`
	} `json:"location" bson:"location"`
	Name     string  `json:"name" bson:"name"`
	Distance float64 `json:"distance" bson:"distance"`
}
const maxMeters = 1000 //meters
var pos = Coordinate{X: -73.93414657, Y: 40.82302903}
var restaurants []*Restaurant
err := e.Model(&restaurants).
        Table("restaurants").
        Limit(10).
        Asc("distance").
        GeoNearByPoint(
            "location", //存储经纬度的字段
            pos, //当前位置数据
            maxMeters, //最大距离数(米)
            "distance"). //距离数据输出字段名
        Aggregate()
if err != nil {
    log.Errorf(err.Error())
    return 
}
for _, restaurant := range restaurants {
    log.Debugf("geo near restaurant [%+v]", restaurant)
}
log.Infof("geo near restaurants total [%d]", len(restaurants))
```

## 切换数据库

```go
  db := e.Use("test2")
```

