package stdlib_datetime

import (
	"fmt"
	"strings"
	"time"

	"github.com/bonkzero404/uddin-lang/interpreter"
)

type datetimeModule struct{}

func (m *datetimeModule) Name() string { return "datetime" }

func (m *datetimeModule) Functions() map[string]interpreter.ModuleFunc {
	return map[string]interpreter.ModuleFunc{
		"now":              nowFunc,
		"time_now":         timeNowFunc,
		"sleep":            sleepFunc,
		"format":           formatFunc,
		"parse":            parseFunc,
		"format_enhanced":  formatEnhancedFunc,
		"add":              addFunc,
		"subtract":         subtractFunc,
		"diff":             diffFunc,
		"between":          betweenFunc,
		"compare":          compareFunc,
	}
}

func init() {
	interpreter.RegisterModule(&datetimeModule{})
}

// nowFunc returns the current date and time in RFC3339 format.
func nowFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	return interpreter.Value(time.Now().Format(time.RFC3339))
}

// timeNowFunc returns the current Unix timestamp in milliseconds.
func timeNowFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	return interpreter.Value(int(time.Now().UnixMilli()))
}

// sleepFunc pauses execution for the specified number of milliseconds.
func sleepFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		panic(interpreter.TypeErrorf(pos, "datetime.sleep() requires exactly 1 argument, got %d", len(args)))
	}
	milliseconds, ok := args[0].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "datetime.sleep() requires an integer argument (milliseconds)"))
	}
	if milliseconds < 0 {
		panic(interpreter.TypeErrorf(pos, "datetime.sleep() requires a non-negative integer"))
	}
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
	return interpreter.Value(nil)
}

// formatFunc formats a date string according to a specified layout.
// Args: t (RFC3339 string), layout (with YYYY/MM/DD/hh/mm/ss placeholders)
func formatFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		panic(interpreter.TypeErrorf(pos, "datetime.format() requires exactly 2 arguments, got %d", len(args)))
	}
	t, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "datetime.format() requires first argument to be a string"))
	}
	layout, ok := args[1].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "datetime.format() requires second argument to be a string"))
	}
	replacer := strings.NewReplacer(
		"YYYY", "2006",
		"MM", "01",
		"DD", "02",
		"hh", "15",
		"mm", "04",
		"ss", "05",
		"ee", "Mon",
		"EE", "Monday",
		"nn", "Jan",
		"NN", "January",
	)
	parsed, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return interpreter.Value(nil)
	}
	layout = replacer.Replace(layout)
	return interpreter.Value(parsed.Format(layout))
}

// parseFunc parses a date string according to the specified format.
// Args: dateStr, format
func parseFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		return interpreter.Value(fmt.Errorf("datetime.parse() requires exactly 2 arguments, got %d", len(args)))
	}
	dateStr, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.parse() requires first argument to be a string"))
	}
	format, ok := args[1].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.parse() requires second argument to be a string"))
	}
	parsedTime, err := time.Parse(format, dateStr)
	if err != nil {
		return interpreter.Value(fmt.Errorf("datetime.parse() failed to parse date: %v", err))
	}
	return interpreter.Value(parsedTime.Format(time.RFC3339))
}

// formatEnhancedFunc formats a date string from one format to another.
// Args: dateStr, inputFormat, outputFormat
func formatEnhancedFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 3 {
		return interpreter.Value(fmt.Errorf("datetime.format_enhanced() requires exactly 3 arguments, got %d", len(args)))
	}
	dateStr, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.format_enhanced() requires first argument to be a string"))
	}
	inputFormat, ok := args[1].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.format_enhanced() requires second argument to be a string"))
	}
	outputFormat, ok := args[2].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.format_enhanced() requires third argument to be a string"))
	}
	parsedTime, err := time.Parse(inputFormat, dateStr)
	if err != nil {
		return interpreter.Value(fmt.Errorf("datetime.format_enhanced() failed to parse date: %v", err))
	}
	return interpreter.Value(parsedTime.Format(outputFormat))
}

