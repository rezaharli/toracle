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

type KeluhanController struct {
	*Base
}

func NewKeluhanController() *KeluhanController {
	return new(KeluhanController)
}

func (c *KeluhanController) ReadExcels() error {
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

func (c *KeluhanController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "Keluhan") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. Keluhan files found:", len(resourceFiles))
	return resourceFiles
}

func (c *KeluhanController) readExcel(filename string) error {
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

func (c *KeluhanController) ReadData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	columnsMapping := clit.Config("keluhan", "columnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "NO" {
			firstDataRow = i + 1
			break
		}
		i++
	}

	var headers []Header
	for key, column := range columnsMapping {
		header := Header{
			DBFieldName: key,
			Column:      column.(string),
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
			if header.Column == "" {
				continue
			}

			if header.DBFieldName == "PERIOD" || header.DBFieldName == "DUE_DATE" {
				style, _ := f.NewStyle(`{"number_format":15}`)
				f.SetCellStyle(sheetName, header.Column+toolkit.ToString(currentRow), header.Column+toolkit.ToString(currentRow), style)
				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				stringData = strings.ReplaceAll(stringData, "'", "")
				stringData = strings.ReplaceAll(stringData, "`", "")

				var t time.Time
				if stringData != "" {
					isRowEmpty = false
					t, err = time.Parse("2-Jan-06", stringData)
					if err != nil {
						t, err = time.Parse("02/01/2006", stringData)
						if err != nil {
							t, err = time.Parse("2/1/2006", stringData)
							if err != nil {
								log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
							}
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
