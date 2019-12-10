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
		if sheetName == "KK Internasional" || sheetName == "KK Domestik" {
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
	// columnsMapping := clit.Config("rkap", "columnsMapping", nil).(map[string]interface{})
	months := clit.Config("rkap", "months", nil).([]interface{})

	tahunCell, err := f.GetCellValue(sheetName, "BD3")
	if err != nil {
		log.Fatal(err)
	}

	splitted := strings.Split(tahunCell, " ")
	currentTahun := splitted[len(splitted)-1]

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "NO." {
			cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i+1))
			if err != nil {
				log.Fatal(err)
			}

			if cellValue == "NO." {
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

	for i := 0; true; i++ {
		//iterate col
		col := "BD"
		colInNumber := helpers.CharStrToNum(col)
		colInNumber = colInNumber + i
		col = helpers.ToCharStr(colInNumber)

		rowBulan := "4"

		//mengambil tahun
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

		emptyCount := 0

		currentCustomer := ""

		//iterate over rows
		for j := 0; true; j++ {
			obj := toolkit.M{}
			currentRow := firstDataRow + j

			cellValueAnalisa, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			//mengambil Value di kolom
			cellValueCustomer, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}
			currentCustomer = cellValueCustomer

			cellValueSatuan, err := f.GetCellValue(sheetName, "C"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			if strings.Contains(cellValueAnalisa, "KEBUTUHAN ANALISA") {
				break
			}

			if strings.ToUpper(strings.TrimSpace(cellValueSatuan)) != strings.ToUpper("Box") {
				emptyCount++
				continue
			} else {
				emptyCount = 0
			}

			obj.Set("TAHUN", currentTahun)
			obj.Set("BULAN", helpers.IndexOf(cellValueBulan, months)+1)

			if strings.Contains(sheetName, "Internasional") {
				obj.Set("D_I", "I")
			} else {
				obj.Set("D_I", "D")
			}

			obj.Set("CUSTOMER", currentCustomer)

			cellValue, err := f.GetCellValue(sheetName, col+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			obj.Set("BOX", cellValue)
			obj.Set("TEUS", "")
			obj.Set("CUKER", "")

			if emptyCount >= 10 {
				break
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
