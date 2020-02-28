package controllers

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/sebar/dbflex"
	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"
)

// RealisasiController is a controller for every kind of Realisasi files.
type RealisasiController struct {
	*Base
}

// New is used to initiate the controller
func (c *RealisasiController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for Realisasi files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *RealisasiController) FileCriteria(file string) bool {
	return strings.Contains(strings.ToUpper(filepath.Base(file)), strings.ToUpper("REALISASI ANGGARAN"))
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *RealisasiController) ReadExcel() {
	for _, sheetName := range c.Engine.GetSheetMap() {
		if strings.EqualFold(sheetName, "NERACA") {
			c.ReadSheet(c.ReadDataNeraca, sheetName)
		}

		if strings.EqualFold(sheetName, "ARUS KAS") {
			c.ReadSheet(c.ReadDataArusKas, sheetName)
		}

		if strings.EqualFold(sheetName, "REKAP LR") {
			c.ReadSheet(c.ReadDataLabaRugi, sheetName)
		}

		if strings.EqualFold(sheetName, "RASIO (PERB.)") {
			c.ReadSheet(c.ReadDataRasioSummary, sheetName)
		}
	}
}

func (c *RealisasiController) ReadDataNeraca(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("realisasiAnggaran", "Neraca", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "KODE" {
			cellValueAfter, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i+1))
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

	months := clit.Config("realisasiAnggaran", "months", nil).([]interface{})

	stringData, err := c.Engine.GetCellValue(sheetName, "A3")
	if err != nil {
		log.Fatal(err)
	}

	splitted := strings.Split(stringData, " ")
	currentBulan := toolkit.ToString(helpers.IndexOf(splitted[len(splitted)-2], months) + 1)
	currentTahun := splitted[len(splitted)-1]

	stringSumber, err := c.Engine.GetCellValue(sheetName, "A2")
	if err != nil {
		log.Fatal(err)
	}

	currentSumber := "TTL"
	if strings.Contains(strings.ToUpper(stringSumber), "KONSOLIDASI") {
		currentSumber = "KONSOLIDASI"
	}

	countEmpty := 0

	tablename := "Neraca"

	// check if data exists
	sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + currentTahun + "' AND bulan = '" + currentBulan + "' AND sumber = '" + currentSumber + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From(tablename).SQL(sqlQuery), nil)
	defer cursor.Close()

	res := make([]toolkit.M, 0)
	err = cursor.Fetchs(&res, 0)

	//only insert if len of datas is 0 / if no data yet
	if len(res) == 0 {
		//iterate over rows
		for index := 0; true; index++ {
			rowData := toolkit.M{}
			currentRow := firstDataRow + index

			stringData, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
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
				stringUraian, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
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
					rowData.Set(header.DBFieldName, currentSumber)
				} else {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
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

			c.InsertRowData(currentRow, rowData, tablename)

			rowCount++
			no++
		}
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *RealisasiController) ReadDataArusKas(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("realisasiAnggaran", "ArusKas", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "KODE" {
			cellValueAfter, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i+1))
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

	months := clit.Config("realisasiAnggaran", "months", nil).([]interface{})

	stringTanggalan, err := c.Engine.GetCellValue(sheetName, "A3")
	if err != nil {
		log.Fatal(err)
	}

	splitted := strings.Split(stringTanggalan, " ")
	currentBulan := toolkit.ToString(helpers.IndexOf(splitted[len(splitted)-2], months) + 1)
	currentTahun := splitted[len(splitted)-1]

	stringSumber, err := c.Engine.GetCellValue(sheetName, "A2")
	if err != nil {
		log.Fatal(err)
	}

	currentSumber := "TTL"
	if strings.Contains(strings.ToUpper(stringSumber), "KONSOLIDASI") {
		currentSumber = "KONSOLIDASI"
	}

	countEmpty := 0

	tablename := "Arus_Kas"

	// check if data exists
	sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + currentTahun + "' AND bulan = '" + currentBulan + "' AND sumber = '" + currentSumber + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From(tablename).SQL(sqlQuery), nil)
	defer cursor.Close()

	res := make([]toolkit.M, 0)
	err = cursor.Fetchs(&res, 0)

	//only insert if len of datas is 0 / if no data yet
	if len(res) == 0 {
		//iterate over rows
		for index := 0; true; index++ {
			rowData := toolkit.M{}
			currentRow := firstDataRow + index

			stringKode, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
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
					stringUraian, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					currentKelompok = stringUraian
					currentSubTipe = ""

					continue
				}
			}

			stringData, err := c.Engine.GetCellValue(sheetName, "AF"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			if strings.TrimSpace(stringData) == "" {
				stringUraian, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
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
					stringData, err := c.Engine.GetCellValue(sheetName, "AF"+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					if strings.TrimSpace(currentSubTipe) == "PENERIMAAN" {
						rowData.Set(header.DBFieldName, stringData)
					} else {
						rowData.Set(header.DBFieldName, "")
					}
				} else if header.DBFieldName == "EXP_YTD" {
					stringData, err := c.Engine.GetCellValue(sheetName, "AF"+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					if strings.TrimSpace(currentSubTipe) == "PENGELUARAN" {
						rowData.Set(header.DBFieldName, stringData)
					} else {
						rowData.Set(header.DBFieldName, "")
					}
				} else if header.DBFieldName == "GROUP_NO" {
					norek := strings.TrimSpace(stringKode)

					rowData.Set(header.DBFieldName, norek[len(norek)-3:])
				} else if header.DBFieldName == "SUMBER" {
					rowData.Set(header.DBFieldName, currentSumber)
				} else {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
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

			c.InsertRowData(currentRow, rowData, tablename)

			rowCount++
			no++
		}
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *RealisasiController) ReadDataLabaRugi(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("realisasiAnggaran", "LabaRugi", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "KODE" {
			cellValueAfter, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i+1))
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

	months := clit.Config("realisasiAnggaran", "months", nil).([]interface{})

	stringTanggalan, err := c.Engine.GetCellValue(sheetName, "A3")
	if err != nil {
		log.Fatal(err)
	}

	splitted := strings.Split(stringTanggalan, " ")
	currentBulan := toolkit.ToString(helpers.IndexOf(splitted[len(splitted)-2], months) + 1)
	currentTahun := splitted[len(splitted)-1]

	stringSumber, err := c.Engine.GetCellValue(sheetName, "A2")
	if err != nil {
		log.Fatal(err)
	}

	currentSumber := "TTL"
	if strings.Contains(strings.ToUpper(stringSumber), "KONSOLIDASI") {
		currentSumber = "KONSOLIDASI"
	}

	countEmpty := 0

	tablename := "Laba_Rugi"

	// check if data exists
	sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + currentTahun + "' AND bulan = '" + currentBulan + "' AND sumber = '" + currentSumber + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From(tablename).SQL(sqlQuery), nil)
	defer cursor.Close()

	res := make([]toolkit.M, 0)
	err = cursor.Fetchs(&res, 0)

	//only insert if len of datas is 0 / if no data yet
	if len(res) == 0 {
		//iterate over rows
		for index := 0; true; index++ {
			rowData := toolkit.M{}
			currentRow := firstDataRow + index

			stringKode, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
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
			if err != nil { //jika error maka tipe
				stringUraian, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				currentTipe = stringUraian

				continue
			}

			for _, header := range headers {
				if header.DBFieldName == "No" {
					rowData.Set(header.DBFieldName, no)
				} else if header.DBFieldName == "Tipe" {
					rowData.Set(header.DBFieldName, currentTipe)
				} else if header.DBFieldName == "TAHUN" {
					rowData.Set(header.DBFieldName, currentTahun)
				} else if header.DBFieldName == "BULAN" {
					rowData.Set(header.DBFieldName, currentBulan)
				} else if header.DBFieldName == "Sumber" {
					rowData.Set(header.DBFieldName, currentSumber)
				} else {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
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

			c.InsertRowData(currentRow, rowData, tablename)

			rowCount++
			no++
		}
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *RealisasiController) ReadDataRasioSummary(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("realisasiAnggaran", "RasioSummary", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "KODE" {
			cellValueAfter, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(i+1))
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

	months := clit.Config("realisasiAnggaran", "months", nil).([]interface{})

	stringTanggalan, err := c.Engine.GetCellValue(sheetName, "A3")
	if err != nil {
		log.Fatal(err)
	}

	splitted := strings.Split(stringTanggalan, " ")
	currentBulan := toolkit.ToString(helpers.IndexOf(splitted[len(splitted)-2], months) + 1)
	currentTahun := splitted[len(splitted)-1]

	stringSumber, err := c.Engine.GetCellValue(sheetName, "A2")
	if err != nil {
		log.Fatal(err)
	}

	currentSumber := "TTL"
	if strings.Contains(strings.ToUpper(stringSumber), "KONSOLIDASI") {
		currentSumber = "KONSOLIDASI"
	}

	countEmpty := 0

	tablename := "Rasio_Summary"

	// check if data exists
	sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + currentTahun + "' AND bulan = '" + currentBulan + "' AND sumber = '" + currentSumber + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From(tablename).SQL(sqlQuery), nil)
	defer cursor.Close()

	res := make([]toolkit.M, 0)
	err = cursor.Fetchs(&res, 0)

	//only insert if len of datas is 0 / if no data yet
	if len(res) == 0 {
		//iterate over rows
		for index := 0; true; index++ {
			rowData := toolkit.M{}
			currentRow := firstDataRow + index

			stringUraian, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			stringSatuan, err := c.Engine.GetCellValue(sheetName, "C"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			if strings.TrimSpace(stringSatuan) == "" || !strings.Contains(stringUraian, ".") { //jika cell satuan kosong maka skip saja ehe
				countEmpty++

				if countEmpty >= 100 {
					break
				}

				continue
			}

			for _, header := range headers {
				if header.DBFieldName == "No" {
					rowData.Set(header.DBFieldName, no)
				} else if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, currentTahun)
				} else if header.DBFieldName == "Bulan" {
					rowData.Set(header.DBFieldName, currentBulan)
				} else if header.DBFieldName == "Sumber" {
					rowData.Set(header.DBFieldName, currentSumber)
				} else {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
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

			c.InsertRowData(currentRow, rowData, tablename)

			rowCount++
			no++
		}
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}
