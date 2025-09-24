package raw

const (
	ObjectType          = "object"
	ArrayType           = "array"
	BooleanType         = "boolean"
	NullType            = "null"
	QuotedStringType    = "quoted_string"
	UnquotedStringType  = "unquoted_string"
	MultilineStringType = "multiline_string"
	ConcatStringType    = "concat_string"
	NumberType          = "number"
	SubstitutionType    = "substitution"
	ConcatType          = "concat"
	AddAssignType       = "add_assign"
)

type Value interface {
	Type() string
	String() string
	isRawValue()
}
