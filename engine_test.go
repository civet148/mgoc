package mgoc

import (
	"github.com/civet148/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

const (
	objectId             = "63f46e3e9599d819bd5aebc4"
	TableNameStudentInfo = "student_info"
	defaultMongoUrl      = "mongodb://root:123456@192.168.20.108:27017/test?authSource=admin"
)

type ExtraData struct {
	IdCard      string   `bson:"id_card"`
	HomeAddress string   `bson:"home_address"`
	Sports      []string `bson:"sports"`
}

type Student struct {
	Id          string    `bson:"_id,omitempty"`
	Name        string    `bson:"name"`
	Sex         string    `bson:"sex"`
	Age         int       `bson:"age"`
	Balance     Decimal   `bson:"balance"`
	ClassNo     string    `bson:"class_no"`
	CreatedTime time.Time `bson:"created_time"`
	ExtraData   ExtraData `bson:"extra_data"`
}

func TestMongoDBCases(t *testing.T) {
	e, err := NewEngine(defaultMongoUrl)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	e.Debug(true)
	Insert(e)
	Query(e)
	Update(e)
	Count(e)
	Delete(e)
	Aggregate(e)
}

func Query(e *Engine) {
	var err error

	var student *Student
	err = e.Model(&student).
		Table(TableNameStudentInfo).
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
		var oid primitive.ObjectID
		oid, err = primitive.ObjectIDFromHex(student.Id)
		if err != nil {
			log.Errorf("decode object id [%s] error [%s]", student.Id, err.Error())
			return
		}
		log.Infof("student %+v oid [%s] create time [%s]", student, oid, student.CreatedTime.Local())
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
}

func Count(e *Engine) {
	rows, err := e.Model().
		Options(&options.CountOptions{}).
		Table(TableNameStudentInfo).
		Filter(bson.M{
			"name": "lory",
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
		Name:        "john",
		Sex:         "male",
		Age:         33,
		ClassNo:     "CLASS-3",
		Balance:     NewDecimal("532.324"),
		CreatedTime: time.Now(),
		ExtraData: ExtraData{
			IdCard:      "2023001",
			HomeAddress: "sz 003",
			Sports:      []string{"football", "badmin"},
		},
	}
	//var students = []*Student{
	//	{
	//		Name:        "lory",
	//		Sex:         "male",
	//		Age:         18,
	//		ClassNo:     "CLASS-1",
	//		CreatedTime: time.Now(),
	//		ExtraData: ExtraData{
	//			IdCard:      "2023002",
	//			HomeAddress: "sz no 101",
	//			Sports:      []string{"football", "baseball"},
	//		},
	//	},
	//	{
	//		Name:        "katy",
	//		Sex:         "female",
	//		Age:         28,
	//		ClassNo:     "CLASS-2",
	//		CreatedTime: time.Now(),
	//		ExtraData: ExtraData{
	//			IdCard:      "2023003",
	//			HomeAddress: "london no 102",
	//			Sports:      []string{"singing", "dance"},
	//		},
	//	},
	//}
	ids, err := e.Model(student).
		Table(TableNameStudentInfo).
		Options(&options.InsertOneOptions{}).
		Insert()
	if err != nil {
		log.Errorf(err.Error())
	}
	log.Infof("[Single] insert ids %+v", ids)
	//ids, err = e.Model(students).
	//	Options(&options.InsertManyOptions{}).
	//	Table(TableNameStudentInfo).
	//	Insert()
	//if err != nil {
	//	log.Errorf(err.Error())
	//}
	//log.Infof("[Many] insert ids %+v", ids)
	//mapStudent := map[string]interface{}{
	//	"name":         "juan",
	//	"sex":          "male",
	//	"age":          58,
	//	"class_no":     "CLASS-22",
	//	"created_time": time.Now(),
	//	"extra_data": ExtraData{
	//		IdCard:      "2023004",
	//		HomeAddress: "berlin no 108",
	//		Sports:      []string{"dance"},
	//	},
	//}
	//ids, err = e.Model(mapStudent).
	//	Options(&options.InsertOneOptions{}).
	//	Table(TableNameStudentInfo).
	//	Insert()
	//if err != nil {
	//	log.Errorf(err.Error())
	//}
	//log.Infof("[Map] insert ids %+v", ids)
}

func Update(e *Engine) {
	//oid, _ := primitive.ObjectIDFromHex(objectId)
	_, err := e.Model().
		Table(TableNameStudentInfo).
		Options(&options.UpdateOptions{}).
		Filter(bson.M{
			"_id": objectId,
		}).
		Set("name", "golang2006").
		Set("sex", "xx").
		Set("balance", "52.01").
		Update()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	var s *Student
	err = e.Model(&s).
		Table(TableNameStudentInfo).
		Id(objectId).
		Query()
	log.Infof("query updated student [%+v]", s)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	student := &Student{
		Id:   objectId,
		Name: "kary",
		Sex:  "female",
		Age:  39,
	}
	//UPDATE student_info SET name='kary', sex='female', age=39 WHERE _id='63e9f16b76527645cc38a815'
	_, err = e.Model(&student).
		Table(TableNameStudentInfo).
		Options(&options.UpdateOptions{}).
		Select("name", "sex", "age").
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
		Filter(bson.M{
			"name": "lory",
			"age":  23,
		}).
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
				{"name", "john"},
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
