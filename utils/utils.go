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
	year2, month2, day2 := now.Date()
	hour2, minute2, second2 := now.Clock()

	year1, month1, day1 := then.Date()
	hour1, minute1, second1 := then.Clock()

	year := math.Abs(float64(int(year2 - year1)))
	month := math.Abs(float64(int(month2 - month1)))
	day := math.Abs(float64(int(day2 - day1)))
	hour := math.Abs(float64(int(hour2 - hour1)))
	minute := math.Abs(float64(int(minute2 - minute1)))
	second := math.Abs(float64(int(second2 - second1)))

	week := math.Floor(day / 7)

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
