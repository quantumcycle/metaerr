package metaerr

import "context"

type Option func(*Error)

func WithLocationSkip(additionalCallerSkip int) Option {
	return func(e *Error) {
		//+1 since this is called from the option
		e.Location = getLocation(additionalCallerSkip)
	}
}

func WithStackTrace(additionalCallerSkip, maxDepth int) Option {
	return func(e *Error) {
		//+1 because we always skip the first frame since it will be the same as the location
		e.Stacktrace = newStacktrace(additionalCallerSkip+1, maxDepth)
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
