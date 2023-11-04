package metaerr

import (
	"bytes"
	"context"
	stderr "errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"slices"
	"sort"
	"strings"
)

func Wrap(err error, msg string, opt ...Option) error {
	if err == nil {
		return nil
	}

	e := Error{
		Reason:   msg,
		Location: getLocation(0),
		Cause:    err,
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
		if metaErr, ok := AsMetaError(err); ok {
			for _, m := range metaErr.Metas {
				values := m(metaErr)
				for _, val := range values {
					//ignore empty metadata
					if len(val.Values) == 0 {
						continue
					}
					if meta[val.Name] == nil {
						meta[val.Name] = make([]string, 0, len(val.Values))
					}
					for _, v := range val.Values {
						if v != "" {
							meta[val.Name] = append(meta[val.Name], v)
						}
					}
				}
			}

			//sort all slices to make output deterministic
			for k, v := range meta {
				sort.Strings(v)
				//remove consecutive duplicates
				meta[k] = slices.Compact(v)
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
	Context    context.Context
	Location   string
	Reason     string
	Stacktrace *Stacktrace
	Cause      error
	Metas      []ErrorMetadata
}

func (e Error) Unwrap() error {
	return e.Cause
}

func (e Error) Error() string {
	buf := new(bytes.Buffer)
	e.printError(buf, false)
	return buf.String()
}

func AsMetaError(err error) (Error, bool) {
	if metaError, ok := err.(Error); ok {
		return metaError, true
	}
	if metaErrorPtr, ok := err.(*Error); ok {
		return *metaErrorPtr, true
	}
	return Error{}, false
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
		var st *Stacktrace

		if metaError, ok := AsMetaError(err); ok {
			message = metaError.Reason
			if len(metaError.Metas) > 0 {
				metasStr := make([]string, 0, len(metaError.Metas))
				metas := GetMeta(metaError, false)
				for k, v := range metas {
					metasStr = append(metasStr, fmt.Sprintf("[%s=%s]", k, strings.Join(v, ",")))
				}
				//To make the output deterministic
				sort.Strings(metasStr)
				metaMsg = strings.Join(metasStr, " ")
			}
			if withLocation && metaError.Location != "" {
				location = metaError.Location
			}
			if withLocation && metaError.Stacktrace != nil {
				st = metaError.Stacktrace
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

func New(reason string, opt ...Option) error {
	e := Error{
		Reason:   reason,
		Location: getLocation(0),
	}

	if len(opt) > 0 {
		for _, o := range opt {
			o(&e)
		}
	}

	return e
}

func getLocation(callerSkip int) string {
	st := newStacktrace(callerSkip, 1)
	if len(st.Frames) == 0 {
		return ""
	}
	return st.Frames[0].String()
}

type Stacktrace struct {
	Frames []Frame
}

type Frame struct {
	File string
	Line int
}

func (frame *Frame) String() string {
	return fmt.Sprintf("%v:%v", frame.File, frame.Line)
}

// Just a struct to be able to get the internal package path of this library to exclude it
type internal struct{}

var internalPath = reflect.TypeOf(internal{}).PkgPath()

func newStacktrace(frameStackSkip, maxDepth int) *Stacktrace {
	var frames []Frame
	var index = 0

	// start by skipping everything related to this package
	for {
		_, file, _, _ := runtime.Caller(index)
		if !strings.Contains(file, internalPath) {
			break
		}
		if !strings.Contains(file, "/builder.go") &&
			!strings.Contains(file, "/errors.go") &&
			!strings.Contains(file, "/options.go") {
			break
		}
		index++
	}

	// We loop until we have StackTraceMaxDepth frames or we run out of frames.
	// Frames from this package are skipped.
	for i := index + frameStackSkip; len(frames) < maxDepth; i++ {
		_, file, line, ok := runtime.Caller(i)
		//Once we find a frame in the stdlib, we stop, since stdlib code won't call back to user code
		if !ok || strings.Contains(file, runtime.GOROOT()) {
			break
		}

		frames = append(frames, Frame{
			File: file,
			Line: line,
		})
	}

	return &Stacktrace{
		Frames: frames,
	}
}
