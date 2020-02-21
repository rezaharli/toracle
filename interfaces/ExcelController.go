package interfaces

type ExcelController interface {
	New(base interface{})

	FileCriteria(file string) bool

	ReadExcel()
}
