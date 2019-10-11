package helpers

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"
)

type Header struct {
	DBFieldName string
	HeaderName  string

	Column string
	Row    string
}

func ReadExcel(filename string) error {
	timeNow := time.Now()

	toolkit.Println("\n================================================================================")
	toolkit.Println("Opening file", filepath.Base(filename), "\n")
	f, err := excelize.OpenFile(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	toolkit.Println("Processing sheets...")
	for i, sheetName := range f.GetSheetMap() {
		if i == 1 {
			err = ReadMonthlyData(f, sheetName)
			if err != nil {
				toolkit.Println("ERROR:", err)
			}
		} else {
			err = ReadDailyData(f, sheetName)
			if err != nil {
				toolkit.Println("ERROR:", err)
			}
		}
	}

	if err == nil {
		toolkit.Println("\nSUCCESS")
	}
	toolkit.Println("Total Process Time:", time.Since(timeNow).Seconds(), "seconds")

	return err
}

func ReadMonthlyData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()
	toolkit.Println("\nReadMonthlyData", sheetName)
	columnsMapping := clit.Config("default", "monthlyColumnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		if f.GetCellValue(sheetName, "A"+toolkit.ToString(i)) == "1" {
			firstDataRow = i
			break
		}
		i++
	}

	headerRow := toolkit.ToString(firstDataRow - 1)

	var headers []Header
	for key, column := range columnsMapping {
		isHeaderDetected := false
		i = 1

		header := Header{
			DBFieldName: key,
			HeaderName:  "",
			Column:      "",
			Row:         "",
		}

		// search for particular header in excel
		for {
			currentCol := ToCharStr(i)
			cellText := f.GetCellValue(sheetName, currentCol+headerRow)

			if isHeaderDetected == false && strings.TrimSpace(cellText) != "" {
				isHeaderDetected = true
			}

			if isHeaderDetected == true && strings.TrimSpace(cellText) == "" {
				break
			}

			if isHeaderDetected {
				if strings.Replace(column.(string), " ", "", -1) == strings.Replace(cellText, " ", "", -1) {
					header.HeaderName = cellText
					header.Column = currentCol
					header.Row = headerRow

					break
				}
			}

			i++
		}

		headers = append(headers, header)
	}

	var rowDatas []toolkit.M
	rowCount := 0
	//iterate over rows
	for index := 0; true; index++ {
		// end jika udah nemu total
		if f.GetCellValue(sheetName, "A"+toolkit.ToString(firstDataRow+index)) == "Total" {
			break
		}

		rowData := toolkit.M{}
		for _, header := range headers {
			currentRow := firstDataRow + index

			stringData := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
			if stringData == "" {
				stringData = "0"
			}

			if header.DBFieldName == "ItemID" {
				resultRows := make([]toolkit.M, 0)
				param := SqlQueryParam{
					ItemName: strings.ReplaceAll(stringData, "-", ""),
					Results:  &resultRows,
				}

				err := selectItemID(param)
				if err != nil {
					toolkit.Println(err)
				}

				rowData.Set(header.DBFieldName, resultRows[0].GetString("ITEMID"))
			} else {
				rowData.Set(header.DBFieldName, stringData)
			}
		}

		rowDatas = append(rowDatas, rowData)
		rowCount++
	}

	param := InsertParam{
		TableName: "F_EquipmentMonthly",
		Data:      rowDatas,
	}

	err := Insert(param)

	if err == nil {
		toolkit.Println("SUCCESS Processing", rowCount, "rows")
	}
	toolkit.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func ReadDailyData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()
	toolkit.Println("\nReadDailyData", sheetName)
	columnsMapping := clit.Config("default", "dailyColumnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		style, _ := f.NewStyle(`{"number_format":15}`)
		f.SetCellStyle(sheetName, "A"+toolkit.ToString(i), "A"+toolkit.ToString(i), style)
		_, err := time.Parse("2-Jan-06", f.GetCellValue(sheetName, "A"+toolkit.ToString(i)))
		if err == nil {
			firstDataRow = i
			break
		}
		i++
	}

	headerRow := toolkit.ToString(firstDataRow - 1)

	var headers []Header
	for key, column := range columnsMapping {
		isHeaderDetected := false
		i = 1

		header := Header{
			DBFieldName: key,
			HeaderName:  "",
			Column:      "",
			Row:         "",
		}

		// search for particular header in excel
		for {
			currentCol := ToCharStr(i)
			cellText := f.GetCellValue(sheetName, currentCol+headerRow)

			if isHeaderDetected == false && strings.TrimSpace(cellText) != "" {
				isHeaderDetected = true
			}

			if isHeaderDetected == true && strings.TrimSpace(cellText) == "" {
				//kalo header ga nemu coba sekali lagi mbok bilih di atasnya
				cellText = f.GetCellValue(sheetName, currentCol+toolkit.ToString(toolkit.ToInt(headerRow, "")-1))
				if strings.TrimSpace(cellText) == "" {
					break
				}
			}

			if isHeaderDetected {
				if strings.Replace(column.(string), " ", "", -1) == strings.Replace(cellText, " ", "", -1) {
					header.HeaderName = cellText
					header.Column = currentCol
					header.Row = headerRow

					break
				}
			}

			i++
		}

		headers = append(headers, header)
	}

	var rowDatas []toolkit.M
	rowCount := 0
	for index := 0; true; index++ {
		// end jika udah nemu total
		if f.GetCellValue(sheetName, "A"+toolkit.ToString(firstDataRow+index)) == "Total" {
			break
		}

		rowData := toolkit.M{}
		for _, header := range headers {
			currentRow := firstDataRow + index

			stringData := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
			if stringData == "" {
				stringData = "0"
			}

			if header.DBFieldName == "ItemID" {
				resultRows := make([]toolkit.M, 0)
				param := SqlQueryParam{
					ItemName: strings.ReplaceAll(sheetName, "-", ""),
					Results:  &resultRows,
				}

				err := selectItemID(param)
				if err != nil {
					toolkit.Println(err)
				}

				rowData.Set(header.DBFieldName, resultRows[0].GetString("ITEMID"))
			} else {
				rowData.Set(header.DBFieldName, stringData)
			}
		}

		rowDatas = append(rowDatas, rowData)
		rowCount++
	}

	param := InsertParam{
		TableName: "F_EquipmentDaily",
		Data:      rowDatas,
	}

	err := Insert(param)

	if err == nil {
		toolkit.Println("SUCCESS Processing", rowCount, "rows")
	}
	toolkit.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}
