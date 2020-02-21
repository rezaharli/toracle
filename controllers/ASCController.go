package controllers

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
)

// AscController is a controller for every kind of ASC files.
type AscController struct {
	*Base
}

// New is used to initiate the controller
func (c *AscController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for ASC files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *AscController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "Equipment Performance ASC")
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *AscController) ReadExcel() {
	for i, sheetName := range c.Engine.GetSheetMap() {
		if i == 1 {
			c.ReadSheet(c.readMonthlyData, sheetName)
		} else {
			c.ReadSheet(c.readDailyData, sheetName)
		}
	}
}

func (c *AscController) readMonthlyData(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadMonthlyData", sheetName)
	columnsMapping := clit.Config("asc", "monthlyColumnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "1" {
			firstDataRow = i
			break
		}
		i++
	}

	headerRow := toolkit.ToString(firstDataRow - 1)

	months := clit.Config("asc", "months", []interface{}{}).([]interface{})

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
			currentCol := helpers.ToCharStr(i)
			cellText, err := c.Engine.GetCellValue(sheetName, currentCol+headerRow)
			if err != nil {
				log.Fatal(err)
			}

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
		cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(firstDataRow+index))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "Total" {
			break
		}

		rowData := toolkit.M{}
		for _, header := range headers {
			currentRow := firstDataRow + index

			if header.DBFieldName == "PERIOD" {
				stringData, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(firstDataRow-4))
				if err != nil {
					log.Fatal(err)
				}

				monthYear := strings.Split(stringData, " ")
				month := monthYear[2]
				year := monthYear[3]

				t, err := time.Parse("2006-1-02", year+"-"+toolkit.ToString(helpers.IndexOf(month, months)+1)+"-01")
				if err != nil {
					log.Fatal(err)
				}

				rowData.Set(header.DBFieldName, t)
			} else if header.DBFieldName == "ITEM_ID" {
				stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				if stringData == "" {
					stringData = "0"
				}

				resultRows := make([]toolkit.M, 0)
				param := SqlQueryParam{
					ItemName: strings.ReplaceAll(stringData, "-", ""),
					Results:  &resultRows,
				}

				err = c.SelectItemID(param)
				if err != nil {
					log.Fatal(err)
				}

				rowData.Set(header.DBFieldName, resultRows[0].GetString("ITEM_ID"))
			} else {
				stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}
				if stringData == "" {
					stringData = "0"
				}

				rowData.Set(header.DBFieldName, stringData)
			}
		}

		rowDatas = append(rowDatas, rowData)
		rowCount++
	}

	param := helpers.InsertParam{
		TableName: "F_ENG_EQUIPMENT_MONTHLY",
		Data:      rowDatas,
	}

	err := helpers.Insert(param)

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *AscController) readDailyData(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadDailyData", sheetName)
	columnsMapping := clit.Config("asc", "dailyColumnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		style, _ := c.Engine.NewStyle(`{"number_format":15}`)
		c.Engine.SetCellStyle(sheetName, "A"+toolkit.ToString(i), "A"+toolkit.ToString(i), style)

		cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		_, err = time.Parse("2-Jan-06", cellValue)
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
			currentCol := helpers.ToCharStr(i)
			cellText, err := c.Engine.GetCellValue(sheetName, currentCol+headerRow)
			if err != nil {
				log.Fatal(err)
			}

			if isHeaderDetected == false && strings.TrimSpace(cellText) != "" {
				isHeaderDetected = true
			}

			if isHeaderDetected == true && strings.TrimSpace(cellText) == "" {
				//kalo header ga nemu coba sekali lagi mbok bilih di atasnya
				cellText, err = c.Engine.GetCellValue(sheetName, currentCol+toolkit.ToString(toolkit.ToInt(headerRow, "")-1))
				if err != nil {
					log.Fatal(err)
				}

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
		cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(firstDataRow+index))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "Total" {
			break
		}

		rowData := toolkit.M{}
		for _, header := range headers {
			currentRow := firstDataRow + index

			if header.DBFieldName == "PERIOD" {
				style, _ := c.Engine.NewStyle(`{"number_format":15}`)
				c.Engine.SetCellStyle(sheetName, header.Column+toolkit.ToString(currentRow), header.Column+toolkit.ToString(currentRow), style)
				stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				if stringData == "" {
					stringData = "0"
				}

				t, err := time.Parse("2-Jan-06", stringData)
				if err != nil {
					log.Fatal(err)
				}

				rowData.Set(header.DBFieldName, t)
			} else if header.DBFieldName == "ITEM_ID" {
				resultRows := make([]toolkit.M, 0)
				param := SqlQueryParam{
					ItemName: strings.ReplaceAll(sheetName, "-", ""),
					Results:  &resultRows,
				}

				err := c.SelectItemID(param)
				if err != nil {
					log.Fatal(err)
				}

				rowData.Set(header.DBFieldName, resultRows[0].GetString("ITEM_ID"))
			} else {
				stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}
				if stringData == "" {
					stringData = "0"
				}

				rowData.Set(header.DBFieldName, stringData)
			}
		}

		rowDatas = append(rowDatas, rowData)
		rowCount++
	}

	param := helpers.InsertParam{
		TableName: "F_ENG_EQUIPMENT_DAILY",
		Data:      rowDatas,
	}

	err := helpers.Insert(param)

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}
