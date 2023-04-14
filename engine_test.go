package mgoc

import (
	"encoding/hex"
	"github.com/civet148/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	//bson2 "gopkg.in/mgo.v2/bson"
	bson2 "gopkg.in/mgo.v2/bson"
	"testing"
	"time"
)

const (
	officialObjectId     = "6438f32fd71fc42e601558aa"
	TableNameStudentInfo = "student_info"
	defaultMongoUrl      = "mongodb://root:123456@192.168.2.9:27017/test?authSource=admin"
)

type ExtraData struct {
	IdCard      string   `json:"id_card" bson:"id_card"`
	HomeAddress string   `json:"home_address" bson:"home_address"`
	Sports      []string `json:"sports" bson:"sports"`
}

type Student struct {
	Id bson2.ObjectId `json:"_id,omitempty" bson:"_id,omitempty"`
	//Id          string    `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string    `json:"name" bson:"name"`
	Sex         string    `json:"sex" bson:"sex"`
	Age         int       `json:"age" bson:"age"`
	Balance     Decimal   `json:"balance" bson:"balance"`
	ClassNo     string    `json:"class_no" bson:"class_no"`
	CreatedTime time.Time `json:"created_time" bson:"created_time"`
	ExtraData   ExtraData `json:"extra_data" bson:"extra_data"`
}

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
	//Insert(e)
	//Query(e)
	//Update(e)
	//Count(e)
	Delete(e)
	//Aggregate(e)
}

func Query(e *Engine) {
	var err error
	var student *Student
	sid := hex.EncodeToString([]byte(officialObjectId)) //mgo.v2 ObjectId string
	err = e.Model(&student).
		Table(TableNameStudentInfo).
		Equal("_id", sid).
		Query()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("single student %+v", student)

	var students []*Student
	err = e.Model(&students).
		Table(TableNameStudentInfo).
		Options(&options.FindOptions{}).
		Filter(bson.M{
			"name": "john",
			//"age":  18,
		}).
		//Equal("extra_data.id_card", "2023001").
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
	var students2 []*Student
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

	var students3 []*Student
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

func Count(e *Engine) {
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

func Insert(e *Engine) {
	var student = &Student{
		Name:        "john1",
		Sex:         "male",
		Age:         3,
		ClassNo:     "CLASS-3",
		Balance:     NewDecimal("532.324"),
		CreatedTime: time.Now(),
		ExtraData: ExtraData{
			IdCard:      "2023001",
			HomeAddress: "sz 003",
			Sports:      []string{"football", "badmin"},
		},
	}
	var students = []*Student{
		{
			Id:          bson2.NewObjectId(),
			Name:        "lory2",
			Sex:         "male",
			Age:         18,
			ClassNo:     "CLASS-1",
			CreatedTime: time.Now(),
			ExtraData: ExtraData{
				IdCard:      "2023002",
				HomeAddress: "sz no 101",
				Sports:      []string{"football", "baseball"},
			},
		},
		{
			Id:          bson2.NewObjectId(),
			Name:        "katy3",
			Sex:         "female",
			Age:         28,
			ClassNo:     "CLASS-2",
			CreatedTime: time.Now(),
			ExtraData: ExtraData{
				IdCard:      "2023003",
				HomeAddress: "london no 102",
				Sports:      []string{"singing", "dance"},
			},
		},
	}
	ids, err := e.Model(student).
		Table(TableNameStudentInfo).
		Options(&options.InsertOneOptions{}).
		Insert()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("[Single] insert ids %+v", ids)
	ids, err = e.Model(students).
		Options(&options.InsertManyOptions{}).
		Table(TableNameStudentInfo).
		Insert()
	if err != nil {
		log.Errorf(err.Error())
	}
	log.Infof("[Many] insert ids %+v", ids)
}

func Update(e *Engine) {
	var err error
	//_, err = e.Model().
	//	Table(TableNameStudentInfo).
	//	Options(&options.UpdateOptions{}).
	//	Filter(bson.M{
	//		"_id": officialObjectId,
	//	}).
	//	Set("name", "golang2006").
	//	Set("sex", "xx").
	//	Set("balance", NewDecimal("52.01")).
	//	Update()
	//if err != nil {
	//	log.Errorf(err.Error())
	//	return
	//}
	//
	var s *Student
	err = e.Model(&s).
		Table(TableNameStudentInfo).
		Id(officialObjectId).
		Query()
	log.Infof("query updated student [%+v]", s)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	var student = &Student{
		Id:          officialObjectId,
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

func Delete(e *Engine) {

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

type StudentAgg struct {
	Age   int `bson:"age"`
	Total int `bson:"total"`
}

func Aggregate(e *Engine) {

	/*
		db.getCollection("student_info").aggregate([
		   {
		     "$match":{
				    "name":"john"
			   },
			 },
			 {
			   "$group":{
			      		"_id":"$name",
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
			"$match", bson.D{
				{"sex", "female"},
			},
		},
	}
	// create group stage
	group := bson.D{
		{"$group", bson.D{
			{"_id", "$name"},
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
