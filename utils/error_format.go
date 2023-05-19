package utils

import "fmt"

type NetUtilError struct {
	errInfo string
}

func (err *NetUtilError) Error() string {
	return fmt.Sprintf("%s\n", err.errInfo)
}

func NewError(err string) *NetUtilError {
	return &NetUtilError{errInfo: err}
}
