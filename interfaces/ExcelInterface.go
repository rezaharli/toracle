package interfaces

import "github.com/360EntSecGroup-Skylar/excelize"

type ExcelController interface {
	New(base interface{})

	FetchFiles(resourcePath string) []string

	FileCriteria(file string) bool

	ReadExcel(f *excelize.File) error
}