// addFunc adds time duration to a date.
// Args: dateStr, durationStr OR dateStr, unit, amount
func addFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 {
		return interpreter.Value(fmt.Errorf("datetime.add() requires at least 2 arguments"))
	}
	dateStr, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.add(): first argument must be a date string"))
	}
	parsedTime, err := parseMultiFormatDate(dateStr)
	if err != nil {
		return interpreter.Value(fmt.Errorf("datetime.add(): failed to parse date: %v", err))
	}
	if len(args) == 2 {
		durationStr, ok := args[1].(string)
		if !ok {
			return interpreter.Value(fmt.Errorf("datetime.add(): second argument must be a duration string"))
		}
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return interpreter.Value(fmt.Errorf("datetime.add(): invalid duration format: %v", err))
		}
		return interpreter.Value(parsedTime.Add(duration).Format(time.RFC3339))
	} else if len(args) == 3 {
		unit, ok := args[1].(string)
		if !ok {
			return interpreter.Value(fmt.Errorf("datetime.add(): second argument must be a unit string"))
		}
		amount, err := toInt64(args[2])
		if err != nil {
			return interpreter.Value(fmt.Errorf("datetime.add(): third argument must be a number"))
		}
		result, err := applyDateUnit(parsedTime, unit, amount)
		if err != nil {
			return interpreter.Value(fmt.Errorf("datetime.add(): %v", err))
		}
		return interpreter.Value(result.Format(time.RFC3339))
	}
	return interpreter.Value(fmt.Errorf("datetime.add(): invalid number of arguments"))
}

// subtractFunc subtracts time duration from a date.
// Args: dateStr, durationStr OR dateStr, unit, amount
func subtractFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 {
		return interpreter.Value(fmt.Errorf("datetime.subtract() requires at least 2 arguments"))
	}
	dateStr, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.subtract(): first argument must be a date string"))
	}
	parsedTime, err := parseMultiFormatDate(dateStr)
	if err != nil {
		return interpreter.Value(fmt.Errorf("datetime.subtract(): failed to parse date: %v", err))
	}
	if len(args) == 2 {
		durationStr, ok := args[1].(string)
		if !ok {
			return interpreter.Value(fmt.Errorf("datetime.subtract(): second argument must be a duration string"))
		}
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return interpreter.Value(fmt.Errorf("datetime.subtract(): invalid duration format: %v", err))
		}
		return interpreter.Value(parsedTime.Add(-duration).Format(time.RFC3339))
	} else if len(args) == 3 {
		unit, ok := args[1].(string)
		if !ok {
			return interpreter.Value(fmt.Errorf("datetime.subtract(): second argument must be a unit string"))
		}
		amount, err := toInt64(args[2])
		if err != nil {
			return interpreter.Value(fmt.Errorf("datetime.subtract(): third argument must be a number"))
		}
		result, err := applyDateUnit(parsedTime, unit, -amount)
		if err != nil {
			return interpreter.Value(fmt.Errorf("datetime.subtract(): %v", err))
		}
		return interpreter.Value(result.Format(time.RFC3339))
	}
	return interpreter.Value(fmt.Errorf("datetime.subtract(): invalid number of arguments"))
}

// diffFunc calculates the difference between two dates.
// Args: date1, date2, unit? (optional, default seconds)
func diffFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 {
		return interpreter.Value(fmt.Errorf("datetime.diff() requires at least 2 arguments"))
	}
	date1Str, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.diff(): first argument must be a date string"))
	}
	date2Str, ok := args[1].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.diff(): second argument must be a date string"))
	}
	date1, err := parseMultiFormatDate(date1Str)
	if err != nil {
		return interpreter.Value(fmt.Errorf("datetime.diff(): failed to parse first date: %v", err))
	}
	date2, err := parseMultiFormatDate(date2Str)
	if err != nil {
		return interpreter.Value(fmt.Errorf("datetime.diff(): failed to parse second date: %v", err))
	}
	diff := date1.Sub(date2)
	if len(args) == 2 {
		return interpreter.Value(int64(diff.Seconds()))
	}
	unit, ok := args[2].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.diff(): third argument must be a unit string"))
	}
	switch strings.ToLower(unit) {
	case "nanoseconds", "nanosecond":
		return interpreter.Value(int64(diff.Nanoseconds()))
	case "microseconds", "microsecond":
		return interpreter.Value(int64(diff.Nanoseconds() / 1000))
	case "milliseconds", "millisecond":
		return interpreter.Value(int64(diff.Nanoseconds() / 1000000))
	case "seconds", "second":
		return interpreter.Value(int64(diff.Seconds()))
	case "minutes", "minute":
		return interpreter.Value(int64(diff.Minutes()))
	case "hours", "hour":
		return interpreter.Value(int64(diff.Hours()))
	case "days", "day":
		return interpreter.Value(int64(diff.Hours() / 24))
	default:
		return interpreter.Value(fmt.Errorf("datetime.diff(): unsupported unit '%s'", unit))
	}
}

