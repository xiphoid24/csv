package form

import (
	"fmt"
	"reflect"
)

func GetOptions(v interface{}) ([]string, error) {
	var options []string

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("csv error: expected a struct or a list of struct\n")
	}
	getOptions(val, &options, "")
	return options, nil
}

func getOptions(strctVal reflect.Value, s *[]string, start string) error {

	strctTyp := strctVal.Type()

	if start != "" {
		start += " "
	}

	for fieldNum := 0; fieldNum < strctVal.NumField(); fieldNum++ {
		tag, ok := strctTyp.Field(fieldNum).Tag.Lookup("csvform")
		if !ok {
			continue
		}

		fld := strctVal.Field(fieldNum)
		name := strctTyp.Field(fieldNum).Name

		if tag == "" {
			tag = name
		}

		if tag == "-" {
			tag = ""
		}

		switch fld.Kind() {
		case reflect.Struct:
			if err := getOptions(reflect.Indirect(fld), s, start+tag); err != nil {
				return err
			}
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
			(*s) = append((*s), start+tag)
		}
	}

	return nil
}

func getColumnName(key string, m map[string][]string) (string, bool) {
	if m == nil {
		return "", false
	}
	ss, ok := m[key]
	if !ok || len(ss) < 1 {
		return "", false
	}
	return ss[0], ss[0] != ""
}
