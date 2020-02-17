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

type CorsecController struct {
	*Base
}

func NewCorsecController() *CorsecController {
	return new(CorsecController)
}

func (c *CorsecController) ReadExcels() error {
	for _, file := range c.FetchFiles() {
		err := c.readExcel(file)
		if err == nil {
			// move file if read succeeded
			c.MoveToArchive(file)
			log.Println("Done.")
		} else {
			return err
		}
	}

	return nil
}

func (c *CorsecController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "RKM") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. CORSEC files found:", len(resourceFiles))
	return resourceFiles
}

func (c *CorsecController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	for _, sheetName := range f.GetSheetMap() {
		if strings.Contains(sheetName, "Usulan RKM") {
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

func (c *CorsecController) ReadData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	columnsMapping := clit.Config("corsec", "columnsMapping", nil).(map[string]interface{})

	dataFound := false
	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i))
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

		styleID, err := f.GetCellStyle(sheetName, "D"+toolkit.ToString(currentRow))
		if err != nil {
			toolkit.Println("1")
			log.Fatal(err)
		}
		fillID := f.Styles.CellXfs.Xf[styleID].FillID
		fgColor := f.Styles.Fills.Fill[fillID].PatternFill.FgColor

		color := fgColor.RGB
		number, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
		if err != nil {
			toolkit.Println("2")
			log.Fatal(err)
		}

		if fgColor.Theme != nil && number == "" {
			srgbClr := f.Theme.ThemeElements.ClrScheme.Children[*fgColor.Theme].SrgbClr.Val
			color = excelize.ThemeColor(srgbClr, fgColor.Tint)
		}

		if color == "FFEAF1DD" || color == "FF00B0F0" {
			newCategory, err := f.GetCellValue(sheetName, "D"+toolkit.ToString(currentRow))
			if err != nil {
				toolkit.Println("3")
				log.Fatal(err)
			}

			if newCategory != currentCategory {
				currentArea = ""
			}

			currentCategory = newCategory

			continue
		} else if color == "FFFFF2CC" {
			currentArea, err = f.GetCellValue(sheetName, "D"+toolkit.ToString(currentRow))
			if err != nil {
				toolkit.Println("4")
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

				filename := trimSuffix(filepath.Base(f.Path), filepath.Ext(f.Path))
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
					isProses, err := f.GetCellValue(sheetName, "AA"+toolkit.ToString(row))
					if err != nil {
						log.Fatal(err)
					}

					isSelesai, err := f.GetCellValue(sheetName, "AB"+toolkit.ToString(row))
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
			break
		}

		toolkit.Println(rowData)
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
