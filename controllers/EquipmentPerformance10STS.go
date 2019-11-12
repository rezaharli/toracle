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

type EquipmentPerformance10STSController struct {
	*Base
}

func NewEquipmentPerformance10STSController() *EquipmentPerformance10STSController {
	return new(EquipmentPerformance10STSController)
}

func (c *EquipmentPerformance10STSController) ReadExcels() error {
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

func (c *EquipmentPerformance10STSController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "1. Equipment Performence 10 UNIT STS") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. EquipmentPerformance10STS files found:", len(resourceFiles))
	return resourceFiles
}

func (c *EquipmentPerformance10STSController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	for _, sheetName := range f.GetSheetMap() {
		if strings.EqualFold(sheetName, "EQUIPMENT PERFORMANCE") {
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

func (c *EquipmentPerformance10STSController) ReadData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("equipmentPerformance", "10STS", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})
	months := config["months"].([]interface{})

	var currentPeriod time.Time

	var err error

	//iterate into groups of data
	row := 1
	firstDataRow := 0
	for {
		stringData, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(row))
		if err != nil {
			log.Fatal(err)
		}

		splitted := strings.Split(stringData, " ")
		monthYear := ""
		if len(splitted) >= 2 {
			monthYear = "1-" + splitted[len(splitted)-2] + "-" + splitted[len(splitted)-1]
		}
		if len(splitted) >= 2 {
			month := toolkit.ToString(helpers.IndexOf(splitted[len(splitted)-2], months) + 1)
			monthYear = "1-" + month + "-" + splitted[len(splitted)-1]
		}

		//check if value is a period
		t, err := time.Parse("2-1-2006", monthYear)
		if err == nil {
			currentPeriod = t
		}

		if strings.EqualFold(strings.TrimSpace(stringData), strings.TrimSpace("NO")) {
			stringDataAfter, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(row+1))
			if err != nil {
				log.Fatal(err)
			}

			if strings.EqualFold(strings.TrimSpace(stringDataAfter), strings.TrimSpace("NO")) {
				row++
				continue
			}

			firstDataRow = row + 1
			break
		}

		row++
	}

	var headers []Header
	for key, column := range columnsMapping {
		header := Header{
			DBFieldName: key,
			Column:      column.(string),
		}

		headers = append(headers, header)
	}

	rowCount := 0
	//iterate over rows
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index
		row = currentRow
		isRowEmpty := true

		stringData, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		if strings.Contains(stringData, "Total") {
			break
		}

		for _, header := range headers {
			if header.DBFieldName == "PERIOD" {
				rowData.Set(header.DBFieldName, currentPeriod)
			} else if header.DBFieldName == "ITEM_ID" {
				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				resultRows := make([]toolkit.M, 0)
				param := SqlQueryParam{
					ItemName: strings.ReplaceAll(stringData, "-", ""),
					Results:  &resultRows,
				}

				err = c.selectItemID(param)
				if err != nil {
					log.Fatal(err)
				}

				if stringData != "" {
					isRowEmpty = false
				}

				if len(resultRows) > 0 {
					rowData.Set(header.DBFieldName, resultRows[0].GetString("ITEM_ID"))
				} else {
					rowData.Set(header.DBFieldName, nil)
				}
			} else {
				if header.Column != "" {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.TrimSpace(strings.ReplaceAll(stringData, "'", "''"))

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					if stringData != "" {
						isRowEmpty = false
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}
		}

		if isRowEmpty {
			break
		}

		param := helpers.InsertParam{
			TableName: "F_ENG_EQUIPMENT_MONTHLY",
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

	row++

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
		toolkit.Println()
	}

	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *EquipmentPerformance10STSController) selectItemID(param SqlQueryParam) error {
	sqlQuery := "SELECT * FROM D_Item WHERE ITEM_NAME = '" + param.ItemName + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From("D_Item").SQL(sqlQuery), nil)
	defer cursor.Close()

	err := cursor.Fetchs(param.Results, 0)

	return err
}
