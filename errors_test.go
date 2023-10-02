package metaerr_test

import (
	stderr "errors"
	"fmt"
	"testing"

	"github.com/quantumcycle/metaerr"
	"github.com/stretchr/testify/assert"
)

// We put these helper here to make sure the line number for all the tests are constant
func CreateError(reason string, meta map[string][]string) metaerr.Error {
	err := metaerr.New(reason)
	for k, values := range meta {
		for _, val := range values {
			err = err.Meta(metaerr.StringMeta(k)(val))
		}

	}
	return err
}

func Wrap(err error, reason string) *metaerr.Error {
	return metaerr.Wrap(err, reason)
}

func SimulateCreateFromLibrary(reason string) metaerr.Error {
	//We create the error in a function but we want to reported location to be here instead
	return libraryCreateNew(reason)
}

func libraryCreateNew(reason string) metaerr.Error {
	return metaerr.New(reason, metaerr.WithLocationSkip(1))
}

const createErrorLocation = 14
const wrapErrorLocation = 25
const simulateCreateFromLibraryLocation = 30

func TestFormatWithoutMeta(t *testing.T) {
	a := assert.New(t)

	err := CreateError("failure", nil)

	a.Equal("failure", err.Error())
	a.Regexp(fmt.Sprintf(`failure
\s+at.+/metaerr/errors_test.go:%d
`, createErrorLocation),
		fmt.Sprintf("%+v", err))
}

func TestFormatWithMeta(t *testing.T) {
	a := assert.New(t)

	err := CreateError("failure", map[string][]string{
		"errorCode": {"code2", "code1"},
		"tag":       {"not_found"},
	})

	a.Equal("failure", err.Error())
	a.Regexp(fmt.Sprintf(`failure \[errorCode=code1, code2\] \[tag=not_found\]
\s+at.+/metaerr/errors_test.go:%d
`, createErrorLocation),
		fmt.Sprintf("%+v", err))
}

func TestFormatWithWrappedMetaError(t *testing.T) {
	a := assert.New(t)

	err := CreateError("failure", nil)
	wrapped := Wrap(err, "wrapped")

	a.Equal("wrapped", wrapped.Error())
	a.Regexp(fmt.Sprintf(`wrapped
\s+at.+/metaerr/errors_test.go:%d
failure
\s+at.+/metaerr/errors_test.go:%d
`, wrapErrorLocation, createErrorLocation),
		fmt.Sprintf("%+v", wrapped))
}

func TestFormatWithWrappedStdError(t *testing.T) {
	a := assert.New(t)

	err := stderr.New("failure")
	wrapped := Wrap(err, "wrapped")

	a.Equal("wrapped", wrapped.Error())
	a.Regexp(fmt.Sprintf(`wrapped
\s+at.+/metaerr/errors_test.go:%d
failure
\s+at.+\[no location\]
`, wrapErrorLocation),
		fmt.Sprintf("%+v", wrapped))
}

func TestFormatLocationWhenCreatedFromLibraryOrHelperFunction(t *testing.T) {
	a := assert.New(t)

	err := SimulateCreateFromLibrary("failure")

	a.Equal("failure", err.Error())
	a.Regexp(fmt.Sprintf(`failure
\s+at.+/metaerr/errors_test.go:%d
`, simulateCreateFromLibraryLocation),
		fmt.Sprintf("%+v", err))
}

func TestFormatAsStringOnlyDisplayError(t *testing.T) {
	a := assert.New(t)

	err := CreateError("failure", nil)

	a.Equal("failure", err.Error())
	a.Equal("failure", fmt.Sprintf("%s", err))
}

func TestFormatDefaultToNoMessageWhenNoReasonProvided(t *testing.T) {
	a := assert.New(t)

	err := CreateError("", nil)

	a.Equal("", err.Error())
	a.Equal("", fmt.Sprintf("%s", err))
	a.Regexp(fmt.Sprintf(`\[no message\]
\s+at.+/metaerr/errors_test.go:%d
`, createErrorLocation),
		fmt.Sprintf("%+v", err))
}

func TestGetMetaReturnsMergeMetaFromWrappedErrors(t *testing.T) {
	a := assert.New(t)

	err := CreateError("failure", map[string][]string{
		"errorCode": {"code2", "code1"},
		"tag":       {"not_found"},
	})
	wrapped := Wrap(err, "wrapped")
	wrappedWithMeta := wrapped.Meta(metaerr.StringMeta("errorCode")("code3"))

	meta := metaerr.GetMeta(wrappedWithMeta, true)

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
	wrapped := Wrap(err, "wrapped")
	wrappedWithMeta := wrapped.Meta(metaerr.StringMeta("errorCode")("code3"))

	meta := metaerr.GetMeta(wrappedWithMeta, false)

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

	a.Regexp(fmt.Sprintf(`.+/metaerr/errors_test.go:%d`, createErrorLocation), err.Location())
}

type MyMetaValue string

func (m MyMetaValue) String() string { return string(m) }

var MetaValue1 MyMetaValue = "value1"

func TestStringerMeta(t *testing.T) {
	a := assert.New(t)

	meta := metaerr.StringerMeta[MyMetaValue]("mymeta")
	err := metaerr.New("failure").Meta(meta(MetaValue1))

	errMetaValues := metaerr.GetMeta(err, false)

	a.Equal(map[string][]string{
		"mymeta": {"value1"},
	}, errMetaValues)

}
