package common

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/afero"
)

// Use afero for file system to allow for easier testing
var fs afero.Fs = afero.NewOsFs()

// Thinking outputs a spinner to indicate that we're thinking about something
var Thinking *spinner.Spinner = spinner.New(spinner.CharSets[9], 100*time.Millisecond)

// DeleteFromSlice removes the supplied single element from a given slice
func DeleteFromSlice(slice []string, selector string) []string {
	var result []string
	for _, element := range slice {
		if element != selector {
			result = append(result, element)
		}
	}
	return result
}

// FileExists returns whether or not the given file exists in the OS
func FileExists(fs afero.Fs, filename string) bool {
	_, err := fs.Stat(filename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// FindInSlice returns whether or not the given slice contains the supplied element
// and, if so, returns its index
func FindInSlice(slice []string, element string) (int, bool) {
	for i, item := range slice {
		if item == element {
			return i, true
		}
	}
	return -1, false
}
