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
