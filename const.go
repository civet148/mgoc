package mgoc

const (
	KeyIn               = "$in"
	KeyEqual            = "$eq"
	KeyAnd              = "$and"
	KeyOr               = "$or"
	KeyGreaterThan      = "$gt"
	KeyGreaterThanEqual = "$gte"
	KeyLessThan         = "$lt"
	KeyLessThanEqual    = "$lte"
	KeyNotEqual         = "$ne"
	KeyExists           = "$exists"
	KeyRegex            = "$regex"
	KeySet              = "$set"
	KeyElemMatch        = "$elemMatch"
	KeyMatch            = "$match"
	KeyGroup            = "$group"
	KeyHaving           = "$having"
	KeyProject          = "$project"
	KeySort             = "$sort"
	KeyLimit            = "$limit"
	KeySum              = "$sum"
	KeyAll              = "$all"
	KeyNear             = "$near"
	KeyGeoNear          = "$geoNear"
	KeyGeoWithin        = "$geoWithin"
	KeyCenter           = "$center"
	KeyCenterSphere     = "$centerSphere"
	KeyGeoIntersects    = "$geoIntersects"
	KeyNearSphere       = "$nearSphere"
	KeyGeoMetry         = "$geometry"
	KeyMaxDistance      = "$maxDistance"
)

const (
	toBool     = "$toBool"
	toDecimal  = "$toDecimal"
	toDouble   = "$toDouble"
	toInt      = "$toInt"
	toLong     = "$toLong"
	toDate     = "$toDate"
	toString   = "$toString"
	toObjectId = "$toObjectId"
	toLower    = "$toLower"
	toUpper    = "$toUpper"
)

const (
	columnNameType          = "type"
	columnNameCoordinates   = "coordinates"
	columnNameNear          = "near"
	columnNameDistanceField = "distanceField"
	columnNameMaxDistance   = "maxDistance"
	columnNameIncludeLocs   = "includeLocs"
	columnNameSpherical     = "spherical"
)

const (
	defaultConnectTimeoutSeconds = 3     // connect timeout seconds
	defaultWriteTimeoutSeconds   = 60    // write timeout seconds
	defaultReadTimeoutSeconds    = 60    // read timeout seconds
	defaultPrimaryKeyName        = "_id" // database primary key name
)
