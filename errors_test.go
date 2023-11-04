package metaerr_test

import (
	"context"
	stderr "errors"
	"fmt"
	"testing"

	"github.com/quantumcycle/metaerr"
	"github.com/stretchr/testify/assert"
)

// We put these helper here to make sure the line number for all the tests are constant
func CreateError(reason string, meta map[string][]string) error {
	metas := make([]metaerr.ErrorMetadata, 0, len(meta))
	for k, values := range meta {
		for _, val := range values {
			metas = append(metas, metaerr.StringMeta(k)(val))
		}

	}

	return metaerr.New(reason, metaerr.WithMeta(metas...))
}

func Wrap(err error, reason string, metas ...metaerr.ErrorMetadata) error {
	return metaerr.Wrap(err, reason, metaerr.WithMeta(metas...))
}

func SimulateCreateFromLibrary(reason string) error {
	//We create the error in a function but we want to reported location to be here instead
	return libraryCreateNew(reason)
}

func libraryCreateNew(reason string) error {
	return metaerr.New(reason, metaerr.WithLocationSkip(1))
}

func SimulateCreateFromLibraryWithStack(reason string) error {
	//We create the error in a function but we want to reported location to be here instead
	return libraryCreateNewWithStack(reason)
}

func libraryCreateNewWithStack(reason string) error {
	return metaerr.New(reason, metaerr.WithLocationSkip(1), metaerr.WithStackTrace(1, 3))
}

func SimulateWrapFromLibrary(err error, reason string) error {
	//We create the error in a function but we want to reported location to be here instead
	return libraryWrap(err, reason)
}

func libraryWrap(err error, reason string) error {
	return metaerr.Wrap(err, reason, metaerr.WithLocationSkip(1))
}

// Same as SimulateCreateFromLibraryWithStack, but with an additional stack frame we want to assert on
func SimulateCreateFromLibraryWithStackLevel2(reason string) error {
	return SimulateCreateFromLibraryWithStack(reason)
}

const createErrorLocation = 23
const wrapErrorLocation = 27
const simulateCreateFromLibraryLocation = 32
const simulateCreateFromLibraryWithStackLocation = 41
const simulateCreateFromLibraryWithStackLevel2Location = 59

func TestFormatWithoutMeta(t *testing.T) {
	a := assert.New(t)

	err := CreateError("failure", nil)

	a.Equal("failure", err.Error())
	a.Regexp(fmt.Sprintf(`failure
\s+at.+/metaerr/errors_test.go:%d
`, createErrorLocation),
		fmt.Sprintf("%+v\n", err))
}

func TestFormatWithMeta(t *testing.T) {
	a := assert.New(t)

	err := CreateError("failure", map[string][]string{
		"errorCode": {"code2", "code1"},
		"tag":       {"not_found"},
	})

	a.Equal("failure [errorCode=code1,code2] [tag=not_found]", err.Error())
	a.Regexp(fmt.Sprintf(`failure \[errorCode=code1,code2\] \[tag=not_found\]
\s+at.+/metaerr/errors_test.go:%d
`, createErrorLocation),
		fmt.Sprintf("%+v\n", err))
}

func TestFormatWithWrappedMetaError(t *testing.T) {
	a := assert.New(t)

	err := CreateError("failure", nil)
	wrapped := Wrap(err, "wrapped")

	a.Equal("wrapped: failure", wrapped.Error())
	a.Regexp(fmt.Sprintf(`wrapped
\s+at.+/metaerr/errors_test.go:%d
failure
\s+at.+/metaerr/errors_test.go:%d
`, wrapErrorLocation, createErrorLocation),
		fmt.Sprintf("%+v\n", wrapped))
}

func TestFormatWithWrappedStdError(t *testing.T) {
	a := assert.New(t)

	err := stderr.New("failure")
	wrapped := Wrap(err, "wrapped")

	a.Equal("wrapped: failure", wrapped.Error())
	a.Regexp(fmt.Sprintf(`wrapped
\s+at.+/metaerr/errors_test.go:%d
failure
`, wrapErrorLocation),
		fmt.Sprintf("%+v\n", wrapped))
}

