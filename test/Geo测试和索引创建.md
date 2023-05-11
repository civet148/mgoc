# 官方参考连接

[Find Restaurants with Geospatial Queries — MongoDB Manual](https://www.mongodb.com/docs/v5.0/tutorial/geospatial-tutorial/)

# 1. 下载地理位置测试数据

- **基于经纬度的测试数据**

restaurants.json是餐馆所在经纬度数据

https://raw.githubusercontent.com/mongodb/docs-assets/geospatial/restaurants.json

```sh
# 下载保存到/tmp目录
$ cd /tmp && wget https://raw.githubusercontent.com/mongodb/docs-assets/geospatial/restaurants.json
```



- **基于GeoJSON的测试数据**

neighborhoods.json数据是一些多边形范围数据(由N个点围成一圈)

https://raw.githubusercontent.com/mongodb/docs-assets/geospatial/neighborhoods.json

```sh
# 下载保存到/tmp目录
$ cd /tmp && wget https://raw.githubusercontent.com/mongodb/docs-assets/geospatial/neighborhoods.json
```

# 2. 导入测试数据

- 导入测试数据到test库

```sh
$ cd /tmp && ls *.json
neighborhoods.json  restaurants.json

$ mongoimport restaurants.json -c restaurants 
2023-05-10T09:20:32.821+0000    connected to: mongodb://localhost/
2023-05-10T09:20:35.517+0000    25359 document(s) imported successfully. 0 document(s) failed to import.

$ mongoimport neighborhoods.json -c neighborhoods
2023-05-10T09:21:05.400+0000    connected to: mongodb://localhost/
2023-05-10T09:21:07.305+0000    195 document(s) imported successfully. 0 document(s) failed to import.
```

- 登录mongo终端创建索引

```sh
$ mongo

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
# 为restaurants表location.coordinates字段创建2D索引
> db.restaurants.createIndex({"location.coordinates":"2d"})
{
        "createdCollectionAutomatically" : false,
        "numIndexesBefore" : 2,
        "numIndexesAfter" : 3,
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

# 3. 地理位置查询类型



| 查询类型                              | 几何类型 | 备注                                                         |
| ------------------------------------- | -------- | ------------------------------------------------------------ |
| $near (GeoJSON点，2dsphere索引)       | 球面     | 输入为GeoJSON时要建立2dsphere索引。 结果已经排序，由近及远。 |
| $near (传统坐标，2d索引)              | 平面     | 输入为普通坐标点时要建立2d索引，结果已经排序，由近及远。     |
| $nearSphere (GeoJSON点，2dsphere索引) | 球面     | 输入为GeoJSON时要建立2dsphere索引。 结果已经排序，由近及远。 |
| $nearSphere (传统坐标，2d索引)        | 平面     | 传统坐标自动转为GeoJSON点，结果已经排序，由近及远。          |
| $geoWithin:{$geometry:...}            | 球面     | 根据GeoMetry范围查询(不排序)                                 |
| $geoWithin:{$box:...}                 | 平面     | 根据方形坐标进行范围查询(不排序)                             |
| $geoWithin:{$polygon:...}             | 平面     | 根据多边形坐标进行范围查询(不排序)                           |
| $geoWithin:{$center:...}              | 平面     | 根据中心点坐标进行范围查询(不排序)                           |
| $geoWithin:{$centerSphere:...}        | 球面     | 根据球面中心点坐标进行范围查询(不排序)                       |
| $geoIntersects                        | 球面     |                                                              |



# 4. 查询附近的餐馆

## 4.1 指定范围查询附近餐馆

- 首先从neighborhoods表查出某个点所在区域范围(比如某某广场)

```sh
> db.neighborhoods.findOne({ geometry: { $geoIntersects: { $geometry: { type: "Point", coordinates: [ -73.93414657, 40.82302903 ] } } } })
{
        "_id" : ObjectId("55cb9c666c522cafdb053a68"),
        "geometry" : {
                "coordinates" : [
                        [
                                [
                                        -73.93383000695911,
                                        40.81949109558767
                                ],
                                [
                                        -73.93411701695138,
                                        40.81955053491088
                                ],
                                [
                                        -73.93431276819767,
                                        40.81962986684897
                                ],
                                [
                                        -73.93440040009484,
                                        40.819667782434465
                                ],
                                ...
                        ]
                ],
                "type" : "Polygon"
        },
        "name" : "Central Harlem North-Polo Grounds"
}
```



- 查询某个范围内的餐馆

```sh
> var neighborhood = db.neighborhoods.findOne( { geometry: { $geoIntersects: { $geometry: { type: "Point", coordinates: [ -73.93414657, 40.82302903 ] } } } } )
> db.restaurants.find( { location: { $geoWithin: { $geometry: neighborhood.geometry } } } ).pretty()

{
        "_id" : ObjectId("55cba2476c522cafdb057f93"),
        "location" : {
                "coordinates" : [
                        -73.93781899999999,
                        40.8073089
                ],
                "type" : "Point"
        },
        "name" : "Perfect Taste"
}
{
        "_id" : ObjectId("55cba2476c522cafdb058b1a"),
        "location" : {
                "coordinates" : [
                        -73.943046,
                        40.807862
                ],
                "type" : "Point"
        },
        "name" : "Event Productions Catering & Food Services"
}
{
        "_id" : ObjectId("55cba2476c522cafdb054824"),
        "location" : {
                "coordinates" : [
                        -73.9446889,
                        40.8087276
                ],
                "type" : "Point"
        },
        "name" : "Sylvia'S Restaurant"
}
{
        "_id" : ObjectId("55cba2476c522cafdb0576a9"),
        "location" : {
                "coordinates" : [
                        -73.9450154,
                        40.808573
                ],
                "type" : "Point"
        },
        "name" : "Corner Social"
}
{
        "_id" : ObjectId("55cba2476c522cafdb0576c1"),
        "location" : {
                "coordinates" : [
                        -73.9449154,
                        40.8086991
                ],
                "type" : "Point"
        },
        "name" : "Cove Lounge"
}
{
        "_id" : ObjectId("55cba2476c522cafdb056a56"),
        "location" : {
                "coordinates" : [
                        -73.9508604,
                        40.8111432
                ],
                "type" : "Point"
        },
        "name" : "Manna'S Restaurant"
}
{
        "_id" : ObjectId("55cba2476c522cafdb0548db"),
        "location" : {
                "coordinates" : [
                        -73.9503386,
                        40.8116759
                ],
                "type" : "Point"
        },
        "name" : "Harlem Bar-B-Q"
}
{
        "_id" : ObjectId("55cba2476c522cafdb05488c"),
        "location" : {
                "coordinates" : [
                        -73.949938,
                        40.812365
                ],
                "type" : "Point"
        },
        "name" : "Hong Cheong"
}
{
        "_id" : ObjectId("55cba2486c522cafdb059376"),
        "location" : {
                "coordinates" : [
                        -73.948542,
                        40.814343
                ],
                "type" : "Point"
        },
        "name" : "Lighthouse Fishmarket"
}
{
        "_id" : ObjectId("55cba2486c522cafdb0592dd"),
        "location" : {
                "coordinates" : [
                        -73.948343,
                        40.8145609
                ],
                "type" : "Point"
        },
        "name" : "Mahalaxmi Food Inc"
}
{
        "_id" : ObjectId("55cba2486c522cafdb0599ce"),
        "location" : {
                "coordinates" : [
                        -73.94826669999999,
                        40.814616
                ],
                "type" : "Point"
        },
        "name" : "Rose Seeds"
}
{
        "_id" : ObjectId("55cba2476c522cafdb058263"),
        "location" : {
                "coordinates" : [
                        -73.9478824,
                        40.8151724
                ],
                "type" : "Point"
        },
        "name" : "J. Restaurant"
}
{
        "_id" : ObjectId("55cba2486c522cafdb059d5e"),
        "location" : {
                "coordinates" : [
                        -73.946651,
                        40.816918
                ],
                "type" : "Point"
        },
        "name" : "Harlem Coral Llc"
}
{
        "_id" : ObjectId("55cba2476c522cafdb0556ec"),
        "location" : {
                "coordinates" : [
                        -73.9462197,
                        40.8169283
                ],
                "type" : "Point"
        },
        "name" : "Baraka Buffet"
}
{
        "_id" : ObjectId("55cba2476c522cafdb054556"),
        "location" : {
                "coordinates" : [
                        -73.9420895,
                        40.8181467
                ],
                "type" : "Point"
        },
        "name" : "Make My Cake"
}
{
        "_id" : ObjectId("55cba2486c522cafdb0597a9"),
        "location" : {
                "coordinates" : [
                        -73.9422781,
                        40.8178187
                ],
                "type" : "Point"
        },
        "name" : "Hyacinth Haven Harlem"
}
{
        "_id" : ObjectId("55cba2476c522cafdb053fec"),
        "location" : {
                "coordinates" : [
                        -73.9419869,
                        40.8175016
                ],
                "type" : "Point"
        },
        "name" : "Mcdonald'S"
}
{
        "_id" : ObjectId("55cba2476c522cafdb054e73"),
        "location" : {
                "coordinates" : [
                        -73.9419869,
                        40.8175016
                ],
                "type" : "Point"
        },
        "name" : "Ihop"
}
{
        "_id" : ObjectId("55cba2476c522cafdb05922f"),
        "location" : {
                "coordinates" : [
                        -73.9421715,
                        40.8172179
                ],
                "type" : "Point"
        },
        "name" : "Island Spice And Southern Cuisine"
}
{
        "_id" : ObjectId("55cba2486c522cafdb059767"),
        "location" : {
                "coordinates" : [
                        -73.94301899999999,
                        40.816936
                ],
                "type" : "Point"
        },
        "name" : "To Your Health & Happiness"
}
Type "it" for more
```



## 4.2 给定坐标查询附近餐馆

- 查询1000米范围内所有的餐馆

弧度计算公式：(radius/1000)/6359.0 = (1000/1000)/6359.0 = 1/6359.0 

```sh
# 查询附近1000米内所有的餐馆总数
> db.restaurants.find(
{ 
location: { 
$geoWithin: { 
			$centerSphere: [[ -73.93414657, 40.82302903 ], 1/6359.0]
		} 
	}
}).count()
172

# 查询附近1000米内所有的餐馆数据
> db.restaurants.find(
{ 
location: { 
$geoWithin: { 
			$centerSphere: [[ -73.93414657, 40.82302903 ], 1/6359.0]
		} 
	}
})

{ "_id" : ObjectId("55cba2476c522cafdb056005"), "location" : { "coordinates" : [ -73.9259201, 40.8278293 ], "type" : "Point" }, "name" : "Nyy Steak" }
{ "_id" : ObjectId("55cba2476c522cafdb056004"), "location" : { "coordinates" : [ -73.9259201, 40.8278293 ], "type" : "Point" }, "name" : "Hard Rock Cafe" }
{ "_id" : ObjectId("55cba2476c522cafdb058ae9"), "location" : { "coordinates" : [ -73.9259245, 40.827435 ], "type" : "Point" }, "name" : "Dunkin Donuts" }
{ "_id" : ObjectId("55cba2476c522cafdb0543cb"), "location" : { "coordinates" : [ -73.92594439999999, 40.8272129 ], "type" : "Point" }, "name" : "Billy'S Sport Bar Restaurant & Lounge" }
{ "_id" : ObjectId("55cba2476c522cafdb056784"), "location" : { "coordinates" : [ -73.9262845, 40.82669569999999 ], "type" : "Point" }, "name" : "Stan'S Sports Bar" }
{ "_id" : ObjectId("55cba2476c522cafdb056eb6"), "location" : { "coordinates" : [ -73.9259928, 40.82713630000001 ], "type" : "Point" }, "name" : "Yankee Bar & Grill" }
{ "_id" : ObjectId("55cba2476c522cafdb057b8a"), "location" : { "coordinates" : [ -73.9250339, 40.8269856 ], "type" : "Point" }, "name" : "Flavas International Grill" }
{ "_id" : ObjectId("55cba2476c522cafdb05553c"), "location" : { "coordinates" : [ -73.9244796, 40.8270316 ], "type" : "Point" }, "name" : "Court Deli Restaurant" }
{ "_id" : ObjectId("55cba2476c522cafdb0581c2"), "location" : { "coordinates" : [ -73.9249324, 40.8274822 ], "type" : "Point" }, "name" : "Us Fried Chicken" }
Type "it" for more
> 
```

## 4.3 查询距离

根据某个点和最大距离搜索附近的数据并返回离当前点的距离

- near 搜索起始点
- distanceField 存放距离的参数(不能跟其他字段名冲突)
- maxDistance 单位：米
- includeLocs 包含位置信息的字段（建立2dsphere索引的字段）
- spherical 是否使用球面几何计算

```javascript
db.restaurants.aggregate(
    {
        $geoNear:{
            "near":{type:"Point", coordinates:[-73.93414657, 40.82302903]},
            "distanceField":"distance",
            "maxDistance": 1000,
            "includeLocs": "location",
            "spherical": true, 
      }
  }
)
```

