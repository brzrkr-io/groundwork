package session

import (
	"maps"
	"slices"
	"time"

	"github.com/pjanderson/groundwork/internal/progress"
)

// Stats summarises the learner's activity for the progress screen.
type Stats struct {
	// Streak counts consecutive active days up to today. A day without
	// practice yet doesn't break it until it's over, so the streak shown in
	// the morning is still yesterday's.
	Streak  int
	Longest int
	// Week counts active days among the last seven, today included.
	Week  int
	Today progress.Day
	// Days holds the activity of the last len(Days) days, oldest first,
	// ending today.
	Days []progress.Day

	Passed, Exercises  int
	GuidesRead, Guides int
	ReviewsDue         int
	ReviewsPassed      int
}

// active reports whether anything was practised that day.
func active(d progress.Day) bool { return d.Checks+d.Passed+d.Reviews+d.Guides > 0 }

// Stats computes the stats, with Days covering the last days days.
func (s *Session) Stats(days int) Stats {
	today := s.now().In(time.Local)
	day := func(back int) progress.Day {
		return s.State.Activity[progress.DateKey(today.AddDate(0, 0, -back))]
	}
	st := Stats{
		Today:      day(0),
		Passed:     s.DoneCount(),
		Exercises:  len(s.Course.Exercises),
		Guides:     len(s.Course.Guides),
		ReviewsDue: len(s.DueReviews()),
	}
	for _, g := range s.Course.Guides {
		if s.Read(g) {
			st.GuidesRead++
		}
	}
	for back := range 7 {
		if active(day(back)) {
			st.Week++
		}
	}
	start := 0
	if !active(st.Today) {
		start = 1 // today isn't over: the streak runs up to yesterday
	}
	for back := start; active(day(back)); back++ {
		st.Streak++
	}
	for i := days - 1; i >= 0; i-- {
		st.Days = append(st.Days, day(i))
	}

	// The longest streak and review total, over all recorded days.
	run := 0
	var prev time.Time
	for _, key := range slices.Sorted(maps.Keys(s.State.Activity)) {
		d := s.State.Activity[key]
		st.ReviewsPassed += d.Reviews
		if !active(d) {
			continue
		}
		t, err := time.ParseInLocation(time.DateOnly, key, time.Local)
		if err != nil {
			continue
		}
		if !prev.IsZero() && t.Equal(prev.AddDate(0, 0, 1)) {
			run++
		} else {
			run = 1
		}
		prev = t
		st.Longest = max(st.Longest, run)
	}
	return st
}
