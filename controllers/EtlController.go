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

type EtlController struct {
	*Base
}

func NewEtlController() *EtlController {
	return new(EtlController)
}

func (c *EtlController) ReadExcels() error {
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

func (c *EtlController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "EnMS") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. Etl files found:", len(resourceFiles))
	return resourceFiles
}

func (c *EtlController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	for _, sheetName := range f.GetSheetMap() {
		if strings.Contains(sheetName, "GRK") {
			err = c.ReadDataGRK(f, sheetName)
			if err != nil {
				log.Println("Error reading data. ERROR:", err)
			}
		}

		if strings.Contains(sheetName, "Konsumsi BBM per Alat") {
			err = c.ReadDataEnergyItemBBM(f, sheetName)
			if err != nil {
				log.Println("Error reading data. ERROR:", err)
			}
		}

		if strings.Contains(sheetName, "Konsumsi Listrik per Alat") {
			err = c.ReadDataEnergyItemListrik(f, sheetName)
			if err != nil {
				log.Println("Error reading data. ERROR:", err)
			}
		}

		if strings.Contains(sheetName, "Energy Performance") {
			err = c.ReadDataPerformance(f, sheetName)
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

func (c *EtlController) ReadDataGRK(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("etl", "GRK", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		_, err = strconv.Atoi(cellValue)
		if err == nil {
			//jika tahun
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

	months := config["months"].([]interface{})

	var err error
	// var rowDatas []toolkit.M
	rowCount := 0

	//iterate over rows
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index
		isRowEmpty := true

		cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue != "" {
			isRowEmpty = false
		}

		_, err = strconv.Atoi(cellValue)
		if isRowEmpty == false {
			if err != nil {
				//jika bukan tahun
				continue
			}
		}

		skipRow := true

		for _, header := range headers {
			if header.DBFieldName == "PERIOD" {
				stringDataYear, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				stringDataMonth, err := f.GetCellValue(sheetName, "C"+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				stringData := "1/" + toolkit.ToString(helpers.IndexOf(stringDataMonth, months)+1) + "/" + stringDataYear

				stringData = strings.ReplaceAll(stringData, "'", "")
				stringData = strings.ReplaceAll(stringData, "`", "")

				var t time.Time
				if stringDataYear != "" {
					isRowEmpty = false
					t, err = time.Parse("2-Jan-06", stringData)
					if err != nil {
						t, err = time.Parse("02/01/2006", stringData)
						if err != nil {
							t, err = time.Parse("2/1/2006", stringData)
							if err != nil {
								log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
							}
						}
					}
				}

				rowData.Set(header.DBFieldName, t)
			} else {
				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				stringData = strings.ReplaceAll(stringData, "'", "''")

				if len(stringData) > 300 {
					stringData = stringData[0:300]
				}

				if stringData != "" {
					isRowEmpty = false
					skipRow = false
				}

				rowData.Set(header.DBFieldName, stringData)
			}
		}

		if isRowEmpty {
			break
		}

		if skipRow {
			continue
		}

		// check if data exists
		sqlQuery := "SELECT * FROM F_QHSSE_ENERGYGRK WHERE trunc(period) = TO_DATE('" + rowData.Get("PERIOD").(time.Time).Format("2006-01-02") + "', 'YYYY-MM-DD')"

		conn := helpers.Database()
		cursor := conn.Cursor(dbflex.From("F_QHSSE_ENERGYGRK").SQL(sqlQuery), nil)
		defer cursor.Close()

		res := make([]toolkit.M, 0)
		err = cursor.Fetchs(&res, 0)

		//only insert if len of datas in currentPeriod is 0 / if no data yet
		if len(res) == 0 {
			param := helpers.InsertParam{
				TableName: "F_QHSSE_ENERGYGRK",
				Data:      rowData,
			}

			err = helpers.Insert(param)
			if err != nil {
				log.Fatal("Error inserting row "+toolkit.ToString(currentRow)+", ERROR:", err.Error())
			} else {
				log.Println("Row", currentRow, "inserted.")
			}
		} else {
			log.Println("Skipping", rowData.Get("PERIOD").(time.Time).Format("2006-01-02"))
		}
		rowCount++
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *EtlController) ReadDataEnergyItemBBM(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)

	config := clit.Config("etl", "EnergyBBM", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		_, err = strconv.Atoi(cellValue)
		if err == nil {
			//jika angka
			firstDataRow = i
			break
		}
		i++
	}

	months := config["months"].([]interface{})

	monthRow := firstDataRow - 2
	var monthHeaders []Header
	isHeaderDetected := false

	i = 1
	prevCell := ""
	for {
		header := Header{
			DBFieldName:  "",
			Column:       "",
			ColumnNumber: i,
		}

		currentCol := helpers.ToCharStr(i)
		cellText, err := f.GetCellValue(sheetName, currentCol+toolkit.ToString(monthRow))
		if err != nil {
			log.Fatal(err)
		}

		if isHeaderDetected == false && strings.TrimSpace(cellText) != "" {
			isHeaderDetected = true
		}

		if isHeaderDetected == true && strings.TrimSpace(cellText) == "" {
			break
		}

		if isHeaderDetected {
			if strings.TrimSpace(cellText) != "" && helpers.ArrayContainsWhitespaceTrimmed(months, cellText) != -1 {
				if strings.TrimSpace(cellText) != strings.TrimSpace(prevCell) {
					header.HeaderName = cellText
					header.Column = currentCol

					monthHeaders = append(monthHeaders, header)

					prevCell = cellText
				}
			}
		}

		i++
	}

	rowCount := 0
	var err error
	for _, monthHeader := range monthHeaders {
		var headers []Header
		for key, column := range columnsMapping {

			header := Header{
				DBFieldName: key,
				Column:      column.(string),
			}

			if key == "PERIOD" {
				header.Value = monthHeader.HeaderName
				header.Column = monthHeader.Column
			}

			if key == "TOTAL_CONSUMPTION" {
				header.Column = monthHeader.Column
			}

			if key == "TOTAL_PRODUCTION" {
				header.Column = helpers.ToCharStr(monthHeader.ColumnNumber + 1)
			}

			headers = append(headers, header)
		}

		isMonthlyAdaIsinya := map[time.Time]bool{}
		rowDatas := make([]toolkit.M, 0)

		//iterate over rows
		for index := 0; true; index++ {
			rowData := toolkit.M{}
			currentRow := firstDataRow + index
			isRowEmpty := true

			cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			if strings.TrimSpace(cellValue) != "" {
				isRowEmpty = false
			}

			_, err = strconv.Atoi(cellValue)
			if isRowEmpty == false {
				if err != nil {
					//jika bukan tahun
					continue
				}
			}

			isRowAdaDataKonsumsiProduksinya := false

			for _, header := range headers {
				if header.DBFieldName == "PERIOD" {
					splitted := strings.Split(sheetName, " ")
					stringDataYear := splitted[len(splitted)-1]

					stringDataMonth := strings.TrimSpace(header.Value)

					stringData := "1/" + toolkit.ToString(helpers.IndexOf(stringDataMonth, months)+1) + "/" + stringDataYear

					stringData = strings.ReplaceAll(stringData, "'", "")
					stringData = strings.ReplaceAll(stringData, "`", "")

					var t time.Time
					if stringData != "" {
						t, err = time.Parse("2-Jan-06", stringData)
						if err != nil {
							t, err = time.Parse("02/01/2006", stringData)
							if err != nil {
								t, err = time.Parse("2/1/2006", stringData)
								if err != nil {
									log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
								}
							}
						}
					}

					rowData.Set(header.DBFieldName, t)
				} else if header.DBFieldName == "ITEM_ID" {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					resultRows := make([]toolkit.M, 0)
					param := SqlQueryParam{
						ItemName: strings.ReplaceAll(stringData, "-", ""),
						Results:  &resultRows,
					}

					err = c.selectItemID(param)
					if err != nil {
						log.Fatal(err)
					}

					if stringData != "" {
						isRowEmpty = false
					}

					if len(resultRows) > 0 {
						rowData.Set("Nama Alat", param.ItemName)
						rowData.Set(header.DBFieldName, resultRows[0].GetString("ITEM_ID"))
					} else {
						rowData.Set(header.DBFieldName, nil)
					}
				} else if header.DBFieldName == "ENERGY_TYPE" {
					splitted := strings.Split(sheetName, " ")
					stringData := splitted[1]

					rowData.Set(header.DBFieldName, stringData)
				} else {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")
					stringData = strings.ReplaceAll(stringData, "-", "")

					stringData = strings.TrimSpace(stringData)

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					if stringData != "" {
						isRowEmpty = false
						isRowAdaDataKonsumsiProduksinya = true
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if isRowEmpty {
				break
			}

			if ok := isMonthlyAdaIsinya[rowData.Get("PERIOD").(time.Time)]; !ok {
				isMonthlyAdaIsinya[rowData.Get("PERIOD").(time.Time)] = false
			}

			if isRowAdaDataKonsumsiProduksinya == true {
				isMonthlyAdaIsinya[rowData.Get("PERIOD").(time.Time)] = true
			}

			rowDatas = append(rowDatas, rowData)

			rowCount++
		}

		for _, rowData := range rowDatas {
			if isMonthlyAdaIsinya[rowData.Get("PERIOD").(time.Time)] == true {
				currentAlat := rowData.GetString("Nama Alat")
				rowData.Unset("Nama Alat")

				// check if data exists
				sqlQuery := "SELECT PERIOD, ENERGY_TYPE FROM F_QHSSE_ENERGY_ITEM WHERE ENERGY_TYPE = '" + rowData.GetString("ENERGY_TYPE") + "' AND trunc(period) = TO_DATE('" + rowData.Get("PERIOD").(time.Time).Format("2006-01-02") + "', 'YYYY-MM-DD')"

				conn := helpers.Database()
				cursor := conn.Cursor(dbflex.From("F_QHSSE_ENERGY_ITEM").SQL(sqlQuery), nil)
				defer cursor.Close()

				res := make([]toolkit.M, 0)
				err = cursor.Fetchs(&res, 0)
				if err != nil {
					log.Println(err)
				}

				//only insert if len of datas in currentPeriod is 0 / if no data yet
				if len(res) == 0 {
					param := helpers.InsertParam{
						TableName: "F_QHSSE_ENERGY_ITEM",
						Data:      rowData,
					}

					err = helpers.Insert(param)
					if err != nil {
						log.Fatal("Error inserting "+monthHeader.HeaderName+",", currentAlat+", ERROR:", err.Error())
					} else {
						log.Println(monthHeader.HeaderName+",", currentAlat+", inserted.")
					}
				} else {
					log.Println("Skipping", rowData.Get("PERIOD").(time.Time).Format("2006-01-02"), currentAlat, ".")
				}
			}
		}
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *EtlController) ReadDataEnergyItemListrik(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)

	config := clit.Config("etl", "EnergyListrik", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		_, err = strconv.Atoi(cellValue)
		if err == nil {
			//jika angka
			firstDataRow = i
			break
		}
		i++
	}

	months := config["months"].([]interface{})

	monthRow := firstDataRow - 2
	var monthHeaders []Header
	isHeaderDetected := false

	i = 1
	prevCell := ""
	for {
		header := Header{
			DBFieldName:  "",
			Column:       "",
			ColumnNumber: i,
		}

		currentCol := helpers.ToCharStr(i)
		cellText, err := f.GetCellValue(sheetName, currentCol+toolkit.ToString(monthRow))
		if err != nil {
			log.Fatal(err)
		}

		if isHeaderDetected == false && strings.TrimSpace(cellText) != "" {
			isHeaderDetected = true
		}

		if isHeaderDetected == true && strings.TrimSpace(cellText) == "" {
			break
		}

		if isHeaderDetected {
			if strings.TrimSpace(cellText) != "" && helpers.ArrayContainsWhitespaceTrimmed(months, strings.TrimSpace(cellText)) != -1 {
				if strings.TrimSpace(cellText) != strings.TrimSpace(prevCell) {
					header.HeaderName = strings.TrimSpace(cellText)
					header.Column = currentCol

					monthHeaders = append(monthHeaders, header)

					prevCell = strings.TrimSpace(cellText)
				}
			}
		}

		i++
	}

	rowCount := 0
	var err error
	for _, monthHeader := range monthHeaders {
		var headers []Header
		for key, column := range columnsMapping {

			header := Header{
				DBFieldName: key,
				Column:      column.(string),
			}

			if key == "PERIOD" {
				header.Value = monthHeader.HeaderName
				header.Column = monthHeader.Column
			}

			if key == "TOTAL_CONSUMPTION" {
				header.Column = helpers.ToCharStr(monthHeader.ColumnNumber + 2)
			}

			if key == "TOTAL_PRODUCTION" {
				header.Column = helpers.ToCharStr(monthHeader.ColumnNumber + 3)
			}

			headers = append(headers, header)
		}

		//iterate over rows
		for index := 0; true; index++ {
			rowData := toolkit.M{}
			currentRow := firstDataRow + index
			isRowEmpty := true

			cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			if strings.TrimSpace(cellValue) != "" {
				isRowEmpty = false
			}

			_, err = strconv.Atoi(cellValue)
			if isRowEmpty == false {
				if err != nil {
					//jika bukan tahun
					continue
				}
			}

			for _, header := range headers {
				if header.DBFieldName == "PERIOD" {
					splitted := strings.Split(sheetName, " ")
					stringDataYear := splitted[len(splitted)-1]

					stringDataMonth := strings.TrimSpace(header.Value)

					stringData := "1/" + toolkit.ToString(helpers.IndexOf(stringDataMonth, months)+1) + "/" + stringDataYear

					stringData = strings.ReplaceAll(stringData, "'", "")
					stringData = strings.ReplaceAll(stringData, "`", "")

					var t time.Time
					if stringData != "" {
						t, err = time.Parse("2-Jan-06", stringData)
						if err != nil {
							t, err = time.Parse("02/01/2006", stringData)
							if err != nil {
								t, err = time.Parse("2/1/2006", stringData)
								if err != nil {
									log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
								}
							}
						}
					}

					rowData.Set(header.DBFieldName, t)
				} else if header.DBFieldName == "ITEM_ID" {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					resultRows := make([]toolkit.M, 0)
					param := SqlQueryParam{
						ItemName: strings.ReplaceAll(stringData, "-", ""),
						Results:  &resultRows,
					}

					err = c.selectItemID(param)
					if err != nil {
						log.Fatal(err)
					}

					if stringData != "" {
						isRowEmpty = false
					}

					if len(resultRows) > 0 {
						rowData.Set(header.DBFieldName, resultRows[0].GetString("ITEM_ID"))
					} else {
						rowData.Set(header.DBFieldName, nil)
					}
				} else if header.DBFieldName == "ENERGY_TYPE" {
					splitted := strings.Split(sheetName, " ")
					stringData := splitted[1]

					rowData.Set(header.DBFieldName, stringData)
				} else {
					stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.ReplaceAll(stringData, "'", "''")
					stringData = strings.ReplaceAll(stringData, "-", "")

					stringData = strings.TrimSpace(stringData)

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
				break
			}

			param := helpers.InsertParam{
				TableName: "F_QHSSE_ENERGY_ITEM",
				Data:      rowData,
			}

			err = helpers.Insert(param)
			if err != nil {
				log.Fatal("Error inserting "+monthHeader.HeaderName+" row "+toolkit.ToString(currentRow)+", ERROR:", err.Error())
			} else {
				log.Println(monthHeader.HeaderName+" Row", currentRow, "inserted.")
			}
			rowCount++
		}
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *EtlController) ReadDataPerformance(f *excelize.File, sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("etl", "EnergyCO2", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		_, err = strconv.Atoi(cellValue)
		if err == nil {
			//jika tahun
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

	months := config["months"].([]interface{})

	var err error
	// var rowDatas []toolkit.M
	rowCount := 0
	//iterate over rows
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index
		isRowEmpty := true

		cellValue, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue != "" {
			isRowEmpty = false
		}

		_, err = strconv.Atoi(cellValue)
		if isRowEmpty == false {
			if err != nil {
				//jika bukan tahun
				continue
			}
		}

		for _, header := range headers {
			if strings.EqualFold(header.DBFieldName, "PERIOD") {
				stringDataYear, err := f.GetCellValue(sheetName, "B"+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				stringDataMonth, err := f.GetCellValue(sheetName, "C"+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				stringData := "1/" + toolkit.ToString(helpers.IndexOf(stringDataMonth, months)+1) + "/" + stringDataYear

				stringData = strings.ReplaceAll(stringData, "'", "")
				stringData = strings.ReplaceAll(stringData, "`", "")

				var t time.Time
				if stringDataYear != "" {
					isRowEmpty = false
					t, err = time.Parse("2-Jan-06", stringData)
					if err != nil {
						t, err = time.Parse("02/01/2006", stringData)
						if err != nil {
							t, err = time.Parse("2/1/2006", stringData)
							if err != nil {
								log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
							}
						}
					}
				}

				rowData.Set(header.DBFieldName, t)
			} else {
				stringData, err := f.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
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
			break
		}

		param := helpers.InsertParam{
			TableName: "F_QHSSE_ENERGY_CO2",
			Data:      rowData,
		}

		err = helpers.Insert(param)
		if err != nil {
			log.Fatal("Error inserting row "+toolkit.ToString(currentRow)+", ERROR:", err.Error())
		} else {
			log.Println("Row", currentRow, "inserted.")
		}
		rowCount++
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *EtlController) selectItemID(param SqlQueryParam) error {
	sqlQuery := "SELECT * FROM D_Item WHERE ITEM_NAME = TRIM('" + param.ItemName + "')"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From("D_Item").SQL(sqlQuery), nil)
	defer cursor.Close()

	err := cursor.Fetchs(param.Results, 0)

	return err
}
