package Utils

import (
	"strconv"
	"time"
)

func getMonth(mm int) time.Month  {
	var month time.Month
	switch mm {
	case 1:
		month = time.January
	case 2:
		month = time.February
	case 3:
		month = time.March
	case 4:
		month = time.April
	case 5:
		month = time.May
	case 6:
		month = time.June
	case 7:
		month = time.July
	case 8:
		month = time.August
	case 9:
		month = time.September
	case 10:
		month = time.October
	case 11:
		month = time.November
	case 12:
		month = time.December
	}
	return month
}

func TimestampToDate(timestamp string) time.Time{
	// date
	yyyy, _ := strconv.Atoi(timestamp[:4])
	mm, _ := strconv.Atoi(timestamp[5:7])
	dd, _ := strconv.Atoi(timestamp[8:10])

	// time
	hh, _ := strconv.Atoi(timestamp[11:13])
	min, _ := strconv.Atoi(timestamp[14:16])
	sec, _ := strconv.Atoi(timestamp[17:19])

	return time.Date(yyyy, getMonth(mm), dd, hh, min, sec, 0, time.UTC)
}
