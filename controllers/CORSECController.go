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

// CorsecController is a controller for every kind of CORSEC files.
type CorsecController struct {
	*Base
}

// New is used to initiate the controller
func (c *CorsecController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for CORSEC files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *CorsecController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "RKM")
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *CorsecController) ReadExcel() {
	for _, sheetName := range c.Engine.GetSheetMap() {
		if strings.Contains(sheetName, "Usulan RKM") {
			c.ReadSheet(c.ReadData, sheetName)
		}
	}
}

func (c *CorsecController) ReadData(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	columnsMapping := clit.Config("corsec", "columnsMapping", nil).(map[string]interface{})

	dataFound := false
	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "NO" {
			dataFound = true
			firstDataRow = i + 1
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
	currentCategory := ""
	currentArea := ""

	//iterate over rows
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index

		styleID, err := c.Engine.GetCellStyle(sheetName, "D"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		fillID := c.Engine.GetFillID(styleID)

		number, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		fgColorTheme := c.Engine.GetFgColorTheme(fillID)
		fgColorRGB := c.Engine.GetFgColorRGB(fillID)
		fgColorTint := c.Engine.GetFgColorTint(fillID)

		if fgColorTheme != nil && number == "" {
			srgbClr := c.Engine.GetSrgbClrVal(fgColorTheme)
			fgColorRGB = excelize.ThemeColor(srgbClr, fgColorTint)
		}

		if fgColorRGB == "FFEAF1DD" || fgColorRGB == "FF00B0F0" {
			newCategory, err := c.Engine.GetCellValue(sheetName, "D"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			if newCategory != currentCategory {
				currentArea = ""
			}

			currentCategory = newCategory

			continue
		} else if fgColorRGB == "FFFFF2CC" {
			currentArea, err = c.Engine.GetCellValue(sheetName, "D"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			continue
		}

		isRowEmpty := true
		for _, header := range headers {
			if header.DBFieldName == "PERIOD" {
				trimSuffix := func(s, suffix string) string {
					if strings.HasSuffix(s, suffix) {
						s = s[:len(s)-len(suffix)]
					}
					return s
				}

				filename := trimSuffix(filepath.Base(c.Engine.GetExcelPath()), filepath.Ext(c.Engine.GetExcelPath()))
				splittedFilename := strings.Split(filename, " ")

				year := splittedFilename[len(splittedFilename)-3]
				month := splittedFilename[len(splittedFilename)-1]

				t, err := time.Parse("02/January/2006", "01/"+month+"/"+year)
				if err != nil {
					log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
				}

				rowData.Set(header.DBFieldName, t)
			} else if header.DBFieldName == "STATUS" {
				var getStringData func(row int) string

				getStringData = func(row int) string {
					isProses, err := c.Engine.GetCellValue(sheetName, "AA"+toolkit.ToString(row))
					if err != nil {
						log.Fatal(err)
					}

					isSelesai, err := c.Engine.GetCellValue(sheetName, "AB"+toolkit.ToString(row))
					if err != nil {
						log.Fatal(err)
					}

					stringData := ""
					if strings.TrimSpace(isProses) != "" {
						stringData = "proses"
					}
					if strings.TrimSpace(isSelesai) != "" {
						stringData = "selesai"
					}

					if strings.TrimSpace(stringData) == "" {
						stringData = getStringData(row - 1)
					}

					return stringData
				}

				stringData := strings.ReplaceAll(getStringData(currentRow), "'", "''")

				if len(stringData) > 300 {
					stringData = stringData[0:300]
				}

				rowData.Set(header.DBFieldName, stringData)
			} else if header.DBFieldName == "CATEGORY" {
				rowData.Set(header.DBFieldName, currentCategory)
			} else if header.DBFieldName == "AREA" {
				rowData.Set(header.DBFieldName, currentArea)
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
			TableName: "F_CORSEC_RKM",
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
