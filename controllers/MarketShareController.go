package controllers

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/sebar/dbflex"
)

// MarketShareController is a controller for every kind of MarketShare files.
type MarketShareController struct {
	*Base
}

// New is used to initiate the controller
func (c *MarketShareController) New(base interface{}) {
	c.Base = base.(*Base)

	log.Println("Scanning for MarketShare files.")
	c.FileExtension = ".xlsx"
}

// FileCriteria is a callback function
// Used to filter file that is going to extract
func (c *MarketShareController) FileCriteria(file string) bool {
	return strings.Contains(filepath.Base(file), "Market Share Curah Kering TTL")
}

// ReadExcel fetch sheets of the excel and call ReadSheet for every sheet that match the condition
func (c *MarketShareController) ReadExcel() {
	c.ReadSheet(c.ReadData, "Market Share")
}

func (c *MarketShareController) ReadData(sheetName string) error {
	//timeNow := time.Now()

	log.Println("Deleting datas.")

	sql := "DELETE FROM F_CBD_MARKET_SHARE_CUKER"

	conn := helpers.Database()
	query, err := conn.Prepare(dbflex.From("F_CBD_MARKET_SHARE_CUKER").SQL(sql))
	if err != nil {
		log.Println(err)
	}

	_, err = query.Execute(toolkit.M{}.Set("data", toolkit.M{}))
	if err != nil {
		log.Println(err)
	}

	log.Println("Data deleted.")

	log.Println("ReadData", sheetName)
	//columnsMapping := clit.Config("marketshare", "columnsMapping", nil).(map[string]interface{})

	year, err := c.Engine.GetCellValue(sheetName, "B44")
	yearstr := strings.Trim(year, "Tahun ")
	yearint, _ := strconv.Atoi(yearstr)
	terminalname := []string{}

	for i := 46; i < 50; i++ {
		val, err := c.Engine.GetCellValue(sheetName, fmt.Sprintf("A%d", i))
		if err != nil {
			toolkit.Println(err)
		}
		terminalname = append(terminalname, val)
	}

	for j, terminal := range terminalname {
		terminalidx := 46 + j
		data := toolkit.M{}
		data.Set("TERMINAL", terminal)
		datedata := time.Date(yearint, time.January, 1, 0, 0, 0, 0, time.UTC)
		data.Set("PERIOD", datedata)

		for i := 46; i < 50; i++ {
			val, err := c.Engine.GetCellValue(sheetName, fmt.Sprintf("B%d", i))
			if err != nil {
				toolkit.Println(err)
			}

			if i == terminalidx {
				data.Set("TONASE", val)
			}
		}

		toolkit.Println(data)
		param := helpers.InsertParam{
			TableName: "F_CBD_MARKET_SHARE_CUKER",
			Data:      data,
		}

		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		} else {
			log.Println("Row inserted.")
		}
	}

	for j, terminal := range terminalname {
		terminalidx := 46 + j
		data := toolkit.M{}
		data.Set("TERMINAL", terminal)
		datedata := time.Date(yearint, time.February, 1, 0, 0, 0, 0, time.UTC)
		data.Set("PERIOD", datedata)

		for i := 46; i < 50; i++ {
			val, err := c.Engine.GetCellValue(sheetName, fmt.Sprintf("C%d", i))
			if err != nil {
				toolkit.Println(err)
			}

			if i == terminalidx {
				data.Set("TONASE", val)
			}
		}

		// toolkit.Println(data)
		// toolkit.Println(data)
		param := helpers.InsertParam{
			TableName: "F_CBD_MARKET_SHARE_CUKER",
			Data:      data,
		}

		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		} else {
			log.Println("Row inserted.")
		}
	}

	for j, terminal := range terminalname {
		terminalidx := 46 + j
		data := toolkit.M{}
		data.Set("TERMINAL", terminal)
		datedata := time.Date(yearint, time.March, 1, 0, 0, 0, 0, time.UTC)
		data.Set("PERIOD", datedata)

		for i := 46; i < 50; i++ {
			val, err := c.Engine.GetCellValue(sheetName, fmt.Sprintf("D%d", i))
			if err != nil {
				toolkit.Println(err)
			}

			if i == terminalidx {
				data.Set("TONASE", val)
			}
		}

		// toolkit.Println(data)
		param := helpers.InsertParam{
			TableName: "F_CBD_MARKET_SHARE_CUKER",
			Data:      data,
		}

		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		} else {
			log.Println("Row inserted.")
		}
	}

	for j, terminal := range terminalname {
		terminalidx := 46 + j
		data := toolkit.M{}
		data.Set("TERMINAL", terminal)
		datedata := time.Date(yearint, time.April, 1, 0, 0, 0, 0, time.UTC)
		data.Set("PERIOD", datedata)

		for i := 46; i < 50; i++ {
			val, err := c.Engine.GetCellValue(sheetName, fmt.Sprintf("E%d", i))
			if err != nil {
				toolkit.Println(err)
			}

			if i == terminalidx {
				data.Set("TONASE", val)
			}
		}

		// toolkit.Println(data)
		// toolkit.Println(data)
		param := helpers.InsertParam{
			TableName: "F_CBD_MARKET_SHARE_CUKER",
			Data:      data,
		}

		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		} else {
			log.Println("Row inserted.")
		}
	}

	for j, terminal := range terminalname {
		terminalidx := 46 + j
		data := toolkit.M{}
		data.Set("TERMINAL", terminal)
		datedata := time.Date(yearint, time.May, 1, 0, 0, 0, 0, time.UTC)
		data.Set("PERIOD", datedata)

		for i := 46; i < 50; i++ {
			val, err := c.Engine.GetCellValue(sheetName, fmt.Sprintf("F%d", i))
			if err != nil {
				toolkit.Println(err)
			}

			if i == terminalidx {
				data.Set("TONASE", val)
			}
		}

		// toolkit.Println(data)
		// toolkit.Println(data)
		param := helpers.InsertParam{
			TableName: "F_CBD_MARKET_SHARE_CUKER",
			Data:      data,
		}

		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		} else {
			log.Println("Row inserted.")
		}
	}

	for j, terminal := range terminalname {
		terminalidx := 46 + j
		data := toolkit.M{}
		data.Set("TERMINAL", terminal)
		datedata := time.Date(yearint, time.June, 1, 0, 0, 0, 0, time.UTC)
		data.Set("PERIOD", datedata)

		for i := 46; i < 50; i++ {
			val, err := c.Engine.GetCellValue(sheetName, fmt.Sprintf("G%d", i))
			if err != nil {
				toolkit.Println(err)
			}

			if i == terminalidx {
				data.Set("TONASE", val)
			}
		}

		// toolkit.Println(data)
		// toolkit.Println(data)
		param := helpers.InsertParam{
			TableName: "F_CBD_MARKET_SHARE_CUKER",
			Data:      data,
		}

		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		} else {
			log.Println("Row inserted.")
		}
	}

	for j, terminal := range terminalname {
		terminalidx := 46 + j
		data := toolkit.M{}
		data.Set("TERMINAL", terminal)
		datedata := time.Date(yearint, time.July, 1, 0, 0, 0, 0, time.UTC)
		data.Set("PERIOD", datedata)

		for i := 46; i < 50; i++ {
			val, err := c.Engine.GetCellValue(sheetName, fmt.Sprintf("H%d", i))
			if err != nil {
				toolkit.Println(err)
			}

			if i == terminalidx {
				data.Set("TONASE", val)
			}
		}

		// toolkit.Println(data)
		// toolkit.Println(data)
		param := helpers.InsertParam{
			TableName: "F_CBD_MARKET_SHARE_CUKER",
			Data:      data,
		}

		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		} else {
			log.Println("Row inserted.")
		}
	}

	for j, terminal := range terminalname {
		terminalidx := 46 + j
		data := toolkit.M{}
		data.Set("TERMINAL", terminal)
		datedata := time.Date(yearint, time.August, 1, 0, 0, 0, 0, time.UTC)
		data.Set("PERIOD", datedata)

		for i := 46; i < 50; i++ {
			val, err := c.Engine.GetCellValue(sheetName, fmt.Sprintf("I%d", i))
			if err != nil {
				toolkit.Println(err)
			}

			if i == terminalidx {
				data.Set("TONASE", val)
			}
		}

		// toolkit.Println(data)
		// toolkit.Println(data)
		param := helpers.InsertParam{
			TableName: "F_CBD_MARKET_SHARE_CUKER",
			Data:      data,
		}

		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		} else {
			log.Println("Row inserted.")
		}
	}

	for j, terminal := range terminalname {
		terminalidx := 46 + j
		data := toolkit.M{}
		data.Set("TERMINAL", terminal)
		datedata := time.Date(yearint, time.September, 1, 0, 0, 0, 0, time.UTC)
		data.Set("PERIOD", datedata)

		for i := 46; i < 50; i++ {
			val, err := c.Engine.GetCellValue(sheetName, fmt.Sprintf("J%d", i))
			if err != nil {
				toolkit.Println(err)
			}

			if i == terminalidx {
				data.Set("TONASE", val)
			}
		}

		// toolkit.Println(data)
		// toolkit.Println(data)
		param := helpers.InsertParam{
			TableName: "F_CBD_MARKET_SHARE_CUKER",
			Data:      data,
		}

		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		} else {
			log.Println("Row inserted.")
		}
	}

	for j, terminal := range terminalname {
		terminalidx := 46 + j
		data := toolkit.M{}
		data.Set("TERMINAL", terminal)
		datedata := time.Date(yearint, time.October, 1, 0, 0, 0, 0, time.UTC)
		data.Set("PERIOD", datedata)

		for i := 46; i < 50; i++ {
			val, err := c.Engine.GetCellValue(sheetName, fmt.Sprintf("K%d", i))
			if err != nil {
				toolkit.Println(err)
			}

			if i == terminalidx {
				data.Set("TONASE", val)
			}
		}

		// toolkit.Println(data)
		// toolkit.Println(data)
		param := helpers.InsertParam{
			TableName: "F_CBD_MARKET_SHARE_CUKER",
			Data:      data,
		}

		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		} else {
			log.Println("Row inserted.")
		}
	}

	for j, terminal := range terminalname {
		terminalidx := 46 + j
		data := toolkit.M{}
		data.Set("TERMINAL", terminal)
		datedata := time.Date(yearint, time.November, 1, 0, 0, 0, 0, time.UTC)
		data.Set("PERIOD", datedata)

		for i := 46; i < 50; i++ {
			val, err := c.Engine.GetCellValue(sheetName, fmt.Sprintf("L%d", i))
			if err != nil {
				toolkit.Println(err)
			}

			if i == terminalidx {
				data.Set("TONASE", val)
			}
		}

		// toolkit.Println(data)
		// toolkit.Println(data)
		param := helpers.InsertParam{
			TableName: "F_CBD_MARKET_SHARE_CUKER",
			Data:      data,
		}

		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		} else {
			log.Println("Row inserted.")
		}
	}

	for j, terminal := range terminalname {
		terminalidx := 46 + j
		data := toolkit.M{}
		data.Set("TERMINAL", terminal)
		datedata := time.Date(yearint, time.December, 1, 0, 0, 0, 0, time.UTC)
		data.Set("PERIOD", datedata)

		for i := 46; i < 50; i++ {
			val, err := c.Engine.GetCellValue(sheetName, fmt.Sprintf("M%d", i))
			if err != nil {
				toolkit.Println(err)
			}

			if i == terminalidx {
				data.Set("TONASE", val)
			}
		}

		// toolkit.Println(data)
		// toolkit.Println(data)
		param := helpers.InsertParam{
			TableName: "F_CBD_MARKET_SHARE_CUKER",
			Data:      data,
		}

		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		} else {
			log.Println("Row inserted.")
		}
	}

	return err
}
