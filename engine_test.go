package mgoc

import (
	"github.com/civet148/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

const (
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
	//Insert(e)
	//Query(e)
	Update(e)
	//Count(e)
	//Delete(e)
}

func Query(e *Engine) {
	var students []*Student
	err := e.Model(&students).
		Table(TableNameStudentInfo).
		Filter(bson.M{
			"name": "libin",
			//"age":  33,
		}).
		Equal("extra_data.id_card", "2023001").
		Desc("created_time").
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
		log.Infof("student %+v oid [%s]", student, oid)
	}
	var total int64
	var students2 []*Student
	total, err = e.Model(&students2).
		Select("name", "sex").
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
		Table(TableNameStudentInfo).
		Filter(bson.M{
			"name": "libin",
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
		Name:        "libin",
		Sex:         "male",
		Age:         33,
		ClassNo:     "CLASS-3",
		CreatedTime: time.Now(),
		ExtraData: ExtraData{
			IdCard:      "2023001",
			HomeAddress: "sz 003",
			Sports:      []string{"football", "badmin"},
		},
	}
	var students = []*Student{
		{
			Name:        "lory",
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
			Name:        "katy sky",
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
	ids, err := e.Model(student).Table(TableNameStudentInfo).Insert()
	if err != nil {
		log.Errorf(err.Error())
	}
	log.Infof("[Single] insert ids %+v", ids)
	ids, err = e.Model(students).Table(TableNameStudentInfo).Insert()
	if err != nil {
		log.Errorf(err.Error())
	}
	log.Infof("[Many] insert ids %+v", ids)
	mapStudent := map[string]interface{}{
		"name":     "covlaent",
		"sex":      "female",
		"age":      58,
		"class_no": "CLASS-22",
		"extra_data": ExtraData{
			IdCard:      "2023004",
			HomeAddress: "berlin no 108",
			Sports:      []string{"dance"},
		},
	}
	ids, err = e.Model(mapStudent).Table(TableNameStudentInfo).Insert()
	if err != nil {
		log.Errorf(err.Error())
	}
	log.Infof("[Map] insert ids %+v", ids)
}

func Update(e *Engine) {
	//oid, _ := primitive.ObjectIDFromHex("63e9f16b76527645cc38a815")
	rows, err := e.Model().
		Table(TableNameStudentInfo).
		Filter(bson.M{
			"_id": "63e9f16b76527645cc38a815",
		}).
		Set("name", "libin815").
		Set("sex", "xx").
		Update()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("rows %d", rows)
}

func Delete(e *Engine) {

	rows, err := e.Model().
		Table(TableNameStudentInfo).
		Filter(bson.M{
			"name": "libin",
			"age":  33,
		}).
		Delete()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("rows %d", rows)
}
