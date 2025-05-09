package metaerr_test

import (
	"context"
	"errors"
	"testing"

	"github.com/quantumcycle/metaerr"
	"github.com/stretchr/testify/assert"
)

func TestBuilderNewWithMeta(t *testing.T) {
	a := assert.New(t)

	tagMeta := metaerr.StringMeta("tags")
	builder := metaerr.NewBuilder().Meta(tagMeta("security"), tagMeta("db"))

	err := builder.New("failure")

	a.Equal("failure [tags=db,security]", err.Error())
}

func TestBuilderNewWithContext(t *testing.T) {
	a := assert.New(t)

	ctxTagMeta := metaerr.StringMetaFromContext("tag", "tag")
	builder := metaerr.NewBuilder().Meta(ctxTagMeta())

	ctx := context.WithValue(context.Background(), "tag", "security")
	err := builder.Context(ctx).New("failure")

	a.Equal("failure [tag=security]", err.Error())
}

func TestBuilderNewfWithMeta(t *testing.T) {
	a := assert.New(t)

	tagMeta := metaerr.StringMeta("tags")
	builder := metaerr.NewBuilder().Meta(tagMeta("security"), tagMeta("db"))

	err := builder.Newf("failure %s", "test")

	a.Equal("failure test [tags=db,security]", err.Error())
}

func TestBuilderNewfWithContext(t *testing.T) {
	a := assert.New(t)

	ctxTagMeta := metaerr.StringMetaFromContext("tag", "tag")
	builder := metaerr.NewBuilder().Meta(ctxTagMeta())

	ctx := context.WithValue(context.Background(), "tag", "security")
	err := builder.Context(ctx).Newf("failure %s", "test")

	a.Equal("failure test [tag=security]", err.Error())
}

func TestBuilderWrapWithMeta(t *testing.T) {
	a := assert.New(t)

	tagMeta := metaerr.StringMeta("tags")
	builder := metaerr.NewBuilder().Meta(tagMeta("security")).Meta(tagMeta("db"))

	err := builder.Wrap(errors.New("failure"), "wrapped")

	a.Equal("wrapped [tags=db,security]: failure", err.Error())
}

func TestBuilderWrapWithoutErr(t *testing.T) {
	a := assert.New(t)

	tagMeta := metaerr.StringMeta("tags")
	builder := metaerr.NewBuilder().Meta(tagMeta("security"), tagMeta("db"))

	err := builder.Wrap(nil, "wrapped")

	a.Nil(err)
}

func TestBuilderWrapfWithMeta(t *testing.T) {
	a := assert.New(t)

	tagMeta := metaerr.StringMeta("tags")
	builder := metaerr.NewBuilder().Meta(tagMeta("security")).Meta(tagMeta("db"))

	err := builder.Wrapf(errors.New("failure"), "wrapped %s", "test")

	a.Equal("wrapped test [tags=db,security]: failure", err.Error())
}

func TestBuilderWrapfWithoutErr(t *testing.T) {
	a := assert.New(t)

	tagMeta := metaerr.StringMeta("tags")
	builder := metaerr.NewBuilder().Meta(tagMeta("security"), tagMeta("db"))

	err := builder.Wrapf(nil, "wrapped %s", "test")

	a.Nil(err)
}
