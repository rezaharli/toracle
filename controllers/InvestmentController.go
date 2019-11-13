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

type InvestmentController struct {
	*Base
}

func NewInvestmentController() *InvestmentController {
	return new(InvestmentController)
}

func (c *InvestmentController) ReadExcels() error {
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

func (c *InvestmentController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "laporan investasi") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. Investment files found:", len(resourceFiles))
	return resourceFiles
}

func (c *InvestmentController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	for _, sheetName := range f.GetSheetMap() {
		if !strings.Contains(sheetName, "LAP INVESTASI ENG") {
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

func (c *InvestmentController) ReadData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	columnsMapping := clit.Config("investment", "columnsMapping", nil).(map[string]interface{})

	dataFound := false
	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "NO" {
			dataFound = true
			firstDataRow = i + 2
		} else {
			if dataFound == true {
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
	// var rowDatas []toolkit.M
	rowCount := 0
	emptyRowCount := 0
	var currentAktiva, currentCategory string
	months := clit.Config("investment", "months", nil).([]interface{})

	//iterate over rows
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index
		isAktiva, isCategory, isProjectName := false, false, false

		number, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		codingMask, err := f.GetCellValue(sheetName, "C"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		namaAktiva, err := f.GetCellValue(sheetName, "D"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		//reset category cekno nemu category baru setiap nemu jumlah
		if strings.Contains(strings.ToUpper(namaAktiva), strings.ToUpper("jumlah")) {
			currentCategory = ""
			continue
		}

		//menentukan apakah aktiva/category/projectname
		if number != "" && codingMask == "" {
			isAktiva = true
		} else {
			isKananEmpty := true
			skipHeaderCheck := []interface{}{"PERIOD", "PROJECT_NAME", "CATEGORY", "AKTIVA"}

			for _, header := range headers {
				if helpers.IndexOf(header.DBFieldName, skipHeaderCheck) != -1 {
					continue
				}

				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				if stringData != "" {
					isKananEmpty = false
				}
			}

			if isKananEmpty == true && currentAktiva != "" && currentCategory == "" {
				isCategory = true
			} else {
				if currentAktiva != "" && currentCategory != "" {
					isProjectName = true
				}
			}
		}

		isRowEmpty := true
		skipRow := false
		for _, header := range headers {
			if header.DBFieldName == "PERIOD" {
				splittedFilename := strings.Split(f.Path, " ")
				year := splittedFilename[len(splittedFilename)-7]

				splitted := strings.Split(sheetName, " ")
				stringDataMonth := splitted[len(splitted)-1]

				stringData := "1/" + toolkit.ToString(helpers.IndexOf(stringDataMonth, months)+1) + "/" + year

				var t time.Time
				if stringDataMonth != "" {
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
			} else if header.DBFieldName == "AKTIVA" {
				if isAktiva {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}
					stringData = strings.ReplaceAll(stringData, "'", "''")

					if stringData != "" {
						isRowEmpty = false
					}

					currentAktiva = stringData
					skipRow = true
					break
				} else if !isCategory && !isProjectName {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}
					stringData = strings.ReplaceAll(stringData, "'", "''")

					if stringData != "" {
						isRowEmpty = false
					}

					skipRow = true
				}
			} else if header.DBFieldName == "CATEGORY" {
				if isCategory {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}
					stringData = strings.ReplaceAll(stringData, "'", "''")

					if stringData != "" {
						isRowEmpty = false
					}

					currentCategory = stringData
					skipRow = true
					break
				} else if !isAktiva && !isProjectName {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}
					stringData = strings.ReplaceAll(stringData, "'", "''")

					if stringData != "" {
						isRowEmpty = false
					}

					skipRow = true
				}
			} else if header.DBFieldName == "PROJECT_NAME" {
				if isProjectName {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}
					stringData = strings.ReplaceAll(stringData, "'", "''")

					if stringData != "" {
						isRowEmpty = false
					}

					rowData.Set("AKTIVA", currentAktiva)
					rowData.Set("CATEGORY", currentCategory)
					rowData.Set(header.DBFieldName, stringData)
				} else if !isAktiva && !isCategory {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}
					stringData = strings.ReplaceAll(stringData, "'", "''")

					if stringData != "" {
						isRowEmpty = false
					}

					skipRow = true
				}
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

		if isRowEmpty {
			emptyRowCount++

			if emptyRowCount >= 10 {
				break
			}
		}

		if skipRow {
			continue
		}

		param := helpers.InsertParam{
			TableName: "F_CORSEC_INVESTMENT",
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
