package controllers

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
)

// RKAPController is a controller for every kind of RKAP files.
type RKAPController struct {
	*Base
}

// New is used to initiate the controller
func (c *RKAPController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for RKAP files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *RKAPController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "MASTER RKAP PRODUKSI TTL 2019 - Arahan BOC 1 - Rapat Teknis Bahas SM ubah kurs")
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *RKAPController) ReadExcel() error {
	var err error

	for _, sheetName := range c.Engine.GetSheetMap() {
		if sheetName == "Arus Rinci (N)" {
			c.ReadSheet(c.ReadData, sheetName)
		}
	}

	return err
}

func (c *RKAPController) ReadData(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	months := clit.Config("rkap", "months", nil).([]interface{})

	tahunCell, err := c.Engine.GetCellValue(sheetName, "S5")
	if err != nil {
		log.Fatal(err)
	}

	splitted := strings.Split(tahunCell, " ")
	currentTahun := splitted[len(splitted)-1]

	objs := make([]toolkit.M, 0)
	for i := 0; true; i++ {
		//iterate col
		col := "S"
		colInNumber := helpers.CharStrToNum(col)
		colInNumber = colInNumber + i
		col = helpers.ToCharStr(colInNumber)

		rowBulan := "6"

		//mengambil bulan
		cellValueBulan, err := c.Engine.GetCellValue(sheetName, col+rowBulan)
		if err != nil {
			log.Fatal(err)
		}

		if cellValueBulan == "" {
			break
		}

		if (helpers.IndexOf(cellValueBulan, months)) < 0 { //yen dudu bulan
			continue //bukan bulan
		}

		DICs := toolkit.M{}
		DICs.Set("D", toolkit.M{}.Set("BOX", "180").Set("TEUS", "181").Set("GT", "24").Set("UNIT", "23"))
		DICs.Set("I", toolkit.M{}.Set("BOX", "119").Set("TEUS", "120").Set("GT", "22").Set("UNIT", "21"))
		DICs.Set("C", toolkit.M{}.Set("CUKER", "38").Set("GT", "26").Set("UNIT", "25"))

		for dic, kinds := range DICs {
			obj := toolkit.M{}

			obj.Set("TAHUN", currentTahun)
			obj.Set("BULAN", helpers.IndexOf(cellValueBulan, months)+1)
			obj.Set("D_I_C", dic)

			for kind, row := range kinds.(toolkit.M) {
				cellValue, err := c.Engine.GetCellValue(sheetName, col+row.(string))
				if err != nil {
					log.Fatal(err)
				}

				obj.Set(kind, cellValue)
			}

			objs = append(objs, obj)
		}
	}

	for _, obj := range objs {
		param := helpers.InsertParam{
			TableName: "F_CBD_RKAP",
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
