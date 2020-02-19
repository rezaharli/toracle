package interfaces

type ExcelEngine interface {
	OpenExcel(filename string) error

	GetExcelPath() string

	GetSheetMap() map[int]string

	GetCellValue(sheet, axis string) (string, error)

	//excelize
	NewStyle(style string) (int, error)
	GetCellStyle(sheet, axis string) (int, error)
	SetCellStyle(sheet, hcell, vcell string, styleID int) error
	GetFillID(styleID int) int
	GetFgColorTheme(fillID int) *int
	GetFgColorRGB(fillID int) string
	GetFgColorTint(fillID int) float64
	GetSrgbClrVal(fgColorTheme *int) string
}
