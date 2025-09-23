package raw

const (
	RAW_OBJECT_TYPE           = "object"
	RAW_ARRAY_TYPE            = "array"
	RAW_BOOLEAN_TYPE          = "boolean"
	RAW_NULL_TYPE             = "null"
	RAW_QUOTED_STRING_TYPE    = "quoted_string"
	RAW_UNQUOTED_STRING_TYPE  = "unquoted_string"
	RAW_MULTILINE_STRING_TYPE = "multiline_string"
	RAW_CONCAT_STRING_TYPE    = "concat_string"
	RAW_NUMBER_TYPE           = "number"
	RAW_SUBSTITUTION_TYPE     = "substitution"
	RAW_CONCAT_TYPE           = "concat"
	RAW_ADD_ASSIGN_TYPE       = "add_assign"
)

type Value interface {
	Type() string
	String() string
	isRawValue()
}
