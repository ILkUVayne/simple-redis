package src

import (
	"bufio"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type configVal struct {
	Bind           string `cfg:"bind"`
	Port           int    `cfg:"port"`
	AppendOnly     bool   `cfg:"appendOnly"`
	RehashNullStep int64  `cfg:"rehashNullStep"`
}

var config *configVal

func SetupConf(confName string) {
	f, err := os.Open(confName)
	if err != nil {
		log.Fatal(err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)

	config = parse(f)
}

func parse(f *os.File) *configVal {
	conf := &configVal{}
	scanner := bufio.NewScanner(f)
	rawMap := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		lLen := len(line) // len for line

		if lLen == 0 || line[0] == '#' {
			continue
		}
		firstIdx := strings.IndexAny(line, " ")
		if firstIdx > 0 && firstIdx < lLen-1 {
			rawMap[strings.ToLower(line[0:firstIdx])] = strings.Trim(line[firstIdx+1:], " ")
		}
	}

	confType := reflect.TypeOf(conf)
	confValue := reflect.ValueOf(conf)

	for i := 0; i < confType.Elem().NumField(); i++ {
		field := confType.Elem().Field(i)
		fieldVal := confValue.Elem().Field(i)
		key, ok := field.Tag.Lookup("cfg")

		if !ok {
			key = field.Name
		}
		value, ok := rawMap[strings.ToLower(key)]

		if ok {
			switch field.Type.Kind() {
			case reflect.String:
				fieldVal.SetString(value)
			case reflect.Int, reflect.Int64:
				intVal, err := strconv.ParseInt(value, 10, 64)
				if err == nil {
					fieldVal.SetInt(intVal)
				}
			case reflect.Bool:
				boolVal := "yes" == value
				fieldVal.SetBool(boolVal)
			case reflect.Slice:
				if field.Type.Elem().Kind() == reflect.String {
					fieldVal.Set(reflect.ValueOf(strings.Split(value, ",")))
				}
			}
		}
	}
	return conf
}
