package mgoc

import (
	"context"
	"github.com/civet148/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"sync"
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
	debug         bool                   // enable debug mode
	engineOpt     *Option                // option for the engine
	options       []interface{}          // mongodb operation options (find/update/delete/insert...)
	client        *mongo.Client          // mongodb client
	db            *mongo.Database        // database instance
	strPkName     string                 // primary key of table, default '_id'
	strTableName  string                 // table name
	modelType     ModelType              // model type
	models        []interface{}          // data model [struct object or struct slice]
	dict          map[string]interface{} // data model dictionary
	selectColumns []string               // select columns to query/update
	exceptColumns map[string]bool        // except columns to query/update
	andConditions map[string]interface{} // AND conditions to query
	orConditions  map[string]interface{} // OR conditions to query
	ascColumns    []string               // columns to order by ASC
	descColumns   []string               // columns to order by DESC
	skip          int64                  // mongodb skip
	limit         int64                  // mongodb limit
	filter        bson.M                 // mongodb filter
	updates       bson.M                 // mongodb updates
	pipeline      mongo.Pipeline         // mongodb pipeline
	locker        sync.RWMutex           // internal locker
}

func NewEngine(strDSN string, opts ...*Option) (*Engine, error) {
	var opt = makeOption(opts...)
	ctx, cancel := ContextWithTimeout(opt.ConnectTimeout)
	defer cancel()
	if opt != nil && opt.SSH != nil {
		var err error
		strDSN, err = opt.SSH.openSSHTunnel(strDSN)
		if err != nil {
			log.Panic(err.Error())
			return nil, err
		}
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(strDSN))
	if err != nil {
		return nil, log.Errorf("connect %s error [%s]", strDSN, err)
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, log.Errorf("ping %s error [%s]", strDSN, err)
	}
	ui := ParseUrl(strDSN)
	var db *mongo.Database
	if ui.Database != "" {
		db = client.Database(ui.Database)
	}
	var debug bool
	if opt != nil && opt.Debug {
		debug = true
	}
	return &Engine{
		debug:         debug,
		engineOpt:     opt,
		db:            db,
		client:        client,
		strPkName:     defaultPrimaryKeyName,
		models:        make([]interface{}, 0),
		exceptColumns: make(map[string]bool),
		dict:          make(map[string]interface{}),
		filter:        make(map[string]interface{}),
		updates:       make(map[string]interface{}),
		andConditions: make(map[string]interface{}),
		orConditions:  make(map[string]interface{}),
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
func (e *Engine) Use(strDatabase string, opts ...*options.DatabaseOptions) *Engine {
	if e.db == nil {
		e.db = e.client.Database(strDatabase, opts...)
		return e
	}
	return e.clone(strDatabase, nil)
}

// Database get database instance specified
func (e *Engine) Database() *mongo.Database {
	return e.db
}

// Collection get collection instance specified
func (e *Engine) Collection(strName string, opts ...*options.CollectionOptions) *mongo.Collection {
	return e.db.Collection(strName, opts...)
}

func (e *Engine) PrimaryKey() string {
	return defaultPrimaryKeyName
}

// Model orm model
// use to get result set, support single struct object or slice [pointer type]
// notice: will clone a new engine object for orm operations(query/update/insert/upsert)
func (e *Engine) Model(args ...interface{}) *Engine {
	if e.db == nil {
		log.Panic("no database specified")
	}
	return e.clone(e.db.Name(), args...)
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
	e.replaceInsertModels()
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
	e.makeUpdates()
	e.debugJson("filter", e.filter, "updates", e.updates)
	if len(e.filter) == 0 {
		return 0, log.Errorf("filter is empty")
	}
	res, err := col.UpdateMany(ctx, e.filter, e.updates, opts...)
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	return res.ModifiedCount, nil
}

// UpdateOne update one document
func (e *Engine) UpdateOne() (rows int64, err error) {
	defer e.clean()
	ctx, cancel := ContextWithTimeout(e.engineOpt.WriteTimeout)
	defer cancel()
	col := e.Collection(e.strTableName)
	var opts []*options.UpdateOptions
	for _, opt := range e.options {
		opts = append(opts, opt.(*options.UpdateOptions))
	}
	e.makeUpdates()
	e.debugJson("filter", e.filter, "updates", e.updates)
	if len(e.filter) == 0 {
		return 0, log.Errorf("filter is empty")
	}
	res, err := col.UpdateOne(ctx, e.filter, e.updates, opts...)
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	return res.ModifiedCount, nil
}

// FindOneUpdate find single document and update
func (e *Engine) FindOneUpdate() (res *mongo.SingleResult, err error) {
	defer e.clean()
	ctx, cancel := ContextWithTimeout(e.engineOpt.WriteTimeout)
	defer cancel()
	col := e.Collection(e.strTableName)
	var opts []*options.FindOneAndUpdateOptions
	for _, opt := range e.options {
		opts = append(opts, opt.(*options.FindOneAndUpdateOptions))
	}
	e.makeUpdates()
	e.debugJson("filter", e.filter, "updates", e.updates)
	if len(e.filter) == 0 {
		return nil, log.Errorf("filter is empty")
	}
	res = col.FindOneAndUpdate(ctx, e.filter, e.updates, opts...)
	err = res.Err()
	if err != nil {
		return nil, log.Errorf(err.Error())
	}
	return res, nil
}

// FindOneReplace find single document and replace
func (e *Engine) FindOneReplace() (res *mongo.SingleResult, err error) {
	defer e.clean()
	ctx, cancel := ContextWithTimeout(e.engineOpt.WriteTimeout)
	defer cancel()
	col := e.Collection(e.strTableName)
	var opts []*options.FindOneAndReplaceOptions
	for _, opt := range e.options {
		opts = append(opts, opt.(*options.FindOneAndReplaceOptions))
	}
	e.makeUpdates()
	e.debugJson("filter", e.filter, "updates", e.updates)
	if len(e.filter) == 0 {
		return nil, log.Errorf("filter is empty")
	}
	res = col.FindOneAndReplace(ctx, e.filter, e.updates, opts...)
	err = res.Err()
	if err != nil {
		return nil, log.Errorf(err.Error())
	}
	return res, nil
}

// FindOneDelete find single document and delete
func (e *Engine) FindOneDelete() (res *mongo.SingleResult, err error) {
	defer e.clean()
	ctx, cancel := ContextWithTimeout(e.engineOpt.WriteTimeout)
	defer cancel()
	col := e.Collection(e.strTableName)
	var opts []*options.FindOneAndDeleteOptions
	for _, opt := range e.options {
		opts = append(opts, opt.(*options.FindOneAndDeleteOptions))
	}
	e.debugJson("filter", e.filter)
	if len(e.filter) == 0 {
		return nil, log.Errorf("filter is empty")
	}
	res = col.FindOneAndDelete(ctx, e.filter, opts...)
	err = res.Err()
	if err != nil {
		return nil, log.Errorf(err.Error())
	}
	return res, nil
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
	if len(e.filter) == 0 {
		return 0, log.Errorf("filter is empty")
	}
	e.debugJson("filter", e.filter, "options", opts)
	res, err := col.DeleteMany(ctx, e.filter, opts...)
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	return res.DeletedCount, nil
}

// Query orm query
// return error if err is not nil must be something wrong
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
	e.makeFilters()

	opts := e.makeFindOptions()
	e.debugJson("filter", e.filter, "options", opts)
	cur, err = col.Find(ctx, e.filter, opts...)
	if err != nil {
		return log.Errorf(err.Error())
	}
	defer cur.Close(ctx)
	err = e.fetchRows(cur)
	if err != nil {
		return log.Errorf(err.Error())
	}
	return nil
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
// return total and error, if err is not nil must be something wrong
// NOTE: Model function is must be called before call this function
func (e *Engine) QueryEx() (total int64, err error) {
	assert(e.models, "query model is nil")
	assert(e.strTableName, "table name not set")
	if len(e.models) == 0 {
		return 0, log.Errorf("no model to fetch records")
	}
	defer e.clean()
	ctx, cancel := ContextWithTimeout(e.engineOpt.ReadTimeout)
	defer cancel()
	col := e.Collection(e.strTableName)
	e.makeFilters()
	opts := e.makeFindOptions()
	var cur *mongo.Cursor
	e.debugJson("filter", e.filter, "options", opts)
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	total, err = col.CountDocuments(ctx, e.filter)
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	cur, err = col.Find(ctx, e.filter, opts...)
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	defer cur.Close(ctx)
	err = e.fetchRows(cur)
	if err != nil {
		return 0, log.Errorf(err.Error())
	}
	return total, nil
}

func (e *Engine) Limit(n int) *Engine {
	e.limit = int64(n)
	return e
}

func (e *Engine) Id(v interface{}) *Engine {
	e.filter[defaultPrimaryKeyName] = ObjectID(v)
	return e
}

// Select orm select columns for projection
func (e *Engine) Select(strColumns ...string) *Engine {
	e.setSelectColumns(strColumns...)
	return e
}

// Except insert/update all except columns
func (e *Engine) Except(strColumns ...string) *Engine {
	e.setExceptColumns(strColumns...)
	return e
}

// Pipeline aggregate pipeline
func (e *Engine) Pipeline(pipelines ...bson.D) *Engine {
	var pipeline = mongo.Pipeline{}
	for _, v := range pipelines {
		if v != nil {
			pipeline = append(pipeline, v)
		}
	}
	e.pipeline = pipeline
	return e
}

// Aggregate execute aggregate pipeline
func (e *Engine) Aggregate() (err error) {
	assert(e.models, "query model is nil")
	assert(e.pipeline, "pipeline is nil")

	defer e.clean()
	var opts []*options.AggregateOptions
	for _, opt := range e.options {
		opts = append(opts, opt.(*options.AggregateOptions))
	}
	ctx, cancel := ContextWithTimeout(e.engineOpt.ReadTimeout)
	defer cancel()
	var cur *mongo.Cursor
	e.debugJson("pipeline", e.pipeline)
	if e.strTableName == "" {
		cur, err = e.db.Aggregate(ctx, e.pipeline, opts...)
	} else {
		col := e.Collection(e.strTableName)
		cur, err = col.Aggregate(ctx, e.pipeline, opts...)
	}
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

// GeoCenterSphere query by coordinate and distance in meters (sphere)
func (e *Engine) GeoCenterSphere(strColumn string, pos Coordinate, distance int) *Engine {
	var rad = Radian(uint64(distance))
	center := []interface{}{
		FloatArray{pos.X, pos.Y},
		rad,
	}
	e.filter[strColumn] = bson.M{
		KeyGeoWithin: bson.M{
			KeyCenterSphere: center,
		},
	}
	return e
}

// Geometry query by geometry
func (e *Engine) Geometry(strColumn string, geometry *Geometry) *Engine {
	e.filter[strColumn] = bson.M{
		KeyGeoWithin: bson.M{
			KeyGeoMetry: geometry,
		},
	}
	return e
}

// GeoNearByPoint query and return matched records with max distance in meters (just one index, 2d or 2dshpere)
// column: the column which include location
// pos: the position to query
// maxDistance: the maximum distance nearby pos (meters)
// disFieldName: distance column name to return
func (e *Engine) GeoNearByPoint(column string, pos Coordinate, maxDistance int, disFieldName string) *Engine {
	/*
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
	*/
	var point = NewGeoPoint(pos)
	match := bson.D{
		{KeyGeoNear, bson.M{
			columnNameNear:          point,
			columnNameDistanceField: disFieldName,
			columnNameMaxDistance:   maxDistance,
			columnNameIncludeLocs:   column,
			columnNameSpherical:     true,
		}},
	}
	e.Pipeline(match, nil)
	return e
}

// Set update columns specified
func (e *Engine) Set(strColumn string, value interface{}) *Engine {
	if strColumn == e.PrimaryKey() {
		return e
	}
	m, ok := e.updates[KeySet]
	if !ok {
		e.updates[KeySet] = bson.M{
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
		KeyIn: value,
	}
	return e
}

func (e *Engine) And(strColumn string, value interface{}) *Engine {
	e.setAndCondition(strColumn, value)
	return e
}

func (e *Engine) Or(strColumn string, value interface{}) *Engine {
	e.setOrCondition(strColumn, value)
	return e
}

func (e *Engine) Equal(strColumn string, value interface{}) *Engine {
	var v interface{}
	v = ConvertValue(strColumn, value)
	e.filter[strColumn] = bson.M{
		KeyEqual: v,
	}
	return e
}

func (e *Engine) NotEqual(strColumn string, value interface{}) *Engine {
	var v interface{}
	v = ConvertValue(strColumn, value)
	e.filter[strColumn] = bson.M{
		KeyNotEqual: v,
	}
	return e
}

func (e *Engine) GreaterThan(strColumn string, value interface{}) *Engine {
	var v interface{}
	v = ConvertValue(strColumn, value)
	e.filter[strColumn] = bson.M{
		KeyGreaterThan: v,
	}
	return e
}

func (e *Engine) GreaterEqual(strColumn string, value interface{}) *Engine {
	var v interface{}
	v = ConvertValue(strColumn, value)
	e.filter[strColumn] = bson.M{
		KeyGreaterEqual: v,
	}
	return e
}

func (e *Engine) LessThan(strColumn string, value interface{}) *Engine {
	var v interface{}
	v = ConvertValue(strColumn, value)
	e.filter[strColumn] = bson.M{
		KeyLessThan: v,
	}
	return e
}

func (e *Engine) LessEqual(strColumn string, value interface{}) *Engine {
	var v interface{}
	v = ConvertValue(strColumn, value)
	e.filter[strColumn] = bson.M{
		KeyLessEqual: v,
	}
	return e
}

func (e *Engine) GreaterLessThan(strColumn string, value1, value2 interface{}) *Engine {
	var v1, v2 interface{}
	v1 = ConvertValue(strColumn, value1)
	v2 = ConvertValue(strColumn, value2)
	e.filter[strColumn] = bson.M{
		KeyGreaterThan: v1,
		KeyLessThan:    v2,
	}
	return e
}

func (e *Engine) GreaterLessEqual(strColumn string, value1, value2 interface{}) *Engine {
	var v1, v2 interface{}
	v1 = ConvertValue(strColumn, value1)
	v2 = ConvertValue(strColumn, value2)
	e.filter[strColumn] = bson.M{
		KeyGreaterEqual: v1,
		KeyLessEqual:    v2,
	}
	return e
}

func (e *Engine) GreaterThanLessEqual(strColumn string, value1, value2 interface{}) *Engine {
	var v1, v2 interface{}
	v1 = ConvertValue(strColumn, value1)
	v2 = ConvertValue(strColumn, value2)
	e.filter[strColumn] = bson.M{
		KeyGreaterThan: v1,
		KeyLessEqual:   v2,
	}
	return e
}

func (e *Engine) GreaterEqualLessThan(strColumn string, value1, value2 interface{}) *Engine {
	var v1, v2 interface{}
	v1 = ConvertValue(strColumn, value1)
	v2 = ConvertValue(strColumn, value2)
	e.filter[strColumn] = bson.M{
		KeyGreaterEqual: v1,
		KeyLessThan:     v2,
	}
	return e
}

func (e *Engine) Regex(strColumn string, value interface{}) *Engine {
	var v interface{}
	v = ConvertValue(strColumn, value)
	e.filter[strColumn] = bson.M{
		KeyRegex: v,
	}
	return e
}

func (e *Engine) Exists(strColumn string, value bool) *Engine {
	var v interface{}
	v = ConvertValue(strColumn, value)
	e.filter[strColumn] = bson.M{
		KeyExists: v,
	}
	return e
}

func (e *Engine) Array(strColumn string, value []interface{}) *Engine {
	var v interface{}
	v = ConvertValue(strColumn, value)
	e.filter[strColumn] = v
	return e
}

func (e *Engine) Page(pageNo, pageSize int) *Engine {
	if pageSize != 0 {
		e.limit = int64(pageSize)
		e.skip = int64(pageSize * pageNo)
	}
	return e
}
