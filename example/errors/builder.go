package errors

import (
	"context"
	"github.com/quantumcycle/metaerr"
)

var errorCode = metaerr.StringMeta("error_code")
var tag = metaerr.StringsMeta("tag")
var userID = metaerr.StringMetaFromContext("user", "user_id")
var builder = metaerr.NewBuilder(
	//skipping one caller to exclude this builder from the stacktraces
	metaerr.WithLocationSkip(1),
	metaerr.WithStackTrace(1, 5)).
	//we can add userID immediately since the value comes from the context
	Meta(userID())

type ErrorBuilder struct {
	errorCode string
	tags      []string
}

func New() *ErrorBuilder {
	return &ErrorBuilder{}
}

func (b *ErrorBuilder) ErrorCode(code string) *ErrorBuilder {
	b.errorCode = code
	return b
}

func (b *ErrorBuilder) Tags(tag ...string) *ErrorBuilder {
	b.tags = append(b.tags, tag...)
	return b
}

func (b *ErrorBuilder) Errorf(ctx context.Context, format string, args ...any) error {
	return builder.Context(ctx).Meta(errorCode(b.errorCode), tag(b.tags...)).Newf(format, args...)
}

func (b *ErrorBuilder) Wrapf(ctx context.Context, err error, format string, args ...any) error {
	return builder.Context(ctx).Meta(errorCode(b.errorCode), tag(b.tags...)).Wrapf(err, format, args...)
}
