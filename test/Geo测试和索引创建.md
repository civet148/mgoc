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

- 查询5000米范围内所有的餐馆

```sh
> db.restaurants.find(
{ 
location: { 
$nearSphere: { 
	$geometry: { 
		type: "Point", 
		coordinates: [ -73.93414657, 40.82302903 ] }, 
		$maxDistance: 5000 
		} 
	}
})

{ "_id" : ObjectId("55cba2476c522cafdb058c83"), "location" : { "coordinates" : [ -73.9316894, 40.8231974 ], "type" : "Point" }, "name" : "Gotham Stadium Tennis Center Cafe" }
{ "_id" : ObjectId("55cba2476c522cafdb05864b"), "location" : { "coordinates" : [ -73.9378967, 40.823448 ], "type" : "Point" }, "name" : "Tia Melli'S Latin Kitchen" }
{ "_id" : ObjectId("55cba2476c522cafdb058c63"), "location" : { "coordinates" : [ -73.9303724, 40.8234978 ], "type" : "Point" }, "name" : "Chuck E. Cheese'S" }
{ "_id" : ObjectId("55cba2476c522cafdb0550aa"), "location" : { "coordinates" : [ -73.93795159999999, 40.823376 ], "type" : "Point" }, "name" : "Domino'S Pizza" }
{ "_id" : ObjectId("55cba2476c522cafdb0548e0"), "location" : { "coordinates" : [ -73.9381738, 40.8224212 ], "type" : "Point" }, "name" : "Red Star Chinese Restaurant" }
{ "_id" : ObjectId("55cba2476c522cafdb056b6a"), "location" : { "coordinates" : [ -73.93011659999999, 40.8219403 ], "type" : "Point" }, "name" : "Applebee'S Neighborhood Grill & Bar" }
{ "_id" : ObjectId("55cba2476c522cafdb0578b3"), "location" : { "coordinates" : [ -73.93011659999999, 40.8219403 ], "type" : "Point" }, "name" : "Marisco Centro Seafood Restaurant  & Bar" }
{ "_id" : ObjectId("55cba2476c522cafdb058dfc"), "location" : { "coordinates" : [ -73.9370572, 40.8206095 ], "type" : "Point" }, "name" : "108 Fast Food Corp" }
{ "_id" : ObjectId("55cba2476c522cafdb0574cd"), "location" : { "coordinates" : [ -73.9365102, 40.8202205 ], "type" : "Point" }, "name" : "Kentucky Fried Chicken" }
{ "_id" : ObjectId("55cba2476c522cafdb057d52"), "location" : { "coordinates" : [ -73.9385009, 40.8222455 ], "type" : "Point" }, "name" : "United Fried Chicken" }
{ "_id" : ObjectId("55cba2476c522cafdb054e83"), "location" : { "coordinates" : [ -73.9373291, 40.8206458 ], "type" : "Point" }, "name" : "Dunkin Donuts" }
{ "_id" : ObjectId("55cba2476c522cafdb05615f"), "location" : { "coordinates" : [ -73.9373291, 40.8206458 ], "type" : "Point" }, "name" : "King'S Pizza" }
{ "_id" : ObjectId("55cba2476c522cafdb05476a"), "location" : { "coordinates" : [ -73.9365637, 40.8201488 ], "type" : "Point" }, "name" : "Papa John'S" }
{ "_id" : ObjectId("55cba2486c522cafdb059a11"), "location" : { "coordinates" : [ -73.9365637, 40.8201488 ], "type" : "Point" }, "name" : "Jimbo'S Hamburgers" }
{ "_id" : ObjectId("55cba2476c522cafdb0580a7"), "location" : { "coordinates" : [ -73.938599, 40.82211110000001 ], "type" : "Point" }, "name" : "Home Garden Chinese Restaurant" }
{ "_id" : ObjectId("55cba2476c522cafdb05814c"), "location" : { "coordinates" : [ -73.9367511, 40.8198978 ], "type" : "Point" }, "name" : "Sweet Mama'S Soul Food" }
{ "_id" : ObjectId("55cba2476c522cafdb056b96"), "location" : { "coordinates" : [ -73.9308109, 40.82594580000001 ], "type" : "Point" }, "name" : "Dunkin Donuts (Inside Gulf Gas Station On North Side Of Maj. Deegan Exwy- After Exit 13 - 233 St.)" }
{ "_id" : ObjectId("55cba2476c522cafdb056ffd"), "location" : { "coordinates" : [ -73.939159, 40.8216897 ], "type" : "Point" }, "name" : "Reggae Sun Delights Natural Juice Bar" }
{ "_id" : ObjectId("55cba2476c522cafdb056b0c"), "location" : { "coordinates" : [ -73.939145, 40.8213757 ], "type" : "Point" }, "name" : "Ho Lee Chinese Restaurant" }
{ "_id" : ObjectId("55cba2486c522cafdb059617"), "location" : { "coordinates" : [ -73.9396354, 40.8220958 ], "type" : "Point" }, "name" : "Ivory D O S  Inc" }
Type "it" for more
```



