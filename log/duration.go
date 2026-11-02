package log

import (
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"
)

type FormattableDuration time.Duration

func (f FormattableDuration) String() string {
	d := time.Duration(f)

	days := int(math.Trunc(d.Hours() / 24))
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := d.Milliseconds() % 1000
	ns := d.Nanoseconds() % int64(time.Millisecond)

	s := new(strings.Builder)

	if days > 0 {
		_, _ = fmt.Fprintf(s, "%dd", days)
	}

	if hours > 0 {
		_, _ = fmt.Fprintf(s, "%dh", hours)
	}

	if minutes > 0 {
		_, _ = fmt.Fprintf(s, "%dm", minutes)
	}

	if seconds > 0 {
		_, _ = fmt.Fprintf(s, "%ds", seconds)
	}

	if milliseconds > 0 {
		_, _ = fmt.Fprintf(s, "%dms", milliseconds)
	}

	if ns > 0 {
		_, _ = fmt.Fprintf(s, "%dns", ns)
	}

	return s.String()
}

func (f FormattableDuration) LogValue() slog.Value {
	return slog.StringValue(f.String())
}
