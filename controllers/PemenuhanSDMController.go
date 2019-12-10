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
	"git.eaciitapp.com/sebar/dbflex"
)

type PemenuhanSDMController struct {
	*Base
}

func NewPemenuhanSDMController() *PemenuhanSDMController {
	return new(PemenuhanSDMController)
}

func (c *PemenuhanSDMController) ReadExcels() error {
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

func (c *PemenuhanSDMController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "PEMENUHAN SDM NOVEMBER 2019 (INTERNAL) - Contoh buat BI") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. PemenuhanSDM files found:", len(resourceFiles))
	return resourceFiles
}

func (c *PemenuhanSDMController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	for _, sheetName := range f.GetSheetMap() {
		if sheetName == "DETAIL" {
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

func (c *PemenuhanSDMController) ReadData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	columnsMapping := clit.Config("pemenuhansdm", "columnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "NO" {
			cellValue, err = f.GetCellValue(sheetName, "A"+toolkit.ToString(i+1))
			if err != nil {
				log.Fatal(err)
			}

			if cellValue == "NO" {
				i++
				continue
			} else {
				firstDataRow = i + 1

				break
			}
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
	rowCount := 0
	emptyRowCount := 0
	var currentSubDir string
	// months := clit.Config("pemenuhansdm", "months", nil).([]interface{})

	//iterate over rows
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index

		isRowEmpty := true
		skipRow := false

		stringData, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		//check if value is a SUB_DIR
		if strings.TrimSpace(stringData) != "" {
			stringSubDir, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			if strings.TrimSpace(stringSubDir) != "" {
				currentSubDir = stringSubDir
			}

			continue
		}

		stringB, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		if currentSubDir == "" || strings.Contains(stringB, "SUB JUMLAH") {
			continue
		}

		for _, header := range headers {
			if header.DBFieldName == "SUB_DIR" {
				rowData.Set(header.DBFieldName, currentSubDir)
			} else {
				if header.Column == "" {
					rowData.Set(header.DBFieldName, "")
				} else {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}
					stringData = strings.ReplaceAll(stringData, "'", "''")
					stringData = strings.ReplaceAll(stringData, "-", "")

					if stringData != "" {
						isRowEmpty = false
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}
		}

		if isRowEmpty {
			emptyRowCount++

			if emptyRowCount >= 10 {
				break
			}

			continue
		}

		if skipRow {
			continue
		}

		// toolkit.Println(currentRow, rowData)
		param := helpers.InsertParam{
			TableName: "F_HC_POSITION",
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

func (c *PemenuhanSDMController) selectItemID(param SqlQueryParam) error {
	sqlQuery := "SELECT * FROM D_Item WHERE ITEM_NAME = '" + param.ItemName + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From("D_Item").SQL(sqlQuery), nil)
	defer cursor.Close()

	err := cursor.Fetchs(param.Results, 0)

	return err
}
