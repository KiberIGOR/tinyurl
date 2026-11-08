package resettest

import "testing"

func TestResetableStructReset(t *testing.T) {
	str := "x"
	s := &ResetableStruct{
		I:    10,
		Str:  "hello",
		StrP: &str,
		S:    []int{1, 2},
		M:    map[string]string{"a": "b"},
		Child: &ResetableStruct{I: 99},
	}
	s.Reset()

	if s.I != 0 || s.Str != "" {
		t.Fatal("primitives not reset")
	}
	if str != "" {
		t.Fatal("strP target not reset")
	}
	if len(s.S) != 0 {
		t.Fatal("slice should be [:0]")
	}
	if len(s.M) != 0 {
		t.Fatal("map should be cleared")
	}
	if s.Child.I != 0 {
		t.Fatal("child should be reset recursively")
	}
}
