package controllers

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
)

type KinerjaTerminalController struct {
	*Base

	FileExtension string
}

func (c *KinerjaTerminalController) New() {
	log.Println("Scanning for KinerjaTerminal files.")
	c.FileExtension = ".xlsx"
}

func (c *KinerjaTerminalController) FetchFiles(resourcePath string) []string {
	return helpers.FetchFilePathsWithExt(resourcePath, c.FileExtension)
}

func (c *KinerjaTerminalController) FileCriteria(file string) bool {
	if strings.Contains(filepath.Base(file), "Data Trafik, Arus, & Kinerja Operasi") {
		return true
	}

	return false
}

func (c *KinerjaTerminalController) ReadExcel(f *excelize.File) error {
	var err error

	for _, sheetName := range f.GetSheetMap() {
		_, err := strconv.Atoi(sheetName)
		if err == nil {
			err = c.readSheet(f, sheetName)
			if err != nil {
				log.Println("Error reading monthly data. ERROR:", err)
			}
		}
	}

	return err
}

func (c *KinerjaTerminalController) readSheet(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	log.Println("\nReadData", sheetName)

	columnsMapping := clit.Config("kinerjaTerminal", "columnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if strings.EqualFold(strings.TrimSpace(cellValue), "Uraian") {
			cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i+1))
			if err != nil {
				log.Fatal(err)
			}

			if !strings.EqualFold(strings.TrimSpace(cellValue), "Uraian") {
				firstDataRow = i + 2
				break
			}
		}

		i++
	}

	monthRow := firstDataRow - 2
	var monthHeaders []Header
	isHeaderDetected := false

	i = 5 //mulai kolom 5 (E)
	prevCell := ""
	for {
		header := Header{
			DBFieldName:  "",
			Column:       "",
			ColumnNumber: i,
		}

		currentCol := helpers.ToCharStr(i)
		cellText, err := f.GetCellValue(sheetName, currentCol+toolkit.ToString(monthRow))
		if err != nil {
			log.Fatal(err)
		}

		_, timeParseErr := time.Parse("Jan-06", strings.TrimSpace(cellText))
		if isHeaderDetected == false && timeParseErr == nil {
			isHeaderDetected = true
		}

		if isHeaderDetected == true && timeParseErr != nil {
			break
		}

		if isHeaderDetected {
			if timeParseErr == nil {
				if strings.TrimSpace(cellText) != strings.TrimSpace(prevCell) {
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
	var err error
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

			if key == "Realisasi" {
				header.Column = monthHeader.Column
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
					rowData.Set(header.DBFieldName, header.Value)
				} else if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, header.Value)
				} else {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")
					stringData = strings.ReplaceAll(stringData, "-", "")

					stringData = strings.TrimSpace(stringData)

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

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
				TableName: "BOD_Kinerja_Terminal",
				Data:      rowData,
			}

			err = helpers.Insert(param)
			if err != nil {
				log.Fatal("Error inserting "+monthHeader.HeaderName+", ERROR:", err.Error())
			} else {
				log.Println(monthHeader.HeaderName + ", inserted.")
			}

			rowCount++
		}
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}
