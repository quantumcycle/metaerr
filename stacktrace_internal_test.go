package metaerr

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func TestPackageOf(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"runtime.main", "runtime"},
		{"testing.tRunner", "testing"},
		{"net/http.(*conn).serve", "net/http"},
		{"github.com/quantumcycle/metaerr.New", "github.com/quantumcycle/metaerr"},
		{"github.com/quantumcycle/metaerr.(*Error).Error", "github.com/quantumcycle/metaerr"},
		{"github.com/x/y.Outer.func1", "github.com/x/y"},
		{"main.main", "main"},
		{"main", "main"},
		{"runtime", "runtime"},
	}
	for _, c := range cases {
		if got := packageOf(c.name); got != c.want {
			t.Errorf("packageOf(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestDefaultRootPackage(t *testing.T) {
	cases := []struct {
		pkg  string
		want bool
	}{
		// standard library / runtime
		{"runtime", true},
		{"testing", true},
		{"net/http", true},
		{"internal/poll", true},
		// conventional user modules (have a domain)
		{"github.com/quantumcycle/metaerr", false},
		{"golang.org/x/net/http2", false},
		{"main", false},
		// KNOWN LIMITATION: a domain-less module path is indistinguishable from
		// the stdlib by this heuristic and is classified as a root. The keep-first
		// rule and WithRootPackageDetector keep this from emptying a stack.
		{"myapp", true},
		{"internal/foo", true},
	}
	for _, c := range cases {
		if got := DefaultRootPackage(c.pkg); got != c.want {
			t.Errorf("DefaultRootPackage(%q) = %v, want %v", c.pkg, got, c.want)
		}
	}
}

func aUserFuncForStackTest() {}

// TestIsRootFrameClassifiesByPackage guards the -trimpath regression: frame
// classification must not depend on file paths or runtime.GOROOT() (both of
// which -trimpath defeats), only on the import-path-based symbol name.
func TestIsRootFrameClassifiesByPackage(t *testing.T) {
	stdlibPC := reflect.ValueOf(fmt.Sprintf).Pointer()
	if !isRootFrame(stdlibPC, DefaultRootPackage) {
		t.Errorf("fmt.Sprintf must be classified as a root frame (name=%q)",
			runtime.FuncForPC(stdlibPC).Name())
	}

	userPC := reflect.ValueOf(aUserFuncForStackTest).Pointer()
	if isRootFrame(userPC, DefaultRootPackage) {
		t.Errorf("a user-package func must not be a root frame (name=%q)",
			runtime.FuncForPC(userPC).Name())
	}
}

func TestIsRootFrameNilFunc(t *testing.T) {
	if isRootFrame(0, DefaultRootPackage) {
		t.Error("zero PC (FuncForPC == nil) must not be a root frame")
	}
}
