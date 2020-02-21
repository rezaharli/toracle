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

// KinerjaController is a controller for every kind of Kinerja files.
type KinerjaController struct {
	*Base
}

// New is used to initiate the controller
func (c *KinerjaController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for Kinerja files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *KinerjaController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "formulirperhitungan kinerja MK3L 2019 with TKBM")
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *KinerjaController) ReadExcel() {
	for _, sheetName := range c.Engine.GetSheetMap() {
		if c.Engine.IsSheetVisible(sheetName) && sheetName != "ytd" && sheetName != "to disnaker" {
			c.ReadSheet(c.ReadData, sheetName)
		}
	}
}

func (c *KinerjaController) ReadData(sheetName string) error {
	timeNow := time.Now()

	toolkit.Println()
	log.Println("ReadData", sheetName)
	columnsMapping := clit.Config("kinerja", "columnsMapping", nil).(map[string]interface{})
	// months := clit.Config("kinerja", "months", []interface{}{}).([]interface{})

	firstDataRow := 0
	i := 1
	for {
		cellValue, err := c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(i))
		if err != nil {
			log.Fatal(err)
		}

		if cellValue == "No" {
			cellValue, err = c.Engine.GetCellValue(sheetName, "B"+toolkit.ToString(i+1))
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
	emptyRowCount := 0
	var currentPeriod time.Time
	months := clit.Config("kinerja", "months", nil).([]interface{})

	isPeriodSkip := false

	//iterate over rows
Rowloop:
	for index := 0; true; index++ {
		rowData := toolkit.M{}
		currentRow := firstDataRow + index

		isRowEmpty := true
		skipRow := false
		for _, header := range headers {
			if header.DBFieldName == "PERIOD" {
				cellID := "B5"
				// style, _ := f.NewStyle(`{"number_format":15}`)
				// f.SetCellStyle(sheetName, cellID, cellID, style)
				stringData, err := c.Engine.GetCellValue(sheetName, cellID)
				if err != nil {
					log.Fatal(err)
				}

				splitted := strings.Split(stringData, " ")
				month := ""
				year := ""
				if len(splitted) >= 2 {
					month = splitted[0]
					year = splitted[1]
				} else {
					splitted = strings.Split(stringData, "-")

					if len(splitted) >= 2 {
						month = splitted[0]
						year = splitted[1]
					} else {
						month = stringData[0 : len(stringData)-4]
						year = stringData[len(stringData)-4:]
					}
				}

				stringTanggal := "1-" + toolkit.ToString(helpers.IndexOf(month, months)+1) + "-" + year

				var t time.Time
				if strings.TrimSpace(stringData) != "" {
					t, err = time.Parse("2-1-2006", stringTanggal)
					if err != nil {
						t, err = time.Parse("2-1-06", stringTanggal)
						if err != nil {
							t, err = time.Parse("2/1/2006", stringTanggal)
							if err != nil {
								log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
							}
						}
					}

					currentPeriod = t
				}

				rowData.Set(header.DBFieldName, currentPeriod)
			} else {
				col := header.Column

				stringData, err := c.Engine.GetCellValue(sheetName, col+toolkit.ToString(currentRow))
				if err != nil {
					log.Fatal(err)
				}

				if strings.TrimSpace(stringData) != "" {
					colNum := helpers.CharStrToNum(col)
					newCol := helpers.ToCharStr(colNum)
					stringData, err = c.Engine.GetCellValue(sheetName, newCol+toolkit.ToString(currentRow))
					if err != nil {
						log.Fatal(err)
					}
				}

				if strings.Contains(strings.ToUpper(stringData), strings.ToUpper("Mitra Kerja")) || (header.DBFieldName == "COMPANY_NAME" && stringData == "") {
					skipRow = true
				}

				stringData = strings.ReplaceAll(stringData, "'", "''")

				if stringData != "" {
					isRowEmpty = false
				}

				if header.DBFieldName == "COMPANY_NAME" && (strings.ToUpper(strings.TrimSpace(stringData)) == strings.ToUpper(strings.TrimSpace("Total"))) {
					break Rowloop
				}

				rowData.Set(header.DBFieldName, stringData)
			}
		}

		if isRowEmpty {
			emptyRowCount++

			if emptyRowCount >= 3 {
				break
			}
		}

		if skipRow {
			continue
		}

		if rowCount == 0 {
			// check if data exists
			sqlQuery := "SELECT PERIOD FROM F_QHSSE_MK3L WHERE trunc(period) = TO_DATE('" + currentPeriod.Format("2006-01-02") + "', 'YYYY-MM-DD')"

			conn := helpers.Database()
			cursor := conn.Cursor(dbflex.From("F_QHSSE_MK3L").SQL(sqlQuery), nil)
			defer cursor.Close()

			res := make([]toolkit.M, 0)
			err = cursor.Fetchs(&res, 0)
			if err != nil {
				log.Println(err)
			}

			//only insert if len of datas in currentPeriod is 0 / if no data yet
			if len(res) != 0 {
				isPeriodSkip = true
				log.Println("Skipping", currentPeriod.Format("2006-01-02")+".")
			}
		}

		if isPeriodSkip == false {
			param := helpers.InsertParam{
				TableName: "F_QHSSE_MK3L",
				Data:      rowData,
			}

			err = helpers.Insert(param)
			if err != nil {
				log.Fatal("Error inserting row "+toolkit.ToString(currentRow)+", ERROR:", err.Error())
			} else {
				log.Println("Row", currentRow, "inserted.")
			}
		}

		rowCount++
	}

	if err == nil {
		log.Println("SUCCESS Processing", rowCount, "rows")
	}
	log.Println("Process time:", time.Since(timeNow).Seconds(), "seconds")
	return err
}

func (c *KinerjaController) selectItemID(param SqlQueryParam) error {
	sqlQuery := "SELECT * FROM D_Item WHERE ITEM_NAME = '" + param.ItemName + "'"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From("D_Item").SQL(sqlQuery), nil)
	defer cursor.Close()

	err := cursor.Fetchs(param.Results, 0)

	return err
}
