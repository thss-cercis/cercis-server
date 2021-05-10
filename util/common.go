package util

import (
	"fmt"
)

func MsgWithError(msg string, err error) string {
	if err != nil {
		return fmt.Sprintf("Msg: %s, Error: %v", msg, err.Error())
	}
	return fmt.Sprintf("Msg: %s", msg)
}

func FirstNCharOfString(s string, n int) string {
	if len(s) < n {
		return s
	} else {
		return s[0:n]
	}
}
