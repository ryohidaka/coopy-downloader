package kikaku

import (
	"fmt"
	"time"
)

// お届け日から企画回を取得する
func CalculateKikakuCode(dateStr string) (string, error) {
	inputDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}

	prevWeek := inputDate.AddDate(0, 0, -14)
	targetWeekday := prevWeek.Weekday()
	year, month, _ := prevWeek.Date()

	count := 0
	for d := time.Date(year, month, 1, 0, 0, 0, 0, time.Local); d.Month() == month; d = d.AddDate(0, 0, 1) {
		if d.Weekday() == targetWeekday {
			count++
			if d.Day() == prevWeek.Day() {
				break
			}
		}
	}

	return fmt.Sprintf("%04d%02d%d", year, int(month), count), nil
}
