package helpers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/eaciit/toolkit"
)

func KillApp(err error) {
	fmt.Printf("error. %s \n", err.Error())
	os.Exit(100)
}

func FetchFilePathsWithExt(resourcePath, ext string) []string {
	var files []string
	filepath.Walk(resourcePath, func(path string, f os.FileInfo, _ error) error {
		// exclude files under archive
		if filepath.Base(filepath.Dir(path)) == "archive" {
			return nil
		}

		if !f.IsDir() {
			if strings.EqualFold(filepath.Ext(path), ext) {
				files = append(files, path)
			}
		}

		return nil
	})

	return files
}

func ReadExcel(filename string) (*excelize.File, error) {
	toolkit.Println("\n================================================================================")
	log.Println("Opening file", filepath.Base(filename))
	toolkit.Println()

	f, err := excelize.OpenFile(filename)
	if err != nil {
		log.Println("Error open file. ERROR:", err)
		return f, err
	}

	return f, err
}

func CharStrToNum(char string) int {
	for index := 1; true; index++ {
		if strings.EqualFold(char, ToCharStr(index)) {
			return index
		}
	}

	return 0
}

func ToCharStr(num int) string {
	var letters = []interface{}{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

	letter := ""
	strNum := strconv.FormatInt(int64(num), 26)
	for i := len(strNum) - 1; i >= 0; i-- {
		angka, err := strconv.Atoi(strNum[i:(i + 1)])

		if err != nil {
			foundIndex := IndexOf(strNum[i:(i+1)], letters)
			if foundIndex != -1 {
				angka = foundIndex + 10
			}
		}

		if i != 0 && angka == 0 {
			tmp, _ := strconv.Atoi(strNum[(i - 1):i])
			tmp = tmp - 1
			strNum = strconv.Itoa(tmp) + strNum[i:(i+1)]

			angka = 26
		}

		if angka != 0 {
			letter = letters[angka-1].(string) + letter
		}
	}

	return letter
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
