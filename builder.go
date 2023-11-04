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
	metas := make([]ErrorMetadata, len(b.metas)+len(meta))
	copy(metas, b.metas)
	copy(metas[len(b.metas):], meta)
	return Builder{
		opts:    b.opts,
		context: b.context,
		metas:   meta,
	}
}

func (b Builder) Context(ctx context.Context) Builder {
	return Builder{
		opts:    b.opts,
		metas:   b.metas,
		context: ctx,
	}
}

func (b Builder) Newf(format string, args ...any) error {
	opts := append(b.opts, WithMeta(b.metas...), WithContext(b.context))
	return New(fmt.Sprintf(format, args...), opts...)
}

func (b Builder) Wrapf(err error, format string, args ...any) error {
	opts := append(b.opts, WithMeta(b.metas...), WithContext(b.context))
	return Wrap(err, fmt.Sprintf(format, args...), opts...)
}

func NewBuilder(opt ...Option) Builder {
	return Builder{
		opts: opt,
	}
}
