package utils

import (
	"reflect"
	"testing"
)

func Test_buildFullSortedFileSlice(t *testing.T) {
	var myslice = []uint32{1, 9, 12, 101}
	pathSlice := buildFullSortedFileSlice(myslice, "/path/to/file/", "M", ".json")

	var expected = []string{"/path/to/file/M00000000000000000001.json",
		"/path/to/file/M00000000000000000009.json",
		"/path/to/file/M00000000000000000012.json",
		"/path/to/file/M00000000000000000101.json"}

	if reflect.DeepEqual(pathSlice, expected) {
		t.Log("Success")
	} else {
		t.Fail()
	}

}

func Test_getSliceOfUintFilenames(t *testing.T) {
	var files = []string{"/path/to/file/M00000000000000000001.json",
		"/path/to/file/M00000000000000000012.json",
		"/path/to/file/M00000000000000000009.json",
		"/path/to/file/M00000000000000000101.json"}
	uint32Filenames, err := getSliceOfUintFilenames(files, "M", ".json")
	if err != nil {
		t.Fail()
	}

	var expected = []uint32{1, 12, 9, 101}
	if reflect.DeepEqual(expected, uint32Filenames) {
		t.Log("Success")
	} else {
		t.Fail()
	}
}

func Test_sortUint32Slice(t *testing.T) {
	var myslice = []uint32{1, 12, 9, 101}
	var expected = []uint32{1, 9, 12, 101}
	sortUint32Slice(myslice)

	if reflect.DeepEqual(expected, myslice) {
		t.Log("Success")
	} else {
		t.Fail()
	}
}
