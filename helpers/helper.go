package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func KillApp(err error) {
	fmt.Printf("error. %s \n", err.Error())
	os.Exit(100)
}

func FetchFilePathsWithExt(resourcePath, ext string) []string {
	var files []string
	filepath.Walk(resourcePath, func(path string, f os.FileInfo, _ error) error {
		// exclude files under archive
		if strings.Contains(path, filepath.Join(resourcePath, "archive")) {
			return nil
		}

		if !f.IsDir() {
			if filepath.Ext(path) == ext {
				files = append(files, path)
			}
		}
		return nil
	})

	return files
}

func ToCharStr(i int) string {
	var arr = [...]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
		"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

	return arr[i-1]
}

func ArrayContainsWhitespaceTrimmed(a []interface{}, x string) int {
	trimmedX := strings.Replace(x, " ", "", -1)

	for i, n := range a {
		if trimmedX == strings.Replace(n.(string), " ", "", -1) {
			return i
		}
	}

	return -1
}
