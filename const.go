package mgoc


const (
	keyIn           = "$in"
	keyEqual        = "$eq"
	keyAnd          = "$and"
	keyOr           = "$or"
	keyGreaterThan  = "$gt"
	keyGreaterEqual = "$gte"
	keyLessThan     = "$lt"
	keyLessEqual    = "$lte"
	keyNotEqual     = "$ne"
	keyExists       = "$exists"
	keyRegex        = "$regex"
	keySet          = "$set"
)

const (
	defaultConnectTimeoutSeconds = 3     // connect timeout seconds
	defaultWriteTimeoutSeconds   = 60    // write timeout seconds
	defaultReadTimeoutSeconds    = 60    // read timeout seconds
	defaultPrimaryKeyName        = "_id" // database primary key name
)
