package metaerr_test

import (
	"context"
	"errors"
	"github.com/quantumcycle/metaerr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuilderNewWithMeta(t *testing.T) {
	a := assert.New(t)

	tagMeta := metaerr.StringMeta("tags")
	builder := metaerr.NewBuilder().Meta(tagMeta("security"), tagMeta("db"))

	err := builder.Newf("failure")

	a.Equal("failure [tags=db,security]", err.Error())
}

func TestBuilderNewWithContext(t *testing.T) {
	a := assert.New(t)

	ctxTagMeta := metaerr.StringMetaFromContext("tag", "tag")
	builder := metaerr.NewBuilder().Meta(ctxTagMeta())

	ctx := context.WithValue(context.Background(), "tag", "security")
	err := builder.Context(ctx).Newf("failure")

	a.Equal("failure [tag=security]", err.Error())
}

func TestBuilderWrapWithMeta(t *testing.T) {
	a := assert.New(t)

	tagMeta := metaerr.StringMeta("tags")
	builder := metaerr.NewBuilder().Meta(tagMeta("security"), tagMeta("db"))

	err := builder.Wrapf(errors.New("failure"), "wrapped")

	a.Equal("wrapped [tags=db,security]: failure", err.Error())
}

func TestBuilderWrapWithoutErr(t *testing.T) {
	a := assert.New(t)

	tagMeta := metaerr.StringMeta("tags")
	builder := metaerr.NewBuilder().Meta(tagMeta("security"), tagMeta("db"))

	err := builder.Wrapf(nil, "wrapped")

	a.Nil(err)
}
