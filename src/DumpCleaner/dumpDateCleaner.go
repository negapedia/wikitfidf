package DumpCleaner

import (
	"../dataStructure"
	"../utils"
	"time"
)

func DataDumpCleaner(page *dataStructure.Page, startDate string, endDate string){

	var startD time.Time
	var endD time.Time
	if startDate != ""{
		startD = utils.TimestampToDate(startDate)
	}
	if endDate != ""{
		endD = utils.TimestampToDate(endDate)
	}

	var newRev []dataStructure.Revision

	for _, rev := range page.Revision {
		timestamp := utils.TimestampToDate(rev.Timestamp)
		if startDate != "" && endDate != "" {
			if timestamp.Sub(startD) >= 0 &&  timestamp.Sub(endD) <= 0 {
				newRev = append(newRev, rev)
			}
		} else if startDate == "" && endDate != "" {
			if timestamp.Sub(endD) <= 0 {
				newRev = append(newRev, rev)
			}
		} else if startDate != "" && endDate == "" {
			if timestamp.Sub(startD) >= 0 {
				newRev = append(newRev, rev)
			}
		}
	}

	page.Revision = newRev
}
