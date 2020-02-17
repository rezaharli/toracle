package controllers

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/sebar/dbflex"
)

type RekapPetikemasController struct {
	*Base
}

func NewRekapPetikemasController() *RekapPetikemasController {
	return new(RekapPetikemasController)
}

func (c *RekapPetikemasController) ReadExcels() error {
	for _, file := range c.FetchFiles() {
		err := c.readExcel(file)
		if err == nil {
			// move file if read succeeded
			c.MoveToArchive(file)
			log.Println("Done.")
		} else {
			return err
		}
	}

	return nil
}

func (c *RekapPetikemasController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "Rekap Petikemas Perak sd September 2019") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. RekapPetikemas files found:", len(resourceFiles))
	return resourceFiles
}

func (c *RekapPetikemasController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	for _, sheetName := range f.GetSheetMap() {
		if strings.EqualFold(sheetName, "Sheet1") {
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

func (c *RekapPetikemasController) ReadData(f *excelize.File, sheetName string) error {
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
		cellValueTahun, err := f.GetCellValue(sheetName, col+rowTahun)
		if err != nil {
			log.Fatal(err)
		}

		//mengambil terminal
		cellValueTerminal, err := f.GetCellValue(sheetName, col+rowTerminal)
		if err != nil {
			log.Fatal(err)
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
			cellValueIntDom, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			cellValue, err := f.GetCellValue(sheetName, col+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			cellValueSat, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
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
