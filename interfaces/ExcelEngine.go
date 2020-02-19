package interfaces

type ExcelEngine interface {
	OpenExcel(filename string) error

	GetExcelPath() string

	GetSheetMap() map[int]string

	GetCellValue(sheet, axis string) (string, error)

	NewStyle(style string) (int, error)

	SetCellStyle(sheet, hcell, vcell string, styleID int) error
}
