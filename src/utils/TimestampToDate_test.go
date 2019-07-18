package utils

import (
	"reflect"
	"testing"
	"time"
)

func TestTimestampToDate(t *testing.T) {
	timestamp1 := "2006-05-15T04:39:39Z"
	timestamp2 := "2006-05-19T18:09:15Z"
	timestamp3 := "2011-01-31T23:30:58Z"

	want1 := time.Date(2006, time.May, 15, 4,39, 39, 0, time.UTC)
	want2 := time.Date(2006, time.May, 19, 18,9, 15,  0, time.UTC)
	want3 := time.Date(2011, time.January, 31, 23,30, 58, 0, time.UTC)

	res1 := TimestampToDate(timestamp1)
	res2 := TimestampToDate(timestamp2)
	res3 := TimestampToDate(timestamp3)

	if reflect.DeepEqual(res1, want1){
		t.Log("timestamp1 OK")
	} else {
		t.Errorf("timestamp3 ERROR expect %v got %v", want1, res1)
	}

	if reflect.DeepEqual(res2, want2){
		t.Log("timestamp2 OK")
	} else {
		t.Errorf("timestamp2 ERROR expect %v got %v", want2, res2)
	}

	if reflect.DeepEqual(res3, want3){
		t.Log("timestamp3 OK")
	} else {
		t.Errorf("timestamp3 ERROR expect %v got %v", want3, res3)
	}
}
