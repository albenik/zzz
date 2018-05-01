package main

import (
	"fmt"
	"io"
	"math"
	"sort"
	"time"

	"github.com/buger/jsonparser"
	"github.com/fatih/color"
)

type entry struct {
	level   string
	message string
	caller  string
	ts      string
	trace   string
	fields  map[string]interface{}
}

func (e *entry) parse(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
	var err error

	switch string(key) {
	case "level":
		if e.level, err = jsonparser.ParseString(value); err != nil {
			return err
		}
	case "msg":
		if e.message, err = jsonparser.ParseString(value); err != nil {
			return err
		}
	case "ts":
		ts, err := jsonparser.ParseFloat(value)
		if err != nil {
			return err
		}
		sec, dec := math.Modf(ts)
		e.ts = time.Unix(int64(sec), int64(dec*(1e9))).Format("2006-01-02T15:04:05.000Z07:00")
	case "caller":
		if e.caller, err = jsonparser.ParseString(value); err != nil {
			return err
		}
	case "stacktrace":
		if e.trace, err = jsonparser.ParseString(value); err != nil {
			return err
		}
	default:
		if e.fields == nil {
			e.fields = make(map[string]interface{})
		}

		switch dataType {

		case jsonparser.Array:
			items := make([]string, 0, 1)
			jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				items = append(items, string(value))
			})
			e.fields[string(key)] = items

		default:
			var str string
			if str, err = jsonparser.ParseString(value); err != nil {
				return err
			}
			e.fields[string(key)] = str

		}
	}

	return nil
}

func (e *entry) print(w io.Writer) {
	var cl *color.Color

	var lvl string
	switch e.level {
	case "debug":
		cl = color.New(color.Bold, color.FgCyan)
		lvl = "[DEBG]"
	case "info":
		cl = color.New(color.Bold, color.FgGreen)
		lvl = "[INFO]"
	case "warn":
		cl = color.New(color.Bold, color.FgYellow)
		lvl = "[WARN]"
	case "error":
		cl = color.New(color.Bold, color.FgRed)
		lvl = "[ERRO]"
	case "dpanic":
		cl = color.New(color.Bold, color.FgHiRed)
		lvl = "[DPAN]"
	case "panic":
		cl = color.New(color.Bold, color.FgHiRed, color.BlinkSlow)
		lvl = "[PANC]"
	case "fatal":
		cl = color.New(color.Bold, color.FgHiRed, color.BlinkRapid)
		lvl = "[FAIL]"
	}

	cl.Fprintf(w, ">>> %s %s: %s", e.ts, lvl, e.message)

	if len(e.caller) > 0 {
		color.New(color.Faint, color.Italic).Fprint(w, " @ ", e.caller)
	}

	if len(e.fields) > 0 {
		keys := make([]string, 0, len(e.fields))
		for key := range e.fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := e.fields[k]

			fmt.Fprint(w, "\n")
			cl.Fprint(w, k, ":")

			switch v := v.(type) {
			case string:
				color.New(color.Faint).Fprint(w, " ", v)
			case []string:
				for _, s := range v {
					color.New(color.Faint).Fprint(w, "\n", s)
				}
			}
		}
		fmt.Fprint(w, "\n")
	}

	if len(e.trace) > 0 {
		fmt.Fprintln(w, "Stacktrace:")
		fmt.Fprintln(w, e.trace)
	}

	fmt.Fprint(w, "\n")
}
