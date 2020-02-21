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

// PemenuhanSDMController is a controller for every kind of PemenuhanSDM files.
type PemenuhanSDMController struct {
	*Base
}

// New is used to initiate the controller
func (c *PemenuhanSDMController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for PemenuhanSDM files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *PemenuhanSDMController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "PEMENUHAN SDM NOVEMBER 2019 (INTERNAL) - Contoh buat BI")
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *PemenuhanSDMController) ReadExcel() error {
	var err error

	for _, sheetName := range c.Engine.GetSheetMap() {
		if sheetName == "DETAIL" {
			c.ReadSheet(c.ReadData, sheetName)
		}
	}

	return err
}

func (c *PemenuhanSDMController) ReadData(sheetName string) error {
	timeNow := time.Now()

	log.Println("Deleting datas.")

	sql := "DELETE FROM F_HC_POSITION"

	conn := helpers.Database()
	query, err := conn.Prepare(dbflex.From("F_HC_POSITION").SQL(sql))
	if err != nil {
		log.Println(err)
	}

	_, err = query.Execute(toolkit.M{}.Set("data", toolkit.M{}))
	if err != nil {
		log.Println(err)
	}

	log.Println("Data deleted.")

	toolkit.Println()
	log.Println("ReadData", sheetName)
	columnsMapping := clit.Config("pemenuhansdm", "columnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "NO" {
			cellValue, err = c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i+1))
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

		stringData, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		//check if value is a SUB_DIR
		if strings.TrimSpace(stringData) != "" {
			stringSubDir, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			if strings.TrimSpace(stringSubDir) != "" {
				currentSubDir = stringSubDir
			}

			continue
		}

		stringB, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
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
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}
					stringData = strings.ReplaceAll(stringData, "'", "''")
					stringData = strings.ReplaceAll(stringData, "-", "")

					if strings.TrimSpace(stringData) != "" {
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
