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

// PencapaianController is a controller for every kind of Pencapaian files.
type PencapaianController struct {
	*Base
}

// New is used to initiate the controller
func (c *PencapaianController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for Pencapaian files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *PencapaianController) FileCriteria(file string) bool {
	return strings.Contains(strings.ToUpper(filepath.Base(file)), strings.ToUpper("PENCAPAIAN"))
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *PencapaianController) ReadExcel() {
	for _, sheetName := range c.Engine.GetSheetMap() {
		if strings.Contains(strings.ToUpper(sheetName), strings.ToUpper("REKAP KONSOL")) {
			c.ReadSheet(c.ReadDataRekapKonsol, sheetName)
			c.ReadSheet(c.ReadDataRekapKonsol2, sheetName)
		}

		if strings.Contains(strings.ToUpper(sheetName), strings.ToUpper("REKAP LEGI")) {
			c.ReadSheet(c.ReadDataRekapLegi, sheetName)
			c.ReadSheet(c.ReadDataRekapLegi2, sheetName)
			c.ReadSheet(c.ReadDataRekapLegi3, sheetName)
		}

		if strings.Contains(strings.ToUpper(sheetName), strings.ToUpper("REKAP TTL")) {
			c.ReadSheet(c.ReadDataRekapTTL, sheetName)
			c.ReadSheet(c.ReadDataRekapTTL2, sheetName)
			c.ReadSheet(c.ReadDataRekapTTL3, sheetName)
		}
	}
}

