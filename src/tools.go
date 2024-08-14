package src

import (
	"github.com/ILkUVayne/utlis-go/v2/flie"
	"github.com/ILkUVayne/utlis-go/v2/math"
	"github.com/ILkUVayne/utlis-go/v2/ulog"
)

//-----------------------------------------------------------------------------
// common function
//-----------------------------------------------------------------------------

func isEmpty(e empty) bool {
	return e.isEmpty()
}

//-----------------------------------------------------------------------------
// sys function
//-----------------------------------------------------------------------------

func absolutePath(file string) string {
	str, err := flie.Home()
	if err != nil {
		ulog.Error(err)
	}
	return str + "/" + file
}

func HistoryFile(file string) string {
	return absolutePath(file)
}

func PersistenceFile(file string) string {
	return absolutePath(file)
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
						start = math.Uint8ToLower(start)
						end = math.Uint8ToLower(end)
						c = math.Uint8ToLower(c)
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
							if math.Uint8ToLower(pattern[pIdx]) == math.Uint8ToLower(str[sIdx]) {
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
				if math.Uint8ToLower(pattern[pIdx]) != math.Uint8ToLower(str[sIdx]) {
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
