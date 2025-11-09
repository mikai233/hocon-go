package parser

import (
	"fmt"
	"hocon-go/raw"
	"testing"
)

func TestParseArray(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{`[1,2,3]`, []string{"number:1", "number:2", "number:3"}},
		{`[true,false,null]`, []string{"boolean:true", "boolean:false", "null:null"}},
		{`[1,2 ,3,
]`, []string{"number:1", "number:2", "number:3"}},
		{`[1
2 ,3, 
]`, []string{"number:1", "number:2", "number:3"}},
		{`[1
2.0001 ,3, 
]`, []string{"number:1", "float:2.0001", "number:3"}},
		{`[1
2.0001f ,3, 
]`, []string{"number:1", "string:2.0001f", "number:3"}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			p := newTestParser(tc.input)
			arr, err := p.parseArray(true)
			if err != nil {
				t.Fatalf("parseArray error: %v", err)
			}
			if got := snapshotValues(arr.Values); !equalSlices(got, tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func snapshotValues(values []raw.Value) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = snapshotValue(v)
	}
	return out
}

func snapshotValue(v raw.Value) string {
	switch val := v.(type) {
	case *raw.PosInt:
		return fmt.Sprintf("number:%d", val.Val)
	case *raw.NegInt:
		return fmt.Sprintf("number:%d", val.Val)
	case *raw.Float:
		return fmt.Sprintf("float:%g", val.Val)
	case *raw.Boolean:
		return fmt.Sprintf("boolean:%t", val.Val)
	case *raw.Null:
		return "null:null"
	case *raw.UnquotedString:
		return fmt.Sprintf("string:%s", val.Value)
	case *raw.QuotedString:
		return fmt.Sprintf("string:%s", val.Value)
	case *raw.MultilineString:
		return fmt.Sprintf("string:%s", val.Value)
	default:
		return fmt.Sprintf("%T:%s", v, v.String())
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
