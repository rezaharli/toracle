package controllers

import (
	"log"
	"os"
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
			helpers.MoveToArchive(file)
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
				log.Println("Error reading monthly data. ERROR:", err)
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

	firstDataRow := 0
	i := 1
	for {
		if f.GetCellValue(sheetName, "A"+toolkit.ToString(i)) == "NO" {
			firstDataRow = i + 1
			break
		}
		i++
	}

	headerRow := toolkit.ToString(firstDataRow - 1)

	var headers []Header
	for key, column := range columnsMapping {
		isHeaderDetected := false
		i = 1

		header := Header{
			DBFieldName: key,
			HeaderName:  "",
			Column:      "",
			Row:         "",
		}

		// search for particular header in excel
		for {
			currentCol := helpers.ToCharStr(i)
			cellText := f.GetCellValue(sheetName, currentCol+headerRow)

			if isHeaderDetected == false && strings.TrimSpace(cellText) != "" {
				isHeaderDetected = true
			}

			if isHeaderDetected == true && strings.TrimSpace(cellText) == "" {
				//kalo header ga nemu coba sekali lagi mbok bilih di atasnya
				cellText = f.GetCellValue(sheetName, currentCol+toolkit.ToString(toolkit.ToInt(headerRow, "")-1))
				if strings.TrimSpace(cellText) == "" {
					cellText = f.GetCellValue(sheetName, currentCol+toolkit.ToString(toolkit.ToInt(headerRow, "")-2))
					if strings.TrimSpace(cellText) == "" {
						break
					}
				}
			}

			if isHeaderDetected {
				if strings.Replace(column.(string), " ", "", -1) == strings.Replace(cellText, " ", "", -1) {
					header.HeaderName = cellText
					header.Column = currentCol
					header.Row = headerRow

					break
				}
			}

			i++
		}

		headers = append(headers, header)
	}

	toolkit.Println(headers)
	var err error
	// var rowDatas []toolkit.M
	rowCount := 0
	//iterate over rows
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index

		isRowEmpty := true
		for _, header := range headers {
			if header.DBFieldName == "PERIOD" {
				// style, _ := f.NewStyle(`{"number_format":15}`)
				// f.SetCellStyle(sheetName, header.Column+toolkit.ToString(currentRow), header.Column+toolkit.ToString(currentRow), style)
				// stringData := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				// stringData = strings.ReplaceAll(stringData, "'", "")

				// var t time.Time
				// if stringData != "" {
				// 	isRowEmpty = false
				// 	t, err = time.Parse("2-Jan-06", stringData)
				// 	if err != nil {
				// 		t, err = time.Parse("02/01/2006", stringData)
				// 		if err != nil {
				// 			log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
				// 		}
				// 	}
				// }

				// rowData.Set(header.DBFieldName, t)
			} else {
				styleID := f.GetCellStyle(sheetName, "D8")
				fillID := f.Styles.CellXfs.Xf[styleID].FillID
				fgColor := f.Styles.Fills.Fill[fillID].PatternFill.FgColor

				toolkit.Println(f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow)))
				toolkit.Println(sheetName, "D8", fgColor, fgColor.RGB)
				// if fgColor.Theme != nil {
				// 	f.
				// 	srgbClr := f.Theme.ThemeElements.ClrScheme.Children[*fgColor.Theme].SrgbClr.Val
				// 	return excelize.ThemeColor(srgbClr, fgColor.Tint)
				// }
				// return fgColor.RGB

				os.Exit(100)
				stringData := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
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
			TableName: "F_CORSEC_INCIDENT",
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
