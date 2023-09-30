package metaerr

import (
	"errors"
	stderr "errors"
	"fmt"
	"runtime"
	"sort"
	"strings"
)

func StringMeta(name string) func(val string) ErrorMetadata {
	return func(val string) ErrorMetadata {
		return func(err Error) []MetaValue {
			return []MetaValue{
				{
					Name:   name,
					Values: []string{val},
				},
			}
		}
	}
}

type MetaValue struct {
	Name   string
	Values []string
}

type ErrorMetadata = func(err Error) []MetaValue

func Wrap(err error, msg string) *Error {
	if err == nil {
		return nil
	}

	return &Error{
		reason:   msg,
		location: getLocation(defaultCallerSkip),
		cause:    err,
	}
}

func GetMeta(err error, nested bool) map[string][]string {
	meta := make(map[string][]string)

	for err != nil {
		var f Error
		if errors.As(err, &f) {
			for _, m := range f.meta {
				values := m(f)
				for _, val := range values {
					if meta[val.Name] == nil {
						meta[val.Name] = make([]string, 0, len(val.Values))
					}
					meta[val.Name] = append(meta[val.Name], val.Values...)
				}
			}

			//sort all slices to make output deterministic
			for _, v := range meta {
				sort.Strings(v)
			}
		}

		if !nested {
			break
		}
		err = stderr.Unwrap(err)
	}

	return meta
}

type Error struct {
	reason   string
	meta     []ErrorMetadata
	location string
	cause    error
}

func (e Error) Meta(metas ...ErrorMetadata) Error {
	return Error{
		reason:   e.reason,
		location: e.location,
		cause:    e.cause,
		meta:     append(e.meta, metas...),
	}
}

func (e Error) Unwrap() error {
	return e.cause
}

func (e Error) Error() string {
	return e.reason
}

func (e Error) Location() string {
	return e.location
}

func (e Error) Format(s fmt.State, verb rune) {
	var err error = e
	switch verb {
	case 'v':
		for err != nil {
			var message string = err.Error()
			var location string = "[no location]"
			var metaMsg string = ""

			if metaError, ok := err.(Error); ok {
				if s.Flag('+') {
					if len(metaError.meta) > 0 {
						metasStr := make([]string, 0, len(metaError.meta))
						metas := GetMeta(metaError, false)
						for k, v := range metas {
							metasStr = append(metasStr, fmt.Sprintf("[%s=%s]", k, strings.Join(v, ", ")))
						}
						//To make the output deterministic
						sort.Strings(metasStr)
						metaMsg = fmt.Sprintf(" %s", strings.Join(metasStr, " "))
					}
					if metaError.location != "" {
						location = metaError.location
					}
				}
			}
			if message == "" {
				message = "[no message]"
			}
			fmt.Fprintf(s, "%s%s\n\tat %s\n", message, metaMsg, location)
			err = stderr.Unwrap(err)
		}
	case 's':
		fmt.Fprint(s, e.Error())
	}
}

type Option func(*Error)

const defaultCallerSkip = 2

func WithLocationSkip(additionalCallerSkip int) Option {
	return func(e *Error) {
		//+1 since this is called from the option
		e.location = getLocation(additionalCallerSkip + defaultCallerSkip + 1)
	}
}

func New(reason string, opt ...Option) Error {
	if len(opt) > 0 {
		e := Error{
			reason: reason,
		}
		for _, o := range opt {
			o(&e)
		}
		return e
	}

	// Default error construction
	return Error{
		reason:   reason,
		location: getLocation(defaultCallerSkip),
	}
}

func getLocation(callerSkip int) string {
	_, file, line, _ := runtime.Caller(callerSkip)
	return fmt.Sprintf("%s:%d", file, line)
}