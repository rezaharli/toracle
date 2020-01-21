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
)

type RealisasiController struct {
	*Base
}

func NewRealisasiController() *RealisasiController {
	return new(RealisasiController)
}

func (c *RealisasiController) ReadExcels() error {
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

func (c *RealisasiController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(strings.ToUpper(filepath.Base(file)), strings.ToUpper("REALISASI ANGGARAN")) {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. Realisasi files found:", len(resourceFiles))
	return resourceFiles
}

func (c *RealisasiController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	for _, sheetName := range f.GetSheetMap() {
		if strings.EqualFold(sheetName, "NERACA") {
			err = c.ReadDataNeraca(f, sheetName)
			if err != nil {
				log.Println("Error reading data. ERROR:", err)
			}
		}

		if strings.EqualFold(sheetName, "ARUS KAS") {
			err = c.ReadDataArusKas(f, sheetName)
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

func (c *RealisasiController) ReadDataNeraca(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("realisasiAnggaran", "Neraca", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "KODE" {
			cellValueAfter, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i+1))
			if err != nil {
				log.Fatal(err)
			}

			if cellValueAfter != "KODE" {
				firstDataRow = i + 2
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

	var err error
	// var rowDatas []toolkit.M
	rowCount := 0
	no := 1

	currentTipe := ""
	currentSubTipe := ""

	stringData, err := f.GetCellValue(sheetName, "A3")
	if err != nil {
		log.Fatal(err)
	}

	splitted := strings.Split(stringData, " ")
	currentBulan := splitted[len(splitted)-2]
	currentTahun := splitted[len(splitted)-1]

	countEmpty := 0

	//iterate over rows
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index

		stringData, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		if strings.TrimSpace(stringData) == "" { //jika cell kode kosong maka skip saja ehe
			countEmpty++

			if countEmpty >= 100 {
				break
			}

			continue
		}

		_, err = strconv.Atoi(stringData)
		if err != nil { //jika error maka tipe atau subtipe
			stringUraian, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			if !strings.Contains(stringData, ".") { //jika tidak mengandung titik maka tipe
				currentTipe = stringUraian
				currentSubTipe = ""
			} else {
				currentSubTipe = stringUraian
			}

			continue
		}

		for _, header := range headers {
			if header.DBFieldName == "No" {
				rowData.Set(header.DBFieldName, no)
			} else if header.DBFieldName == "Tipe" {
				rowData.Set(header.DBFieldName, currentTipe)
			} else if header.DBFieldName == "SubTipe" {
				rowData.Set(header.DBFieldName, currentSubTipe)
			} else if header.DBFieldName == "Tahun" {
				rowData.Set(header.DBFieldName, currentTahun)
			} else if header.DBFieldName == "Bulan" {
				rowData.Set(header.DBFieldName, currentBulan)
			} else if header.DBFieldName == "Sumber" {
				rowData.Set(header.DBFieldName, "KONSOLIDASI / TTL")
			} else {
				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				stringData = strings.ReplaceAll(stringData, "'", "''")

				if len(stringData) > 300 {
					stringData = stringData[0:300]
				}

				rowData.Set(header.DBFieldName, stringData)
			}
		}

		param := helpers.InsertParam{
			TableName: "Neraca",
			Data:      rowData,
		}

		err = helpers.Insert(param)
		if err != nil {
			log.Fatal("Error inserting row "+toolkit.ToString(currentRow)+", ERROR:", err.Error())
		} else {
			log.Println("Row", currentRow, "inserted.")
		}

		rowCount++
		no++
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *RealisasiController) ReadDataArusKas(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("realisasiAnggaran", "ArusKas", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "KODE" {
			cellValueAfter, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(i+1))
			if err != nil {
				log.Fatal(err)
			}

			if cellValueAfter != "KODE" {
				firstDataRow = i + 2
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

	var err error
	// var rowDatas []toolkit.M
	rowCount := 0
	no := 1

	currentKelompok := ""
	currentSubTipe := ""

	stringTanggalan, err := f.GetCellValue(sheetName, "A3")
	if err != nil {
		log.Fatal(err)
	}

	splitted := strings.Split(stringTanggalan, " ")
	currentBulan := splitted[len(splitted)-2]
	currentTahun := splitted[len(splitted)-1]

	countEmpty := 0

	//iterate over rows
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index

		stringKode, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		if strings.TrimSpace(stringKode) == "" { //jika cell kode kosong maka skip saja ehe
			countEmpty++

			if countEmpty >= 100 {
				break
			}

			continue
		}

		_, err = strconv.Atoi(stringKode)
		if err != nil { //jika string maka kelompok (atau NO_REK)
			if !strings.Contains(stringKode, ".") {
				stringUraian, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				currentKelompok = stringUraian
				currentSubTipe = ""

				continue
			}
		}

		stringData, err := f.GetCellValue(sheetName, "AF"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		if strings.TrimSpace(stringData) == "" {
			stringUraian, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			currentSubTipe = stringUraian

			continue
		}

		for _, header := range headers {
			if header.DBFieldName == "TAHUN" {
				rowData.Set(header.DBFieldName, currentTahun)
			} else if header.DBFieldName == "BULAN" {
				rowData.Set(header.DBFieldName, currentBulan)
			} else if header.DBFieldName == "KELOMPOK" {
				rowData.Set(header.DBFieldName, currentKelompok)
			} else if header.DBFieldName == "INCOME_YTD" {
				stringData, err := f.GetCellValue(sheetName, "AF"+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				if strings.TrimSpace(currentSubTipe) == "PENERIMAAN" {
					rowData.Set(header.DBFieldName, stringData)
				} else {
					rowData.Set(header.DBFieldName, "")
				}
			} else if header.DBFieldName == "EXP_YTD" {
				stringData, err := f.GetCellValue(sheetName, "AF"+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				if strings.TrimSpace(currentSubTipe) == "PENGELUARAN" {
					rowData.Set(header.DBFieldName, stringData)
				} else {
					rowData.Set(header.DBFieldName, "")
				}
			} else if header.DBFieldName == "GRUP" {
				norek := strings.TrimSpace(stringKode)

				rowData.Set(header.DBFieldName, norek[len(norek)-3:])
			} else if header.DBFieldName == "SUMBER" {
				rowData.Set(header.DBFieldName, "KONSOLIDASI / TTL")
			} else {
				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				stringData = strings.ReplaceAll(stringData, "'", "''")

				if len(stringData) > 300 {
					stringData = stringData[0:300]
				}

				rowData.Set(header.DBFieldName, stringData)
			}
		}

		param := helpers.InsertParam{
			TableName: "Arus_Kas",
			Data:      rowData,
		}

		err = helpers.Insert(param)
		if err != nil {
			log.Fatal("Error inserting row "+toolkit.ToString(currentRow)+", ERROR:", err.Error())
		} else {
			log.Println("Row", currentRow, "inserted.")
		}

		rowCount++
		no++
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}
