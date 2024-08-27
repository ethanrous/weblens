package internal

import (
	"fmt"
	"slices"
	"time"

	"github.com/ethrousseau/weblens/internal/log"
)

type lap struct {
	tag  string
	time time.Time
}

type Stopwatch interface {
	Lap(tag ...any)
	Stop() time.Duration
	PrintResults(firstLapIsStart bool)
	GetTotalTime(firstLapIsStart bool) time.Duration
}

type sw struct {
	name  string
	start time.Time
	laps  []lap
	stop  time.Time
}

type prod_sw struct {
	start time.Time
	stop  time.Time
}

func (sw prod_sw) Stop() (t time.Duration)         { sw.stop = time.Now(); return sw.stop.Sub(sw.start) }
func (sw prod_sw) Lap(tag ...any)                  {}
func (sw prod_sw) PrintResults(bool)               {}
func (sw prod_sw) GetTotalTime(bool) time.Duration { return sw.stop.Sub(sw.start) }

func NewStopwatch(name string) Stopwatch {
	if IsDevMode() {
		return &sw{name: name, start: time.Now()}
	}
	return prod_sw{start: time.Now()}
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
	// if l.tag != "" {
	// 	Debug.Println(l.tag)
	// }
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

func (s *sw) PrintResults(firstLapIsStart bool) {
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
		longest := len(slices.MaxFunc(s.laps, func(a, b lap) int { return len(a.tag) - len(b.tag) }).tag)
		lapFmt := fmt.Sprintf("\t%%-%ds %%-15s (%%s since start -- %%s since creation)", longest+5)
		for i, l := range s.laps {
			var sinceLast time.Duration
			if i != 0 {
				sinceLast = s.laps[i].time.Sub(s.laps[i-1].time)
			} else {
				sinceLast = s.laps[i].time.Sub(s.start)
			}

			if l.tag != "" {
				res = fmt.Sprintf(
					"%s\n%s", res, fmt.Sprintf(lapFmt, l.tag, sinceLast, l.time.Sub(startTime), l.time.Sub(s.start)),
				)
			}
		}
	}

	fmt.Printf("%s\n%s\n", res, fmt.Sprintf("Stopped at %s", s.stop.Sub(startTime)))
}