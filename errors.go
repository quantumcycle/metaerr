package metaerr

import (
	"bytes"
	"errors"
	stderr "errors"
	"fmt"
	"io"
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

func StringerMeta[T fmt.Stringer](name string) func(val T) ErrorMetadata {
	return func(val T) ErrorMetadata {
		return func(err Error) []MetaValue {
			return []MetaValue{
				{
					Name:   name,
					Values: []string{val.String()},
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

func Wrap(err error, msg string, opt ...Option) *Error {
	if err == nil {
		return nil
	}

	e := Error{
		reason:   msg,
		location: getLocation(defaultCallerSkip),
		cause:    err,
	}

	if len(opt) > 0 {
		for _, o := range opt {
			o(&e)
		}
	}

	return &e
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
	reason     string
	meta       []ErrorMetadata
	location   string
	cause      error
	stacktrace *stacktrace
}

func (e Error) Meta(metas ...ErrorMetadata) Error {
	return Error{
		reason:     e.reason,
		location:   e.location,
		cause:      e.cause,
		stacktrace: e.stacktrace,
		meta:       append(e.meta, metas...),
	}
}

func (e Error) Unwrap() error {
	return e.cause
}

func (e Error) Error() string {
	buf := new(bytes.Buffer)
	e.printError(buf, false)
	return buf.String()
}

func (e Error) Reason() string {
	return e.reason
}

func (e Error) Location() string {
	return e.location
}

func getError(err error) (Error, bool) {
	if metaError, ok := err.(Error); ok {
		return metaError, true
	}
	if metaErrorPtr, ok := err.(*Error); ok {
		return *metaErrorPtr, true
	}
	return Error{}, false
}

type errorWriter interface {
	Error(msg, metadata, location string, stacktrace *stacktrace)
}

type stackErrorWriter struct {
	writer           io.Writer
	firstLinePrinted bool
}

func (ew *stackErrorWriter) Error(msg, metadata, location string, st *stacktrace) {
	if msg == "" && metadata == "" && location == "" {
		return
	}
	if ew.firstLinePrinted {
		fmt.Fprint(ew.writer, "\n")
	}
	if msg != "" {
		fmt.Fprint(ew.writer, msg)
		ew.firstLinePrinted = true
	}

	if metadata != "" {
		if msg != "" {
			fmt.Fprint(ew.writer, " ")
		}
		fmt.Fprint(ew.writer, metadata)
		ew.firstLinePrinted = true
	}

	if location == "" {
		return
	}
	if ew.firstLinePrinted {
		fmt.Fprintf(ew.writer, "\n")
	}
	fmt.Fprintf(ew.writer, "\tat %s", location)
	ew.firstLinePrinted = true

	if st != nil && len(st.frames) > 0 {
		fmt.Fprintf(ew.writer, "\n")
		for i, frame := range st.frames {
			fmt.Fprintf(ew.writer, "\tat %s", frame.String())
			if i < len(st.frames)-1 {
				fmt.Fprintf(ew.writer, "\n")
			}
		}
	}

}

type lineErrorWriter struct {
	writer            io.Writer
	firstErrorPrinted bool
}

func (ew *lineErrorWriter) Error(msg, metadata, location string, st *stacktrace) {
	if msg == "" && metadata == "" {
		return
	}
	if ew.firstErrorPrinted {
		fmt.Fprint(ew.writer, ": ")
	}
	if msg != "" {
		fmt.Fprint(ew.writer, msg)
		ew.firstErrorPrinted = true
	}
	if metadata != "" {
		if msg != "" {
			fmt.Fprint(ew.writer, " ")
		}
		fmt.Fprint(ew.writer, metadata)
		ew.firstErrorPrinted = true
	}
}

func (e Error) printError(w io.Writer, withLocation bool) {
	var err error = e
	var errWriter errorWriter
	if withLocation {
		errWriter = &stackErrorWriter{
			writer: w,
		}
	} else {
		errWriter = &lineErrorWriter{
			writer: w,
		}
	}
	for err != nil {
		var message string = ""
		var location string = ""
		var metaMsg string = ""
		var st *stacktrace

		if metaError, ok := getError(err); ok {
			message = metaError.Reason()
			if len(metaError.meta) > 0 {
				metasStr := make([]string, 0, len(metaError.meta))
				metas := GetMeta(metaError, false)
				for k, v := range metas {
					metasStr = append(metasStr, fmt.Sprintf("[%s=%s]", k, strings.Join(v, ",")))
				}
				//To make the output deterministic
				sort.Strings(metasStr)
				metaMsg = strings.Join(metasStr, " ")
			}
			if withLocation && metaError.location != "" {
				location = metaError.location
			}
			if withLocation && metaError.stacktrace != nil {
				st = metaError.stacktrace
			}
		} else {
			message = err.Error()
		}
		errWriter.Error(message, metaMsg, location, st)
		err = stderr.Unwrap(err)
	}
}

func (e Error) Format(s fmt.State, verb rune) {
	detailledPrint := s.Flag('+')

	switch verb {
	case 'v':
		e.printError(s, detailledPrint)
	case 's':
		e.printError(s, false)
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

func WithStackTrace(additionalCallerSkip, maxDepth int) Option {
	return func(e *Error) {
		//+1 since this is called from the option
		//+1 because we always skip the first frame since it will be the same as the location
		e.stacktrace = newStacktrace(additionalCallerSkip+defaultCallerSkip+2, maxDepth)
	}
}

func New(reason string, opt ...Option) Error {
	e := Error{
		reason:   reason,
		location: getLocation(defaultCallerSkip),
	}

	if len(opt) > 0 {
		for _, o := range opt {
			o(&e)
		}
	}

	return e
}

func getLocation(callerSkip int) string {
	_, file, line, _ := runtime.Caller(callerSkip)
	return fmt.Sprintf("%s:%d", file, line)
}
