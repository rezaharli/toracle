package controllers

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/sebar/dbflex"
	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"
)

var (
	dataPT   []string
	dataPTKV toolkit.M
)

type RKMController struct {
	*Base
}

func (c *RKMController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for RKM files.")
	c.FileExtension = ".xlsx"
}

func (c *RKMController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "KK BOD BOC")
}

func (c *RKMController) ReadExcel() {
	sheetAllowed := []string{"RKM", "INVESTASI", "SDM"}

	dataPT = []string{}
	dataPTKV = make(toolkit.M, 0)
	for _, sheetName := range c.Engine.GetSheetMap() {
		if sheetName == sheetAllowed[0] {
			c.ReadSheet(c.ReadDataRKM, sheetName)
		} else if sheetName == sheetAllowed[1] {
			c.ReadSheet(c.ReadDataInvestasi, sheetName)
		} else if sheetName == sheetAllowed[2] {
			c.ReadSheet(c.ReadDataSDM, sheetName)
		}
	}

	err := c.SetDataPTHub()
	if err != nil {
		log.Println(err)
	}
}

func (c *RKMController) ReadDataRKM(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	file := filepath.Base(c.Engine.GetExcelPath())
	monthsToNumber := clit.Config("kk_bod_boc_rkm", "monthsToNumber", nil).(map[string]interface{})

	monthNumber := -1
	monthStr := ""
	for k, v := range monthsToNumber {
		month := strings.Split(file, " ")
		monFound := false

		for _, mon := range month {
			if strings.ToLower(mon) == k {
				monthStr = mon
				monFound = true
			}
		}

		if monFound && monthStr != "" {
			monthNumber = int(v.(float64))
			break
		}
	}

	if monthNumber == -1 {
		helpers.HandleError(errors.New("month number not detected from filename"))
	}

	firstDataRow := 0
	for index := 2; true; index++ {
		PROGRAM_STRATEGIS, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(index))
		if err != nil {
			helpers.HandleError(err)
			break
		}

		if index > 20 {
			break
		}

		if PROGRAM_STRATEGIS != "PROGRAM STRATEGIS" {
			firstDataRow = index
			break
		}
	}

	columnInt := []string{"JUMLAH_PROGRAM", "PROSES", "BELUM_PROSES", "SELESAI"}
	columnsMapping := clit.Config("kk_bod_boc_rkm", "columnsMapping", nil).(map[string]interface{})
	mappingForPTTTL := map[string]interface{}{}
	mappingForPTLEGI := map[string]interface{}{}

	if mapping, ok := columnsMapping["PT TTL"]; ok {
		mappingForPTTTL = mapping.(map[string]interface{})
	}

	if mapping, ok := columnsMapping["PT LEGI"]; ok {
		mappingForPTLEGI = mapping.(map[string]interface{})
	}

	var headersTTL []Header
	var headersLEGI []Header

	for key, column := range mappingForPTTTL {
		header := Header{
			DBFieldName: key,
			Column:      column.(string),
		}

		headersTTL = append(headersTTL, header)
	}

	for key, column := range mappingForPTLEGI {
		header := Header{
			DBFieldName: key,
			Column:      column.(string),
		}

		headersLEGI = append(headersLEGI, header)
	}

	tahunStr, err := c.Engine.GetCellValue(sheetName, "C2")
	if err != nil {
		helpers.HandleError(err)
	}

	tahunStr = strings.Replace(tahunStr, "RKM ", "", -1)
	tipeTTL, err := c.Engine.GetCellValue(sheetName, "B1")
	if err != nil {
		helpers.HandleError(err)
	}

	if tipeTTL != "" {
		if _, ok := dataPTKV[tipeTTL]; !ok {
			dataPT = append(dataPT, tipeTTL)
			dataPTKV.Set(tipeTTL, tipeTTL)
		}
	}

	tipeLEGI, err := c.Engine.GetCellValue(sheetName, "H1")
	if err != nil {
		helpers.HandleError(err)
	}

	if tipeLEGI != "" {
		if _, ok := dataPTKV[tipeLEGI]; !ok {
			dataPT = append(dataPT, tipeLEGI)
			dataPTKV.Set(tipeLEGI, tipeLEGI)
		}
	}

	tahun := toolkit.ToInt(tahunStr, "")
	if err != nil {
		helpers.HandleError(err)
	}

	emptyRowCount := 0
	var rowDatas []toolkit.M
	rowCount := 0

	for index := 0; true; index++ {
		rowDataTTL := toolkit.M{"TAHUN": tahun, "TIPE": tipeTTL, "BULAN": monthNumber}
		rowDataLEGI := toolkit.M{"TAHUN": tahun, "TIPE": tipeLEGI, "BULAN": monthNumber}

		currentRow := firstDataRow + index
		isRowEmpty := true
		isBreak := false

		rowProcess := func(header Header, rowData toolkit.M) toolkit.M {
			stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
			if err != nil {
				helpers.HandleError(err)
			}
			stringData = strings.ReplaceAll(stringData, "'", "''")
			stringData = strings.ReplaceAll(stringData, "-", "")

			if strings.TrimSpace(stringData) != "" {
				isRowEmpty = false
			}

			if strings.Contains(strings.ToLower(stringData), "jumlah") {
				isBreak = true
			}

			isIntRow := false
			for _, columnint := range columnInt {
				if header.DBFieldName == columnint {
					isIntRow = true
					break
				}
			}

			if isIntRow {
				rowData.Set(header.DBFieldName, toolkit.ToInt(stringData, ""))
			} else {
				rowData.Set(header.DBFieldName, stringData)
			}

			return rowData
		}

		for _, header := range headersTTL {
			rowDataTTL = rowProcess(header, rowDataTTL)
		}

		for _, header := range headersLEGI {
			rowDataLEGI = rowProcess(header, rowDataLEGI)
		}

		if isBreak {
			break
		}

		if isRowEmpty {
			emptyRowCount++

			if emptyRowCount >= 10 {
				break
			}

			continue
		}

		rowDatas = append(rowDatas, rowDataTTL)
		rowDatas = append(rowDatas, rowDataLEGI)
		rowCount++
	}

	log.Println("Deleting data for ", monthStr+" "+toolkit.ToString(tahun))

	sql := "DELETE FROM BOD_RKM WHERE TAHUN = " + toolkit.ToString(tahun) + " AND (BULAN = " + toolkit.ToString(monthNumber) + " OR BULAN = -1)"

	conn := c.Conn
	query, err := conn.Prepare(dbflex.From("BOD_RKM").SQL(sql))
	if err != nil {
		log.Println(err)
	}

	fmt.Println("Done prepare")
	_, err = query.Execute(toolkit.M{}.Set("data", toolkit.M{}))
	if err != nil {
		log.Println(err)
	}

	log.Println("Period ", monthStr+" "+toolkit.ToString(tahun), "deleted.")

	param := helpers.InsertParam{
		TableName: "BOD_RKM",
		Data:      rowDatas,
	}

	err = helpers.InsertWithConn(param, c.Conn)

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *RKMController) ReadDataInvestasi(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	file := filepath.Base(c.Engine.GetExcelPath())
	monthsToNumber := clit.Config("kk_bod_boc_rkm", "monthsToNumber", nil).(map[string]interface{})

	monthNumber := -1
	monthYearTemp := strings.Split(file, " ")
	if len(monthYearTemp) != 5 {
		helpers.HandleError(errors.New("filename format wrong"))
	}

	monthStr := monthYearTemp[3]
	yearStr := strings.Replace(monthYearTemp[4], ".xlsx", "", -1)
	tahun := toolkit.ToInt(strings.Replace(yearStr, ".xls", "", -1), "")

	for k, v := range monthsToNumber {
		if strings.ToLower(monthStr) == k {
			monthNumber = int(v.(float64))
		}
	}

	if monthNumber == -1 {
		helpers.HandleError(errors.New("month number not detected from filename"))
	}

	firstDataRow := 4
	columnInt := []string{"PROGRAM_ANGGARAN", "NILAI_PROYEK", "RKAP", "REVISI", "KONTRAK_KESELURUHAN", "KONTRAK_BERJALAN", "PROGRAM", "FISIK", "NILAI_SERAPAN"}
	columnsMapping := clit.Config("kk_bod_boc_investasi", "columnsMapping", nil).(map[string]interface{})

	headers := []Header{}

	for key, column := range columnsMapping {
		header := Header{
			DBFieldName: key,
			Column:      column.(string),
		}

		headers = append(headers, header)
	}

	emptyRowCount := 0
	var rowDatas []toolkit.M
	rowCount := 0

	for index := 0; true; index++ {
		rowData := toolkit.M{"TAHUN": tahun, "BULAN": monthNumber}
		currentRow := firstDataRow + index
		isRowEmpty := true

		footer, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
		if err != nil {
			helpers.HandleError(err)
		}

		if strings.Contains(footer, "JUMLAH") {
			break
		}

		if index > 200 {
			break
		}

		rowProcess := func(header Header, rowData toolkit.M) toolkit.M {
			stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
			if err != nil {
				helpers.HandleError(err)
			}
			stringData = strings.ReplaceAll(stringData, "'", "''")
			stringData = strings.ReplaceAll(stringData, "-", "")

			if strings.TrimSpace(stringData) != "" {
				isRowEmpty = false
			}

			isIntRow := false
			for _, columnint := range columnInt {
				if header.DBFieldName == columnint {
					isIntRow = true
					break
				}
			}

			if isIntRow {
				stringData = strings.ReplaceAll(stringData, ",", "")
				rowData.Set(header.DBFieldName, toolkit.ToInt(stringData, ""))
			} else {
				rowData.Set(header.DBFieldName, stringData)
			}

			return rowData
		}

		for _, header := range headers {
			rowData = rowProcess(header, rowData)
		}

		if isRowEmpty {
			emptyRowCount++

			if emptyRowCount >= 10 {
				break
			}

			continue
		}

		rowDatas = append(rowDatas, rowData)
		rowCount++
	}

	log.Println("Deleting data for ", monthStr+" "+toolkit.ToString(tahun))
	sql := "DELETE FROM BOD_INVESTASI WHERE TAHUN = " + toolkit.ToString(tahun) + " AND (BULAN = " + toolkit.ToString(monthNumber) + " OR BULAN = -1)"
	conn := c.Conn
	query, err := conn.Prepare(dbflex.From("BOD_INVESTASI").SQL(sql))
	if err != nil {
		log.Println(err)
	}

	_, err = query.Execute(toolkit.M{}.Set("data", toolkit.M{}))
	if err != nil {
		log.Println(err)
	}

	log.Println("Period ", monthStr+" "+toolkit.ToString(tahun), "deleted.")

	param := helpers.InsertParam{
		TableName: "BOD_INVESTASI",
		Data:      rowDatas,
	}

	err = helpers.InsertWithConn(param, c.Conn)

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *RKMController) ReadDataSDM(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	file := filepath.Base(c.Engine.GetExcelPath())
	monthsToNumber := clit.Config("kk_bod_boc_sdm", "monthsToNumber", nil).(map[string]interface{})

	monthNumber := -1
	monthYearTemp := strings.Split(file, " ")
	if len(monthYearTemp) != 5 {
		helpers.HandleError(errors.New("filename format wrong"))
	}

	monthStr := monthYearTemp[3]
	yearStr := strings.Replace(monthYearTemp[4], ".xlsx", "", -1)
	tahun := toolkit.ToInt(strings.Replace(yearStr, ".xls", "", -1), "")

	for k, v := range monthsToNumber {
		if strings.ToLower(monthStr) == k {
			monthNumber = int(v.(float64))
		}
	}

	if monthNumber == -1 {
		helpers.HandleError(errors.New("month number not detected from filename"))
	}

	firstDataRow := 5

	columnInt := []string{"RKAP", "REALISASI"}
	columnsMapping := clit.Config("kk_bod_boc_sdm", "columnsMapping", nil).(map[string]interface{})

	var headers []Header

	for key, column := range columnsMapping {
		header := Header{
			DBFieldName: key,
			Column:      column.(string),
		}

		headers = append(headers, header)
	}

	emptyRowCount := 0
	var rowDatas []toolkit.M
	rowCount := 0

	for index := 0; true; index++ {
		rowData := toolkit.M{"TAHUN": tahun, "BULAN": monthNumber}

		currentRow := firstDataRow + index
		isRowEmpty := true
		isBreak := false
		isSkippedRow := false

		rowProcess := func(header Header, rowData toolkit.M) toolkit.M {
			stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
			if err != nil {
				helpers.HandleError(err)
			}
			stringData = strings.ReplaceAll(stringData, "'", "''")
			stringData = strings.ReplaceAll(stringData, "-", "")

			if strings.TrimSpace(stringData) != "" {
				isRowEmpty = false
			}

			if strings.Contains(stringData, "JUMLAH") {
				isSkippedRow = true
			} else if strings.Contains(stringData, "TOTAL") {
				isBreak = true
			} else {
				if header.DBFieldName == "URAIAN" {

					if _, ok := dataPTKV[stringData]; !ok {
						dataPT = append(dataPT, stringData)
						dataPTKV.Set(stringData, stringData)
					}
				}

				rowData.Set(header.DBFieldName, stringData)
			}

			isIntRow := false
			for _, columnint := range columnInt {
				if header.DBFieldName == columnint {
					isIntRow = true
					break
				}
			}

			if isIntRow {
				rowData.Set(header.DBFieldName, toolkit.ToInt(stringData, ""))
			}

			return rowData
		}

		for _, header := range headers {
			rowData = rowProcess(header, rowData)
		}

		if isSkippedRow {
			continue
		} else if isBreak {
			break
		}

		if isRowEmpty {
			emptyRowCount++

			if emptyRowCount >= 10 {
				break
			}

			continue
		}

		rowDatas = append(rowDatas, rowData)
		rowCount++
	}

	log.Println("Deleting data for ", monthStr+" "+toolkit.ToString(tahun))
	sql := "DELETE FROM BOD_SDM WHERE TAHUN = " + toolkit.ToString(tahun) + " AND (BULAN = " + toolkit.ToString(monthNumber) + " OR BULAN = -1)"
	conn := c.Conn
	query, err := conn.Prepare(dbflex.From("BOD_SDM").SQL(sql))
	if err != nil {
		log.Println(err)
	}

	_, err = query.Execute(toolkit.M{}.Set("data", toolkit.M{}))
	if err != nil {
		log.Println(err)
	}

	log.Println("Period ", monthStr+" "+toolkit.ToString(tahun), "deleted.")

	param := helpers.InsertParam{
		TableName: "BOD_SDM",
		Data:      rowDatas,
	}

	err = helpers.InsertWithConn(param, c.Conn)

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *RKMController) SetDataPTHub() error {
	sqlQuery := "SELECT * FROM BOD_PT_HUB"

	conn := c.Conn
	cursor := conn.Cursor(dbflex.From("FROM BOD_PT_HUB").SQL(sqlQuery), nil)
	defer cursor.Close()

	datas := []toolkit.M{}

	err := cursor.Fetchs(datas, 0)

	for _, data := range datas {
		pt := data.GetString("PT")
		if pt != "" {
			if _, ok := dataPTKV[pt]; !ok {
				dataPT = append(dataPT, pt)
				dataPTKV.Set(pt, pt)
			}
		}
	}

	log.Println("Deleting data")

	sql := "DELETE FROM BOD_PT_HUB"
	query, err := conn.Prepare(dbflex.From("BOD_PT_HUB").SQL(sql))
	if err != nil {
		log.Println(err)
	}

	fmt.Println("Done prepare")
	_, err = query.Execute(toolkit.M{}.Set("data", toolkit.M{}))
	if err != nil {
		log.Println(err)
	}

	log.Println("Data deleted.")

	rowDatas := []toolkit.M{}
	for _, pt := range dataPT {
		rowData := toolkit.M{"PT": pt}
		rowDatas = append(rowDatas, rowData)
	}

	param := helpers.InsertParam{
		TableName: "BOD_PT_HUB",
		Data:      rowDatas,
	}

	err = helpers.InsertWithConn(param, c.Conn)
	if err == nil {
		log.Println("data inserted")
	}

	return err
}
