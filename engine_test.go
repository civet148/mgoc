package mgoc

import (
	"github.com/civet148/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"testing"
	"time"
)

const (
	TableNameStudentInfo   = "student_info"
	TableNameRestaurants   = "restaurants"
	TableNameNeighborhoods = "neighborhoods"
)

type extraData struct {
	IdCard      string   `json:"id_card" bson:"id_card"`
	HomeAddress string   `json:"home_address" bson:"home_address"`
	Sports      []string `json:"sports" bson:"sports"`
}

type docStudent struct {
	Id          ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string    `json:"name" bson:"name"`
	Sex         string    `json:"sex" bson:"sex"`
	Age         int       `json:"age" bson:"age"`
	Balance     Decimal   `json:"balance" bson:"balance"`
	ClassNo     string    `json:"class_no" bson:"class_no"`
	CreatedTime time.Time `json:"created_time" bson:"created_time"`
	ExtraData   extraData `json:"extra_data" bson:"extra_data"`
}

type docRestaurant struct {
	Id       string `json:"_id" bson:"_id,omitempty"`
	Location struct {
		Type        string    `json:"type" bson:"type"`
		Coordinates []float64 `json:"coordinates" bson:"coordinates"`
	} `json:"location" bson:"location"`
	Name     string  `json:"name" bson:"name"`
	Distance float64 `json:"distance" bson:"distance"`
}

type docNeighborhood struct {
	Id       string   `json:"_id" bson:"_id,omitempty"`
	Geometry Geometry `json:"geometry" bson:"geometry"`
	Name     string   `json:"name" bson:"name"`
}

const (
	officialObjectId = "6438f32fd71fc42e601558aa"
	defaultMongoUrl  = "mongodb://root:123456@192.168.2.146:27017/test?authSource=admin"
)

var opt = &Option{
	Debug: true,
	Max:   100,
	Idle:  5,
	//SSH: &SSH{
	//	User:     "root",
	//	Password: "123456",
	//	Host:     "192.168.2.19:22",
	//},
	ConnectTimeout: 3,
	WriteTimeout:   60,
	ReadTimeout:    60,
	DatabaseOpt:    nil,
}

func TestMongoDBCases(t *testing.T) {
	e, err := NewEngine(defaultMongoUrl, opt)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	//e.Use("test") //switch to other database
	e.Debug(true)
	OrmInsert(e)
	OrmQuery(e)
	GeoQuery(e)
	OrmUpdate(e)
	OrmUpsert(e)
	OrmCount(e)
	//OrmDelete(e)
	OrmAggregate(e)
	PipelineAggregate(e)
}

