package src

import (
	"bufio"
	"log"
	"os"
	"reflect"
	"simple-redis/utils"
	"strings"
)

type complexConfFunc func(val string)

type saveParam struct {
	seconds int
	changes int
}

// rdb save conf
func appendServerSaveParams(val string) {
	firstIdx := strings.IndexAny(val, " ")
	if val == "\"\"" {
		return
	}
	if firstIdx <= 0 || firstIdx >= len(val)-1 {
		utils.Error("Invalid save parameters: ", val)
	}

	seconds, changes := val[0:firstIdx], val[firstIdx+1:]

	var intVal int64
	sp := new(saveParam)
	if utils.String2Int64(&seconds, &intVal) == REDIS_ERR {
		utils.Error("Invalid save seconds parameters: ", seconds)
	}
	sp.seconds = int(intVal)
	if utils.String2Int64(&changes, &intVal) == REDIS_ERR {
		utils.Error("Invalid save changes parameters: ", changes)
	}
	sp.changes = int(intVal)

	if config.saveParams == nil {
		config.saveParams = make([]*saveParam, 1)
		config.saveParams[0] = sp
		return
	}
	config.saveParams = append(config.saveParams, sp)
}

var complexConfFuncMaps = map[string]complexConfFunc{
	"save": appendServerSaveParams,
}

type configVal struct {
	Bind           string `cfg:"bind"`
	Port           int    `cfg:"port"`
	AppendOnly     bool   `cfg:"appendOnly"`
	RehashNullStep int64  `cfg:"rehashNullStep"`
	// complex conf
	saveParams []*saveParam
}

var config *configVal

func newConfig() {
	config = new(configVal)
}

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

	//config = parse(f)
	parse(f)
}

// return true complexConf,or false simpleConf
func complexConfHandle(key, val string) bool {
	fn, ok := complexConfFuncMaps[strings.ToLower(key)]
	// simpleConf return false
	if !ok {
		return false
	}
	fn(val)
	return true
}

func parse(f *os.File) {
	//conf := &configVal{}
	newConfig()
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
				if utils.String2Int64(&value, &intVal) == REDIS_OK {
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
	//return conf
}
