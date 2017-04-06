package form

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

type CSVRelationEncoder struct {
	HeaderMap   map[string]int
	RelationMap map[string][]string
	Rows        [][]byte
	RowCache    []string
}

func Marshal(v interface{}, rel map[string][]string) ([]byte, error) {
	val := reflect.ValueOf(v)

	if val.Kind() == reflect.Slice {
		if val.Type().Elem().Kind() != reflect.Struct {
			return nil, fmt.Errorf("csv error: expected a struct or a list of struct\n")
		}
	} else if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("csv error: expected a struct or a list of struct\n")
	}

	encoder, err := NewCSVRelationEncoder(val, rel)
	if err != nil {
		return nil, err
	}

	b, err := encoder.Encode(val)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *CSVRelationEncoder) String() string {
	s := "Header Fields:\n"
	/*for k, v := range c.HeaderMap {
		s += fmt.Sprintf("\t%q >> %s\n", k, strings.Join(v, ", "))
	}*/
	s += "Rows:\n"
	for _, v := range c.Rows {
		s += fmt.Sprintf("\t%s\n", v)
	}
	return s
}

func NewCSVRelationEncoder(v reflect.Value, rel map[string][]string) (*CSVRelationEncoder, error) {
	if rel == nil {
		return nil, fmt.Errorf("csv/form: nil relationship map")
	}
	exporter := &CSVRelationEncoder{
		HeaderMap:   map[string]int{},
		RelationMap: rel,
		Rows:        [][]byte{},
	}

	if err := exporter.EncodeHeader(); err != nil {
		return nil, err
	}
	return exporter, nil
}

func (c *CSVRelationEncoder) EncodeHeader() error {

	c.RowCache = []string{}
	i := 0
	for k, _ := range c.RelationMap {
		if header, ok := getColumnName(k, c.RelationMap); ok {
			c.RowCache = append(c.RowCache, header)
			c.HeaderMap[k] = i
			i++
		}
	}
	if len(c.RowCache) < 1 {
		return fmt.Errorf("csv/form: Empty relationship map")
	}
	c.Rows = append(c.Rows, []byte(strings.Join(c.RowCache, ",")))

	return nil
}

func (c *CSVRelationEncoder) Encode(v reflect.Value) ([]byte, error) {
	if v.Kind() == reflect.Struct {
		if err := c.EncodeRelationRow(v); err != nil {
			return nil, err
		}
		return bytes.Join(c.Rows, []byte("\n")), nil
	}

	for i := 0; i < v.Len(); i++ {

		strctVal := v.Index(i)
		c.RowCache = make([]string, len(c.HeaderMap))
		if err := c.EncodeRelationRow(strctVal); err != nil {
			return nil, err
		}
		fmt.Printf("\n\n%v\n\n", len(c.HeaderMap))
		c.Rows = append(c.Rows, []byte(strings.Join(c.RowCache, ",")))
	}

	return bytes.Join(c.Rows, []byte("\n")), nil
}

func (c *CSVRelationEncoder) EncodeRelationRow(strctVal reflect.Value) error {
	if strctVal.Kind() != reflect.Struct {
		return fmt.Errorf("csv error: expected a struct or a list of struct\n")
	}

	strctTyp := strctVal.Type()

	for fieldNum := 0; fieldNum < strctVal.NumField(); fieldNum++ {
		fld := strctVal.Field(fieldNum)
		name := strctTyp.Field(fieldNum).Name

		formtag, ok := strctTyp.Field(fieldNum).Tag.Lookup("csvform")
		if !ok {
			continue
		}
		if formtag == "" {
			formtag = name
		}

		switch fld.Kind() {
		case reflect.Struct:
			if err := c.EncodeRelationRow(reflect.Indirect(fld)); err != nil {
				return err
			}
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
			// c.RowCache = append(c.RowCache, fmt.Sprintf("%v", fld.Interface()))
			if i, ok := c.HeaderMap[formtag]; ok {
				c.RowCache[i] = fmt.Sprintf("%v", fld.Interface())
			}
		}
	}
	return nil
}
