package controllers

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"
	"github.com/xuri/excelize"
)

type AscController struct {
}

func NewAscController() *AscController {
	return new(AscController)
}

type Header struct {
	DBFieldName string
	HeaderName  string

	Column string
	Row    string
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

func (c *AscController) readExcel(filename string) error {
	timeNow := time.Now()

	toolkit.Println("\n================================================================================")
	log.Println("Opening file", filepath.Base(filename), "\n")
	f, err := excelize.OpenFile(filename)
	if err != nil {
		log.Println("Error open file. ERROR:", err)
		return err
	}

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
			currentCol := helpers.ToCharStr(i)
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
				param := helpers.SqlQueryParam{
					ItemName: strings.ReplaceAll(stringData, "-", ""),
					Results:  &resultRows,
				}

				err := helpers.SelectItemID(param)
				if err != nil {
					log.Fatal(err)
				}

				rowData.Set(header.DBFieldName, resultRows[0].GetString("ITEMID"))
			} else {
				rowData.Set(header.DBFieldName, stringData)
			}
		}

		rowDatas = append(rowDatas, rowData)
		rowCount++
	}

	param := helpers.InsertParam{
		TableName: "F_EquipmentMonthly",
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
			currentCol := helpers.ToCharStr(i)
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
				param := helpers.SqlQueryParam{
					ItemName: strings.ReplaceAll(sheetName, "-", ""),
					Results:  &resultRows,
				}

				err := helpers.SelectItemID(param)
				if err != nil {
					log.Fatal(err)
				}

				rowData.Set(header.DBFieldName, resultRows[0].GetString("ITEMID"))
			} else {
				rowData.Set(header.DBFieldName, stringData)
			}
		}

		rowDatas = append(rowDatas, rowData)
		rowCount++
	}

	param := helpers.InsertParam{
		TableName: "F_EquipmentDaily",
		Data:      rowDatas,
	}

	err := helpers.Insert(param)

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}
