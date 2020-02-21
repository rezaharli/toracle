package controllers

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/sebar/dbflex"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
)

// CttController is a controller for every kind of CTT files.
type CttController struct {
	*Base
}

// New is used to initiate the controller
func (c *CttController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for CTT files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *CttController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "CTT")
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *CttController) ReadExcel() {
	for _, sheetName := range c.Engine.GetSheetMap() {
		if strings.EqualFold(sheetName, "Daily") {
			c.ReadSheet(c.ReadDataDaily, sheetName)
		} else if strings.EqualFold(sheetName, "EQP") {
			c.ReadSheet(c.ReadDataMonthly, sheetName)
		}
	}
}

func (c *CttController) ReadDataDaily(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("ctt", "Daily", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	var currentPeriod time.Time

	var err error

	//iterate into groups of data
	notPeriodCount := 0
	periodFound := false
	row := 1
	for {
		firstDataRow := 0
		notPeriodCount = 0
		for {
			style, _ := c.Engine.NewStyle(`{"number_format":15}`)
			c.Engine.SetCellStyle(sheetName, "A"+toolkit.ToString(row), "A"+toolkit.ToString(row), style)
			stringData, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(row))
			if err != nil {
				log.Fatal(err)
			}

			//check if value is a period
			t, err := time.Parse("2-Jan-06", stringData)
			if err == nil {
				currentPeriod = t
				firstDataRow = row + 7

				stringData, err = c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(firstDataRow))
				if err != nil {
					log.Fatal(err)
				}

				if strings.EqualFold(strings.TrimSpace(stringData), strings.TrimSpace("UNIT")) {
					firstDataRow = firstDataRow + 1
				}

				periodFound = true
				break
			}

			if notPeriodCount > 100 {
				periodFound = false
				break
			}

			notPeriodCount++
			row++
		}

		if !periodFound {
			break
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
		currentItemID := ""

		tablename := "F_ENG_CTT_DAILY"

		// check if data exists
		sqlQuery := "SELECT * FROM " + tablename + " WHERE trunc(period) = TO_DATE('" + currentPeriod.Format("2006-01-02") + "', 'YYYY-MM-DD')"

		conn := helpers.Database()
		cursor := conn.Cursor(dbflex.From(tablename).SQL(sqlQuery), nil)
		defer cursor.Close()

		res := make([]toolkit.M, 0)
		err := cursor.Fetchs(&res, 0)

		//only insert if len of datas in currentPeriod is 0 / if no data yet
		if len(res) == 0 {
			//iterate over rows
			for index := 0; true; index++ {
				rowData := toolkit.M{}
				currentRow := firstDataRow + index
				row = currentRow
				isRowEmpty := true
				dontInsert := false

				for _, header := range headers {
					if header.DBFieldName == "PERIOD" {
						rowData.Set(header.DBFieldName, currentPeriod)
					} else if header.DBFieldName == "ITEM_ID" {
						stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
						if err != nil {
							log.Fatal(err)
						}

						if currentItemID == stringData {
							dontInsert = true
						}

						currentItemID = stringData

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
					} else if header.DBFieldName == "START_TIME" {
						cellID := header.Column + toolkit.ToString(currentRow)
						style, _ := c.Engine.NewStyle(`{"number_format":15}`)
						c.Engine.SetCellStyle(sheetName, cellID, cellID, style)
						startDate, err := c.Engine.GetCellValue(sheetName, cellID)
						if err != nil {
							log.Fatal(err)
						}

						cellID = helpers.ToCharStr(helpers.CharStrToNum(header.Column)+2) + toolkit.ToString(currentRow)
						startTime, err := c.Engine.GetCellValue(sheetName, cellID)
						if err != nil {
							log.Fatal(err)
						}

						if startDate != "" && startTime != "" {
							isRowEmpty = false
						}

						t, _ := time.Parse("2-Jan-06-15:04", startDate+"-"+startTime)

						rowData.Set(header.DBFieldName, t)
					} else if header.DBFieldName == "END_TIME" {
						cellID := header.Column + toolkit.ToString(currentRow)
						style, _ := c.Engine.NewStyle(`{"number_format":15}`)
						c.Engine.SetCellStyle(sheetName, cellID, cellID, style)
						startDate, err := c.Engine.GetCellValue(sheetName, cellID)
						if err != nil {
							log.Fatal(err)
						}

						cellID = helpers.ToCharStr(helpers.CharStrToNum(header.Column)+2) + toolkit.ToString(currentRow)
						startTime, err := c.Engine.GetCellValue(sheetName, cellID)
						if err != nil {
							log.Fatal(err)
						}

						if startDate != "" && startTime != "" {
							isRowEmpty = false
						}

						t, _ := time.Parse("2-Jan-06-15:04", startDate+"-"+startTime)

						rowData.Set(header.DBFieldName, t)
					} else if header.DBFieldName == "QTY" {
						stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
						if err != nil {
							log.Fatal(err)
						}

						stringData = strings.ReplaceAll(stringData, "-", "")

						if stringData != "" {
							isRowEmpty = false
						}

						intData, err := strconv.Atoi(stringData)
						if err != nil {
							intData = 0
						}

						rowData.Set(header.DBFieldName, intData)
					} else {
						stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
						if err != nil {
							log.Fatal(err)
						}

						stringData = strings.TrimSpace(strings.ReplaceAll(stringData, "'", "''"))

						if len(stringData) > 300 {
							stringData = stringData[0:300]
						}

						if stringData != "" {
							isRowEmpty = false
						}

						rowData.Set(header.DBFieldName, stringData)
					}
				}

				if dontInsert {
					continue
				}

				if isRowEmpty {
					break
				}

				param := helpers.InsertParam{
					TableName: tablename,
					Data:      rowData,
				}

				toolkit.Println("Inserting...")
				err = helpers.Insert(param)
				if err != nil {
					log.Fatal("Error inserting row "+toolkit.ToString(currentRow)+", ERROR:", err.Error())
				} else {
					log.Println("Row", currentRow, "inserted.")
				}
				rowCount++
			}

			if err == nil {
				log.Println("SUCCESS Processing", rowCount, "rows\n")
			}
		} else {
			log.Println("Skipping", currentPeriod.Format("2006-01-02"))
		}

		row++
	}

	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *CttController) ReadDataMonthly(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	config := clit.Config("ctt", "Monthly", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	var currentPeriod time.Time

	var err error

	//iterate into groups of data
	notPeriodCount := 0
	periodFound := false
	row := 1
	isPeriodDeleted := map[time.Time]bool{}

	for {
		firstDataRow := 0
		notPeriodCount = 0

		//search for period
		for {
			if row >= 1 {
				style, _ := c.Engine.NewStyle(`{"number_format":15}`)
				c.Engine.SetCellStyle(sheetName, "A"+toolkit.ToString(row), "A"+toolkit.ToString(row), style)
				stringData, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(row))
				if err != nil {
					log.Fatal(err)
				}

				splitted := strings.Split(stringData, " ")
				monthYear := ""
				if len(splitted) >= 2 {
					monthYear = "1-" + splitted[len(splitted)-2] + "-" + splitted[len(splitted)-1]
				}

				//check if value is a period
				t, err := time.Parse("2-January-2006", monthYear)
				if err == nil {
					notPeriodCount = 0

					stringDataAcuan, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(row-1))
					if err != nil {
						log.Fatal(err)
					}

					if !strings.Contains(stringDataAcuan, "EQUIPMENT PERFORMANCE DATA") {
						row++
						continue
					}

					currentPeriod = t
					firstDataRow = row + 5

					stringData, err = c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(firstDataRow))
					if err != nil {
						log.Fatal(err)
					}

					if strings.EqualFold(strings.TrimSpace(stringData), strings.TrimSpace("NO")) {
						firstDataRow = firstDataRow + 1
					}

					periodFound = true
					break
				}
			}

			if notPeriodCount > 200 {
				periodFound = false
				break
			}

			notPeriodCount++
			row++
		}

		if !periodFound {
			break
		}

		var headers []Header
		for key, column := range columnsMapping {
			header := Header{
				DBFieldName: key,
				Column:      column.(string),
			}

			headers = append(headers, header)
		}

		// check, delete if data exists
		if isPeriodDeleted[currentPeriod] == false {
			log.Println("Deleting period", currentPeriod.Format("2006-01-02"))

			sql := "DELETE FROM F_ENG_EQUIPMENT_MONTHLY WHERE trunc(period) = TO_DATE('" + currentPeriod.Format("2006-01-02") + "', 'YYYY-MM-DD')"

			conn := helpers.Database()
			query, err := conn.Prepare(dbflex.From("F_ENG_EQUIPMENT_MONTHLY").SQL(sql))
			if err != nil {
				log.Println(err)
			}

			_, err = query.Execute(toolkit.M{}.Set("data", toolkit.M{}))
			if err != nil {
				log.Println(err)
			}

			log.Println("Period", currentPeriod.Format("2006-01-02"), "deleted.")
			isPeriodDeleted[currentPeriod] = true
		}

		//iterate over rows
		rowCount := 0
		currentItemID := ""
		for index := 0; true; index++ {
			rowData := toolkit.M{}
			currentRow := firstDataRow + index
			row = currentRow
			dontInsert := false

			stringData, err := c.Engine.GetCellValue(sheetName, "A"+toolkit.ToString(currentRow))
			if err != nil {
				log.Fatal(err)
			}

			if strings.Contains(stringData, "Total") {
				break
			}

			for _, header := range headers {
				if header.DBFieldName == "PERIOD" {
					rowData.Set(header.DBFieldName, currentPeriod)
				} else if header.DBFieldName == "ITEM_ID" {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					if currentItemID == stringData {
						dontInsert = true
					}

					currentItemID = stringData

					resultRows := make([]toolkit.M, 0)
					param := SqlQueryParam{
						ItemName: strings.ReplaceAll(stringData, "-", ""),
						Results:  &resultRows,
					}

					err = c.selectItemID(param)
					if err != nil {
						log.Fatal(err)
					}

					if len(resultRows) > 0 {
						rowData.Set(header.DBFieldName, resultRows[0].GetString("ITEM_ID"))
					} else {
						rowData.Set(header.DBFieldName, nil)
					}
				} else if header.Column == "" {
					rowData.Set(header.DBFieldName, "")
				} else {
					stringData, err := c.Engine.GetCellValue(sheetName, header.Column+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}

					stringData = strings.TrimSpace(strings.ReplaceAll(stringData, "'", "''"))

					if len(stringData) > 300 {
						stringData = stringData[0:300]
					}

					rowData.Set(header.DBFieldName, stringData)
				}
			}

			if dontInsert {
				continue
			}

			param := helpers.InsertParam{
				TableName: "F_ENG_EQUIPMENT_MONTHLY",
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

		row++

		if err == nil {
			log.Println("SUCCESS Processing", rowCount, "rows\n")
		}
	}

	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *CttController) selectItemID(param SqlQueryParam) error {
	sqlQuery := "SELECT * FROM D_Item WHERE ITEM_NAME = '" + param.ItemName + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From("D_Item").SQL(sqlQuery), nil)
	defer cursor.Close()

	err := cursor.Fetchs(param.Results, 0)

	return err
}
