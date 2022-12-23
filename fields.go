package flog

// Field
type Field struct {
	Key string
	Val any
}

// NewField
func NewField(k string, v any) Field {
	return Field{
		Key: k,
		Val: v,
	}
}
