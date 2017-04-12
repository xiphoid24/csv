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
	count       int
	added       bool
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
		count:       0,
	}

	if err := exporter.EncodeHeader(v); err != nil {
		return nil, err
	}
	return exporter, nil
}

func (c *CSVRelationEncoder) EncodeHeader(v reflect.Value) error {
	if v.Kind() == reflect.Slice {
		v = reflect.Zero(v.Type().Elem())
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("csv error: expected a struct or a list of struct\n")
	}
	c.RowCache = []string{}
	/*i := 0
	for k, _ := range c.RelationMap {
		if header, ok := getColumnName(k, c.RelationMap); ok {
			c.RowCache = append(c.RowCache, header)
			c.HeaderMap[k] = i
			i++
		}
	}*/
	if err := c.encodeHeader(v, ""); err != nil {
		return err
	}

	if len(c.RowCache) < 1 {
		return fmt.Errorf("csv/form: Empty relationship map")
	}
	c.Rows = append(c.Rows, []byte(strings.Join(c.RowCache, ",")))
	return nil
}

func (c *CSVRelationEncoder) encodeHeader(strctVal reflect.Value, start string) error {
	if strctVal.Kind() != reflect.Struct {
		return fmt.Errorf("csv error: expected a struct or a list of struct\n")
	}

	if start != "" {
		start += " "
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
		if formtag == "-" {
			formtag = ""
		}
		switch fld.Kind() {
		case reflect.Struct:
			if err := c.encodeHeader(reflect.Indirect(fld), start+formtag); err != nil {
				return err
			}
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
			if header, ok := getColumnName(start+formtag, c.RelationMap); ok {
				c.RowCache = append(c.RowCache, header)
				c.HeaderMap[start+formtag] = c.count
				c.count++
			}
		}
	}

	return nil

}

func (c *CSVRelationEncoder) Encode(v reflect.Value) ([]byte, error) {
	if v.Kind() == reflect.Struct {
		if err := c.EncodeRelationRow(v, ""); err != nil {
			return nil, err
		}
		return bytes.Join(c.Rows, []byte("\n")), nil
	}

	for i := 0; i < v.Len(); i++ {
		c.added = false
		strctVal := v.Index(i)
		c.RowCache = make([]string, len(c.HeaderMap))
		if err := c.EncodeRelationRow(strctVal, ""); err != nil {
			return nil, err
		}
		if c.added {
			c.Rows = append(c.Rows, []byte(strings.Join(c.RowCache, ",")))
		}
	}
	return bytes.Join(c.Rows, []byte("\n")), nil
}

func (c *CSVRelationEncoder) EncodeRelationRow(strctVal reflect.Value, start string) error {
	if strctVal.Kind() != reflect.Struct {
		return fmt.Errorf("csv error: expected a struct or a list of struct\n")
	}

	if start != "" {
		start += " "
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
		if formtag == "-" {
			formtag = ""
		}
		switch fld.Kind() {
		case reflect.Struct:
			if err := c.EncodeRelationRow(reflect.Indirect(fld), start+formtag); err != nil {
				return err
			}
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
			if i, ok := c.HeaderMap[start+formtag]; ok {
				if s := fmt.Sprintf("%v", fld.Interface()); s != "" {
					c.RowCache[i] = s
					c.added = true
				}
			}
		}
	}
	return nil
}
