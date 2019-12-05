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

type InduksiController struct {
	*Base
}

func NewInduksiController() *InduksiController {
	return new(InduksiController)
}

func (c *InduksiController) ReadExcels() error {
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

func (c *InduksiController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "induksi K3L") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. Induksi files found:", len(resourceFiles))
	return resourceFiles
}

func (c *InduksiController) readExcel(filename string) error {
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

func (c *InduksiController) ReadData(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	months := clit.Config("induksi", "months", []interface{}{}).([]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "No" {
			cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i+1))
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
		cellValueBulan, err := f.GetCellValue(sheetName, col+rowBulan)
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

		//iterate over rows
		for j := 0; j < 10; j++ {
			obj := toolkit.M{}
			currentRow := firstDataRow + j

			obj.Set("PERIOD", period)

			//mengambil Value di kolom
			cellValueJenisInduksi, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			cellValue, err := f.GetCellValue(sheetName, col+toolkit.ToString(currentRow))
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
