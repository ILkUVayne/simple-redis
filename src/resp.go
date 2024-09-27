package src

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// resp type mapping
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

// e.g. queryLen = "get name\r\n" return 8
//
// queryLen = "$3\r\nget\r\n$4\r\nname\r\n" return 2
func getQueryLine(buf []byte, queryLen int) (int, error) {
	idx := strings.Index(string(buf), "\r\n")
	if idx < 0 && queryLen > SREDIS_MAX_INLINE {
		return idx, errors.New("inline cmd is too long")
	}
	return idx, nil
}

//-----------------------------------------------------------------------------
// Parse resp data
//-----------------------------------------------------------------------------

// ================================ tool func =================================

// 检查数组是否为空
func respArrayCheckEmpty(aLen int) (string, bool) {
	if aLen == -1 {
		return "(nil)\r\n", true
	}
	if aLen == 0 {
		return "(empty array)\r\n", true
	}
	return "", false
}

// 获取数组长度
func respArrayLength(buf []byte, length int) ([]byte, int, error) {
	idx, err := getQueryLine(buf, length)
	if err != nil || idx < 0 {
		return buf, -1, err
	}
	aLen, err := strconv.Atoi(string(buf[1:idx]))
	if err != nil {
		return buf, -1, err
	}
	buf = buf[idx+2:]
	return buf, aLen, nil
}

// 递归解析数组数据
func respParseArrays1(buf []byte, length int, spaceNum int) ([]byte, string, error) {
	buf, aLen, err := respArrayLength(buf, length)
	if err != nil || aLen < 0 {
		return buf, "", err
	}
	if emptyStr, ok := respArrayCheckEmpty(aLen); ok {
		return buf, emptyStr + spaces(spaceNum), nil
	}

	str, subStr, idx := "", "", 0

	for i := 0; i < aLen; i++ {
		if buf[0] == '*' {
			str += strconv.Itoa(i+1) + ") "
			buf, subStr, err = respParseArrays1(buf, length, spaceNum+3)

			if err != nil || subStr == "" {
				return buf, "", err
			}
			str += subStr[:len(subStr)-(spaceNum+3)]
			continue
		}
		idx, err = getQueryLine(buf, length)
		if err != nil || idx < 0 {
			return buf, "", err
		}
		bulkLen, err := strconv.Atoi(string(buf[1:idx]))
		if err != nil {
			return buf, "", err
		}
		if bulkLen == -1 {
			str += strconv.Itoa(i+1) + ") (nil)\r\n" + spaces(spaceNum)
			buf = buf[idx+2:]
			continue
		}
		if bulkLen == 0 {
			str += strconv.Itoa(i+1) + ") \r\n" + spaces(spaceNum)
			buf = buf[idx+2:]
			continue
		}
		buf = buf[idx+2:]
		idx, err = getQueryLine(buf, length)
		if err != nil || idx < 0 {
			return buf, "", err
		}
		str += strconv.Itoa(i+1) + ") \"" + string(buf[:idx]) + "\"\r\n" + spaces(spaceNum)
		buf = buf[idx+2:]
	}
	return buf, str, nil
}

// ================================ respParseFunc Impl =================================

// e.g. "+OK\r\n" => "OK"
func respParseSimpleStr(buf []byte, length int) (string, error) {
	idx, err := getQueryLine(buf, length)
	if err != nil || idx < 0 {
		return "", err
	}
	str := string(buf[1:idx])
	return str, nil
}

// e.g. "$5\r\nhello\r\n" => "hello"
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

// e.g. ":3\r\n" => 3
func respParseIntegers(buf []byte, length int) (string, error) {
	idx, err := getQueryLine(buf, length)
	if err != nil || idx < 0 {
		return "", err
	}
	str := string(buf[1:idx])
	return str, nil
}

// e.g. *2\r\n$2\r\nxx\r\n$3\r\nccc\r\n => "1) \"xx\"\r\n2) \"ccc\"\r\n"
func respParseArrays(buf []byte, length int) (string, error) {
	_, str, err := respParseArrays1(buf, length, 0)
	if str != "" {
		str = str[:len(str)-2]
	}
	return str, err
}

// ================================ format response string =================================

// e.g. "hello" => "\"hello\""
func bulkStrFormat(s string) string {
	if s != NIL_STR {
		return fmt.Sprintf("\"%s\"", s)
	}
	return ""
}

// e.g. "ERR: xxxx" => "(error) ERR: xxxx"
func simpleErrStrFormat(s string) string {
	return fmt.Sprintf("(error) %s", s)
}

// e.g. "15" => "(integer) 15"
func intStrFormat(s string) string {
	return fmt.Sprintf("(integer) %s", s)
}
