package mgoc

import (
	"github.com/civet148/log"
	"testing"
)

const (
	TableNameStudentInfo = "student_info"
	defaultMongoUrl      = "mongodb://root:123456@192.168.20.108:27017/test?authSource=admin"
)

type ExtraData struct {
	HomeAddress string   `bson:"home_address"`
	Sports      []string `bson:"sports"`
}

type Student struct {
	id        string    `bson:"_id"`
	Name      string    `bson:"name"`
	Sex       string    `bson:"sex"`
	Age       int       `bson:"age"`
	ClassNo   string    `bson:"class_no"`
	ExtraData ExtraData `bson:"extra_data"`
}

func TestMongoDBCases(t *testing.T) {
	e, err := NewEngine(defaultMongoUrl)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	_ = e
	Insert(e)
}

func Insert(e *Engine) {
	var student = &Student{
		id:      "",
		Name:    "libin",
		Sex:     "male",
		Age:     43,
		ClassNo: "CLASS-3",
		ExtraData: ExtraData{
			HomeAddress: "sz 003",
			Sports:      []string{"football", "badmin"},
		},
	}
	var students = []*Student{
		{
			Name:    "lory",
			Sex:     "male",
			Age:     18,
			ClassNo: "CLASS-1",
			ExtraData: ExtraData{
				HomeAddress: "sz no 101",
				Sports:      []string{"football", "baseball"},
			},
		},
		{
			Name:    "katy sky",
			Sex:     "female",
			Age:     28,
			ClassNo: "CLASS-2",
			ExtraData: ExtraData{
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
