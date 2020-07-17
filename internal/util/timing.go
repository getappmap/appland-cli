package util

import (
	"fmt"
	"time"
)

var currentTiming Timing

// Timing allows measuring time spent in code. Each Start() starts a new
// step and returns a Timing which can be used in turn to measure substeps.
// Last started Timing is also available to timing-ignorant code through util.Time().
type Timing interface {
	Start(name string) Timing
	Finish()
	Print()
}

type timing struct {
	startTime time.Time
	duration  time.Duration
	name      string
	steps     []*timing
}

func (t *timing) Start(name string) Timing {
	t.finishCurrentStep()
	nt := newTiming(name)
	t.steps = append(t.steps, nt)
	currentTiming = nt
	return nt
}

func (t *timing) finishCurrentStep() {
	if l := len(t.steps); l > 0 {
		t.steps[l-1].Finish()
	}
}

func (t *timing) Finish() {
	if t.duration == 0 {
		t.finishCurrentStep()
		t.duration = time.Since(t.startTime)
		if currentTiming == t {
			currentTiming = nil
		}
	}
}

// Time allows code which doesn't want to deal with timing explicitly
// to nonetheless allow timing events if the calling code requests.
// It uses a hidden global reference to the last explicitly Start()ed Timing.
func Time(name string) {
	if t := currentTiming; t != nil {
		t.Start(name)
		currentTiming = t
	}
}

func newTiming(name string) *timing {
	t := timing{}
	t.startTime = time.Now()
	t.name = name
	return &t
}

// NewTiming makes a new top-level Timing.
func NewTiming(name string) Timing {
	t := newTiming(name)
	return t
}

func (t *timing) Print() {
	t.print("")
}

func (t *timing) print(prefix string) {
	t.Finish()
	fmt.Printf("%s%s: %s\n", prefix, t.name, t.duration)
	for _, step := range t.steps {
		step.print(prefix + "  ")
	}
}
