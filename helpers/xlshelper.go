package helpers

import (
	"errors"
)

type XlsHelper struct{}

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

func (c *XlsHelper) GetStyle(style string) (int, error) {
	return 0, errors.New("Not yet implemented.")
}

func (c *XlsHelper) GetCellStyle(sheet, axis string) (int, error) {
	return 0, errors.New("Not yet implemented.")
}

func (c *XlsHelper) SetCellStyle(sheet, hcell, vcell string, styleID int) error {
	return errors.New("Not yet implemented.")
}

func (c *XlsHelper) GetFillID(styleID int) int {
	return 0
}

func (c *XlsHelper) GetFgColorTheme(fillID int) *int {
	ret := 0
	return &ret
}

func (c *XlsHelper) GetFgColorRGB(fillID int) string {
	return ""
}

func (c *XlsHelper) GetFgColorTint(fillID int) float64 {
	return 0
}

func (c *XlsHelper) GetSrgbClrVal(fgColorTheme *int) string {
	return ""
}
