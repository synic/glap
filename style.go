package glap

import (
	"os"
)

// ColorMode controls whether ANSI color codes are used in output.
type ColorMode int

const (
	ColorAuto   ColorMode = iota
	ColorAlways
	ColorNever
)

func (m ColorMode) enabled() bool {
	switch m {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default:
		if _, ok := os.LookupEnv("NO_COLOR"); ok {
			return false
		}
		return isTTY()
	}
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

type styler struct {
	enabled bool
}

func newStyler(mode ColorMode) styler {
	return styler{enabled: mode.enabled()}
}

func (s styler) bold(text string) string {
	if !s.enabled {
		return text
	}
	return "\033[1m" + text + "\033[0m"
}

func (s styler) dim(text string) string {
	if !s.enabled {
		return text
	}
	return "\033[2m" + text + "\033[0m"
}

func (s styler) green(text string) string {
	if !s.enabled {
		return text
	}
	return "\033[32m" + text + "\033[0m"
}

func (s styler) cyan(text string) string {
	if !s.enabled {
		return text
	}
	return "\033[36m" + text + "\033[0m"
}

func (s styler) yellow(text string) string {
	if !s.enabled {
		return text
	}
	return "\033[33m" + text + "\033[0m"
}
