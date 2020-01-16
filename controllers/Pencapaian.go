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

type PencapaianController struct {
	*Base
}

func NewPencapaianController() *PencapaianController {
	return new(PencapaianController)
}

func (c *PencapaianController) ReadExcels() error {
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

func (c *PencapaianController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(strings.ToUpper(filepath.Base(file)), strings.ToUpper("PENCAPAIAN")) {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. Pencapaian files found:", len(resourceFiles))
	return resourceFiles
}

func (c *PencapaianController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	for _, sheetName := range f.GetSheetMap() {
		if strings.Contains(strings.ToUpper(sheetName), strings.ToUpper("REKAP KONSOL")) {
			err = c.ReadData(f, sheetName)
			if err != nil {
				log.Println("Error reading data. ERROR:", err)
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

func (c *PencapaianController) ReadData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	columnsMapping := clit.Config("rekapKonsol2", "columnsMapping", nil).(map[string]interface{})

	firstDataRow := 53

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
	no := 1
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index

		if currentRow > 57 {
			break
		}

		isRowEmpty := true
		for _, header := range headers {
			if header.DBFieldName == "No" {
				rowData.Set(header.DBFieldName, no)
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
			continue
		}

		param := helpers.InsertParam{
			TableName: "Rekap_Konsol2",
			Data:      rowData,
		}

		err = helpers.Insert(param)
		if err != nil {
			log.Fatal("Error inserting row "+toolkit.ToString(currentRow)+", ERROR:", err.Error())
		} else {
			log.Println("Row", currentRow, "inserted.")
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
