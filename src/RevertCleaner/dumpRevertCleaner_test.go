package RevertCleaner

import (
	"reflect"
	"testing"
)

func Test_getAllOccurencePosition(t *testing.T) {
	sha1Array := []string{"a", "b", "a", "c", "b"}

	want := []int{0, 1, 2, 4}

	result := getAllOccurencePosition(sha1Array)

	if reflect.DeepEqual(result, want) {
		t.Logf("OK Test_getAllOccurencePosition")
	} else {
		t.Errorf("Error: get:%v expected: %v", result, want)
	}
}

func Test_getRevertsPosition(t *testing.T) {
	array := []int{1, 2, 3, 5, 6, 9}
	want := []int{4, 7, 8}

	result := getRevertsPosition(array)

	if reflect.DeepEqual(result, want) {
		t.Logf("OK Test_getRevertsPosition")
	} else {
		t.Errorf("Error: get:%v expected: %v", result, want)
	}
}
