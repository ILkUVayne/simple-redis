package utils

import (
	"bytes"
	"errors"
	"github.com/mattn/go-isatty"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
)

const (
	REDIS_OK  = 0
	REDIS_ERR = 1
)

// StrToHost string type host to []byte host
// e.g. "127.0.0.1" -> []byte{127,0,0,1}
func StrToHost(host string) [4]byte {
	hosts := strings.Split(host, ".")
	if len(hosts) != 4 {
		Error("str2host error: host is bad, host == ", host)
	}
	var h [4]byte
	for idx, v := range hosts {
		i, err := strconv.Atoi(v)
		if err != nil {
			Error("str2host error: host is bad: ", host)
		}
		h[idx] = uint8(i)
	}
	return h
}

// Home returns the home directory for the executing user.
//
// This uses an OS-specific method for discovering the home directory.
// An error is returned if a home directory cannot be detected.
func Home() (string, error) {
	usr, err := user.Current()
	if nil == err {
		return usr.HomeDir, nil
	}

	// cross compile support

	if "windows" == runtime.GOOS {
		return homeWindows()
	}

	// Unix-like system, so just assume Unix
	return homeUnix()
}

func homeUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}

func Isatty() bool {
	if isatty.IsTerminal(os.Stdin.Fd()) {
		return true
	}
	if isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		return true
	}
	return false
}

func Exit(code int) {
	os.Exit(code)
}

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
