package controllers

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/sebar/dbflex"
)

// EquipmentPerformance10STSController is a controller for every kind of EquipmentPerformance10STS files.
type EquipmentPerformance10STSController struct {
	*Base
}

// New is used to initiate the controller
func (c *EquipmentPerformance10STSController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for EquipmentPerformance10STS files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *EquipmentPerformance10STSController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "1. Equipment Performence 10 UNIT STS")
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *EquipmentPerformance10STSController) ReadExcel() {
	for _, sheetName := range c.Engine.GetSheetMap() {
		if strings.EqualFold(sheetName, "EQUIPMENT PERFORMANCE") {
			c.ReadSheet(c.ReadData, sheetName)
		}
	}
}

func (c *EquipmentPerformance10STSController) ReadData(sheetName string) error {
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
		stringData, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(row))
		if err != nil {
			helpers.HandleError(err)
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
			stringDataAfter, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(row+1))
			if err != nil {
				helpers.HandleError(err)
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

		stringData, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
		if err != nil {
			helpers.HandleError(err)
		}

		if strings.Contains(stringData, "Total") {
			break
		}

		for _, header := range headers {
			if header.DBFieldName == "PERIOD" {
				rowData.Set(header.DBFieldName, currentPeriod)
			} else if header.DBFieldName == "ITEM_ID" {
				stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					helpers.HandleError(err)
				}

				resultRows := make([]toolkit.M, 0)
				param := SqlQueryParam{
					ItemName: strings.ReplaceAll(stringData, "-", ""),
					Results:  &resultRows,
				}

				err = c.selectItemID(param)
				if err != nil {
					helpers.HandleError(err)
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
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						helpers.HandleError(err)
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
			helpers.HandleError(err)
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
