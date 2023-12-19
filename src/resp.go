package src

import (
	"errors"
	"strconv"
	"strings"
)

const (
	RESP_NIL_VAL      = "$-1\r\n"
	RESP_TYP_ERR      = "-ERR: wrong type\r\n"
	RESP_OK           = "+OK\r\n"
	RESP_UNKOWN       = "-ERR: unknow command\r\n"
	RESP_ARGS_NUM_ERR = "-ERR: wrong number of args\r\n"

	RESP_BULK = "$%d\r\n%v\r\n"
)

const (
	SIMPLE_STR   = iota + 1 // +OK\r\n
	SIMPLE_ERROR            // -Error message\r\n
	INTEGERS                // :[<+|->]<value>\r\n
	BULK_STR                // $<length>\r\n<data>\r\n
	ARRAYS                  // *<number-of-elements>\r\n<element-1>...<element-n>
	NULLS                   // _\r\n
	BOOLEANS                // #<t|f>\r\n
	DOUBLE                  // ,[<+|->]<integral>[.<fractional>][<E|e>[sign]<exponent>]\r\n e.g. ,1.23\r\n
	BIG_NUMBERS             // ([+|-]<number>\r\n
	BULK_ERR                // !<length>\r\n<error>\r\n
	VERBATIM_STR            // =<length>\r\n<encoding>:<data>\r\n
	MAPS                    // %<number-of-entries>\r\n<key-1><value-1>...<key-n><value-n>
	SETS                    // ~<number-of-elements>\r\n<element-1>...<element-n>
	PUSHES                  // ><number-of-elements>\r\n<element-1>...<element-n>
	// more
)

type respParseFunc func([]byte, int) (string, error)

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

var respParseFuncs = map[int]respParseFunc{
	SIMPLE_STR:   respParseSimpleStr,
	SIMPLE_ERROR: respParseSimpleStr,
	BULK_STR:     respParseBulk,
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
	if sLen == 0 && len(buf) == 6 {
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
