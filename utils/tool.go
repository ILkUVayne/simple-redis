package utils

import (
	"bytes"
	"encoding/binary"
	"github.com/ILkUVayne/utlis-go/v2/flie"
	"os"
	"strconv"
	"strings"
)

const (
	REDIS_OK  = 0
	REDIS_ERR = 1
)

//-----------------------------------------------------------------------------
// sys function
//-----------------------------------------------------------------------------

func absolutePath(file string) string {
	str, err := flie.Home()
	if err != nil {
		Error(err)
	}
	return str + "/" + file
}

func HistoryFile(file string) string {
	return absolutePath(file)
}

func PersistenceFile(file string) string {
	return absolutePath(file)
}

func Exit(code int) {
	os.Exit(code)
}

//-----------------------------------------------------------------------------
// transform function
//-----------------------------------------------------------------------------

func String2Int64(s *string, intVal *int64) int {
	i, err := strconv.ParseInt(*s, 10, 64)
	if err != nil {
		return REDIS_ERR
	}
	if intVal != nil {
		*intVal = i
	}
	return REDIS_OK
}

func String2Float64(s *string, intVal *float64) int {
	i, err := strconv.ParseFloat(*s, 64)
	if err != nil {
		return REDIS_ERR
	}
	if intVal != nil {
		*intVal = i
	}
	return REDIS_OK
}

func uint8ToLower(n uint8) uint8 {
	return []byte(strings.ToLower(string(n)))[0]
}

func Int2Bytes(i int) []byte {
	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, binary.BigEndian, int64(i))
	if err != nil {
		Error("Int2Bytes err: ", err)
	}
	return buf.Bytes()
}

func Bytes2Int64(buff []byte) int64 {
	var i int64
	buf := bytes.NewBuffer(buff)
	err := binary.Read(buf, binary.BigEndian, &i)
	if err != nil {
		Error("Bytes2Int64 err: ", err)
	}
	return i
}

//-----------------------------------------------------------------------------
// match function
//-----------------------------------------------------------------------------

func StringMatchLen(pattern, str string, patternLen, strLen int, noCase bool) bool {
	pIdx, sIdx := 0, 0
	for patternLen > 0 {
		switch pattern[pIdx] {
		case '*':
			if patternLen == 1 {
				return true
			}
			for pattern[pIdx+1] == '*' {
				pIdx++
				patternLen--
			}
			if patternLen == 1 {
				return true
			}
			for strLen > 0 {
				if StringMatchLen(pattern[pIdx+1:], str[sIdx:], patternLen-1, strLen, noCase) {
					return true
				}
				sIdx++
				strLen--
			}
			return false
		case '?':
			if strLen == 0 {
				return false
			}
			sIdx++
			strLen--
			break
		case '[':
			pIdx++
			patternLen--
			not, match := false, false
			if pattern[pIdx] == '^' {
				not = true
				pIdx++
				patternLen--
			}
			for {
				if pattern[pIdx] == '\\' {
					pIdx++
					patternLen--
					if pattern[pIdx] == str[sIdx] {
						match = true
					}
				}
				if pattern[pIdx] == ']' {
					break
				}
				if patternLen == 0 {
					pIdx--
					patternLen++
					break
				}
				if pattern[pIdx+1] == '-' && patternLen >= 3 {
					start := pattern[pIdx]
					end := pattern[pIdx+2]
					c := str[sIdx]
					if start > end {
						t := start
						start = end
						end = t
					}
					if noCase {
						start = uint8ToLower(start)
						end = uint8ToLower(end)
						c = uint8ToLower(c)
					}
					pIdx += 2
					patternLen -= 2
					if c >= start && c <= end {
						match = true
					}
				} else {
					if !noCase {
						if pattern[pIdx] == str[sIdx] {
							match = true
						} else {
							if uint8ToLower(pattern[pIdx]) == uint8ToLower(str[sIdx]) {
								match = true
							}
						}
					}
				}
				pIdx++
				patternLen--
			}
			if not {
				match = !match
			}
			if !match {
				return false
			}
			sIdx++
			strLen--
			break
		case '\\':
			if patternLen >= 2 {
				pIdx++
				patternLen--
			}
			fallthrough
		default:
			if !noCase {
				if pattern[pIdx] != str[sIdx] {
					return false
				}
			} else {
				if uint8ToLower(pattern[pIdx]) != uint8ToLower(str[sIdx]) {
					return false
				}
			}
			sIdx++
			strLen--
			break
		}
		pIdx++
		patternLen--
		if strLen == 0 {
			for pattern[pIdx:] == "*" {
				pIdx++
				patternLen--
			}
			break
		}
	}
	if patternLen == 0 && strLen == 0 {
		return true
	}
	return false
}

func StringMatch(pattern, str string, noCase bool) bool {
	return StringMatchLen(pattern, str, len(pattern), len(str), noCase)
}
