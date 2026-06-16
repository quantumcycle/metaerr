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
		Location: getLocation(0, nil),
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
	// rootDetector classifies whether an import path is a stack-terminating
	// "root" package (stdlib/runtime). nil means DefaultRootPackage. Set via
	// WithRootPackageDetector; must be applied before WithLocationSkip /
	// WithStackTrace to take effect (those capture eagerly).
	rootDetector func(pkg string) bool
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
		Location: getLocation(0, nil),
	}

	if len(opt) > 0 {
		for _, o := range opt {
			o(&e)
		}
	}

	return e
}

func getLocation(callerSkip int, isRoot func(pkg string) bool) string {
	st := newStacktrace(callerSkip, 1, isRoot)
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

func newStacktrace(frameStackSkip, maxDepth int, isRoot func(pkg string) bool) *Stacktrace {
	if isRoot == nil {
		isRoot = DefaultRootPackage
	}
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

	// Collect frames up to maxDepth, stopping once we reach the stdlib/runtime
	// (it won't call back into user code).
	//
	// We always keep the FIRST frame and only apply the stdlib check from the
	// second one on. The first frame is the site that created the error, which is
	// user code by construction (the stdlib never calls into metaerr). This makes
	// the worst case graceful: the root classifier can misclassify user code in
	// a domain-less module (see DefaultRootPackage), and without this guard such
	// a frame would be dropped, leaving an empty stack / location.
	for i := index + frameStackSkip; len(frames) < maxDepth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if len(frames) > 0 && isRootFrame(pc, isRoot) {
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

// packageOf returns the import path of the package a fully-qualified function
// name belongs to.
//
//	"net/http.(*conn).serve"              -> "net/http"
//	"runtime.main"                        -> "runtime"
//	"github.com/quantumcycle/metaerr.New" -> "github.com/quantumcycle/metaerr"
func packageOf(funcName string) string {
	// The function/method part starts at the first '.' that follows the last '/'.
	if slash := strings.LastIndexByte(funcName, '/'); slash >= 0 {
		if dot := strings.IndexByte(funcName[slash:], '.'); dot >= 0 {
			return funcName[:slash+dot]
		}
		return funcName
	}
	if dot := strings.IndexByte(funcName, '.'); dot >= 0 {
		return funcName[:dot]
	}
	return funcName
}

func isRootFrame(pc uintptr, isRoot func(pkg string) bool) bool {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return false
	}
	return isRoot(packageOf(fn.Name()))
}

// DefaultRootPackage is the default root-package classifier, used when none is
// configured via WithRootPackageDetector. It reports whether an import path
// belongs to the standard library or runtime — i.e. where stack capture stops.
//
// Stdlib import paths have no domain (no '.') in their first
// path segment ("runtime", "net/http", "testing"), whereas conventional module
// paths do ("github.com/...", "golang.org/x/..."). This is not fullproof.
// Modules declared with a domain-less path (e.g. `go mod init myapp` → "myapp",
// "internal/foo") are indistinguishable from the stdlib by this rule and are
// classified as a root.
func DefaultRootPackage(pkg string) bool {
	if pkg == "main" {
		return false
	}
	first := pkg
	if slash := strings.IndexByte(pkg, '/'); slash >= 0 {
		first = pkg[:slash]
	}
	return !strings.Contains(first, ".")
}
