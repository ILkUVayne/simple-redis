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

//-----------------------------------------------------------------------------
// Parse resp data
//-----------------------------------------------------------------------------

// ================================ tool func =================================

// e.g. buf = "get name\r\n" return 8
//
// buf = "$3\r\nget\r\n$4\r\nname\r\n" return 2
func getQueryLine(buf []byte, queryLen int) (int, error) {
	idx := strings.Index(string(buf), "\r\n")
	if idx < 0 && queryLen > SREDIS_MAX_INLINE {
		return idx, errors.New("inline cmd is too long")
	}
	return idx, nil
}

// buf = "$3\r\nget\r\n$4\r\nname\r\n" return 3
func getQueryNum(buf []byte, queryLen int, idx int) ([]byte, int, error) {
	var err error
	if idx == 0 {
		idx, err = getQueryLine(buf, queryLen)
		if idx < 0 {
			return buf, INCOMPLETE_STR, err
		}
	}
	qLen, err := strconv.Atoi(string(buf[1:idx]))
	return buf[idx+2:], qLen, err
}

// buf = "$3\r\nget\r\n$4\r\nname\r\n" return ["$4\r\nname\r\n", 3, get, nil]
func getBulk(buf []byte, length int, idx int) ([]byte, int, string, error) {
	newBuf, bLen, err := getQueryNum(buf, length, idx)
	if err != nil || bLen <= 0 {
		return newBuf, bLen, "", err
	}
	idx, err = getQueryLine(newBuf, length)
	if err != nil || idx < 0 {
		return newBuf, INCOMPLETE_STR, "", err
	}
	return newBuf[idx+2:], bLen, string(newBuf[:idx]), nil
}

// 递归解析数组数据
func respParseArrays1(buf []byte, length int, spaceNum int) ([]byte, string, error) {
	// get array length
	newBuf, aLen, err := getQueryNum(buf, length, 0)
	if err != nil || aLen == INCOMPLETE_STR {
		return buf, "", err
	}
	buf = newBuf
	// check empty
	if aLen == -1 {
		return buf, "(nil)\r\n" + spaces(spaceNum), nil
	}
	if aLen == 0 {
		return buf, "(empty array)\r\n" + spaces(spaceNum), nil
	}

	str, subStr := "", ""

	for i := 0; i < aLen; i++ {
		// 递归解析嵌套数组
		if buf[0] == '*' {
			str += strconv.Itoa(i+1) + ") "
			buf, subStr, err = respParseArrays1(buf, length, spaceNum+3)

			if err != nil || subStr == "" {
				return buf, "", err
			}
			str += subStr[:len(subStr)-(spaceNum+3)]
			continue
		}
		// 解析bulk
		newBuf, bulkLen, bulk, err := getBulk(buf, length, 0)
		if err != nil || bulkLen == INCOMPLETE_STR {
			return buf, "", err
		}
		buf = newBuf
		if bulkLen == -1 {
			str += strconv.Itoa(i+1) + ") (nil)\r\n" + spaces(spaceNum)
			continue
		}
		if bulkLen == 0 {
			str += strconv.Itoa(i+1) + ") \r\n" + spaces(spaceNum)
			continue
		}
		str += strconv.Itoa(i+1) + ") \"" + bulk + "\"\r\n" + spaces(spaceNum)
	}
	return buf, str, nil
}

// ================================ respParseFunc Impl =================================

// e.g. "+OK\r\n" => "OK"
func respParseSimpleStr(buf []byte, _ int, idx int) (string, error) {
	return string(buf[1:idx]), nil
}

// e.g. "$5\r\nhello\r\n" => "hello"
func respParseBulk(buf []byte, length int, idx int) (string, error) {
	_, bulkLen, bulk, err := getBulk(buf, length, idx)
	if err != nil || bulkLen == INCOMPLETE_STR {
		return "", err
	}
	if bulkLen == -1 {
		return "(nil)", nil
	}
	if bulkLen == 0 && length == 4 {
		return "empty string", nil
	}
	return bulk, nil
}

// e.g. ":3\r\n" => 3
func respParseIntegers(buf []byte, _ int, idx int) (string, error) {
	return string(buf[1:idx]), nil
}

// e.g. *2\r\n$2\r\nxx\r\n$3\r\nccc\r\n => "1) \"xx\"\r\n2) \"ccc\"\r\n"
func respParseArrays(buf []byte, length int, _ int) (string, error) {
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
