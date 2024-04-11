package logs

import (
	"github.com/rs/zerolog"
	"runtime"
	strconv "strconv"
	"strings"
)

func newCallerHook(nameModule, ignorePrefix string, callerSkipFrameCount int, ignorePrefixes []string) callerHook {
	return callerHook{
		nameModule:           nameModule,
		ignorePrefix:         ignorePrefix,
		ignorePrefixes:       ignorePrefixes,
		callerSkipFrameCount: callerSkipFrameCount,
	}
}

type callerHook struct {
	nameModule           string
	ignorePrefix         string
	ignorePrefixes       []string
	callerSkipFrameCount int
}

func (ch callerHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if e == nil {
		return
	}
	stack := NewStacktrace()
	var filterFrames []frame
	for _, v := range stack.Frames {
		if ch.checkIgnorePrefix(v.function) {
			continue
		}
		if strings.HasPrefix(v.function, ch.nameModule) {
			filterFrames = append(filterFrames, v)
		}
	}
	if len(filterFrames) == 0 {
		return
	}
	e.Str(zerolog.CallerFieldName, filterFrames[0].file+":"+strconv.Itoa(filterFrames[0].line))
}

func (ch callerHook) checkIgnorePrefix(function string) bool {
	if strings.HasPrefix(function, ch.ignorePrefix) {
		return true
	}
	for _, prefix := range ch.ignorePrefixes {
		if strings.HasPrefix(function, prefix) {
			return true
		}
	}
	return false
}

type frame struct {
	file     string
	line     int
	function string
}
type Stacktrace struct {
	Frames []frame
}

// NewStacktrace creates a stacktrace using runtime.Callers.
func NewStacktrace() *Stacktrace {
	pcs := make([]uintptr, 100)
	n := runtime.Callers(1, pcs)

	if n == 0 {
		return nil
	}

	frames := extractFrames(pcs[:n])

	stacktrace := Stacktrace{
		Frames: frames,
	}

	return &stacktrace
}

func extractFrames(pcs []uintptr) []frame {
	var frames = make([]frame, 0, len(pcs))
	callersFrames := runtime.CallersFrames(pcs)

	for {
		callerFrame, more := callersFrames.Next()

		frames = append(frames, frame{
			file:     callerFrame.File,
			line:     callerFrame.Line,
			function: callerFrame.Function,
		})

		if !more {
			break
		}
	}

	return frames
}
