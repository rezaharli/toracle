package helpers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/eaciit/clit"
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

func MoveToArchive(filePath string) {
	log.Println("Moving file to archive...")
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)

	archivePath := filepath.Join(resourcePath, "archive")
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		os.Mkdir(archivePath, 0755)
	}

	err := os.Rename(filePath, filepath.Join(resourcePath, "archive", filepath.Base(filePath)))
	if err != nil {
		log.Fatal(err)
	}
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

func IndexOf(element string, data []interface{}) int {
	for k, v := range data {
		if strings.EqualFold(element, v.(string)) {
			return k
		}
	}
	return -1 //not found.
}
