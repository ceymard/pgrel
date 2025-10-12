package main

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/fatih/color"
	"gitlab.com/tozd/go/errors"
)

func printStackTrace(err error) {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	if err, ok := err.(stackTracer); ok {
		frames := runtime.CallersFrames(err.StackTrace())
		frameNum := 1
		for {
			if frame, ok := frames.Next(); !ok {
				break
			} else {
				fmt.Printf("%s:%s %s\n", color.CyanString(frame.File), color.GreenString(strconv.Itoa(frame.Line)), frame.Function)
				frameNum++
			}
		}
	}
}
