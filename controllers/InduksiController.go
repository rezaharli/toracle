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

// InduksiController is a controller for every kind of Induksi files.
type InduksiController struct {
	*Base
}

// New is used to initiate the controller
func (c *InduksiController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for Induksi files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *InduksiController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "induksi K3L")
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *InduksiController) ReadExcel() error {
	var err error

	for _, sheetName := range c.Engine.GetSheetMap() {
		c.ReadSheet(c.ReadData, sheetName)
	}

	return err
}

func (c *InduksiController) ReadData(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	months := clit.Config("induksi", "months", []interface{}{}).([]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "No" {
			cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i+1))
			if err != nil {
				log.Fatal(err)
			}

			if cellValue == "No" {
				i++
				continue
			} else {
				firstDataRow = i + 1

				break
			}
		}
		i++
	}

	objs := make([]toolkit.M, 0)

	var err error
	for i := 0; true; i++ {
		//iterate col
		col := "C"
		colInNumber := helpers.CharStrToNum(col)
		colInNumber = colInNumber + i
		col = helpers.ToCharStr(colInNumber)

		rowBulan := toolkit.ToString(firstDataRow - 1)

		//mengambil BULAN
		cellValueBulan, err := c.Engine.GetCellValue(sheetName, col+rowBulan)
		if err != nil {
			log.Fatal(err)
		}

		if cellValueBulan == "" {
			break
		}

		sheetnameSplitted := strings.Split(sheetName, " ")
		tahun := sheetnameSplitted[len(sheetnameSplitted)-1]

		period, err := time.Parse("2-1-2006", "1-"+toolkit.ToString(helpers.IndexOf(cellValueBulan, months)+1)+"-"+tahun)
		if err != nil {
			log.Fatal(err)
		}

		emptyCount := 0

		// check if data exists
		sqlQuery := "SELECT PERIOD FROM F_QHSSE_INDUKSI WHERE trunc(period) = TO_DATE('" + period.Format("2006-01-02") + "', 'YYYY-MM-DD')"

		conn := helpers.Database()
		cursor := conn.Cursor(dbflex.From("F_QHSSE_INDUKSI").SQL(sqlQuery), nil)
		defer cursor.Close()

		res := make([]toolkit.M, 0)
		err = cursor.Fetchs(&res, 0)
		if err != nil {
			log.Println(err)
		}

		//only insert if len of datas in currentPeriod is 0 / if no data yet
		if len(res) == 0 {
			//iterate over rows
			for j := 0; j < 10; j++ {
				obj := toolkit.M{}
				currentRow := firstDataRow + j

				obj.Set("PERIOD", period)

				//mengambil Value di kolom
				cellValueJenisInduksi, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				cellValue, err := c.Engine.GetCellValue(sheetName, col+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				if cellValueJenisInduksi == "" {
					emptyCount++
					continue
				}

				obj.Set("JENIS_INDUKSI", cellValueJenisInduksi)
				obj.Set("JUMLAH_INDUKSI", cellValue)

				objs = append(objs, obj)

				if emptyCount >= 10 {
					break
				}
			}
		} else {
			log.Println("Skipping", period.Format("2006-01-02")+".")
		}
	}

	for _, obj := range objs {
		param := helpers.InsertParam{
			TableName: "F_QHSSE_INDUKSI",
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
