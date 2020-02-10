package controllers

import (
	"log"
	"path/filepath"
	"time"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/rezaharli/toracle/interfaces"
	"git.eaciitapp.com/sebar/dbflex"
	"github.com/eaciit/clit"
)

type Base struct {
	interfaces.ExcelController
}

func (c *Base) Extract() {
	c.New()

	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	filePaths := c.FetchFiles(resourcePath)

	filenames := []string{}
	for _, file := range filePaths {
		if c.FileCriteria(file) {
			filenames = append(filenames, file)
		}
	}

	log.Println("Scanning finished. files found:", len(filenames))

	for _, file := range filenames {
		log.Println("Processing sheets...")
		timeNow := time.Now()

		f, err := helpers.ReadExcel(file)
		if err != nil {
			log.Fatal(err.Error())
		}

		err = c.ReadExcel(f)
		if err != nil {
			log.Fatal(err.Error())
		}

		if err == nil {
			log.Println("\nSUCCESS")
		}

		log.Println("Total Process Time:", time.Since(timeNow).Seconds(), "seconds")

		// move file if read succeeded
		helpers.MoveToArchive(file)
		log.Println("Done.")
	}
}

func (c *Base) SelectItemID(param SqlQueryParam) error {
	sqlQuery := "SELECT * FROM D_Item WHERE ITEM_NAME = TRIM('" + param.ItemName + "')"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From("D_Item").SQL(sqlQuery), nil)
	defer cursor.Close()

	err := cursor.Fetchs(param.Results, 0)

	return err
}

//------------------------------------------------------------------------------------------------------------------------

type SqlQueryParam struct {
	ItemName string
	Results  interface{}
}

type Header struct {
	DBFieldName string
	HeaderName  string

	Column       string
	ColumnNumber int
	Row          string

	Value string
}
