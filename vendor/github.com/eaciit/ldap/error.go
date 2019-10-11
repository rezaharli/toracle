package ldap

import (
	"fmt"
)

type ValueMismatchError struct {
	got		interface{}
}

func (v *ValueMismatchError) Error() string {
	return fmt.Sprintf("Unexpected value type: %T", v.got)
}

func NewValueMismatchError(got interface{}) *ValueMismatchError {
	return &ValueMismatchError{got: got}
}