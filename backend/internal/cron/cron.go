package cron

import (
	"time"

	"github.com/robfig/cron/v3"
)

// NextRun parses a cron expression and returns the next run time after the given time.
// Returns an error if the expression is invalid.
func NextRun(expression string, after time.Time) (time.Time, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(expression)
	if err != nil {
		return time.Time{}, err
	}
	return schedule.Next(after), nil
}
