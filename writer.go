package metaerr

import (
	"fmt"
	"io"
)

type errorWriter interface {
	Error(msg, metadata, location string, stacktrace *Stacktrace)
}

type stackErrorWriter struct {
	writer           io.Writer
	firstLinePrinted bool
}

func (ew *stackErrorWriter) Error(msg, metadata, location string, st *Stacktrace) {
	if msg == "" && metadata == "" && location == "" {
		return
	}
	if ew.firstLinePrinted {
		fmt.Fprint(ew.writer, "\n")
	}
	if msg != "" {
		fmt.Fprint(ew.writer, msg)
		ew.firstLinePrinted = true
	}

	if metadata != "" {
		if msg != "" {
			fmt.Fprint(ew.writer, " ")
		}
		fmt.Fprint(ew.writer, metadata)
		ew.firstLinePrinted = true
	}

	if location == "" {
		return
	}
	if ew.firstLinePrinted {
		fmt.Fprintf(ew.writer, "\n")
	}
	fmt.Fprintf(ew.writer, "\tat %s", location)
	ew.firstLinePrinted = true

	if st != nil && len(st.Frames) > 0 {
		fmt.Fprintf(ew.writer, "\n")
		for i, frame := range st.Frames {
			fmt.Fprintf(ew.writer, "\tat %s", frame.String())
			if i < len(st.Frames)-1 {
				fmt.Fprintf(ew.writer, "\n")
			}
		}
	}

}

type lineErrorWriter struct {
	writer            io.Writer
	firstErrorPrinted bool
}

func (ew *lineErrorWriter) Error(msg, metadata, location string, st *Stacktrace) {
	if msg == "" && metadata == "" {
		return
	}
	if ew.firstErrorPrinted {
		fmt.Fprint(ew.writer, ": ")
	}
	if msg != "" {
		fmt.Fprint(ew.writer, msg)
		ew.firstErrorPrinted = true
	}
	if metadata != "" {
		if msg != "" {
			fmt.Fprint(ew.writer, " ")
		}
		fmt.Fprint(ew.writer, metadata)
		ew.firstErrorPrinted = true
	}
}
