package rsv

import (
	"bytes"
	"testing"
)

type TestCase struct {
	s Record
	b []byte
}

var TestCases = []TestCase{
	{
		s: Record{
			[]byte("Hello"),
			[]byte("🌎"),
		},
		b: []byte{72, 101, 108, 108, 111, ValueTerminator, 240, 159, 140, 142, ValueTerminator, RowTerminator},
	},
	{
		s: Record{},
		b: []byte{RowTerminator},
	},
	{
		s: Record{nil, []byte{}},
		b: []byte{NullValue, ValueTerminator, ValueTerminator, RowTerminator},
	},
}

func TestRSV(t *testing.T) {
	for i, testcase := range TestCases {
		buffer := new(bytes.Buffer)
		writer := NewWriter(buffer)
		if err := writer.Write(testcase.s); err != nil {
			t.Fatalf("error writing: %v\n", err)
		} else if err := writer.Flush(); err != nil {
			t.Fatalf("error flushing: %v\n", err)
		} else if want, got := testcase.b, buffer.Bytes(); !bytes.Equal(want, got) {
			t.Fatalf("error: wanted %v; got %v\n", want, got)
		}
		reader := NewReader(buffer)
		records, err := reader.ReadAll()
		if err != nil {
			t.Fatalf("error from readall: %v\n", err)
		} else if want, got := len(testcase.s), len(records[0]); want != got {
			t.Fatalf("error for test case %d: wanted %d values; got %d\n", i, want, got)
		}
	}
}
