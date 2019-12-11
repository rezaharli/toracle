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

type RKAPController struct {
	*Base
}

func NewRKAPController() *RKAPController {
	return new(RKAPController)
}

func (c *RKAPController) ReadExcels() error {
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

func (c *RKAPController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "MASTER RKAP PRODUKSI TTL 2019 - Arahan BOC 1 - Rapat Teknis Bahas SM ubah kurs") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. RKAP files found:", len(resourceFiles))
	return resourceFiles
}

func (c *RKAPController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	for _, sheetName := range f.GetSheetMap() {
		if sheetName == "Arus Rinci (N)" {
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

func (c *RKAPController) ReadData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	months := clit.Config("rkap", "months", nil).([]interface{})

	tahunCell, err := f.GetCellValue(sheetName, "S5")
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
		cellValueBulan, err := f.GetCellValue(sheetName, col+rowBulan)
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
				cellValue, err := f.GetCellValue(sheetName, col+row.(string))
				if err != nil {
					log.Fatal(err)
				}

				obj.Set(kind, cellValue)
			}

			toolkit.Println(obj)
			objs = append(objs, obj)
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
