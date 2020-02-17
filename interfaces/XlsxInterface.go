package interfaces

import "github.com/360EntSecGroup-Skylar/excelize"

type XlsxController interface {
	ReadExcel(filename string) (*excelize.File, error)
}
