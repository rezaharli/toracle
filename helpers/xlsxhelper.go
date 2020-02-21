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

func (c *XlsxHelper) GetFile(style string) *excelize.File {
	return c.F
}

func (c *XlsxHelper) GetSheetMap() map[int]string {
	return c.F.GetSheetMap()
}

func (c *XlsxHelper) GetCellValue(sheet, axis string) (string, error) {
	return c.F.GetCellValue(sheet, axis)
}

func (c *XlsxHelper) GetCellStyle(sheet, axis string) (int, error) {
	return c.F.GetCellStyle(sheet, axis)
}

func (c *XlsxHelper) NewStyle(style string) (int, error) {
	return c.F.NewStyle(style)
}

func (c *XlsxHelper) SetCellStyle(sheet, hcell, vcell string, styleID int) error {
	return c.F.SetCellStyle(sheet, hcell, vcell, styleID)
}

func (c *XlsxHelper) GetFillID(styleID int) int {
	return c.F.Styles.CellXfs.Xf[styleID].FillID
}

func (c *XlsxHelper) GetFgColorTheme(fillID int) *int {
	return c.F.Styles.Fills.Fill[fillID].PatternFill.FgColor.Theme
}

func (c *XlsxHelper) GetFgColorRGB(fillID int) string {
	return c.F.Styles.Fills.Fill[fillID].PatternFill.FgColor.RGB
}

func (c *XlsxHelper) GetFgColorTint(fillID int) float64 {
	return c.F.Styles.Fills.Fill[fillID].PatternFill.FgColor.Tint
}

func (c *XlsxHelper) GetSrgbClrVal(fgColorTheme *int) string {
	return c.F.Theme.ThemeElements.ClrScheme.Children[*fgColorTheme].SrgbClr.Val
}

func (c *XlsxHelper) IsSheetVisible(name string) bool {
	return c.F.GetSheetVisible(name)
}
