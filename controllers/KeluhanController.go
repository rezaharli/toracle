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

// KeluhanController is a controller for every kind of Keluhan files.
type KeluhanController struct {
	*Base
}

// New is used to initiate the controller
func (c *KeluhanController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for Keluhan files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *KeluhanController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "Keluhan")
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *KeluhanController) ReadExcel() {
	for _, sheetName := range c.Engine.GetSheetMap() {
		if strings.Contains(sheetName, "Rekap") {
			c.ReadSheet(c.ReadData, sheetName)
		}
	}
}

func (c *KeluhanController) ReadData(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	columnsMapping := clit.Config("keluhan", "columnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i))
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
				if header.DBFieldName == "TYPE" {
					rowData.Set(header.DBFieldName, "Keluhan Pelanggan")
				}

				continue
			}

			if header.DBFieldName == "PERIOD" || header.DBFieldName == "DUE_DATE" {
				style, _ := c.Engine.NewStyle(`{"number_format":15}`)
				c.Engine.SetCellStyle(sheetName, header.Column+toolkit.ToString(currentRow), header.Column+toolkit.ToString(currentRow), style)
				stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
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
				stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
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
