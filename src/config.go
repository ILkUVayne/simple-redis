package src

import (
	"bufio"
	"github.com/ILkUVayne/utlis-go/v2/str"
	"github.com/ILkUVayne/utlis-go/v2/ulog"
	"os"
	"reflect"
	"strings"
)

// rdb save param
type saveParam struct {
	seconds int
	changes int
}

// rdb append saveParams conf
func appendServerSaveParams(val string) {
	firstIdx := strings.IndexAny(val, " ")
	if val == "\"\"" {
		return
	}
	if firstIdx <= 0 || firstIdx >= len(val)-1 {
		ulog.Error("Invalid save parameters: ", val)
	}

	seconds, changes := val[0:firstIdx], val[firstIdx+1:]

	var intVal int64
	sp := new(saveParam)
	if str.String2Int64(&seconds, &intVal) != nil {
		ulog.Error("Invalid save seconds parameters: ", seconds)
	}
	sp.seconds = int(intVal)
	if str.String2Int64(&changes, &intVal) != nil {
		ulog.Error("Invalid save changes parameters: ", changes)
	}
	sp.changes = int(intVal)

	if config.saveParams == nil {
		config.saveParams = make([]*saveParam, 1)
		config.saveParams[0] = sp
		return
	}
	config.saveParams = append(config.saveParams, sp)
}

// 配置信息映射结构
type configVal struct {
	Bind           string `cfg:"bind"`
	Port           int    `cfg:"port"`
	AppendOnly     bool   `cfg:"appendOnly"`
	RehashNullStep int64  `cfg:"rehashNullStep"`

	// complex conf

	saveParams []*saveParam // rdb save params
}

// 配置信息
var config *configVal

// SetupConf 加载配置信息
func SetupConf(confName string) {
	f, err := os.Open(confName)
	if err != nil {
		ulog.Error(err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			ulog.Error(err)
		}
	}(f)

	parse(f)
}

// 解析每行配置，并更新到config中
func parse(f *os.File) {
	config = new(configVal)
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
			key := strings.ToLower(line[0:firstIdx])
			val := strings.Trim(line[firstIdx+1:], " ")
			if complexConfHandle(key, val) {
				continue
			}
			rawMap[key] = val
		}
	}

	confType := reflect.TypeOf(config)
	confValue := reflect.ValueOf(config)

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
				var intVal int64
				if str.String2Int64(&value, &intVal) == nil {
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
}
