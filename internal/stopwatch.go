package internal

import (
	"fmt"
	"slices"
	"time"

	"github.com/ethanrous/weblens/internal/log"
)

type lap struct {
	time time.Time
	tag  string
}

type Stopwatch interface {
	Lap(tag ...any)
	Stop() time.Duration
	PrintResults(firstLapIsStart bool)
	GetTotalTime(firstLapIsStart bool) time.Duration
}

type sw struct {
	start time.Time
	stop  time.Time
	name  string
	laps  []lap
}

func NewStopwatch(name string) Stopwatch {
	return &sw{name: name, start: time.Now()}
}

func (s *sw) Stop() time.Duration {
	s.stop = time.Now()
	return s.stop.Sub(s.start)
}

func (s *sw) Lap(tag ...any) {
	l := lap{
		tag:  fmt.Sprint(tag...),
		time: time.Now(),
	}
	s.laps = append(s.laps, l)
}

func (s *sw) GetTotalTime(firstLapIsStart bool) time.Duration {
	var start time.Time
	var end time.Time

	if s.stop.Unix() < 0 {
		end = time.Now()
	} else {
		end = s.stop
	}

	if firstLapIsStart && len(s.laps) > 0 {
		start = s.laps[0].time
	} else {
		start = s.start
	}

	return end.Sub(start)
}

var (
	red    = "\u001b[31m"
	green  = "\u001b[34m"
	orange = "\u001b[36m"
	reset  = "\u001B[0m"
)

func (s *sw) PrintResults(firstLapIsStart bool) {
	if log.GetLogLevel() != log.TRACE {
		return
	}

	if s.stop.Unix() < 0 {
		log.Error.Println("Stopwatch cannot provide results before being stopped")
		return
	}

	var res = fmt.Sprintf("--- %s Stopwatch ---", s.name)

	var startTime time.Time
	if firstLapIsStart {
		if len(s.laps) <= 1 {
			return
		}
		startTime = s.laps[0].time
	} else {
		startTime = s.start
	}

	if len(s.laps) != 0 {
		longestNameLen := len(slices.MaxFunc(s.laps, func(a, b lap) int { return len(a.tag) - len(b.tag) }).tag)
		lapFmt := fmt.Sprintf("\t%%-%ds %%-15s (%%s since start -- %%s since creation)", longestNameLen+5)

		var lapTimes []time.Duration
		longestLap := time.Duration(0)

		for i := range s.laps {
			var sinceLast time.Duration
			if i != 0 {
				sinceLast = s.laps[i].time.Sub(s.laps[i-1].time)
			} else {
				sinceLast = s.laps[i].time.Sub(s.start)
			}
			lapTimes = append(lapTimes, sinceLast)
			if sinceLast > longestLap {
				longestLap = sinceLast
			}
		}

		for i, sinceLast := range lapTimes {
			l := s.laps[i]
			if l.tag != "" {
				color := ""
				timeStr := sinceLast.String()
				if float64(sinceLast) > float64(longestLap)*0.75 {
					color = red
				} else if float64(sinceLast) > float64(longestLap)*0.50 {
					color = orange
				} else {
					color = green
				}
				res = fmt.Sprintf(
					"%s%s\n%s%s", res, color, fmt.Sprintf(lapFmt, l.tag, timeStr, l.time.Sub(startTime), l.time.Sub(s.start)), reset,
				)
			}
		}
	}

	fmt.Printf("%s\n%s\n", res, fmt.Sprintf("Stopped at %s", s.stop.Sub(startTime)))
}
