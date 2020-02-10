package interfaces

import "github.com/360EntSecGroup-Skylar/excelize"

type ExcelController interface {
	New()
	FetchFiles(resourcePath string) []string
	FileCriteria(file string) bool
	ReadExcel(f *excelize.File) error
}
