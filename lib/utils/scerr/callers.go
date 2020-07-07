package scerr

import "runtime"

// caller helpers

func GetCallerFunctionName() string {
	return getFrame(2).Function
}

func GetCallerFileLine() int {
	return getFrame(2).Line
}

func GetCallerFileName() string {
	return getFrame(2).File
}

// current function helpers

func GetCurrentFunctionName() string {
	return getFrame(1).Function
}

func GetCurrentFileLine() int {
	return getFrame(1).Line
}

func GetCurrentFileName() string {
	return getFrame(1).File
}

func getFrame(skipFrames uint) runtime.Frame {
	targetFrameIndex := skipFrames + 2 // skip 2 extra frames: runtime.Callers and getFrame

	pcs := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, pcs)

	frame := runtime.Frame{
		Function: "unknown",
	}

	if n > 0 {
		frames := runtime.CallersFrames(pcs[:n])
		for more, frameIndex := true, uint(0); more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}
