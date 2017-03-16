package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"strings"
)

func main() {

	users := []User{
		{
			Name:     "name-1",
			Age:      1,
			Active:   true,
			Email:    "1@1.com",
			Password: "1",
			Address: Address{
				Street:  "1 - street",
				City:    "1 - city",
				County:  "1 - county",
				State:   "1 - state",
				Zip:     "11111",
				Country: "1 - country",
			},
		},
	}

	exporter, err := NewCSVExporter(users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", exporter)
	err = exporter.Export("users.csv")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", exporter)
}

type User struct {
	Name     string  `csv:"name"`
	Age      int     `csv:"age"`
	Active   bool    `csv:"active"`
	Email    string  `csv:"email"`
	Password string  `csv:"-"`
	Address  Address `csv:"address"`
}

type Address struct {
	Street  string `csv:"street"`
	City    string `csv:"city"`
	County  string `csv:"-"`
	State   string `csv:"state"`
	Zip     string `csv:"zip"`
	Country string `csv:"-"`
}

type CSVExporter struct {
	HeaderFields map[string][]string
	Rows         [][]string
	V            interface{}
}

func (c *CSVExporter) String() string {
	s := "Header Fields:\n"
	for k, v := range c.HeaderFields {
		s += fmt.Sprintf("\t%q >> %s\n", k, strings.Join(v, ", "))
	}
	s += "Rows:\n"
	for _, v := range c.Rows {
		s += fmt.Sprintf("\t%s\n", strings.Join(v, ", "))
	}
	return s
}

func NewCSVExporter(v interface{}) (*CSVExporter, error) {
	exporter := &CSVExporter{
		HeaderFields: map[string][]string{},
		Rows:         [][]string{},
		V:            v,
	}

	err := exporter.SetHeader()
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func (c *CSVExporter) SetHeader() error {
	slTyp := reflect.TypeOf(c.V)
	if slTyp.Kind() != reflect.Slice {
		return fmt.Errorf("csv error: expected a list of struct\n")
	}

	strctTyp := slTyp.Elem()
	if strctTyp.Kind() != reflect.Struct {
		return fmt.Errorf("csv error: expected a list of struct\n")
	}
	strctVal := reflect.Zero(strctTyp)

	h, err := c.setHeader([]string{}, strctVal, "")
	if err != nil {
		return err
	}
	c.Rows = append(c.Rows, h)
	return nil
}

func (c *CSVExporter) setHeader(header []string, strctVal reflect.Value, start string) ([]string, error) {

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
			st := reflect.Indirect(fld)
			h, err := c.setHeader(header, st, start+tag+":")
			if err != nil {
				return nil, err
			}
			header = h
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
			c.HeaderFields[start] = append(c.HeaderFields[start], name)
			header = append(header, strings.Replace(start+tag, ":", " ", -1))
		}
	}
	return header, nil
}

func (c *CSVExporter) Export(path string) error {
	slVal := reflect.ValueOf(c.V)
	for i := 0; i < slVal.Len(); i++ {
		strctVal := slVal.Index(i)
		r, err := c.setRow([]string{}, strctVal, "")
		if err != nil {
			return err
		}
		c.Rows = append(c.Rows, r)
	}
	var out []string
	for _, r := range c.Rows {
		out = append(out, strings.Join(r, ","))
	}

	err := ioutil.WriteFile(path, []byte(strings.Join(out, "\n")), 0666)
	if err != nil {
		return err
	}
	return nil
}

func (c *CSVExporter) setRow(row []string, strctVal reflect.Value, start string) ([]string, error) {
	if strctVal.Kind() != reflect.Struct {
		return nil, fmt.Errorf("csv error: expected a list of struct\n")
	}
	for _, field := range c.HeaderFields[start] {
		fld := strctVal.FieldByName(field)
		fldTyp, ok := strctVal.Type().FieldByName(field)
		tag := fldTyp.Tag.Get("csv")
		if !ok {
			return nil, fmt.Errorf("csv error: failed to find struct field\n")
		}
		name := fldTyp.Name
		if tag == "" {
			tag = name
		}

		switch fld.Kind() {
		case reflect.Struct:
			fmt.Println("hit struct")
			st := reflect.Indirect(fld)
			r, err := c.setRow(row, st, start+tag+":")
			if err != nil {
				return nil, err
			}
			row = r
		case reflect.String:
			row = append(row, fld.String())

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			row = append(row, fmt.Sprintf("%v", fld.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			row = append(row, fmt.Sprintf("%v", fld.Uint()))
		case reflect.Float32, reflect.Float64:
			row = append(row, fmt.Sprintf("%v", fld.Float()))
		case reflect.Bool:
			row = append(row, fmt.Sprintf("%v", fld.Bool()))
		}
	}
	return row, nil
}
