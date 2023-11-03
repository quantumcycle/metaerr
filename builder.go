package metaerr

import (
	"context"
	"fmt"
)

type Builder struct {
	opts    []Option
	metas   []ErrorMetadata
	context context.Context
}

func (b Builder) Meta(meta ...ErrorMetadata) Builder {
	return Builder{
		opts:    b.opts,
		context: b.context,
		metas:   append(b.metas, meta...),
	}
}

func (b Builder) Context(ctx context.Context) Builder {
	return Builder{
		opts:    b.opts,
		metas:   b.metas,
		context: ctx,
	}
}

func (b Builder) Newf(format string, args ...any) Error {
	return New(fmt.Sprintf(format, args...), b.opts...).WithMeta(b.metas...).WithContext(b.context)
}

func (b Builder) Wrapf(err error, format string, args ...any) *Error {
	w := Wrap(err, fmt.Sprintf(format, args...), b.opts...)
	if w == nil {
		return nil
	}
	w2 := (*w).WithMeta(b.metas...).WithContext(b.context)
	return &w2
}

func NewBuilder(opt ...Option) Builder {
	return Builder{
		opts: opt,
	}
}