func TestFormatLocationWhenCreatedFromLibraryOrHelperFunction(t *testing.T) {
	a := assert.New(t)

	err := SimulateCreateFromLibrary("failure")

	a.Equal("failure", err.Error())
	a.Regexp(fmt.Sprintf(`failure
\s+at.+/metaerr/errors_test.go:%d
`, simulateCreateFromLibraryLocation),
		fmt.Sprintf("%+v\n", err))
}

func TestFormatLocationWhenWrappedFromLibraryOrHelperFunction(t *testing.T) {
	a := assert.New(t)

	err := stderr.New("failure")
	wrapped := SimulateWrapFromLibrary(err, "unknown failure")

	a.Equal("unknown failure: failure", wrapped.Error())
}

func TestFormatAsStringOnlyDisplayError(t *testing.T) {
	a := assert.New(t)

	err := CreateError("failure", nil)

	a.Equal("failure", err.Error())
	a.Equal("failure", fmt.Sprintf("%s", err))
}

func TestFormatWithoutMessage(t *testing.T) {
	a := assert.New(t)

	err := CreateError("", nil)

	a.Equal("", err.Error())
	a.Equal("", fmt.Sprintf("%s", err))
	a.Regexp(fmt.Sprintf(`\s+at.+/metaerr/errors_test.go:%d
`, createErrorLocation),
		fmt.Sprintf("%+v\n", err))
}

func TestFormatMultipleWrappedWithoutMessage(t *testing.T) {
	a := assert.New(t)

	err := CreateError("", nil)
	err2 := Wrap(err, "")
	err3 := Wrap(err2, "root")

	a.Regexp(fmt.Sprintf(`root
\s+at.+/metaerr/errors_test.go:%d
\s+at.+/metaerr/errors_test.go:%d
\s+at.+/metaerr/errors_test.go:%d
`, wrapErrorLocation, wrapErrorLocation, createErrorLocation),
		fmt.Sprintf("%+v\n", err3))

}

func TestGetMetaReturnsMergeMetaFromWrappedErrors(t *testing.T) {
	a := assert.New(t)

	err := CreateError("failure", map[string][]string{
		"errorCode": {"code2", "code1"},
		"tag":       {"not_found"},
	})
	wrapped := Wrap(err, "wrapped", metaerr.StringMeta("errorCode")("code3"))

	meta := metaerr.GetMeta(wrapped, true)

	a.Equal(map[string][]string{
		"errorCode": {"code1", "code2", "code3"},
		"tag":       {"not_found"},
	}, meta)
}

func TestGetMetaReturnsNonNestedMeta(t *testing.T) {
	a := assert.New(t)

	err := CreateError("failure", map[string][]string{
		"errorCode": {"code2", "code1"},
		"tag":       {"not_found"},
	})
	wrapped := Wrap(err, "wrapped", metaerr.StringMeta("errorCode")("code3"))

	meta := metaerr.GetMeta(wrapped, false)

	a.Equal(map[string][]string{
		"errorCode": {"code3"},
	}, meta)
}

func TestWrappingNilErrorReturnsNil(t *testing.T) {
	a := assert.New(t)

	err := Wrap(nil, "wrapped")

	a.Nil(err)
}

func TestGetLocation(t *testing.T) {
	a := assert.New(t)

	err := CreateError("", nil)

	merr, ok := metaerr.AsMetaError(err)
	a.True(ok)
	a.Regexp(fmt.Sprintf(`.+/metaerr/errors_test.go:%d`, createErrorLocation), merr.Location)
}

type MyMetaValue string

func (m MyMetaValue) String() string { return string(m) }

var MetaValue1 MyMetaValue = "value1"

func TestStringerMeta(t *testing.T) {
	a := assert.New(t)

	meta := metaerr.StringerMeta[MyMetaValue]("mymeta")
	err := metaerr.New("failure", metaerr.WithMeta(meta(MetaValue1)))

	errMetaValues := metaerr.GetMeta(err, false)

	a.Equal(map[string][]string{
		"mymeta": {"value1"},
	}, errMetaValues)

}

