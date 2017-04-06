package csv

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

type CSVEncoder struct {
	HeaderFields map[string][]string
	Rows         [][]byte
	RowCache     []string
}

func Marshal(v interface{}) ([]byte, error) {
	val := reflect.ValueOf(v)

	if val.Kind() == reflect.Slice {
		if val.Type().Elem().Kind() != reflect.Struct {
			return nil, fmt.Errorf("csv error: expected a struct or a list of struct\n")
		}
	} else if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("csv error: expected a struct or a list of struct\n")
	}

	encoder, err := NewCSVEncoder(val)
	if err != nil {
		return nil, err
	}

	b, err := encoder.Encode(val)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *CSVEncoder) String() string {
	s := "Header Fields:\n"
	for k, v := range c.HeaderFields {
		s += fmt.Sprintf("\t%q >> %s\n", k, strings.Join(v, ", "))
	}
	s += "Rows:\n"
	for _, v := range c.Rows {
		s += fmt.Sprintf("\t%s\n", v)
	}
	return s
}

func NewCSVEncoder(v reflect.Value) (*CSVEncoder, error) {
	exporter := &CSVEncoder{
		HeaderFields: map[string][]string{},
		Rows:         [][]byte{},
	}

	if err := exporter.EncodeHeader(v); err != nil {
		return nil, err
	}
	return exporter, nil
}

func (c *CSVEncoder) EncodeHeader(v reflect.Value) error {
	if v.Kind() == reflect.Slice {
		v = reflect.Zero(v.Type().Elem())
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("csv error: expected a struct or a list of struct\n")
	}
	c.RowCache = []string{}
	if err := c.encodeHeader(v, ""); err != nil {
		return err
	}

	c.Rows = append(c.Rows, []byte(strings.Join(c.RowCache, ",")))
	return nil
}

func (c *CSVEncoder) encodeHeader(strctVal reflect.Value, start string) error {

	strctTyp := strctVal.Type()

	for fieldNum := 0; fieldNum < strctVal.NumField(); fieldNum++ {
		tag := strctTyp.Field(fieldNum).Tag.Get("csv")
		if tag == "-" {
			continue
		}

		fld := strctVal.Field(fieldNum)
		name := strctTyp.Field(fieldNum).Name

		if tag == "" {
			tag = name
		}
		switch fld.Kind() {
		case reflect.Struct:
			c.HeaderFields[start] = append(c.HeaderFields[start], name)
			if err := c.encodeHeader(reflect.Indirect(fld), start+tag+"."); err != nil {
				return err
			}
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
			c.HeaderFields[start] = append(c.HeaderFields[start], name)
			c.RowCache = append(c.RowCache, start+tag)
		}
	}
	return nil
}

func (c *CSVEncoder) Encode(v reflect.Value) ([]byte, error) {
	if v.Kind() == reflect.Struct {
		if err := c.EncodeRow(v, ""); err != nil {
			return nil, err
		}
		return bytes.Join(c.Rows, []byte("\n")), nil
	}

	for i := 0; i < v.Len(); i++ {

		strctVal := v.Index(i)
		c.RowCache = []string{}
		if err := c.EncodeRow(strctVal, ""); err != nil {
			return nil, err
		}
		c.Rows = append(c.Rows, []byte(strings.Join(c.RowCache, ",")))
	}

	return bytes.Join(c.Rows, []byte("\n")), nil
}

func (c *CSVEncoder) EncodeRow(strctVal reflect.Value, start string) error {
	if strctVal.Kind() != reflect.Struct {
		return fmt.Errorf("csv error: expected a struct or a list of struct\n")
	}
	for _, field := range c.HeaderFields[start] {
		fld := strctVal.FieldByName(field)
		fldTyp, ok := strctVal.Type().FieldByName(field)
		tag := fldTyp.Tag.Get("csv")
		if !ok {
			return fmt.Errorf("csv error: failed to find struct field\n")
		}
		name := fldTyp.Name
		if tag == "" {
			tag = name
		}

		switch fld.Kind() {
		case reflect.Struct:
			if err := c.EncodeRow(reflect.Indirect(fld), start+tag+"."); err != nil {
				return err
			}
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
			c.RowCache = append(c.RowCache, fmt.Sprintf("%v", fld.Interface()))
		}
	}
	return nil
}