func (c *PencapaianController) ReadDataRekapKonsol(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("pencapaian", "rekapKonsol", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "KODE" {
			cellValueAfter, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(i+1))
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

	months := clit.Config("pencapaian", "months", nil).([]interface{})

	filename := strings.ReplaceAll(filepath.Base(c.Engine.GetExcelPath()), ".xlsx", "")
	splitted := strings.Split(filename, " ")
	bulan := toolkit.ToString(helpers.IndexOf(splitted[1], months) + 1)
	tahun := splitted[2]

	var err error
	// var rowDatas []toolkit.M
	rowCount := 0
	no := 1
	emptyRowCount := 0

	tablename := "Rekap_Konsol"

	// check if data exists
	sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + tahun + "' AND bulan = '" + bulan + "'"

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

			isRowEmpty := true
			skipRow := true
			for _, header := range headers {
				if header.DBFieldName == "NO" {
					rowData.Set(header.DBFieldName, no)
				} else if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, tahun)
				} else if header.DBFieldName == "Bulan" {
					rowData.Set(header.DBFieldName, bulan)
				} else {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					if stringData != "" {
						isRowEmpty = false
					}

					if header.DBFieldName != "KODE" && header.DBFieldName != "URAIAN" {
						if strings.TrimSpace(stringData) != "" {
							skipRow = false
						}
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if strings.TrimSpace(rowData.GetString("KODE")) == "" && strings.TrimSpace(rowData.GetString("URAIAN")) == "" {
				skipRow = true
			}

			if isRowEmpty {
				emptyRowCount++
			} else {
				emptyRowCount = 0
			}

			if skipRow {
				if isRowEmpty && emptyRowCount > 1 {
					break
				}

				continue
			}

			if emptyRowCount > 1 {
				break
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

func (c *PencapaianController) ReadDataRekapKonsol2(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("pencapaian", "rekapKonsol2", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 53

	var headers []Header
	for key, column := range columnsMapping {
		header := Header{
			DBFieldName: key,
			Column:      column.(string),
		}

		headers = append(headers, header)
	}

	months := clit.Config("pencapaian", "months", nil).([]interface{})

	filename := strings.ReplaceAll(filepath.Base(c.Engine.GetExcelPath()), ".xlsx", "")
	splitted := strings.Split(filename, " ")
	bulan := toolkit.ToString(helpers.IndexOf(splitted[1], months) + 1)
	tahun := splitted[2]

	var err error
	// var rowDatas []toolkit.M
	rowCount := 0

	tablename := "Rekap_Konsol2"

	// check if data exists
	sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + tahun + "' AND bulan = '" + bulan + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From(tablename).SQL(sqlQuery), nil)
	defer cursor.Close()

	res := make([]toolkit.M, 0)
	err = cursor.Fetchs(&res, 0)

	//only insert if len of datas is 0 / if no data yet
	if len(res) == 0 {
		//iterate over rows
		no := 1
		for index := 0; true; index++ {
			rowData := toolkit.M{}
			currentRow := firstDataRow + index

			if currentRow > 57 {
				break
			}

			isRowEmpty := true
			for _, header := range headers {
				if header.DBFieldName == "No" {
					rowData.Set(header.DBFieldName, no)
				} else if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, tahun)
				} else if header.DBFieldName == "Bulan" {
					rowData.Set(header.DBFieldName, bulan)
				} else {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					if stringData != "" {
						isRowEmpty = false
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if isRowEmpty {
				continue
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

func (c *PencapaianController) ReadDataRekapLegi(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("pencapaian", "rekapLegi", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "NO" {
			cellValueAfter, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(i+1))
			if err != nil {
				log.Fatal(err)
			}

			if cellValueAfter != "NO" {
				firstDataRow = i + 3
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

	months := clit.Config("pencapaian", "months", nil).([]interface{})

	filename := strings.ReplaceAll(filepath.Base(c.Engine.GetExcelPath()), ".xlsx", "")
	splitted := strings.Split(filename, " ")
	bulan := toolkit.ToString(helpers.IndexOf(splitted[1], months) + 1)
	tahun := splitted[2]

	var err error
	// var rowDatas []toolkit.M
	rowCount := 0
	no := 1
	emptyRowCount := 0

	tablename := "Rekap_Legi"

	// check if data exists
	sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + tahun + "' AND bulan = '" + bulan + "'"

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

			isRowEmpty := true
			skipRow := true
			for _, header := range headers {
				if header.DBFieldName == "NO" {
					rowData.Set(header.DBFieldName, no)
				} else if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, tahun)
				} else if header.DBFieldName == "Bulan" {
					rowData.Set(header.DBFieldName, bulan)
				} else {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					if stringData != "" {
						isRowEmpty = false
					}

					if header.DBFieldName != "KODE" && header.DBFieldName != "URAIAN" {
						if strings.TrimSpace(stringData) != "" {
							skipRow = false
						}
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if strings.TrimSpace(rowData.GetString("KODE")) == "" && strings.TrimSpace(rowData.GetString("URAIAN")) == "" {
				skipRow = true
			}

			if isRowEmpty {
				emptyRowCount++
			} else {
				emptyRowCount = 0
			}

			if skipRow {
				continue
			}

			cellValueAfter, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow+1))
			if err != nil {
				log.Fatal(err)
			}

			if cellValueAfter == "NO" {
				break
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

func (c *PencapaianController) ReadDataRekapLegi2(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("pencapaian", "rekapLegi2", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 51

	var headers []Header
	for key, column := range columnsMapping {
		header := Header{
			DBFieldName: key,
			Column:      column.(string),
		}

		headers = append(headers, header)
	}

	months := clit.Config("pencapaian", "months", nil).([]interface{})

	filename := strings.ReplaceAll(filepath.Base(c.Engine.GetExcelPath()), ".xlsx", "")
	splitted := strings.Split(filename, " ")
	bulan := toolkit.ToString(helpers.IndexOf(splitted[1], months) + 1)
	tahun := splitted[2]

	var err error
	// var rowDatas []toolkit.M
	rowCount := 0

	tablename := "Rekap_Legi2"

	// check if data exists
	sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + tahun + "' AND bulan = '" + bulan + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From(tablename).SQL(sqlQuery), nil)
	defer cursor.Close()

	res := make([]toolkit.M, 0)
	err = cursor.Fetchs(&res, 0)

	//only insert if len of datas is 0 / if no data yet
	if len(res) == 0 {
		//iterate over rows
		no := 1
		for index := 0; true; index++ {
			rowData := toolkit.M{}
			currentRow := firstDataRow + index

			if currentRow > 53 {
				break
			}

			isRowEmpty := true
			for _, header := range headers {
				if header.DBFieldName == "No" {
					rowData.Set(header.DBFieldName, no)
				} else if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, tahun)
				} else if header.DBFieldName == "Bulan" {
					rowData.Set(header.DBFieldName, bulan)
				} else {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					if stringData != "" {
						isRowEmpty = false
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if isRowEmpty {
				continue
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

func (c *PencapaianController) ReadDataRekapLegi3(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("pencapaian", "rekapLegi3", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 57

	var headers []Header
	for key, column := range columnsMapping {
		header := Header{
			DBFieldName: key,
			Column:      column.(string),
		}

		headers = append(headers, header)
	}

	months := clit.Config("pencapaian", "months", nil).([]interface{})

	filename := strings.ReplaceAll(filepath.Base(c.Engine.GetExcelPath()), ".xlsx", "")
	splitted := strings.Split(filename, " ")
	bulan := toolkit.ToString(helpers.IndexOf(splitted[1], months) + 1)
	tahun := splitted[2]

	var err error
	// var rowDatas []toolkit.M
	rowCount := 0
	tablename := "Rekap_Legi3"

	// check if data exists
	sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + tahun + "' AND bulan = '" + bulan + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From(tablename).SQL(sqlQuery), nil)
	defer cursor.Close()

	res := make([]toolkit.M, 0)
	err = cursor.Fetchs(&res, 0)

	//only insert if len of datas is 0 / if no data yet
	if len(res) == 0 {
		//iterate over rows
		no := 1
		for index := 0; true; index++ {
			rowData := toolkit.M{}
			currentRow := firstDataRow + index

			if currentRow > 60 {
				break
			}

			isRowEmpty := true
			for _, header := range headers {
				if header.DBFieldName == "No" {
					rowData.Set(header.DBFieldName, no)
				} else if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, tahun)
				} else if header.DBFieldName == "Bulan" {
					rowData.Set(header.DBFieldName, bulan)
				} else {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					if stringData != "" {
						isRowEmpty = false
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if isRowEmpty {
				continue
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

func (c *PencapaianController) ReadDataRekapTTL(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("pencapaian", "rekapTTL", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "NO" {
			cellValueAfter, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(i+1))
			if err != nil {
				log.Fatal(err)
			}

			if cellValueAfter != "NO" {
				firstDataRow = i + 3
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

	months := clit.Config("pencapaian", "months", nil).([]interface{})

	filename := strings.ReplaceAll(filepath.Base(c.Engine.GetExcelPath()), ".xlsx", "")
	splitted := strings.Split(filename, " ")
	bulan := toolkit.ToString(helpers.IndexOf(splitted[1], months) + 1)
	tahun := splitted[2]

	var err error
	// var rowDatas []toolkit.M
	rowCount := 0
	no := 1
	emptyRowCount := 0

	tablename := "Rekap_TTL"

	// check if data exists
	sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + tahun + "' AND bulan = '" + bulan + "'"

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

			isRowEmpty := true
			skipRow := true
			for _, header := range headers {
				if header.DBFieldName == "NO" {
					rowData.Set(header.DBFieldName, no)
				} else if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, tahun)
				} else if header.DBFieldName == "Bulan" {
					rowData.Set(header.DBFieldName, bulan)
				} else {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					if stringData != "" {
						isRowEmpty = false
					}

					if header.DBFieldName != "KODE" && header.DBFieldName != "URAIAN" {
						if strings.TrimSpace(stringData) != "" {
							skipRow = false
						}
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if strings.TrimSpace(rowData.GetString("KODE")) == "" && strings.TrimSpace(rowData.GetString("URAIAN")) == "" {
				skipRow = true
			}

			if isRowEmpty {
				emptyRowCount++
			} else {
				emptyRowCount = 0
			}

			if skipRow {
				continue
			}

			cellValueAfter, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow+1))
			if err != nil {
				log.Fatal(err)
			}

			if cellValueAfter == "NO" {
				break
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

func (c *PencapaianController) ReadDataRekapTTL2(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("pencapaian", "rekapTTL2", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 51

	var headers []Header
	for key, column := range columnsMapping {
		header := Header{
			DBFieldName: key,
			Column:      column.(string),
		}

		headers = append(headers, header)
	}

	months := clit.Config("pencapaian", "months", nil).([]interface{})

	filename := strings.ReplaceAll(filepath.Base(c.Engine.GetExcelPath()), ".xlsx", "")
	splitted := strings.Split(filename, " ")
	bulan := toolkit.ToString(helpers.IndexOf(splitted[1], months) + 1)
	tahun := splitted[2]

	var err error
	// var rowDatas []toolkit.M
	rowCount := 0

	tablename := "Rekap_TTL2"

	// check if data exists
	sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + tahun + "' AND bulan = '" + bulan + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From(tablename).SQL(sqlQuery), nil)
	defer cursor.Close()

	res := make([]toolkit.M, 0)
	err = cursor.Fetchs(&res, 0)

	//only insert if len of datas is 0 / if no data yet
	if len(res) == 0 {
		//iterate over rows
		no := 1
		for index := 0; true; index++ {
			rowData := toolkit.M{}
			currentRow := firstDataRow + index

			if currentRow > 53 {
				break
			}

			isRowEmpty := true
			for _, header := range headers {
				if header.DBFieldName == "No" {
					rowData.Set(header.DBFieldName, no)
				} else if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, tahun)
				} else if header.DBFieldName == "Bulan" {
					rowData.Set(header.DBFieldName, bulan)
				} else {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					if stringData != "" {
						isRowEmpty = false
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if isRowEmpty {
				continue
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

func (c *PencapaianController) ReadDataRekapTTL3(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("pencapaian", "rekapTTL3", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 57

	var headers []Header
	for key, column := range columnsMapping {
		header := Header{
			DBFieldName: key,
			Column:      column.(string),
		}

		headers = append(headers, header)
	}

	months := clit.Config("pencapaian", "months", nil).([]interface{})

	filename := strings.ReplaceAll(filepath.Base(c.Engine.GetExcelPath()), ".xlsx", "")
	splitted := strings.Split(filename, " ")
	bulan := toolkit.ToString(helpers.IndexOf(splitted[1], months) + 1)
	tahun := splitted[2]

	var err error
	// var rowDatas []toolkit.M
	rowCount := 0

	tablename := "Rekap_TTL3"

	// check if data exists
	sqlQuery := "SELECT tahun FROM " + tablename + " WHERE tahun = '" + tahun + "' AND bulan = '" + bulan + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From(tablename).SQL(sqlQuery), nil)
	defer cursor.Close()

	res := make([]toolkit.M, 0)
	err = cursor.Fetchs(&res, 0)

	//only insert if len of datas is 0 / if no data yet
	if len(res) == 0 {
		//iterate over rows
		no := 1
		for index := 0; true; index++ {
			rowData := toolkit.M{}
			currentRow := firstDataRow + index

			if currentRow > 60 {
				break
			}

			isRowEmpty := true
			for _, header := range headers {
				if header.DBFieldName == "No" {
					rowData.Set(header.DBFieldName, no)
				} else if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, tahun)
				} else if header.DBFieldName == "Bulan" {
					rowData.Set(header.DBFieldName, bulan)
				} else {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					if stringData != "" {
						isRowEmpty = false
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if isRowEmpty {
				continue
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