func TestStringerMetaWithEmptyValue(t *testing.T) {
	a := assert.New(t)

	meta := metaerr.StringerMeta[MyMetaValue]("mymeta")
	err := metaerr.New("failure", metaerr.WithMeta(meta("")))

	errMetaValues := metaerr.GetMeta(err, false)

	a.Equal(map[string][]string{}, errMetaValues)

}

func TestStringsMeta(t *testing.T) {
	a := assert.New(t)

	meta := metaerr.StringsMeta("mymeta")
	err := metaerr.New("failure", metaerr.WithMeta(meta("v1", "v2")))

	errMetaValues := metaerr.GetMeta(err, false)

	a.Equal(map[string][]string{
		"mymeta": {"v1", "v2"},
	}, errMetaValues)

}

func TestStringsMetaWithoutValues(t *testing.T) {
	a := assert.New(t)

	meta := metaerr.StringsMeta("mymeta")
	err := metaerr.New("failure", metaerr.WithMeta(meta()))

	errMetaValues := metaerr.GetMeta(err, false)

	a.Equal(map[string][]string{}, errMetaValues)

}

func TestStandardWrapping(t *testing.T) {
	a := assert.New(t)

	err := metaerr.New("failure")
	wrapped := fmt.Errorf("wrapped: %w", err)

	a.Equal("wrapped: failure", wrapped.Error())
}

func TestErrorStringIncludesMetadata(t *testing.T) {
	a := assert.New(t)

	err := CreateError("failure", map[string][]string{
		"errorCode": {"code2", "code1"},
		"tag":       {"not_found"},
	})

	a.Equal("failure [errorCode=code1,code2] [tag=not_found]", err.Error())
}

func TestFormatWrappedEmptyError(t *testing.T) {
	a := assert.New(t)

	err := stderr.New("")
	wrapped := Wrap(err, "something went wrong")

	a.Regexp(fmt.Sprintf(`something went wrong
\s+at.+/metaerr/errors_test.go:%d
`, wrapErrorLocation),
		fmt.Sprintf("%+v\n", wrapped))

}

func TestErrorWithStacktrace(t *testing.T) {
	a := assert.New(t)

	err := SimulateCreateFromLibraryWithStackLevel2("failure")

	a.Regexp(fmt.Sprintf(`failure
\s+at.+/metaerr/errors_test.go:%d
\s+.+/metaerr/errors_test.go:%d
\s+.+/metaerr/errors_test.go:.*
`,
		simulateCreateFromLibraryWithStackLocation,
		simulateCreateFromLibraryWithStackLevel2Location),
		fmt.Sprintf("%+v\n", err))
}

func TestErrorWithMetaFromContextWithValue(t *testing.T) {
	a := assert.New(t)

	ctx := context.WithValue(context.Background(), "user", "123")
	userFromCtx := metaerr.StringMetaFromContext("user", "user")
	err := metaerr.New("failure", metaerr.WithContext(ctx), metaerr.WithMeta(userFromCtx()))

	meta := metaerr.GetMeta(err, false)

	a.Equal(map[string][]string{
		"user": {"123"},
	}, meta)
}

func TestErrorWithMetaFromContextWithoutValue(t *testing.T) {
	a := assert.New(t)

	ctx := context.Background()
	userFromCtx := metaerr.StringMetaFromContext("user", "user")
	err := metaerr.New("failure", metaerr.WithContext(ctx), metaerr.WithMeta(userFromCtx()))

	meta := metaerr.GetMeta(err, false)

	a.Equal(map[string][]string{}, meta)
}

func TestErrorWithMetaFromContextWithoutContext(t *testing.T) {
	a := assert.New(t)

	userFromCtx := metaerr.StringMetaFromContext("user", "user")
	err := metaerr.New("failure", metaerr.WithMeta(userFromCtx()))

	meta := metaerr.GetMeta(err, false)

	a.Equal(map[string][]string{}, meta)
}
