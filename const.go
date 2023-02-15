package mgoc

const (
	KeyIn           = "$in"
	KeyEqual        = "$eq"
	KeyAnd          = "$and"
	KeyOr           = "$or"
	KeyGreaterThan  = "$gt"
	KeyGreaterEqual = "$gte"
	KeyLessThan     = "$lt"
	KeyLessEqual    = "$lte"
	KeyNotEqual     = "$ne"
	KeyExists       = "$exists"
	KeyRegex        = "$regex"
	KeySet          = "$set"
	KeyMatch        = "$match"
	KeyGroup        = "$group"
	KeyHaving       = "$having"
	KeyProject      = "$project"
	KeySort         = "$sort"
	KeyLimit        = "$limit"
	KeySum          = "$sum"
)

const (
	defaultConnectTimeoutSeconds = 3     // connect timeout seconds
	defaultWriteTimeoutSeconds   = 60    // write timeout seconds
	defaultReadTimeoutSeconds    = 60    // read timeout seconds
	defaultPrimaryKeyName        = "_id" // database primary key name
)
