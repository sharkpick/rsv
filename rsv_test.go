package rsv

import (
	"bufio"
	"bytes"
	"math/rand"
	"os"
	"slices"
	"strings"
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
		b: []byte{72, 101, 108, 108, 111,
			ValueTerminator,
			240, 159, 140, 142,
			ValueTerminator, RowTerminator},
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

var TheWords = func() []string {
	f, err := os.Open("/usr/share/dict/words")
	if err != nil {
		panic("error opening words file: " + err.Error())
	}
	defer f.Close()
	results := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); len(line) > 0 {
			results = append(results, line)
		}
	}
	slices.Sort(results)
	return results
}()

var TheWordRecords = func() []Record {
	results := make([]Record, 0, len(TheWords))
	for _, word := range TheWords {
		results = append(results, Record{[]byte(word)})
	}
	return results
}()

func RandomRecord() Record { return TheWordRecords[rand.Intn(len(TheWordRecords))] }

func BenchmarkWrite(b *testing.B) {
	writer := NewWriter(new(bytes.Buffer))
	for i := 0; i < b.N; i++ {
		writer.Write(RandomRecord())
	}
}
