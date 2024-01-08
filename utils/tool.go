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

func String2Int64(s *string, intVal *int64) bool {
	i, err := strconv.ParseInt(*s, 10, 64)
	if err != nil {
		return false
	}
	if intVal != nil {
		*intVal = i
	}
	return true
}

func String2Float64(s *string, intVal *float64) bool {
	i, err := strconv.ParseFloat(*s, 64)
	if err != nil {
		return false
	}
	if intVal != nil {
		*intVal = i
	}
	return true
}
