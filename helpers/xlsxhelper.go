package helpers

import (
	"log"
	"path/filepath"

	"github.com/eaciit/toolkit"
	"github.com/360EntSecGroup-Skylar/excelize"
)

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
