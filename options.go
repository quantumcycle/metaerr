package metaerr

import "context"

type Option func(*Error)

func WithLocationSkip(additionalCallerSkip int) Option {
	return func(e *Error) {
		//+1 since this is called from the option
		e.Location = getLocation(additionalCallerSkip, e.rootDetector)
	}
}

func WithStackTrace(additionalCallerSkip, maxDepth int) Option {
	return func(e *Error) {
		//+1 because we always skip the first frame since it will be the same as the location
		e.Stacktrace = newStacktrace(additionalCallerSkip+1, maxDepth, e.rootDetector)
	}
}

// WithRootPackageDetector overrides how stack capture decides a frame's package
// is a stack-terminating "root" (standard library / runtime). isRoot receives an
// import path (e.g. "net/http", "github.com/x/y") and returns true to stop the
// walk there. When unset, DefaultRootPackage is used.
//
// Use it for projects whose module path has no domain (e.g. `go mod init myapp`),
// which DefaultRootPackage would otherwise mistake for the standard library:
//
//	metaerr.WithRootPackageDetector(func(pkg string) bool {
//		if strings.HasPrefix(pkg, "myapp") {
//			return false // my code, keep walking
//		}
//		return metaerr.DefaultRootPackage(pkg)
//	})
//
// Capture is eager, so this option must be applied BEFORE WithLocationSkip /
// WithStackTrace. A Builder applies its options in order, so list it first.
func WithRootPackageDetector(isRoot func(pkg string) bool) Option {
	return func(e *Error) {
		e.rootDetector = isRoot
	}
}

func WithContext(ctx context.Context) Option {
	return func(e *Error) {
		e.Context = ctx
	}
}

func WithMeta(metas ...ErrorMetadata) Option {
	return func(e *Error) {
		e.Metas = append(e.Metas, metas...)
	}
}
