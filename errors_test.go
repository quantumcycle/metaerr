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

func SimulateWrapFromLibrary(err error, reason string) metaerr.Error {
	//We create the error in a function but we want to reported location to be here instead
	return libraryWrap(err, reason)
}

func libraryWrap(err error, reason string) metaerr.Error {
	return *metaerr.Wrap(err, reason, metaerr.WithLocationSkip(1))
}

const createErrorLocation = 14
const wrapErrorLocation = 25
const simulateCreateFromLibraryLocation = 30
const simulateWrapFromLibraryLocation = 39

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
	s := fmt.Sprintf("%+v\n", err)
	fmt.Println(s)
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
