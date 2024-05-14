package src

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var respType = map[byte]int{
	'+': SIMPLE_STR,
	'-': SIMPLE_ERROR,
	':': INTEGERS,
	'$': BULK_STR,
	'*': ARRAYS,
	'_': NULLS,
	'#': BOOLEANS,
	',': DOUBLE,
	'(': BIG_NUMBERS,
	'!': BULK_ERR,
	'=': VERBATIM_STR,
	'%': MAPS,
	'~': SETS,
	'>': PUSHES,
}

// return -1 if invalid type
func getRespType(t byte) int {
	typ, ok := respType[t]
	if !ok {
		return -1
	}
	return typ
}

func getQueryLine(buf []byte, queryLen int) (int, error) {
	idx := strings.Index(string(buf), "\r\n")
	if idx < 0 && queryLen > SREDIS_MAX_INLINE {
		return idx, errors.New("inline cmd is too long")
	}
	return idx, nil
}

// ================================ Parse resp data =================================

func respParseSimpleStr(buf []byte, length int) (string, error) {
	idx, err := getQueryLine(buf, length)
	if err != nil || idx < 0 {
		return "", err
	}
	str := string(buf[1:idx])
	return str, nil
}

func respParseBulk(buf []byte, length int) (string, error) {
	idx, err := getQueryLine(buf, length)
	if err != nil || idx < 0 {
		return "", err
	}
	sLen, err := strconv.Atoi(string(buf[1:idx]))
	if err != nil {
		return "", err
	}
	if sLen == -1 {
		return "(nil)", nil
	}
	if sLen == 0 && length == 4 {
		return "empty string", nil
	}
	buf = buf[idx+2:]
	idx, err = getQueryLine(buf, length-(idx+2))
	if err != nil || idx < 0 {
		return "", err
	}
	str := string(buf[:idx])
	return str, nil
}

func respParseIntegers(buf []byte, length int) (string, error) {
	idx, err := getQueryLine(buf, length)
	if err != nil || idx < 0 {
		return "", err
	}
	str := string(buf[1:idx])
	return str, nil
}

// *2\r\n$2\r\nxx\r\n$3\r\nccc\r\n
func respParseArrays(buf []byte, length int) (string, error) {
	idx, err := getQueryLine(buf, length)
	if err != nil || idx < 0 {
		return "", err
	}
	aLen, err := strconv.Atoi(string(buf[1:idx]))
	if err != nil {
		return "", err
	}
	if aLen == -1 {
		return "(nil)", nil
	}
	if aLen == 0 && length == 4 {
		return "(empty array)", nil
	}
	buf = buf[idx+2:]
	str := ""
	for i := 0; i < aLen; i++ {
		idx, err = getQueryLine(buf, length)
		if err != nil || idx < 0 {
			return "", err
		}
		bulkLen, err := strconv.Atoi(string(buf[1:idx]))
		if err != nil {
			return "", err
		}
		if bulkLen == -1 {
			str += strconv.Itoa(i+1) + ") (nil)\r\n"
			buf = buf[idx+2:]
			continue
		}
		if bulkLen == 0 {
			str += strconv.Itoa(i+1) + ") \r\n"
			buf = buf[idx+2:]
			continue
		}
		buf = buf[idx+2:]
		idx, err = getQueryLine(buf, length)
		if err != nil || idx < 0 {
			return "", err
		}
		str += strconv.Itoa(i+1) + ") \"" + string(buf[:idx]) + "\"\r\n"
		buf = buf[idx+2:]
	}
	return str[:len(str)-2], nil
}

// ================================ format response string =================================

func bulkStrFormat(s string) string {
	if s != NIL_STR {
		return fmt.Sprintf("\"%s\"", s)
	}
	return ""
}

func simpleErrStrFormat(s string) string {
	return fmt.Sprintf("(error) %s", s)
}

func intStrFormat(s string) string {
	return fmt.Sprintf("(integer) %s", s)
}
