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
		if !ok {
			break
		}
		//Skip frames from the standard library
		if strings.Contains(file, runtime.GOROOT()) {
			continue
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
