package helpers

import (
	"log"
	"path/filepath"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/eaciit/toolkit"
)

type XlsxHelper struct{}

func (c XlsxHelper) ReadExcel(filename string) (*excelize.File, error) {
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
