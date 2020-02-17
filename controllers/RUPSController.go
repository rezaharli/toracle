package controllers

import (
	"errors"
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

type RUPSController struct {
	FileExtension string
}

func (c *RUPSController) New() {
	log.Println("Scanning for RUPS files.")
	c.FileExtension = ".xlsx"
}

func (c *RUPSController) FetchFiles(resourcePath string) []string {
	return helpers.FetchFilePathsWithExt(resourcePath, c.FileExtension)
}

func (c *RUPSController) FileCriteria(file string) bool {
	if strings.Contains(filepath.Base(file), "KK RKAP - 2020 FIN2") {
		return true
	}

	return false
}

func (c *RUPSController) ReadExcel(f *excelize.File) error {
	var err error

	for _, sheetName := range f.GetSheetMap() {
		if strings.EqualFold(sheetName, "Asumsi") {
			err = c.readAsumsi(f, sheetName)
			if err != nil {
				log.Println("Error reading monthly data. ERROR:", err)
			}
		}

		if strings.EqualFold(sheetName, "Highlight") {
			err = c.readHighlight(f, sheetName)
			if err != nil {
				log.Println("Error reading monthly data. ERROR:", err)
			}
		}

		if strings.EqualFold(sheetName, "RKM") {
			err = c.readRKM(f, sheetName)
			if err != nil {
				log.Println("Error reading monthly data. ERROR:", err)
			}
		}

		if strings.EqualFold(sheetName, "LR KONSOL") {
			err = c.readFinancialReport(f, sheetName)
			if err != nil {
				log.Println("Error reading monthly data. ERROR:", err)
			}
		}

		if strings.EqualFold(sheetName, "INVES") {
			err = c.readInvestasi(f, sheetName)
			if err != nil {
				log.Println("Error reading monthly data. ERROR:", err)
			}
		}

		if strings.EqualFold(sheetName, "SDM REKAP") {
			err = c.readSDM(f, sheetName)
			if err != nil {
				log.Println("Error reading monthly data. ERROR:", err)
			}
		}

		if strings.EqualFold(sheetName, "FINANCIAL RATIO") {
			err = c.readFinancialRatio(f, sheetName)
			if err != nil {
				log.Println("Error reading monthly data. ERROR:", err)
			}
		}
	}

	return err
}

func (c *RUPSController) readAsumsi(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("RUPS", "Asumsi", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	filename := filepath.Base(f.Path)
	splitted := strings.Split(filename, " ")
	tahun := splitted[3]

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "NO." {
			firstDataRow = i
			break
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
	rowCount := 0
	no := 1
	emptyCount := 0
	currentTipe := ""

	tablename := "RUPS_Asumsi"

	// check if data exists
	sqlQuery := "SELECT * FROM " + tablename + " WHERE tahun = '" + tahun + "'"

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
			isDataRow := true

			cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			_, errConvert := strconv.Atoi(cellValue)

			if cellValue == "NO." {
				currentTipe, err = f.GetCellValue(sheetName, "C"+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				isDataRow = false
			} else if errConvert != nil {
				isDataRow = false
			}

			for _, header := range headers {
				if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, tahun)
				} else if header.DBFieldName == "Tipe" {
					rowData.Set(header.DBFieldName, currentTipe)
				} else {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")

					if header.DBFieldName == "RKAP" || header.DBFieldName == "Taksasi" || header.DBFieldName == "Usulan" {
						stringData = strings.ReplaceAll(stringData, "%", "")
					}

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					if strings.TrimSpace(stringData) != "" {
						isRowEmpty = false
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if emptyCount >= 10 {
				break
			}

			if isRowEmpty {
				emptyCount++
				continue
			}

			if !isDataRow {
				continue
			}

			param := helpers.InsertParam{
				TableName: tablename,
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
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *RUPSController) readHighlight(f *excelize.File, sheetName string) error {
	var err error

	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	configs := clit.Config("RUPS", "Highlight", nil).(map[string]interface{})

	for tipe, config := range configs {
		columnsMapping := config.(map[string]interface{})["columnsMapping"].(map[string]interface{})

		filename := filepath.Base(f.Path)
		splitted := strings.Split(filename, " ")
		tahun := splitted[3]

		firstDataRow := 0
		i := 1
		for {
			cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i))
			if err != nil {
				log.Fatal(err)
			}

			if cellValue == "Uraian" {
				firstDataRow = i + 2
				break
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

		rowCount := 0
		no := 1
		emptyCount := 0

		tablename := "RUPS_Highlight"

		// check if data exists
		sqlQuery := "SELECT * FROM " + tablename + " WHERE tahun = '" + tahun + "'"

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

				// kalau col A ada isinya, berenti
				cellValue, err := f.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				if cellValue != "" {
					break
				}

				for _, header := range headers {
					if header.DBFieldName == "Tahun" {
						rowData.Set(header.DBFieldName, tahun)
					} else if header.DBFieldName == "Tipe" {
						rowData.Set(header.DBFieldName, tipe)
					} else {
						stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
						if err != nil {
							log.Fatal(err)
						}

						stringData = strings.ReplaceAll(stringData, "'", "''")

						if header.DBFieldName == "Nilai" { //ambil integernya doang
							getNumVal := func(str string) []string {
								charToNum := func(r rune) (int, error) {
									intval := int(r) - '0'
									if 0 <= intval && intval <= 9 {
										return intval, nil
									}

									return -1, errors.New("type: rune was not int")
								}

								var nums []string
								for _, val := range str {
									_, err := charToNum(val)
									if err != nil {
										continue
									}

									nums = append(nums, string(val))
								}

								return nums
							}

							stringData = strings.Join(getNumVal(stringData), "")
						}

						if len(stringData) > 300 {
							stringData = stringData[0:300]
						}

						if strings.TrimSpace(stringData) != "" {
							isRowEmpty = false
						}

						rowData.Set(header.DBFieldName, stringData)
					}
				}

				if emptyCount >= 10 {
					break
				}

				if isRowEmpty {
					emptyCount++
					continue
				}

				param := helpers.InsertParam{
					TableName: tablename,
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
		}
	}

	return err
}

func (c *RUPSController) readRKM(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("RUPS", "RKM", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	filename := filepath.Base(f.Path)
	splitted := strings.Split(filename, " ")
	tahun := splitted[3]

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if strings.TrimSpace(cellValue) == "PROGRAM STRATEGIS" {
			cellValueAfter, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i+1))
			if err != nil {
				log.Fatal(err)
			}

			if cellValueAfter != "PROGRAM STRATEGIS" {
				firstDataRow = i + 1
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
	rowCount := 0
	no := 1

	tablename := "RUPS_RKM"

	// check if data exists
	sqlQuery := "SELECT * FROM " + tablename + " WHERE tahun = '" + tahun + "'"

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
			isDataRow := true

			for _, header := range headers {
				if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, tahun)
				} else {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")

					if header.DBFieldName == "Taksasi_Jumlah" || header.DBFieldName == "Taksasi_Selesai" || header.DBFieldName == "RKAP" {
						stringData = strings.ReplaceAll(stringData, "*", "")
					}

					if strings.TrimSpace(stringData) != "" {
						isRowEmpty = false
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if isRowEmpty {
				break
			}

			if !isDataRow {
				continue
			}

			param := helpers.InsertParam{
				TableName: tablename,
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
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *RUPSController) readFinancialReport(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("RUPS", "FinancialReport", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	filename := filepath.Base(f.Path)
	splitted := strings.Split(filename, " ")
	tahun := splitted[3]

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "L"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if strings.TrimSpace(cellValue) == "Uraian" {
			cellValueAfter, err := f.GetCellValue(sheetName, "L"+toolkit.ToString(i+1))
			if err != nil {
				log.Fatal(err)
			}

			if strings.TrimSpace(cellValueAfter) != "Uraian" {
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
	rowCount := 0
	emptyCount := 0
	no := 1

	tablename := "RUPS_Financial_Report"

	// check if data exists
	sqlQuery := "SELECT * FROM " + tablename + " WHERE tahun = '" + tahun + "'"

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
			isDataRow := true

			for _, header := range headers {
				if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, tahun)
				} else {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")

					if strings.TrimSpace(stringData) != "" {
						isRowEmpty = false
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if emptyCount >= 10 {
				break
			}

			if isRowEmpty {
				emptyCount++
				continue
			}

			if !isDataRow {
				continue
			}

			param := helpers.InsertParam{
				TableName: tablename,
				Data:      rowData,
			}

			err = helpers.Insert(param)
			if err != nil {
				log.Println("Error inserting row "+toolkit.ToString(currentRow)+", ERROR:", err.Error())
			} else {
				log.Println("Row", currentRow, "inserted.")
			}

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

func (c *RUPSController) readInvestasi(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("RUPS", "Investasi", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	filename := filepath.Base(f.Path)
	splitted := strings.Split(filename, " ")
	tahun := splitted[3]

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if strings.TrimSpace(cellValue) == "URAIAN INVESTASI" {
			cellValueAfter, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i+1))
			if err != nil {
				log.Fatal(err)
			}

			if strings.TrimSpace(cellValueAfter) != "URAIAN INVESTASI" {
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
	rowCount := 0
	emptyCount := 0
	no := 1

	tablename := "RUPS_Investasi"

	// check if data exists
	sqlQuery := "SELECT * FROM " + tablename + " WHERE tahun = '" + tahun + "'"

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
			isDataRow := true

			for _, header := range headers {
				if header.DBFieldName == "Tahun" {
					rowData.Set(header.DBFieldName, tahun)
				} else {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")

					if strings.TrimSpace(stringData) != "" {
						isRowEmpty = false
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if emptyCount >= 2 {
				break
			}

			if isRowEmpty {
				emptyCount++
				continue
			}

			if !isDataRow {
				continue
			}

			param := helpers.InsertParam{
				TableName: tablename,
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
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *RUPSController) readSDM(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("RUPS", "SDM", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	filename := filepath.Base(f.Path)
	splitted := strings.Split(filename, " ")
	tahun := splitted[3]

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if strings.TrimSpace(cellValue) == "SDM Organik" {
			cellValueAfter, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i+1))
			if err != nil {
				log.Fatal(err)
			}

			if strings.TrimSpace(cellValueAfter) != "SDM Organik" {
				firstDataRow = i + 1
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
	rowCount := 0
	emptyCount := 0
	no := 1

	tablename := "RUPS_SDM"

	// check if data exists
	sqlQuery := "SELECT * FROM " + tablename + " WHERE tahun = '" + tahun + "'"

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
			isDataRow := true

			for _, header := range headers {
				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				stringData = strings.ReplaceAll(stringData, "'", "''")

				if strings.TrimSpace(stringData) != "" {
					isRowEmpty = false
				}

				rowData.Set(header.DBFieldName, stringData)
			}

			if emptyCount >= 1 {
				break
			}

			if isRowEmpty {
				emptyCount++
				continue
			}

			if !isDataRow {
				continue
			}

			param := helpers.InsertParam{
				TableName: tablename,
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
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *RUPSController) readFinancialRatio(f *excelize.File, sheetName string) error {
	var err error

	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	gridsConfig := clit.Config("RUPS", "FinancialRatio", nil).(map[string]interface{})

	filename := filepath.Base(f.Path)
	splitted := strings.Split(filename, " ")
	tahun := splitted[3]

	tablename := "RUPS_Financial_Ratio"

	// check if data exists
	sqlQuery := "SELECT * FROM " + tablename + " WHERE tahun = '" + tahun + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From(tablename).SQL(sqlQuery), nil)
	defer cursor.Close()

	res := make([]toolkit.M, 0)
	err = cursor.Fetchs(&res, 0)

	//only insert if len of datas is 0 / if no data yet
	if len(res) == 0 {
		for tipe, configEachType := range gridsConfig {
			gridIdentifiers := map[string]string{}
			gridIdentifiers["PendapatanBeban"] = "PENDAPATAN VS BEBAN USAHA"
			gridIdentifiers["EatLaba"] = "EAT VS LABA USAHA"
			gridIdentifiers["RoaRoeEbitda"] = "ROA, ROE, EBITDA MARGIN"
			gridIdentifiers["DebtOrCash"] = "DEBT TO EQUITY, OR &CASH RATIO"

			records := map[string]interface{}{}

			for gridName, gridIdentifier := range gridIdentifiers {
				tableConfig := configEachType.(map[string]interface{})[gridName].(map[string]interface{})
				columnsMapping := tableConfig["columnsMapping"].(map[string]interface{})

				tableFound := false
				firstDataRow := 0
				i := 1
				for {
					if tableFound == false {
						cellValue, err := f.GetCellValue(sheetName, columnsMapping["Uraian"].(string)+toolkit.ToString(i))
						if err != nil {
							log.Fatal(err)
						}

						if strings.Contains(strings.ToUpper(cellValue), strings.ToUpper(gridIdentifier)) {
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

				rowCount := 0
				no := 1
				emptyCount := 0

				//iterate over rows
				for index := 0; true; index++ {
					rowData := toolkit.M{}
					currentRow := firstDataRow + index
					isRowEmpty := true

					for _, header := range headers {
						stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
						if err != nil {
							log.Fatal(err)
						}

						stringData = strings.ReplaceAll(stringData, "'", "''")

						if len(stringData) > 300 {
							stringData = stringData[0:300]
						}

						if strings.TrimSpace(stringData) != "" {
							isRowEmpty = false
						}

						rowData.Set(header.DBFieldName, stringData)
					}

					if emptyCount >= 2 {
						break
					}

					if isRowEmpty {
						emptyCount++
						continue
					}

					var data map[string]interface{}
					if records[rowData.GetString("Uraian")] == nil {
						data = map[string]interface{}{}
						data["Tahun"] = tahun
						data["Tipe"] = tipe
					} else {
						data = records[rowData.GetString("Uraian")].(map[string]interface{})
					}

					for key, val := range rowData {
						data[key] = val
					}

					records[rowData.GetString("Uraian")] = data

					rowCount++
					no++
				}
			}

			for _, rowData := range records {
				param := helpers.InsertParam{
					TableName: tablename,
					Data:      rowData,
				}

				err = helpers.Insert(param)
				if err != nil {
					log.Fatal("Error inserting row, ERROR:", err.Error())
				} else {
					log.Println("Row inserted.")
				}
			}
		}
	}

	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}
