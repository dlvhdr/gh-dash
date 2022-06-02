package utils

import (
	"fmt"
	"log"
	"math"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	ApproxDaysInYear  = 365
	ApproxDaysInMonth = 28
	DaysInWeek        = 7
)

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func openInLinuxBrowser(url string) error {
	var err error
	providers := []string{"xdg-open", "x-www-browser", "www-browser", "wslview"}

	for _, provider := range providers {
		if _, err = exec.LookPath(provider); err == nil {
			err = exec.Command(provider, url).Start()
			if err != nil {
				return err
			}
			return nil
		}
	}

	return &exec.Error{Name: strings.Join(providers, ","), Err: exec.ErrNotFound}
}

func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = openInLinuxBrowser(url)
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	case "android":
		err = exec.Command("termux-open-url", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func TruncateString(str string, num int) string {
	truncated := str
	if num <= 3 {
		return str
	}
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		truncated = str[0:num] + "…"
	}
	return truncated
}

func TruncateStringTrailing(str string, num int) string {
	truncated := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		skipped := len(str) - num
		truncated = "..." + str[skipped:]
	}
	return truncated
}

func TimeElapsed(then time.Time) string {
	var parts []string
	var text string

	now := time.Now()
	diff := now.Sub(then)
	day := math.Floor(diff.Hours() / 24)
	year := math.Floor(day / ApproxDaysInYear)
	month := math.Floor(day / ApproxDaysInMonth)
	week := math.Floor(day / DaysInWeek)
	hour := math.Floor(math.Abs(diff.Hours()))
	minute := math.Floor(math.Abs(diff.Minutes()))
	second := math.Floor(math.Abs(diff.Seconds()))

	if year > 0 {
		parts = append(parts, strconv.Itoa(int(year))+"y")
	}

	if month > 0 {
		parts = append(parts, strconv.Itoa(int(month))+"mo")
	}

	if week > 0 {
		parts = append(parts, strconv.Itoa(int(week))+"w")
	}

	if day > 0 {
		parts = append(parts, strconv.Itoa(int(day))+"d")
	}

	if hour > 0 {
		parts = append(parts, strconv.Itoa(int(hour))+"h")
	}

	if minute > 0 {
		parts = append(parts, strconv.Itoa(int(minute))+"m")
	}

	if second > 0 {
		parts = append(parts, strconv.Itoa(int(second))+"s")
	}

	if now.After(then) {
		text = " ago"
	} else {
		text = " after"
	}

	if len(parts) == 0 {
		return "just now"
	}

	return parts[0] + text
}

func BoolPtr(b bool) *bool       { return &b }
func StringPtr(s string) *string { return &s }
func UintPtr(u uint) *uint       { return &u }
func IntPtr(u int) *int          { return &u }
