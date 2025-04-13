package errors

import (
	"context"
	"fmt"
	"github.com/quantumcycle/metaerr"
)

var errorCode = metaerr.StringMeta("error_code")
var tag = metaerr.StringsMeta("tags")
var userID = metaerr.StringMetaFromContext("user", "user_id")
var builder = metaerr.NewBuilder(
	metaerr.WithStackTrace(0, 2)).
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

func buildMeta(errCode string, tags []string, metaKeyValues ...any) []metaerr.ErrorMetadata {
	metas := make([]metaerr.ErrorMetadata, 0, len(metaKeyValues)+2)
	if errCode != "" {
		metas = append(metas, errorCode(errCode))
	}
	if len(tags) > 0 {
		metas = append(metas, tag(tags...))
	}
	if len(metaKeyValues) > 0 {
		//make sure we have an even number of key/values
		if len(metaKeyValues)%2 != 0 {
			metaKeyValues = append(metaKeyValues, "")
		}
		for i := 0; i < len(metaKeyValues); i += 2 {
			key := fmt.Sprintf("%s", metaKeyValues[i])
			value := fmt.Sprintf("%s", metaKeyValues[i+1])
			metas = append(metas, metaerr.StringMeta(key)(value))
		}
	}
	return metas
}

func (b *ErrorBuilder) Error(ctx context.Context, msg string, metaKeyValues ...any) error {
	return builder.Context(ctx).Meta(buildMeta(b.errorCode, b.tags, metaKeyValues...)...).Newf(msg)
}

func (b *ErrorBuilder) Wrap(ctx context.Context, err error, msg string, metaKeyValues ...any) error {
	return builder.Context(ctx).Meta(buildMeta(b.errorCode, b.tags, metaKeyValues...)...).Wrapf(err, msg)
}

func (b *ErrorBuilder) Errorf(ctx context.Context, format string, args ...any) error {
	return builder.Context(ctx).Meta(errorCode(b.errorCode), tag(b.tags...)).Newf(format, args...)
}

func (b *ErrorBuilder) Wrapf(ctx context.Context, err error, format string, args ...any) error {
	return builder.Context(ctx).Meta(errorCode(b.errorCode), tag(b.tags...)).Wrapf(err, format, args...)
}