// betweenFunc checks if a date is between two other dates (inclusive).
// Args: date, startDate, endDate
func betweenFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 3 {
		return interpreter.Value(fmt.Errorf("datetime.between() requires 3 arguments (date, startDate, endDate)"))
	}
	dateStr, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.between(): first argument must be a date string"))
	}
	startDateStr, ok := args[1].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.between(): second argument must be a date string"))
	}
	endDateStr, ok := args[2].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.between(): third argument must be a date string"))
	}
	date, err := parseMultiFormatDate(dateStr)
	if err != nil {
		return interpreter.Value(fmt.Errorf("datetime.between(): failed to parse date: %v", err))
	}
	startDate, err := parseMultiFormatDate(startDateStr)
	if err != nil {
		return interpreter.Value(fmt.Errorf("datetime.between(): failed to parse start date: %v", err))
	}
	endDate, err := parseMultiFormatDate(endDateStr)
	if err != nil {
		return interpreter.Value(fmt.Errorf("datetime.between(): failed to parse end date: %v", err))
	}
	isBetween := (date.Equal(startDate) || date.After(startDate)) &&
		(date.Equal(endDate) || date.Before(endDate))
	return interpreter.Value(isBetween)
}

// compareFunc compares two dates. Returns -1, 0, or 1.
func compareFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 {
		return interpreter.Value(fmt.Errorf("datetime.compare() requires 2 arguments (date1, date2)"))
	}
	date1Str, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.compare(): first argument must be a date string"))
	}
	date2Str, ok := args[1].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("datetime.compare(): second argument must be a date string"))
	}
	date1, err := parseMultiFormatDate(date1Str)
	if err != nil {
		return interpreter.Value(fmt.Errorf("datetime.compare(): failed to parse first date: %v", err))
	}
	date2, err := parseMultiFormatDate(date2Str)
	if err != nil {
		return interpreter.Value(fmt.Errorf("datetime.compare(): failed to parse second date: %v", err))
	}
	if date1.Before(date2) {
		return interpreter.Value(int64(-1))
	} else if date1.After(date2) {
		return interpreter.Value(int64(1))
	}
	return interpreter.Value(int64(0))
}

// parseMultiFormatDate tries multiple date formats and returns the first successful parse.
func parseMultiFormatDate(s string) (time.Time, error) {
	formats := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02", "2006-01-02 15:04:05"}
	var lastErr error
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

// toInt64 converts an interpreter Value to int64.
func toInt64(v interpreter.Value) (int64, error) {
	switch val := v.(type) {
	case int64:
		return val, nil
	case float64:
		return int64(val), nil
	case int:
		return int64(val), nil
	default:
		return 0, fmt.Errorf("expected a number, got %T", v)
	}
}

// applyDateUnit adds amount of unit to t.
func applyDateUnit(t time.Time, unit string, amount int64) (time.Time, error) {
	switch strings.ToLower(unit) {
	case "years", "year":
		return t.AddDate(int(amount), 0, 0), nil
	case "months", "month":
		return t.AddDate(0, int(amount), 0), nil
	case "days", "day":
		return t.AddDate(0, 0, int(amount)), nil
	case "hours", "hour":
		return t.Add(time.Duration(amount) * time.Hour), nil
	case "minutes", "minute":
		return t.Add(time.Duration(amount) * time.Minute), nil
	case "seconds", "second":
		return t.Add(time.Duration(amount) * time.Second), nil
	default:
		return t, fmt.Errorf("unsupported unit '%s'", unit)
	}
}
