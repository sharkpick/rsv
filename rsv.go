package rsv

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

const (
	RowTerminator   byte = 253
	NullValue       byte = 254
	ValueTerminator byte = 255
)

type Value []byte

func (v Value) String() string {
	switch v {
	case nil:
		return "NULL"
	default:
		return string(v)
	}
}

type Record []Value

func (r Record) String() string {
	buffer := strings.Builder{}
	for i, value := range r {
		buffer.WriteString(value.String())
		if i < len(r) {
			buffer.WriteByte(' ')
		}
	}
	return buffer.String()
}

type Writer struct {
	w *bufio.Writer
	binary.ByteOrder
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: bufio.NewWriter(w),
	}
}

func (w *Writer) Flush() error { return w.w.Flush() }

func (w *Writer) Write(record Record) error {
	for _, field := range record {
		if field == nil {
			if err := binary.Write(w.w, w.ByteOrder, NullValue); err != nil {
				return err
			}
		} else {
			for _, c := range field {
				if err := binary.Write(w.w, w.ByteOrder, c); err != nil {
					return err
				}
			}
		}
		if err := binary.Write(w.w, w.ByteOrder, ValueTerminator); err != nil {
			return err
		}
	}
	return binary.Write(w.w, w.ByteOrder, RowTerminator)
}

func (w *Writer) WriteAll(records []Record) error {
	for _, record := range records {
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return w.Flush()
}

type Reader struct {
	r *bufio.Reader
	binary.ByteOrder
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r: bufio.NewReader(r),
	}
}

func (r *Reader) Read() (Record, error) {
	record := make(Record, 0)
	buffer := make(Value, 0)
loop:
	for {
		var c byte
		if err := binary.Read(r.r, r.ByteOrder, &c); err != nil {
			return nil, err
		}
		switch c {
		case NullValue:
			record = append(record, nil)
			// next value will be the value terminator
			if err := binary.Read(r.r, r.ByteOrder, &c); err != nil {
				return nil, err
			} else if c != ValueTerminator {
				return nil, errors.New("malformed value - expected value terminator after null value")
			}
		case RowTerminator:
			break loop
		case ValueTerminator:
			record = append(record, bytes.Clone(buffer))
			buffer = make(Value, 0, cap(buffer))
		default:
			buffer = append(buffer, c)
		}
	}
	return record, nil
}

func (r *Reader) ReadAll() ([]Record, error) {
	records := make([]Record, 0)
loop:
	for {
		record, err := r.Read()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, err
			}
			break loop
		} else {
			records = append(records, record)
		}
	}
	return records, nil
}
