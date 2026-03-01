package wlog

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

const asyncLogger = true

// WLConsoleLogger is a development console logger with colored output.
type WLConsoleLogger struct {
	msgs    chan []byte
	writeTo io.Writer
}

func newDevLogger(writeTo io.Writer) io.Writer {
	l := WLConsoleLogger{
		msgs:    make(chan []byte, 1000),
		writeTo: writeTo,
	}

	go func() {
		for msg := range l.msgs {
			_, _ = l.write(msg)
		}
	}()

	return l
}

func (l WLConsoleLogger) Write(p []byte) (n int, err error) {
	if asyncLogger {
		// copy so orig buffer (p) can be reused
		cpyP := make([]byte, len(p))
		copy(cpyP, p)

		l.msgs <- cpyP

		return len(p), nil
	}

	return l.write(p)
}

func formatStack(traceback []any) string {
	stackStr := "\n"

	for _, block := range traceback {
		blockMap, ok := block.(map[string]any)

		var line string

		var function string

		var source string

		var okSource bool

		if ok {
			function, _ = blockMap["func"].(string)
			line, _ = blockMap["line"].(string)

			source, okSource = blockMap["source"].(string)
		} else {
			blockMap, ok := block.(map[string]string)
			if !ok {
				continue
			}

			function = blockMap["func"]
			line = blockMap["line"]
			source, okSource = blockMap["source"]
		}

		if strings.HasSuffix(function, "ServeHTTP") {
			break
		}

		if okSource {
			var found bool
			if source, found = strings.CutPrefix(source, projectPrefix); !found {
				source = ".../" + filepath.Base(source)
			}

			fileAndLine := source + ":" + line
			function += "()"
			stackStr += fmt.Sprintf("\t%s%-40s %s%30s\n", BLUE, fileAndLine, RESET, function)
		}
	}

	return stackStr
}

func (l WLConsoleLogger) write(p []byte) (n int, err error) {
	defer recoverLogger()

	var target map[string]any

	err = json.Unmarshal(p, &target)
	if err != nil {
		return n, err
	}

	callerI, ok := target["caller"]
	caller := "[NO CALLER]"

	if ok {
		caller = callerI.(string)
		caller = strings.TrimPrefix(caller, projectPrefix)

		if len(caller) > 30 {
			caller = "..." + caller[len(caller)-27:]
		}
	}

	level, _ := target["level"].(string)
	msgErr, _ := target["error"].(string)
	logMsg, _ := target["message"].(string)
	traceback, _ := target["traceback"].([]any)

	stackStr := ""
	if len(traceback) != 0 {
		stackStr = formatStack(traceback)
	}

	levelColor := ""

	switch level {
	case "trace":
		levelColor = RESET
	case "debug":
		levelColor = ORANGE
	case "info":
		levelColor = BLUE
	case "warn":
		levelColor = YELLOW
	case "error", "fatal":
		levelColor = RED
	}

	extras := ""

	for key, val := range target {
		switch key {
		case "caller", "level", "error", "message", "traceback", "time", "weblens_build_version", "@timestamp", "referer", "ip", "requester", "req_id":
			continue
		default:
			if extras == "" {
				extras = "["
			} else {
				extras += ", "
			}

			extras += fmt.Sprintf("%s=%v", key, val)
		}
	}

	if extras != "" {
		extras += "]"
	}

	timeStr := time.Now().Format(time.TimeOnly + ".000")
	msg := fmt.Sprintf("[ %s %s%30.30s %s%5s%s ]%s %s %s%s%s%s\n", timeStr, BLUE, caller, levelColor, level, RESET, extras, logMsg, RED, msgErr, stackStr, RESET)

	_, err = l.writeTo.Write([]byte(msg))
	if err != nil {
		return n, err
	}

	return len(p), nil
}

func recoverLogger() {
	if r := recover(); r != nil {
		fmt.Println("Panic in async logger:", r)
	}
}
