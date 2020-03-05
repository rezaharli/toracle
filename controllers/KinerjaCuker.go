package controllers

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/sebar/dbflex"
	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"
)

type KinerjaCukerController struct {
	*Base
}

func (c *KinerjaCukerController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for KinerjaCuker files.")
	c.FileExtension = ".xlsx"
}

func (c *KinerjaCukerController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "Dry Bulk Handling")
}

func (c *KinerjaCukerController) ReadExcel() {
	for _, sheetName := range c.Engine.GetSheetMap() {
		c.ReadSheet(c.readSheet, sheetName)
	}
}

func (c *KinerjaCukerController) readSheet(sheetName string) error {
	var err error

	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	columnsMapping := clit.Config("kinerjaCuker", "columnsMapping", nil).(map[string]interface{})
	months := clit.Config("kinerjaCuker", "months", []interface{}{}).([]interface{})

	rowCount := 0

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "NO" {
			firstDataRow = i + 2
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

	no := 1
	emptyCount := 0
	//iterate over rows
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index
		isRowEmpty := true

		for _, header := range headers {
			if header.DBFieldName == "Tahun" {
				stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				if strings.TrimSpace(stringData) != "" {
					isRowEmpty = false
				}

				splitted := strings.Split(stringData, "-")
				if len(splitted) > 1 {
					rowData.Set(header.DBFieldName, splitted[1])
				} else {
					rowData.Set(header.DBFieldName, "")
				}
			} else if header.DBFieldName == "Bulan" {
				stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				if strings.TrimSpace(stringData) != "" {
					isRowEmpty = false
				}

				splitted := strings.Split(stringData, "-")
				if len(splitted) > 0 {
					rowData.Set(header.DBFieldName, toolkit.ToString(helpers.IndexOf(splitted[0], months)+1))
				} else {
					rowData.Set(header.DBFieldName, "")
				}
			} else {
				stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				stringData = strings.ReplaceAll(stringData, "'", "''")

				if len(stringData) > 300 {
					stringData = stringData[0:300]
				}

				if strings.TrimSpace(stringData) != "" {
					isRowEmpty = false
				}

				rowData.Set(header.DBFieldName, stringData)
			}
		}

		if emptyCount >= 2 {
			break
		}

		if isRowEmpty {
			emptyCount++
			continue
		}

		tablename := "BOD_Kinerja_Cuker"

		// check if data exists
		sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + rowData["Tahun"].(string) + "' AND bulan = '" + rowData["Bulan"].(string) + "' AND VESSEL_ID = '" + rowData["VESSEL_ID"].(string) + "'"

		conn := helpers.Database()
		cursor := conn.Cursor(dbflex.From(tablename).SQL(sqlQuery), nil)
		defer cursor.Close()

		res := make([]toolkit.M, 0)
		err = cursor.Fetchs(&res, 0)

		//only insert if len of datas is 0 / if no data yet
		if len(res) == 0 {
			c.InsertRowData(currentRow, rowData, tablename)
		}

		rowCount++
		no++
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}

	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")

	return err
}