func GeoQuery(e *Engine) {
	const maxMeters = 1000 //meters
	var pos = Coordinate{X: -73.93414657, Y: 40.82302903}
	//query restaurants near by distance 1000 meters
	var restaurants []*docRestaurant
	err := e.Model(&restaurants).
		Table(TableNameRestaurants).
		GeoCenterSphere("location", pos, maxMeters).
		Query()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	//for _, restaurant := range restaurants {
	//	log.Debugf("center restaurant [%+v]", restaurant)
	//}
	log.Infof("center restaurants total [%d]", len(restaurants))
	var neighbor *docNeighborhood
	err = e.Model(&neighbor).
		Table(TableNameNeighborhoods).
		Filter(bson.M{
			"geometry": bson.M{
				KeyGeoIntersects: bson.M{
					KeyGeoMetry: NewGeoMetry(GeoTypePoint, FloatArray{-73.93414657, 40.82302903}),
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
	var restaurants2 []*docRestaurant
	err = e.Model(&restaurants2).
		Table(TableNameRestaurants).
		Geometry("location", &neighbor.Geometry).
		Query()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	//for _, restaurant := range restaurants2 {
	//	log.Debugf("neighborhood restaurant [%+v]", restaurant)
	//}
	log.Infof("neighborhood restaurants total [%d]", len(restaurants2))

	var restaurants3 []*docRestaurant
	err = e.Model(&restaurants3).
		Table(TableNameRestaurants).
		//Limit(10).
		//Asc("distance").
		GeoNearByPoint("location",
			pos,
			maxMeters,
			"distance").
		Query()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	for _, restaurant := range restaurants3 {
		log.Debugf("geo near restaurant [%+v]", restaurant)
	}
	log.Infof("geo near restaurants total [%d]", len(restaurants3))
}

func OrmQuery(e *Engine) {
	var err error
	var student *docStudent
	err = e.Model(&student).
		Table(TableNameStudentInfo).
		Query()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("single student %+v", student)

	var students []*docStudent
	err = e.Model(&students).
		Table(TableNameStudentInfo).
		Options(&options.FindOptions{}).
		Desc("created_time").
		Limit(2).
		Query()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("Query rows %d students %+v", len(students), students)
	for _, student := range students {
		log.Infof("student %+v", student)
	}
	var total int64
	var students2 []*docStudent
	total, err = e.Model(&students2).
		Select("name", "sex", "balance", "created_time").
		Options(&options.FindOptions{}).
		Table(TableNameStudentInfo).
		Page(0, 5).
		QueryEx()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("QueryEx total %d rows %d students %+v", total, len(students2), students2)
	for _, student := range students2 {
		log.Infof("student %+v", student)
	}

	var students3 []*docStudent
	err = e.Model(&students3).
		Select("name", "sex", "balance", "created_time").
		Options(&options.FindOptions{}).
		Table(TableNameStudentInfo).
		And("sex", "female").
		And("name", "kary").
		Or("age", bson.M{"$gte": 31}).
		//Page(0, 5).
		Query()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("Query rows %d students %+v", len(students3), students3)
	for _, student := range students3 {
		log.Infof("student %+v", student)
	}
}

func OrmCount(e *Engine) {
	rows, err := e.Model().
		Options(&options.CountOptions{}).
		Table(TableNameStudentInfo).
		Filter(bson.M{
			"name": "lory2",
		}).
		Count()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("rows %d", rows)
}

func OrmInsert(e *Engine) {
	var student = &docStudent{
		Id:          NewObjectID(),
		Name:        "john1",
		Sex:         "male",
		Age:         3,
		ClassNo:     "CLASS-3",
		Balance:     NewDecimal("532.324"),
		CreatedTime: time.Now(),
		ExtraData: extraData{
			IdCard:      "2023001",
			HomeAddress: "sz 003",
			Sports:      []string{"football", "badmin"},
		},
	}
	var students = []*docStudent{
		{
			//Id:        NewObjectID(), //auto generated
			Name:        "lory2",
			Sex:         "male",
			Age:         18,
			ClassNo:     "CLASS-1",
			CreatedTime: time.Now(),
			ExtraData: extraData{
				IdCard:      "2023002",
				HomeAddress: "sz no 101",
				Sports:      []string{"football", "baseball"},
			},
		},
		{
			//Id:        NewObjectID(), //auto generated
			Name:        "katy3",
			Sex:         "female",
			Age:         28,
			ClassNo:     "CLASS-2",
			CreatedTime: time.Now(),
			ExtraData: extraData{
				IdCard:      "2023003",
				HomeAddress: "london no 102",
				Sports:      []string{"singing", "dance"},
			},
		},
	}
	ids, err := e.Model(&student).
		Table(TableNameStudentInfo).
		Options(&options.InsertOneOptions{}).
		Insert()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("[Single] insert ids %+v", ids)
	ids, err = e.Model(&students).
		Options(&options.InsertManyOptions{}).
		Table(TableNameStudentInfo).
		Insert()
	if err != nil {
		log.Errorf(err.Error())
	}
	log.Infof("[Many] insert ids %+v", ids)
}

func OrmUpdate(e *Engine) {
	var err error
	_, err = e.Model().
		Table(TableNameStudentInfo).
		Options(&options.UpdateOptions{}).
		Id(officialObjectId).
		Set("name", "golang2006").
		Set("sex", "xx").
		Set("balance", NewDecimal("52.01")).
		Update()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	oid, err := NewObjectIDFromString(officialObjectId)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	var student = &docStudent{
		Id:          oid,
		Name:        "kary",
		Sex:         "female",
		Age:         39,
		Balance:     NewDecimal("123.456"),
		CreatedTime: time.Now(),
	}
	//UPDATE student_info SET name='kary', sex='female', age=39, created_time=NOW() WHERE _id='63e9f16b76527645cc38a815'
	_, err = e.Model(&student).
		Table(TableNameStudentInfo).
		Options(&options.UpdateOptions{}).
		Select("name", "sex", "age", "balance", "created_time").
		Update()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
}

func OrmUpsert(e *Engine) {
	var err error
	_, err = e.Model().
		Table(TableNameStudentInfo).
		Id(officialObjectId).
		Set("name", "rose").
		Set("sex", "female").
		Set("age", 18).
		Set("created_time", time.Now()).
		Set("balance", NewDecimal("520.1314")).
		Upsert()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
}

func OrmDelete(e *Engine) {

	rows, err := e.Model().
		Table(TableNameStudentInfo).
		Options(&options.DeleteOptions{}).
		//Filter(bson.M{
		//	"name": "lory2",
		//	"age":  18,
		//}).
		Id(officialObjectId).
		Delete()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("rows %d", rows)
}

type AggID struct {
	Name string `bson:"name"`
}

type StudentAgg struct {
	ID      AggID   `bson:"_id"`
	Age     float64 `bson:"age"`
	Total   int     `bson:"total"`
	Balance Decimal `bson:"balance"`
}

func OrmAggregate(e *Engine) {
	var agg []*StudentAgg
	err := e.Model(&agg).
		Table(TableNameStudentInfo).
		Avg("age").
		Sum("total", 1).
		Sum("balance").
		Eq("sex", "female").
		GroupBy("name", "age").
		Query()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("aggregate records %+v", len(agg))
	for _, a := range agg {
		log.Infof("%+v", a)
	}
}

func PipelineAggregate(e *Engine) {

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

	var agg []*StudentAgg
	// create match stage
	match := bson.D{
		{
			"$match", bson.M{
				"sex": "female",
			},
		},
	}
	// create group stage
	group := bson.D{
		{"$group", bson.M{
			"_id":   nil,
			"age":   bson.M{"$avg": "$age"},
			"total": bson.M{"$sum": 1},
		}}}
	//create projection stage
	project := bson.D{
		{
			"$project", bson.M{
				"_id":   0,
				"age":   1,
				"total": 1,
			},
		},
	}
	err := e.Model(&agg).
		Table(TableNameStudentInfo).
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
}
