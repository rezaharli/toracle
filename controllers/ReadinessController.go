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

// ReadinessController is a controller for every kind of Readiness files.
type ReadinessController struct {
	*Base
}

// New is used to initiate the controller
func (c *ReadinessController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for Readiness files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *ReadinessController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "Readiness")
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *ReadinessController) ReadExcel() {
	for _, sheetName := range c.Engine.GetSheetMap() {
		c.ReadSheet(c.ReadData, sheetName)
	}
}

func (c *ReadinessController) ReadData(sheetName string) error {
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
				stringData, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(row))
				if err != nil {
					log.Fatal(err)
				}

				if strings.Contains(stringData, "Shift") {
					notPeriodCount = 0

					var t time.Time

					currentPeriod = t
					firstDataRow = row + 3

					stringData, err = c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(firstDataRow))
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

			stringData, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
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
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.TrimSpace(strings.ReplaceAll(stringData, "'", "''"))

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					//try next column if empty
					if stringData == "" {
						stringData, err := c.Engine.GetCellValue(sheetName, helpers.ToCharStr(helpers.CharStrToNum(header.Column)+1)+toolkit.ToString(currentRow))
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
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
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
