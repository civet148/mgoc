package mgoc

import (
	"context"
	"github.com/civet148/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

const (
	defaultConnectTimeoutSeconds = 3     // connect timeout seconds
	defaultWriteTimeoutSeconds   = 60    // write timeout seconds
	defaultReadTimeoutSeconds    = 60    // read timeout seconds
	defaultPrimaryKeyName        = "_id" // database primary key name
)

type Option struct {
	Debug          bool                     // enable debug mode
	Max            int                      // max active connections
	Idle           int                      // max idle connections
	SSH            *SSH                     // SSH tunnel server config
	ConnectTimeout int                      // connect timeout
	WriteTimeout   int                      // write timeout seconds
	ReadTimeout    int                      // read timeout seconds
	DatabaseOpt    *options.DatabaseOptions // database options
}

type Engine struct {
	opt             *Option                // option for the engine
	client          *mongo.Client          // mongodb client
	db              *mongo.Database        // database instance
	strDatabaseName string                 // database name
	strPkName       string                 // primary key of table, default '_id'
	strTableName    string                 // table name
	modelType       ModelType              // model type
	models          []interface{}          // data model [struct object or struct slice]
	dict            map[string]interface{} // data model dictionary
	selected        bool                   // already selected, just append it
	selectColumns   []string               // columns to query: select
	dbTags          []string               // custom db tag names
}

func NewEngine(strDSN string, opts ...*Option) (*Engine, error) {
	var opt = makeOption(opts...)
	ctx, cancel := ContextWithTimeout(opt.ConnectTimeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(strDSN))
	if err != nil {
		return nil, log.Errorf("connect %s error [%s]", strDSN, err)
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, log.Errorf("ping %s error [%s]", strDSN, err)
	}
	var dbTags []string
	dbTags = append(dbTags, TAG_NAME_BSON, TAG_NAME_DB, TAG_NAME_JSON)
	ui := ParseUrl(strDSN)
	if ui.Database == "" {
		panic("no database found")
	}
	db := client.Database(ui.Database)
	return &Engine{
		opt:             opt,
		db:              db,
		client:          client,
		strDatabaseName: ui.Database,
		strPkName:       defaultPrimaryKeyName,
		models:          make([]interface{}, 0),
		dict:            make(map[string]interface{}),
		dbTags:          dbTags,
	}, nil
}

func makeOption(opts ...*Option) *Option {
	var opt *Option
	if len(opts) != 0 {
		opt = opts[0]
	} else {
		opt = &Option{
			ConnectTimeout: defaultConnectTimeoutSeconds,
			WriteTimeout:   defaultWriteTimeoutSeconds,
			ReadTimeout:    defaultReadTimeoutSeconds,
		}
	}
	return opt
}

func ContextWithTimeout(timeoutSeconds int) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
}

// Use clone another instance and switch to database specified
func (e *Engine) Use(strDatabase string) *Engine {
	return e.clone(strDatabase)
}

// Database get database instance specified
func (e *Engine) Database() *mongo.Database {
	return e.db
}

// Collection get collection instance specified
func (e *Engine) Collection(strName string, opts ...*options.CollectionOptions) *mongo.Collection {
	return e.db.Collection(strName, opts...)
}

// Model orm model
// use to get result set, support single struct object or slice [pointer type]
// notice: will clone a new engine object for orm operations(query/update/insert/upsert)
func (e *Engine) Model(args ...interface{}) *Engine {
	return e.clone(e.strDatabaseName, args...)
}

// Table set orm query table name
func (e *Engine) Table(strName string) *Engine {
	assert(strName, "table name is empty")
	e.setTableName(strName)
	return e
}

// Select orm select columns for filter
func (e *Engine) Select(strColumns ...string) *Engine {
	if e.setSelectColumns(strColumns...) {
		e.selected = true
	}
	return e
}

// orm query
// return rows affected and error, if err is not nil must be something wrong
// NOTE: Model function is must be called before call this function
func (e *Engine) Query() (rowsAffected int64, err error) {
	assert(e.models, "query model is nil")
	assert(e.strTableName, "table name not set")
	return
}

// Insert batch insert and returns id list
func (e *Engine) Insert() ([]interface{}, error) {
	assert(e.strTableName, "table name not set")
	if len(e.models) == 0 {
		return nil, log.Errorf("no document to insert")
	}
	var ids []interface{}
	ctx, cancel := ContextWithTimeout(e.opt.WriteTimeout)
	defer cancel()
	col := e.Collection(e.strTableName)
	if e.modelType == ModelType_Slice {
		res, err := col.InsertMany(ctx, e.models)
		if err != nil {
			return nil, log.Errorf(err.Error())
		}
		ids = res.InsertedIDs
	} else {
		res, err := col.InsertOne(ctx, e.models[0])
		if err != nil {
			return nil, log.Errorf(err.Error())
		}
		ids = append(ids, res.InsertedID)
	}
	return ids, nil
}

// Update update records
//func (e *Engine) Update(strName string) ([]interface{}, error) {
//	ctx, cancel := ContextWithTimeout(e.opt.WriteTimeout)
//	defer cancel()
//	col := e.Collection(e.strTableName)
//	res, err := col.UpdateMany(ctx, objects)
//	if err != nil {
//		return nil, log.Errorf(err.Error())
//	}
//}

// orm where condition
//func (e *Engine) Where(strWhere string, args ...interface{}) *Engine {
//	assert(strWhere, "string is nil")
//	//strWhere = e.formatString(strWhere, args...)
//	//e.setWhere(strWhere)
//	return e
//}

func (e *Engine) And(strColumn string, args ...interface{}) *Engine {
	//e.andConditions = append(e.andConditions, e.formatString(strColumn, args...))
	return e
}

func (e *Engine) Or(strColumn string, args ...interface{}) *Engine {
	//e.orConditions = append(e.orConditions, e.formatString(strColumn, args...))
	return e
}

func (e *Engine) Equal(strColumn string, value interface{}) *Engine {
	//e.And("%s='%v'", strColumn, value)
	return e
}

func (e *Engine) GreaterThan(strColumn string, value interface{}) *Engine {
	//e.And("%s>'%v'", strColumn, value)
	return e
}

func (e *Engine) GreaterEqual(strColumn string, value interface{}) *Engine {
	//e.And("%s>='%v'", strColumn, value)
	return e
}

func (e *Engine) LessThan(strColumn string, value interface{}) *Engine {
	//e.And("%s<'%v'", strColumn, value)
	return e
}

func (e *Engine) LessEqual(strColumn string, value interface{}) *Engine {
	//e.And("%s<='%v'", strColumn, value)
	return e
}
