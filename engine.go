package mgoc

import (
	"context"
	"github.com/civet148/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
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
	debug           bool                   // enable debug mode
	engineOpt       *Option                // option for the engine
	options         []interface{}          // mongodb operation options (find/update/delete/insert...)
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
	ascColumns      []string               // columns to order by ASC
	descColumns     []string               // columns to order by DESC
	dbTags          []string               // custom db tag names
	skip            int64                  // mongodb skip
	limit           int64                  // mongodb limit
	filter          bson.M                 // mongodb filter
	updates         bson.M                 // mongodb updates
	pipeline        mongo.Pipeline         // mongo pipeline
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
		engineOpt:       opt,
		db:              db,
		client:          client,
		strDatabaseName: ui.Database,
		strPkName:       defaultPrimaryKeyName,
		models:          make([]interface{}, 0),
		dict:            make(map[string]interface{}),
		filter:          make(map[string]interface{}),
		updates:         make(map[string]interface{}),
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

func (e *Engine) Close() error {
	return e.client.Disconnect(context.TODO())
}

func (e *Engine) Debug(on bool) {
	e.debug = on
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

// Options set operation options for find/update/delete/insert...
func (e *Engine) Options(options ...interface{}) *Engine {
	e.options = options
	return e
}

// Insert batch insert and returns id list
func (e *Engine) Insert() ([]interface{}, error) {
	assert(e.strTableName, "table name not set")
	if len(e.models) == 0 {
		return nil, log.Errorf("no document to insert")
	}
	defer e.clean()
	var ids []interface{}
	ctx, cancel := ContextWithTimeout(e.engineOpt.WriteTimeout)
	defer cancel()
	col := e.Collection(e.strTableName)
	if e.modelType == ModelType_Slice {
		var opts []*options.InsertManyOptions
		for _, opt := range e.options {
			opts = append(opts, opt.(*options.InsertManyOptions))
		}
		res, err := col.InsertMany(ctx, e.models, opts...)
		if err != nil {
			return nil, log.Errorf(err.Error())
		}
		ids = res.InsertedIDs
	} else {
		var opts []*options.InsertOneOptions
		for _, opt := range e.options {
			opts = append(opts, opt.(*options.InsertOneOptions))
		}
		res, err := col.InsertOne(ctx, e.models[0], opts...)
		if err != nil {
			return nil, log.Errorf(err.Error())
		}
		ids = append(ids, res.InsertedID)
	}
	return ids, nil
}

// Update update records
func (e *Engine) Update() (rows int64, err error) {
	defer e.clean()
	ctx, cancel := ContextWithTimeout(e.engineOpt.WriteTimeout)
	defer cancel()
	col := e.Collection(e.strTableName)
	var opts []*options.UpdateOptions
	for _, opt := range e.options {
		opts = append(opts, opt.(*options.UpdateOptions))
	}
	e.debugJson("filter", e.filter, "updates", e.updates)
	res, err := col.UpdateMany(ctx, e.filter, e.updates, opts...)
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	return res.ModifiedCount, nil
}

// Delete delete many records
func (e *Engine) Delete() (rows int64, err error) {
	defer e.clean()
	ctx, cancel := ContextWithTimeout(e.engineOpt.WriteTimeout)
	defer cancel()
	col := e.Collection(e.strTableName)
	var opts []*options.DeleteOptions
	for _, opt := range e.options {
		opts = append(opts, opt.(*options.DeleteOptions))
	}
	e.debugJson("filter", e.filter, "options", opts)
	res, err := col.DeleteMany(ctx, e.filter, opts...)
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	return res.DeletedCount, nil
}

// Query orm query
// return rows affected and error, if err is not nil must be something wrong
// NOTE: Model function is must be called before call this function
func (e *Engine) Query() (err error) {
	assert(e.models, "query model is nil")
	assert(e.strTableName, "table name not set")
	if len(e.models) == 0 {
		return log.Errorf("no model to fetch records")
	}
	defer e.clean()
	ctx, cancel := ContextWithTimeout(e.engineOpt.ReadTimeout)
	defer cancel()
	col := e.Collection(e.strTableName)
	var cur *mongo.Cursor
	opts := e.makeFindOptions()
	e.debugJson("filter", e.filter, "options", opts)
	cur, err = col.Find(ctx, e.filter, opts...)
	if err != nil {
		return log.Errorf(err.Error())
	}
	defer cur.Close(ctx)
	err = cur.All(ctx, e.models[0])
	if err != nil {
		return log.Errorf(err.Error())
	}
	return
}

// Count orm count documents
func (e *Engine) Count() (rows int64, err error) {
	assert(e.strTableName, "table name not set")
	defer e.clean()
	ctx, cancel := ContextWithTimeout(e.engineOpt.ReadTimeout)
	defer cancel()
	col := e.Collection(e.strTableName)
	var opts []*options.CountOptions
	for _, opt := range e.options {
		opts = append(opts, opt.(*options.CountOptions))
	}
	e.debugJson("filter", e.filter, "options", opts)
	rows, err = col.CountDocuments(ctx, e.filter, opts...)
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	return rows, nil
}

// QueryEx orm query and return total records count
// return rows affected and error, if err is not nil must be something wrong
// NOTE: Model function is must be called before call this function
func (e *Engine) QueryEx() (rows int64, err error) {
	assert(e.models, "query model is nil")
	assert(e.strTableName, "table name not set")
	if len(e.models) == 0 {
		return 0, log.Errorf("no model to fetch records")
	}
	defer e.clean()
	ctx, cancel := ContextWithTimeout(e.engineOpt.ReadTimeout)
	defer cancel()
	col := e.Collection(e.strTableName)

	opts := e.makeFindOptions()
	var cur *mongo.Cursor
	e.debugJson("filter", e.filter, "options", opts)
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	rows, err = col.CountDocuments(ctx, e.filter)
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	cur, err = col.Find(ctx, e.filter, opts...)
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	defer cur.Close(ctx)
	err = cur.All(ctx, e.models[0])
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	return
}

// Select orm select columns for projection
func (e *Engine) Select(strColumns ...string) *Engine {
	if e.setSelectColumns(strColumns...) {
		e.selected = true
	}
	return e
}

// Pipeline aggregate pipeline
func (e *Engine) Pipeline(match, group bson.D, args ...bson.D) *Engine {
	var pipeline = mongo.Pipeline{}
	pipeline = append(pipeline, match)
	pipeline = append(pipeline, group)
	for _, v := range args {
		if v != nil {
			pipeline = append(pipeline, v)
		}
	}
	e.pipeline = pipeline
	return e
}

func (e *Engine) Aggregate() (err error) {
	assert(e.models, "query model is nil")
	assert(e.strTableName, "table name not set")
	assert(e.pipeline, "pipeline is nil")

	defer e.clean()
	var opts []*options.AggregateOptions
	for _, opt := range e.options {
		opts = append(opts, opt.(*options.AggregateOptions))
	}
	ctx, cancel := ContextWithTimeout(e.engineOpt.ReadTimeout)
	defer cancel()
	col := e.Collection(e.strTableName)
	e.debugJson("pipeline", e.pipeline)
	var cur *mongo.Cursor
	cur, err = col.Aggregate(ctx, e.pipeline, opts...)
	if err != nil {
		return log.Errorf(err.Error())
	}
	defer cur.Close(ctx)
	err = cur.All(ctx, e.models[0])
	if err != nil {
		return log.Errorf(err.Error())
	}
	return nil
}

// Asc orm select columns for ORDER BY ASC
func (e *Engine) Asc(strColumns ...string) *Engine {
	e.setAscColumns(strColumns...)
	return e
}

// Desc orm select columns for ORDER BY DESC
func (e *Engine) Desc(strColumns ...string) *Engine {
	e.setDescColumns(strColumns...)
	return e
}

// Filter orm condition
func (e *Engine) Filter(filter bson.M) *Engine {
	assert(filter, "filter cannot be nil")
	e.filter = e.replaceObjectID(filter)
	return e
}

// Set update columns specified
func (e *Engine) Set(strColumn string, value interface{}) *Engine {
	m, ok := e.updates[keySet]
	if !ok {
		e.updates[keySet] = bson.M{
			strColumn: value,
		}
	} else {
		bm := m.(bson.M)
		bm[strColumn] = value
	}
	return e
}

func (e *Engine) In(strColumn string, value interface{}) *Engine {
	e.filter[strColumn] = bson.M{
		keyIn: value,
	}
	return e
}

func (e *Engine) And(value bson.A) *Engine {
	e.filter[keyAnd] = value
	return e
}

func (e *Engine) Or(value bson.A) *Engine {
	e.filter[keyOr] = value
	return e
}

func (e *Engine) Equal(strColumn string, value interface{}) *Engine {
	e.filter[strColumn] = bson.M{
		keyEqual: value,
	}
	return e
}

func (e *Engine) NotEqual(strColumn string, value interface{}) *Engine {
	e.filter[strColumn] = bson.M{
		keyNotEqual: value,
	}
	return e
}

func (e *Engine) GreaterThan(strColumn string, value interface{}) *Engine {
	e.filter[strColumn] = bson.M{
		keyGreaterThan: value,
	}
	return e
}

func (e *Engine) GreaterEqual(strColumn string, value interface{}) *Engine {
	e.filter[strColumn] = bson.M{
		keyGreaterEqual: value,
	}
	return e
}

func (e *Engine) LessThan(strColumn string, value interface{}) *Engine {
	e.filter[strColumn] = bson.M{
		keyLessThan: value,
	}
	return e
}

func (e *Engine) LessEqual(strColumn string, value interface{}) *Engine {
	e.filter[strColumn] = bson.M{
		keyLessEqual: value,
	}
	return e
}

func (e *Engine) Regex(strColumn string, value interface{}) *Engine {
	e.filter[strColumn] = bson.M{
		keyRegex: value,
	}
	return e
}

func (e *Engine) Exists(strColumn string, value bool) *Engine {
	e.filter[strColumn] = bson.M{
		keyExists: value,
	}
	return e
}

func (e *Engine) Page(pageNo, pageSize int) *Engine {
	if pageSize != 0 {
		e.limit = int64(pageSize)
		e.skip = int64(pageSize * pageNo)
	}
	return e
}
