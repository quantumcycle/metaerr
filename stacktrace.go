package metaerr

import (
	"fmt"
	"runtime"
	"strings"
)

type stacktrace struct {
	frames []frame
}

type frame struct {
	file string
	line int
}

func (frame *frame) String() string {
	return fmt.Sprintf("%v:%v", frame.file, frame.line)
}

func newStacktrace(frameStackSkip, maxDepth int) *stacktrace {
	var frames []frame

	// We loop until we have StackTraceMaxDepth frames or we run out of frames.
	// Frames from this package are skipped.
	for i := frameStackSkip; len(frames) < maxDepth; i++ {
		_, file, line, ok := runtime.Caller(i)
		//Once we find a frame in the stdlib, we stop, since stdlib code won't call back to user code
		if !ok || strings.Contains(file, runtime.GOROOT()) {
			break
		}

		frames = append(frames, frame{
			file: file,
			line: line,
		})
	}

	return &stacktrace{
		frames: frames,
	}
}
