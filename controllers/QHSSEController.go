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
)

type QhsseController struct {
	*Base
}

func NewQhsseController() *QhsseController {
	return new(QhsseController)
}

func (c *QhsseController) ReadExcels() error {
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

func (c *QhsseController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "QHSSE") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. QHSSE files found:", len(resourceFiles))
	return resourceFiles
}

func (c *QhsseController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	for _, sheetName := range f.GetSheetMap() {
		err = c.ReadData(f, sheetName)
		if err != nil {
			log.Println("Error reading monthly data. ERROR:", err)
		}
	}

	if err == nil {
		toolkit.Println()
		log.Println("SUCCESS")
	}
	log.Println("Total Process Time:", time.Since(timeNow).Seconds(), "seconds")

	return err
}

func (c *QhsseController) ReadData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	columnsMapping := clit.Config("qhsse", "columnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "NO." {
			firstDataRow = i + 1
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

	var err error
	// var rowDatas []toolkit.M
	rowCount := 0
	//iterate over rows
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index

		isRowEmpty := true
		for _, header := range headers {
			if header.DBFieldName == "PERIOD" {
				style, _ := f.NewStyle(`{"number_format":15}`)
				f.SetCellStyle(sheetName, header.Column+toolkit.ToString(currentRow), header.Column+toolkit.ToString(currentRow), style)
				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				stringData = strings.ReplaceAll(stringData, "'", "")

				var t time.Time
				if stringData != "" {
					isRowEmpty = false
					t, err = time.Parse("2-Jan-06", stringData)
					if err != nil {
						t, err = time.Parse("02/01/2006", stringData)
						if err != nil {
							log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
						}
					}
				}

				rowData.Set(header.DBFieldName, t)
			} else {
				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				stringData = strings.ReplaceAll(stringData, "'", "''")

				if len(stringData) > 300 {
					stringData = stringData[0:300]
				}

				if stringData != "" {
					isRowEmpty = false
				}

				rowData.Set(header.DBFieldName, stringData)
			}
		}

		if isRowEmpty {
			break
		}

		param := helpers.InsertParam{
			TableName: "F_QHSSE_INCIDENT",
			Data:      rowData,
		}

		err = helpers.Insert(param)
		if err != nil {
			log.Fatal("Error inserting row "+toolkit.ToString(currentRow)+", ERROR:", err.Error())
		} else {
			log.Println("Row", currentRow, "inserted.")
		}
		rowCount++
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}
