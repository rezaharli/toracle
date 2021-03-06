package controllers

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/sebar/dbflex"
)

type KinerjaTerminalController struct {
	*Base
}

func (c *KinerjaTerminalController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for KinerjaTerminal files.")
	c.FileExtension = ".xlsx"
}

func (c *KinerjaTerminalController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "Data Trafik, Arus, & Kinerja Operasi")
}

func (c *KinerjaTerminalController) ReadExcel() {
	for _, sheetName := range c.Engine.GetSheetMap() {
		_, err := strconv.Atoi(sheetName)
		if err == nil {
			c.ReadSheet(c.readSheet, sheetName)
		}
	}
}

func (c *KinerjaTerminalController) readSheet(sheetName string) error {
	var err error

	timeNow := time.Now()

	log.Println("\nReadData", sheetName)

	columnsMapping := clit.Config("kinerjaTerminal", "columnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			helpers.HandleError(err)
		}

		if strings.EqualFold(strings.TrimSpace(cellValue), "Uraian") {
			cellValue, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(i+1))
			if err != nil {
				helpers.HandleError(err)
			}

			if !strings.EqualFold(strings.TrimSpace(cellValue), "Uraian") {
				firstDataRow = i + 2
				break
			}
		}

		i++
	}

	var monthHeaders []Header

	monthRow := firstDataRow - 2
	monthsFound := []interface{}{}
	isHeaderDetected := false
	currentTahun := ""

	i = 5 //mulai kolom 5 (E)
	prevCell := ""
	for {
		header := Header{
			DBFieldName:  "",
			Column:       "",
			ColumnNumber: i,
		}

		currentCol := helpers.ToCharStr(i)
		cellText, err := c.Engine.GetCellValue(sheetName, currentCol+toolkit.ToString(monthRow))
		if err != nil {
			helpers.HandleError(err)
		}

		dateString := strings.TrimSpace(cellText)
		_, timeParseErr := time.Parse("Jan-06", strings.TrimSpace(cellText))
		if isHeaderDetected == false && timeParseErr == nil {
			isHeaderDetected = true

			currentTahun = strings.Split(strings.TrimSpace(cellText), "-")[1]
		}

		if isHeaderDetected == true && timeParseErr != nil {
			break
		}

		if isHeaderDetected {
			if timeParseErr == nil {
				if strings.TrimSpace(cellText) != strings.TrimSpace(prevCell) {
					if helpers.IndexOf(dateString, monthsFound) != -1 { //jika bulan sudah ditemukan sebelumnya
						break
					} else {
						monthsFound = append(monthsFound, dateString)
					}

					header.HeaderName = cellText
					header.Column = currentCol

					monthHeaders = append(monthHeaders, header)

					prevCell = cellText
				}
			}
		}

		i++
	}

	rowCount := 0
	months := clit.Config("kinerjaTerminal", "months", nil).([]interface{})

	tablename := "BOD_Kinerja_Terminal"
	
	log.Println("Deleting datas.")

	sql := "DELETE FROM " + tablename + " WHERE tahun = '" + currentTahun + "'"

	conn := helpers.Database()
	query, err := conn.Prepare(dbflex.From(tablename).SQL(sql))
	if err != nil {
		log.Println(err)
	}

	_, err = query.Execute(toolkit.M{}.Set("data", toolkit.M{}))
	if err != nil {
		log.Println(err)
	}

	log.Println("Data deleted.")

	toolkit.Println()

	// check if data exists
	sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + currentTahun + "'"

	cursor := conn.Cursor(dbflex.From(tablename).SQL(sqlQuery), nil)
	defer cursor.Close()

	res := make([]toolkit.M, 0)
	err = cursor.Fetchs(&res, 0)

	// only insert if len of datas is 0 / if no data yet
	if len(res) == 0 {
		for _, monthHeader := range monthHeaders {
			var headers []Header
			for key, column := range columnsMapping {
				header := Header{
					DBFieldName: key,
					Column:      column.(string),
				}

				if key == "Bulan" {
					header.Value = strings.Split(strings.TrimSpace(monthHeader.HeaderName), "-")[0]
					header.Column = monthHeader.Column
				}

				if key == "Tahun" {
					header.Value = strings.Split(strings.TrimSpace(monthHeader.HeaderName), "-")[1]
					header.Column = monthHeader.Column
				}

				if key == "RKAP_Bulanan" {
					header.Column = monthHeader.Column
				}

				if key == "Realisasi" {
					header.Column = helpers.ToCharStr(helpers.CharStrToNum(monthHeader.Column) + 12)
				}

				headers = append(headers, header)
			}

			//iterate over rows
			rowEmptyCount := 0
			for index := 0; true; index++ {
				rowData := toolkit.M{}
				currentRow := firstDataRow + index
				isRowEmpty := true

				for _, header := range headers {
					if header.DBFieldName == "Bulan" {
						rowData.Set(header.DBFieldName, toolkit.ToString(helpers.IndexOf(header.Value, months)+1))
					} else if header.DBFieldName == "Tahun" {
						rowData.Set(header.DBFieldName, header.Value)
					} else {
						stringData := c.readCell(sheetName, header.Column+toolkit.ToString(currentRow))

						if header.DBFieldName != "Uraian" && stringData != "" {
							isRowEmpty = false
						}

						rowData.Set(header.DBFieldName, stringData)
					}
				}

				if rowEmptyCount >= 10 {
					break
				}

				if isRowEmpty {
					rowEmptyCount++
					continue
				}

				param := helpers.InsertParam{
					TableName: tablename,
					Data:      rowData,
				}

				err = helpers.Insert(param)
				if err != nil {
					helpers.HandleError(err)
				} else {
					log.Println(monthHeader.HeaderName + ", inserted.")
				}

				rowCount++
			}
		}
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}

	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *KinerjaTerminalController) readCell(sheetName, cellID string) string {
	stringData, err := c.Engine.GetCellValue(sheetName, cellID)
	if err != nil {
		helpers.HandleError(err)
	}

	stringData = strings.ReplaceAll(stringData, "'", "''")
	stringData = strings.ReplaceAll(stringData, "-", "")

	stringData = strings.TrimSpace(stringData)

	if len(stringData) > 300 {
		stringData = stringData[0:300]
	}

	return stringData
}
