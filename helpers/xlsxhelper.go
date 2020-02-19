package helpers

import (
	"log"
	"path/filepath"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/eaciit/toolkit"
)

type XlsxHelper struct {
	F *excelize.File
}

func (c *XlsxHelper) OpenExcel(filename string) error {
	var err error

	toolkit.Println("\n================================================================================")
	log.Println("Opening file", filepath.Base(filename))
	toolkit.Println()

	c.F, err = excelize.OpenFile(filename)
	if err != nil {
		log.Println("Error open file. ERROR:", err)
		return err
	}

	return err
}

func (c *XlsxHelper) GetExcelPath() string {
	return c.F.Path
}

func (c *XlsxHelper) GetSheetMap() map[int]string {
	return c.F.GetSheetMap()
}

func (c *XlsxHelper) GetCellValue(sheet, axis string) (string, error) {
	return c.F.GetCellValue(sheet, axis)
}

func (c *XlsxHelper) NewStyle(style string) (int, error) {
	return c.F.NewStyle(style)
}

func (c *XlsxHelper) SetCellStyle(sheet, hcell, vcell string, styleID int) error {
	return c.F.SetCellStyle(sheet, hcell, vcell, styleID)
}
