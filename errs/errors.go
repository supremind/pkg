package errs

import "fmt"

type Errors []error

func (e Errors) Error() string {
	return fmt.Sprint("errors: ", e)
}
