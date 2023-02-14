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
	IdCard      string   `bson:"id_card"`
	HomeAddress string   `bson:"home_address"`
	Sports      []string `bson:"sports"`
}

type Student struct {
	Id        string    `bson:"_id,omitempty"`
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
	Query(e)
}

func Query(e *Engine) {
	var students  []*Student
	err := e.Model(&students).Table(TableNameStudentInfo).
		Equal("name", "libin").
		//Equal("age", 33).
		Query()
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("rows %d students %+v", len(students), students)
	for _, student := range students {
		log.Infof("student %+v", student)
	}
}

func Insert(e *Engine) {
	var student = &Student{
		Name:    "libin",
		Sex:     "male",
		Age:     43,
		ClassNo: "CLASS-3",
		ExtraData: ExtraData{
			IdCard:      "2023001",
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
				IdCard:      "2023002",
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
