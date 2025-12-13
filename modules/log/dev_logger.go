package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const asyncLogger = true

type WLConsoleLogger struct {
	msgs chan []byte
}

func newDevLogger() io.Writer {
	l := WLConsoleLogger{
		msgs: make(chan []byte, 1000),
	}

	go func() {
		for msg := range l.msgs {
			_, _ = l.write_(msg)
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
	} else {
		return l.write_(p)
	}
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

func (l WLConsoleLogger) write_(p []byte) (n int, err error) {
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

	timeStr := time.Now().Format(time.TimeOnly + ".000")
	msg := fmt.Sprintf("[ %s %s%30.30s %s%5s%s ] %s %s%s%s%s\n", timeStr, BLUE, caller, levelColor, level, RESET, logMsg, RED, msgErr, stackStr, RESET)
	_, err = os.Stdout.Write([]byte(msg))

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
