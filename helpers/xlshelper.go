package helpers

import (
	"errors"

	"github.com/360EntSecGroup-Skylar/excelize"
)

type XlsHelper struct {
	F *excelize.File
}

func (c *XlsHelper) OpenExcel(filename string) error {
	return errors.New("Not yet implemented.")
}

func (c *XlsHelper) GetExcelPath() string {
	return ""
}

func (c *XlsHelper) GetSheetMap() map[int]string {
	return nil
}

func (c *XlsHelper) GetCellValue(sheet, axis string) (string, error) {
	return "", errors.New("Not yet implemented.")
}

func (c *XlsHelper) NewStyle(style string) (int, error) {
	return 0, errors.New("Not yet implemented.")
}

func (c *XlsHelper) SetCellStyle(sheet, hcell, vcell string, styleID int) error {
	return errors.New("Not yet implemented.")
}
