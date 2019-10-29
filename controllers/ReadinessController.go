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

type ReadinessController struct {
	*Base
}

func NewReadinessController() *ReadinessController {
	return new(ReadinessController)
}

func (c *ReadinessController) ReadExcels() error {
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

func (c *ReadinessController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "Readiness") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. Readiness files found:", len(resourceFiles))
	return resourceFiles
}

func (c *ReadinessController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	for _, sheetName := range f.GetSheetMap() {
		err = c.ReadData(f, sheetName)
		if err != nil {
			log.Println("Error reading data. ERROR:", err)
		}
	}

	if err == nil {
		toolkit.Println()
		log.Println("SUCCESS")
	}
	log.Println("Total Process Time:", time.Since(timeNow).Seconds(), "seconds")

	return err
}

func (c *ReadinessController) ReadData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	columnsMapping := clit.Config("readiness", "columnsMapping", nil).(map[string]interface{})

	var currentPeriod time.Time

	var err error

	//iterate into groups of data
	notPeriodCount := 0
	periodFound := false
	row := 1
	for {
		firstDataRow := 0
		notPeriodCount = 0

		//search for period
		for {
			if row >= 1 {
				stringData, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(row))
				if err != nil {
					log.Fatal(err)
				}

				if strings.Contains(stringData, "Shift") {
					notPeriodCount = 0

					var t time.Time

					currentPeriod = t
					firstDataRow = row + 3

					stringData, err = f.GetCellValue(sheetName, "A"+toolkit.ToString(firstDataRow))
					if err != nil {
						log.Fatal(err)
					}

					if strings.EqualFold(strings.TrimSpace(stringData), strings.TrimSpace("No")) {
						firstDataRow = firstDataRow + 1
					}

					periodFound = true
					break
				}
			}

			if notPeriodCount > 200 {
				periodFound = false
				break
			}

			notPeriodCount++
			row++
		}

		if !periodFound {
			break
		}

		var headers []Header
		for key, column := range columnsMapping {
			header := Header{
				DBFieldName: key,
				Column:      column.(string),
			}

			headers = append(headers, header)
		}

		//iterate over rows
		rowCount := 0
		for index := 0; true; index++ {
			rowData := toolkit.M{}
			currentRow := firstDataRow + index
			row = currentRow
			isRowEmpty := true

			stringData, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			if strings.Contains(stringData, "Total") {
				break
			}

			for _, header := range headers {
				if header.DBFieldName == "PERIOD" || header.DBFieldName == "LAST_UDPATE" {
					rowData.Set(header.DBFieldName, currentPeriod)
				} else if header.DBFieldName == "STATUS" {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.TrimSpace(strings.ReplaceAll(stringData, "'", "''"))

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					//try next column if empty
					if stringData == "" {
						stringData, err := f.GetCellValue(sheetName, helpers.ToCharStr(helpers.CharStrToNum(header.Column)+1)+toolkit.ToString(currentRow))
						if err != nil {
							log.Fatal(err)
						}

						stringData = strings.TrimSpace(strings.ReplaceAll(stringData, "'", "''"))

						if len(stringData) > 300 {
							stringData = stringData[0:300]
						}
					}

					if stringData != "" {
						isRowEmpty = false
					}

					rowData.Set(header.DBFieldName, stringData)
				} else if header.Column == "" {
					rowData.Set(header.DBFieldName, "")
				} else {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.TrimSpace(strings.ReplaceAll(stringData, "'", "''"))

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
				TableName: "F_ENG_EQUIPMENT_STATUS",
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

		row++

		if err == nil {
			log.Println("SUCCESS Processing", rowCount, "rows\n")
		}
	}

	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}
