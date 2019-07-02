package DumpCleaner

import (
	"../DataStructure"
	"../Utils"
	"time"
)

func DataDumpCleaner(page *DataStructure.Page, startDate string, endDate string){

	var startD time.Time
	var endD time.Time
	if startDate != ""{
		startD = Utils.TimestampToDate(startDate)
	}
	if endDate != ""{
		endD = Utils.TimestampToDate(endDate)
	}

	var newRev []DataStructure.Revision

	for _, rev := range page.Revision {
		timestamp := Utils.TimestampToDate(rev.Timestamp)
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
