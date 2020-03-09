package controllers

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/sebar/dbflex"
)

// RekapPetikemasController is a controller for every kind of RekapPetikemas files.
type RekapPetikemasController struct {
	*Base
}

// New is used to initiate the controller
func (c *RekapPetikemasController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for RekapPetikemas files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *RekapPetikemasController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "Rekap Petikemas Perak sd September 2019")
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *RekapPetikemasController) ReadExcel() {
	for _, sheetName := range c.Engine.GetSheetMap() {
		if strings.EqualFold(sheetName, "Sheet1") {
			c.ReadSheet(c.ReadData, sheetName)
		}
	}
}

func (c *RekapPetikemasController) ReadData(sheetName string) error {
	timeNow := time.Now()

	log.Println("Deleting datas.")

	sql := "DELETE FROM F_CBD_MARKET_SHARE_CTR"

	conn := helpers.Database()
	query, err := conn.Prepare(dbflex.From("F_CBD_MARKET_SHARE_CTR").SQL(sql))
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
	// columnsMapping := clit.Config("petikemas", "columnsMapping", nil).(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			helpers.HandleError(err)
		}

		if cellValue == "No" {
			cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i+1))
			if err != nil {
				helpers.HandleError(err)
			}

			if cellValue == "No" {
				i++
				continue
			} else {
				firstDataRow = i + 2

				break
			}
		}
		i++
	}

	objs := make([]toolkit.M, 0)

	for i := 0; true; i++ {
		//iterate col
		col := "C"
		colInNumber := helpers.CharStrToNum(col)
		colInNumber = colInNumber + i
		col = helpers.ToCharStr(colInNumber)

		rowTahun := "5"
		rowTerminal := "3"

		//mengambil tahun
		cellValueTahun, err := c.Engine.GetCellValue(sheetName, col+rowTahun)
		if err != nil {
			helpers.HandleError(err)
		}

		//mengambil terminal
		cellValueTerminal, err := c.Engine.GetCellValue(sheetName, col+rowTerminal)
		if err != nil {
			helpers.HandleError(err)
		}

		if cellValueTahun == "" && cellValueTerminal == "" {
			break
		}

		intVal, err := strconv.Atoi(cellValueTahun)
		if err != nil { //yen dudu tahun
			continue //bukan tahun
		}

		emptyCount := 0

		obj := toolkit.M{}

		//iterate over rows
		for j := 0; j < 10; j++ {
			currentRow := firstDataRow + j

			obj.Set("TAHUN", intVal)
			obj.Set("TERMINAL", cellValueTerminal)

			//mengambil Value di kolom
			cellValueIntDom, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
			if err != nil {
				helpers.HandleError(err)
			}

			cellValue, err := c.Engine.GetCellValue(sheetName, col+toolkit.ToString(currentRow))
			if err != nil {
				helpers.HandleError(err)
			}

			cellValueSat, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
			if err != nil {
				helpers.HandleError(err)
			}

			if cellValueSat == "BOX" {
				obj.Set("BOX", cellValue)
			} else if cellValueSat == "TEUS" {
				obj.Set("TEUS", cellValue)
			}

			for k := currentRow; k > 0; k-- {
				if strings.TrimSpace(cellValueIntDom) == "" {
					continue
				} else {
					obj.Set("DOM_INT", cellValueIntDom)
					break
				}
			}

			if cellValue == "" && cellValueIntDom == "" {
				objs = append(objs, obj)
				obj = toolkit.M{}

				emptyCount++
				continue
			}

			if emptyCount >= 10 {
				break
			}
		}
	}

	for _, obj := range objs {
		param := helpers.InsertParam{
			TableName: "F_CBD_MARKET_SHARE_CTR",
			Data:      obj,
		}

		err = helpers.Insert(param)
		if err != nil {
			log.Println("Error inserting row, ERROR:", err.Error())
		} else {
			log.Println("Row inserted.")
		}
	}

	if err == nil {
		log.Println("SUCCESS Processing", len(objs), "rows")
	}

	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *RekapPetikemasController) selectItemID(param SqlQueryParam) error {
	sqlQuery := "SELECT * FROM D_Item WHERE ITEM_NAME = '" + param.ItemName + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From("D_Item").SQL(sqlQuery), nil)
	defer cursor.Close()

	err := cursor.Fetchs(param.Results, 0)

	return err
}
