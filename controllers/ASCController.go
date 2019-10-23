package controllers

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/sebar/dbflex"
)

type AscController struct {
	*Base
}

func NewAscController() *AscController {
	return new(AscController)
}

func (c *AscController) ReadExcels() error {
	for _, file := range c.FetchFiles() {
		err := c.readExcel(file)
		if err == nil {
			// move file if read succeeded
			helpers.MoveToArchive(file)
			log.Println("Done.")
		} else {
			return err
		}
	}

	return nil
}

func (c *AscController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "ASC") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. ASC files found:", len(resourceFiles))
	return resourceFiles
}

func (c *AscController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	for i, sheetName := range f.GetSheetMap() {
		if i == 1 {
			err = c.ReadMonthlyData(f, sheetName)
			if err != nil {
				log.Println("Error reading monthly data. ERROR:", err)
			}
		} else {
			err = c.ReadDailyData(f, sheetName)
			if err != nil {
				log.Println("Error reading daily data. ERROR:", err)
			}
		}
	}

	if err == nil {
		toolkit.Println()
		log.Println("SUCCESS")
	}
	log.Println("Total Process Time:", time.Since(timeNow).Seconds(), "seconds")

	return err
}

func (c *AscController) ReadMonthlyData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadMonthlyData", sheetName)
	columnsMapping := clit.Config("asc", "monthlyColumnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i))
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
			cellText, err := f.GetCellValue(sheetName, currentCol+headerRow)
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
		cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(firstDataRow+index))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "Total" {
			break
		}

		rowData := toolkit.M{}
		for _, header := range headers {
			currentRow := firstDataRow + index

			stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			if stringData == "" {
				stringData = "0"
			}

			if header.DBFieldName == "PERIOD" {
				stringData, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(firstDataRow-4))
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
				resultRows := make([]toolkit.M, 0)
				param := SqlQueryParam{
					ItemName: strings.ReplaceAll(stringData, "-", ""),
					Results:  &resultRows,
				}

				err := c.selectItemID(param)
				if err != nil {
					log.Fatal(err)
				}

				rowData.Set(header.DBFieldName, resultRows[0].GetString("ITEM_ID"))
			} else {
				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
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

func (c *AscController) ReadDailyData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadDailyData", sheetName)
	columnsMapping := clit.Config("asc", "dailyColumnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		style, _ := f.NewStyle(`{"number_format":15}`)
		f.SetCellStyle(sheetName, "A"+toolkit.ToString(i), "A"+toolkit.ToString(i), style)

		cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i))
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
			cellText, err := f.GetCellValue(sheetName, currentCol+headerRow)
			if err != nil {
				log.Fatal(err)
			}

			if isHeaderDetected == false && strings.TrimSpace(cellText) != "" {
				isHeaderDetected = true
			}

			if isHeaderDetected == true && strings.TrimSpace(cellText) == "" {
				//kalo header ga nemu coba sekali lagi mbok bilih di atasnya
				cellText, err = f.GetCellValue(sheetName, currentCol+toolkit.ToString(toolkit.ToInt(headerRow, "")-1))
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
		cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(firstDataRow+index))
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
				style, _ := f.NewStyle(`{"number_format":15}`)
				f.SetCellStyle(sheetName, header.Column+toolkit.ToString(currentRow), header.Column+toolkit.ToString(currentRow), style)
				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
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

				err := c.selectItemID(param)
				if err != nil {
					log.Fatal(err)
				}

				rowData.Set(header.DBFieldName, resultRows[0].GetString("ITEM_ID"))
			} else {
				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
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

func (c *AscController) selectItemID(param SqlQueryParam) error {
	sqlQuery := "SELECT * FROM D_Item WHERE ITEM_NAME = '" + param.ItemName + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From("D_Item").SQL(sqlQuery), nil)
	defer cursor.Close()

	err := cursor.Fetchs(param.Results, 0)

	return err
}
