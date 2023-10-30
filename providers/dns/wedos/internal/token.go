package internal

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

func authToken(userName string, wapiPass string) string {
	return sha1string(userName + sha1string(wapiPass) + czechHourString())
}

func sha1string(txt string) string {
	h := sha1.New()
	_, _ = io.WriteString(h, txt)
	return hex.EncodeToString(h.Sum(nil))
}

func czechHourString() string {
	return formatHour(czechHour())
}

func czechHour() int {
	tryZones := []string{"Europe/Prague", "Europe/Paris", "CET"}

	for _, zoneName := range tryZones {
		loc, err := time.LoadLocation(zoneName)
		if err == nil {
			return time.Now().In(loc).Hour()
		}
	}

	// hopefully this will never be used
	// this is fallback for containers without tzdata installed
	return utcToCet(time.Now().UTC()).Hour()
}

func utcToCet(utc time.Time) time.Time {
	// https://en.wikipedia.org/wiki/Central_European_Time
	// As of 2011, all member states of the European Union observe Summer Time (daylight saving time),
	// from the last Sunday in March to the last Sunday in October.
	// States within the CET area switch to Central European Summer Time (CEST -- UTC+02:00) for the summer.[1]
	utcMonth := utc.Month()
	if utcMonth < time.March || utcMonth > time.October {
		return utc.Add(time.Hour)
	}
	if utcMonth > time.March && utcMonth < time.October {
		return utc.Add(time.Hour * 2)
	}

	dayOff := 0
	breaking := time.Date(utc.Year(), utcMonth+1, dayOff, 1, 0, 0, 0, time.UTC)
	for {
		if breaking.Weekday() == time.Sunday {
			break
		}
		dayOff--
		breaking = time.Date(utc.Year(), utcMonth+1, dayOff, 1, 0, 0, 0, time.UTC)
		if dayOff < -7 {
			panic("safety exit to avoid infinite loop")
		}
	}

	if (utcMonth == time.March && utc.Before(breaking)) || (utcMonth == time.October && utc.After(breaking)) {
		return utc.Add(time.Hour)
	}
	return utc.Add(time.Hour * 2)
}

func formatHour(hour int) string {
	return fmt.Sprintf("%02d", hour)
}
