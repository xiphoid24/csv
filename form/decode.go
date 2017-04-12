package form

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "csv: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "csv: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	if e.Type.Elem().Kind() != reflect.Slice {
		return "csv: Unmarshal(non-slice " + e.Type.String() + ")"
	}
	if e.Type.Elem().Elem().Kind() != reflect.Struct {
		return "csv: Unmarshal(non-struct slice " + e.Type.String() + ")"
	}
	return "csv: Unmarshal(nil " + e.Type.String() + ")"
}

var EMPTYROW = errors.New("CSV EMPTY ROW")

type CSVRelationDecoder struct {
	Rdr         *csv.Reader
	Rows        [][]string
	HeaderMap   map[string]int
	RelationMap map[string][]string
	RowFilled   bool
}

func NewCSVRelationDecoder(b []byte, rel map[string][]string) (*CSVRelationDecoder, error) {
	if rel == nil {
		return nil, fmt.Errorf("csv: nil relationship map")
	}

	c := new(CSVRelationDecoder)
	c.Rdr = csv.NewReader(bytes.NewBuffer(b))
	for {
		row, err := c.Rdr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		c.Rows = append(c.Rows, row)
	}
	if len(c.Rows) < 1 {
		return nil, fmt.Errorf("csv: error reading rows")
	}

	c.HeaderMap = make(map[string]int)

	for i, h := range c.Rows[0] {
		c.HeaderMap[h] = i
	}

	c.RelationMap = rel
	return c, nil
}

func Unmarshal(b []byte, v interface{}, rel map[string][]string) error {

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	if rv.Elem().Kind() != reflect.Slice {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	if rv.Elem().Type().Elem().Kind() != reflect.Struct {
		fmt.Println(reflect.TypeOf(v))
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	decoder, err := NewCSVRelationDecoder(b, rel)
	if err != nil {
		return err
	}

	if err := decoder.Decode(v); err != nil {
		return err
	}
	return nil
}

func UnmarshalRow(row int, b []byte, v interface{}, rel map[string][]string) error {

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	if rv.Elem().Kind() != reflect.Struct {
		fmt.Println(reflect.TypeOf(v))
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	decoder, err := NewCSVRelationDecoder(b, rel)
	if err != nil {
		return err
	}

	if row > len(decoder.Rows) || row < 1 {
		return fmt.Errorf("csv: Invalid row")
	}

	if err := decoder.DecodeRelationRow(row, rv.Elem(), ""); err != nil {
		return err
	}

	return nil
}

func (c *CSVRelationDecoder) GetFieldInRow(r, f int) string {
	if len(c.Rows) < r {
		return ""
	}
	if len(c.Rows[r]) < f {
		return ""
	}
	return c.Rows[r][f]
}

func (c *CSVRelationDecoder) Decode(ptr interface{}) error {
	if len(c.Rows) < 2 {
		return errors.New("csv: not enough rows in csv file")
	}

	// derefrencing pointer
	val := reflect.Indirect(reflect.ValueOf(ptr))

	// get type of single element
	strctTyp := val.Type().Elem()

	for rowNum := 1; rowNum < len(c.Rows); rowNum++ {
		c.RowFilled = false
		strct := reflect.Indirect(reflect.New(strctTyp))
		if err := c.DecodeRelationRow(rowNum, strct, ""); err != nil {
			return err
		}

		if c.RowFilled {
			val.Set(reflect.Append(val, strct))
		}
	}
	return nil
}

func (c *CSVRelationDecoder) DecodeRelationRow(rowNum int, strct reflect.Value, start string) error {
	strctTyp := strct.Type()

	if start != "" {
		start += " "
	}

	for fieldNum := 0; fieldNum < strct.NumField(); fieldNum++ {

		fld := strct.Field(fieldNum)
		name := strctTyp.Field(fieldNum).Name

		formtag, ok := strctTyp.Field(fieldNum).Tag.Lookup("csvform")
		if !ok {
			continue
		}

		if formtag == "" {
			formtag = name
		}

		if formtag == "-" {
			formtag = ""
		}

		if fld.Kind() == reflect.Struct {
			st := reflect.Indirect(fld)
			if err := c.DecodeRelationRow(rowNum, st, start+formtag); err != nil {
				return err
			}
			fld.Set(st)
			continue
		}
		columnName, ok := getColumnName(start+formtag, c.RelationMap)
		if !ok {
			continue
		}
		columnNum, ok := c.HeaderMap[columnName]
		if !ok {
			continue
		}
		csvVal := c.GetFieldInRow(rowNum, columnNum)
		if csvVal == "" {
			continue
		}
		c.RowFilled = true
		switch fld.Kind() {
		case reflect.String:
			fld.SetString(csvVal)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			in, err := strconv.ParseInt(csvVal, 10, 64)
			if err != nil {
				return fmt.Errorf("csv: %s +  Must be a a number", name)
			}
			fld.SetInt(in)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u, err := strconv.ParseUint(csvVal, 10, 64)
			if err != nil {
				return fmt.Errorf("csv: %s +  Must be a a number", name)
			}
			fld.SetUint(u)
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(csvVal, 64)
			if err != nil {
				return fmt.Errorf("csv: %s +  Must be a a number", name)
			}
			fld.SetFloat(f)
		case reflect.Bool:
			b, err := strconv.ParseBool(csvVal)
			if err != nil {
				return fmt.Errorf("csv: %s +  Must be either true or false", name)
			}
			fld.SetBool(b)
		}
	}

	return nil
}
