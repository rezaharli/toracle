package controllers

import (
	"fmt"
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

type MarketShareController struct {
	*Base
}

func NewMarketShareController() *MarketShareController {
	return new(MarketShareController)
}

func (c *MarketShareController) ReadExcels() error {
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

func (c *MarketShareController) FetchFiles() []string {
	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	files := helpers.FetchFilePathsWithExt(resourcePath, ".xlsx")

	resourceFiles := []string{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), "~") {
			continue
		}

		if strings.Contains(filepath.Base(file), "Market Share Curah Kering TTL") {
			resourceFiles = append(resourceFiles, file)
		}
	}

	log.Println("Scanning finished. Market Share Curah Kering TTL files found:", len(resourceFiles))
	return resourceFiles
}

func (c *MarketShareController) readExcel(filename string) error {
	timeNow := time.Now()

	f, err := helpers.ReadExcel(filename)

	log.Println("Processing sheets...")
	// for _, sheetName := range f.GetSheetMap() {
	// 	err = c.ReadData(f, sheetName)
	// 	if err != nil {
	// 		log.Println("Error reading data. ERROR:", err)
	// 	}
	// }

	err = c.ReadData(f, "Market Share")
	if err != nil {
		log.Println("Error reading data. ERROR:", err)
	}

	if err == nil {
		toolkit.Println()
		log.Println("SUCCESS")
	}
	log.Println("Total Process Time:", time.Since(timeNow).Seconds(), "seconds")

	return err
}

func (c *MarketShareController) ReadData(f *excelize.File, sheetName string) error {
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

	year, err := f.GetCellValue(sheetName, "B44")
	yearstr := strings.Trim(year, "Tahun ")
	yearint, _ := strconv.Atoi(yearstr)
	terminalname := []string{}

	for i := 46; i < 50; i++ {
		val, err := f.GetCellValue(sheetName, fmt.Sprintf("A%d", i))
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
			val, err := f.GetCellValue(sheetName, fmt.Sprintf("B%d", i))
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
			log.Fatal("Error inserting row, ERROR:", err.Error())
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
			val, err := f.GetCellValue(sheetName, fmt.Sprintf("C%d", i))
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
			log.Fatal("Error inserting row, ERROR:", err.Error())
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
			val, err := f.GetCellValue(sheetName, fmt.Sprintf("D%d", i))
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
			log.Fatal("Error inserting row, ERROR:", err.Error())
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
			val, err := f.GetCellValue(sheetName, fmt.Sprintf("E%d", i))
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
			log.Fatal("Error inserting row, ERROR:", err.Error())
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
			val, err := f.GetCellValue(sheetName, fmt.Sprintf("F%d", i))
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
			log.Fatal("Error inserting row, ERROR:", err.Error())
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
			val, err := f.GetCellValue(sheetName, fmt.Sprintf("G%d", i))
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
			log.Fatal("Error inserting row, ERROR:", err.Error())
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
			val, err := f.GetCellValue(sheetName, fmt.Sprintf("H%d", i))
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
			log.Fatal("Error inserting row, ERROR:", err.Error())
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
			val, err := f.GetCellValue(sheetName, fmt.Sprintf("I%d", i))
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
			log.Fatal("Error inserting row, ERROR:", err.Error())
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
			val, err := f.GetCellValue(sheetName, fmt.Sprintf("J%d", i))
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
			log.Fatal("Error inserting row, ERROR:", err.Error())
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
			val, err := f.GetCellValue(sheetName, fmt.Sprintf("K%d", i))
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
			log.Fatal("Error inserting row, ERROR:", err.Error())
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
			val, err := f.GetCellValue(sheetName, fmt.Sprintf("L%d", i))
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
			log.Fatal("Error inserting row, ERROR:", err.Error())
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
			val, err := f.GetCellValue(sheetName, fmt.Sprintf("M%d", i))
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
			log.Fatal("Error inserting row, ERROR:", err.Error())
		} else {
			log.Println("Row inserted.")
		}
	}

	return err
}
