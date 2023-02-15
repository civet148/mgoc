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
	keyMatch        = "$match"
	keyGroup        = "$group"
	keyHaving       = "$having"
	keyProject      = "$project"
	keySort         = "$sort"
	keyLimit        = "$limit"
	keySum          = "$sum"
)

const (
	defaultConnectTimeoutSeconds = 3     // connect timeout seconds
	defaultWriteTimeoutSeconds   = 60    // write timeout seconds
	defaultReadTimeoutSeconds    = 60    // read timeout seconds
	defaultPrimaryKeyName        = "_id" // database primary key name
)
