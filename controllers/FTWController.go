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

type FTWController struct {
	*Base
}

func NewFTWController() *FTWController {
	return new(FTWController)
}

func (c *FTWController) ReadExcels() error {
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

func (c *FTWController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "FTW rekap 2019") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. FTW files found:", len(resourceFiles))
	return resourceFiles
}

func (c *FTWController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	for _, sheetName := range f.GetSheetMap() {
		err = c.ReadData(f, sheetName)
		if err != nil {
			log.Println("Error reading data. ERROR:", err)
		}
	}

	if err == nil {
		toolkit.Println()
		log.Println("SUCCESS")
	}
	log.Println("Total Process Time:", time.Since(timeNow).Seconds(), "seconds")

	return err
}

func (c *FTWController) ReadData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	columnsMapping := clit.Config("ftw", "columnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "TANGGAL" {
			cellValue, err = f.GetCellValue(sheetName, "A"+toolkit.ToString(i+1))
			if err != nil {
				log.Fatal(err)
			}

			if cellValue == "TANGGAL" {
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
	var currentPeriod time.Time
	// months := clit.Config("ftw", "months", nil).([]interface{})

	isPeriodSkipped := map[time.Time]bool{}
	isPeriodChecked := map[time.Time]bool{}

	//iterate over rows
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index

		isRowEmpty := true
		skipRow := false
		for _, header := range headers {
			if header.DBFieldName == "PERIOD" {
				style, _ := f.NewStyle(`{"number_format":15}`)
				f.SetCellStyle(sheetName, header.Column+toolkit.ToString(currentRow), header.Column+toolkit.ToString(currentRow), style)
				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				var t time.Time
				if strings.TrimSpace(stringData) != "" {
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

					currentPeriod = t
				}

				rowData.Set(header.DBFieldName, currentPeriod)
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

		insertRow := func() {
			param := helpers.InsertParam{
				TableName: "F_QHSSE_FTW",
				Data:      rowData,
			}

			err = helpers.Insert(param)
			if err != nil {
				log.Fatal("Error inserting row "+toolkit.ToString(currentRow)+", ERROR:", err.Error())
			} else {
				log.Println("Row", currentRow, "inserted.")
			}
		}

		if ok := isPeriodChecked[rowData.Get("PERIOD").(time.Time)]; !ok {
			isPeriodChecked[rowData.Get("PERIOD").(time.Time)] = false
		}

		if ok := isPeriodSkipped[rowData.Get("PERIOD").(time.Time)]; !ok {
			isPeriodSkipped[rowData.Get("PERIOD").(time.Time)] = false
		}

		if isPeriodChecked[rowData.Get("PERIOD").(time.Time)] == true {
			if isPeriodSkipped[rowData.Get("PERIOD").(time.Time)] == false {
				insertRow()
			}
		} else {
			// check if data exists
			sqlQuery := "SELECT PERIOD FROM F_QHSSE_FTW WHERE trunc(period) = TO_DATE('" + rowData.Get("PERIOD").(time.Time).Format("2006-01-02") + "', 'YYYY-MM-DD')"

			conn := helpers.Database()
			cursor := conn.Cursor(dbflex.From("F_QHSSE_FTW").SQL(sqlQuery), nil)
			defer cursor.Close()

			res := make([]toolkit.M, 0)
			err = cursor.Fetchs(&res, 0)
			if err != nil {
				log.Println(err)
			}

			isPeriodChecked[rowData.Get("PERIOD").(time.Time)] = true
			//only insert if len of datas in currentPeriod is 0 / if no data yet
			if len(res) == 0 {
				insertRow()
			} else {
				log.Println("Skipping", rowData.Get("PERIOD").(time.Time).Format("2006-01-02")+".")
				isPeriodSkipped[rowData.Get("PERIOD").(time.Time)] = true
			}
		}

		rowCount++
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *FTWController) selectItemID(param SqlQueryParam) error {
	sqlQuery := "SELECT * FROM D_Item WHERE ITEM_NAME = '" + param.ItemName + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From("D_Item").SQL(sqlQuery), nil)
	defer cursor.Close()

	err := cursor.Fetchs(param.Results, 0)

	return err
}
