package metaerr_test

import (
	"testing"

	"github.com/quantumcycle/metaerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureWith(detector func(string) bool) *metaerr.Stacktrace {
	err := metaerr.New("boom",
		metaerr.WithRootPackageDetector(detector),
		metaerr.WithStackTrace(0, 10),
	)
	me, _ := metaerr.AsMetaError(err)
	return me.Stacktrace
}

func TestWithRootPackageDetectorControlsCapture(t *testing.T) {
	// Treating every package as a root: the walk stops at the second frame, so
	// the keep-first guarantee leaves exactly the one captured call-site frame.
	allRoot := captureWith(func(string) bool { return true })
	require.NotNil(t, allRoot)
	assert.Len(t, allRoot.Frames, 1,
		"an all-root detector should yield only the kept first frame")

	// Treating nothing as a root: the walk continues past the first frame (into
	// the test runner / runtime) up to maxDepth.
	noneRoot := captureWith(func(string) bool { return false })
	require.NotNil(t, noneRoot)
	assert.Greater(t, len(noneRoot.Frames), 1,
		"a no-root detector should keep walking past the first frame")
}

// TestDefaultRootPackageComposable documents that the default classifier is
// exported so a custom detector can delegate to it for everything it doesn't
// special-case (the common pattern for a domain-less module).
func TestDefaultRootPackageComposable(t *testing.T) {
	assert.True(t, metaerr.DefaultRootPackage("net/http"))
	assert.False(t, metaerr.DefaultRootPackage("github.com/x/y"))
	assert.False(t, metaerr.DefaultRootPackage("main"))
}
