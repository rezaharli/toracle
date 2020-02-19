package controllers

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/rezaharli/toracle/interfaces"
	"git.eaciitapp.com/sebar/dbflex"
)

// Base is a base controller for every other controller.
type Base struct {
	interfaces.ExcelController

	FileExtension string
}

func (c *Base) Decide() interfaces.XlsxController {
	switch c.FileExtension {
	case ".xlsx":
		return helpers.XlsxHelper{}
	default:
		return nil
	}
}

func (c *Base) Extract() {
	c.New(c)
	engine := c.Decide()

	resourcePath := clit.Config("default", "resourcePath", filepath.Join(clit.ExeDir(), "resource")).(string)
	filePaths := helpers.FetchFilePathsWithExt(resourcePath, c.FileExtension)

	filenames := []string{}
	for _, file := range filePaths {
		if c.FileCriteria(file) {
			filenames = append(filenames, file)
		}
	}

	log.Println("Scanning finished. files found:", len(filenames))

	for _, filePath := range filenames {
		log.Println("Processing sheets...")
		timeNow := time.Now()

		f, err := engine.ReadExcel(filePath)
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
		c.MoveToArchive(filePath)
		log.Println("Done.")
	}
}

func (c *Base) ReadSheet(f *excelize.File, sheetToRead string, readSheet readSheet) {
	err := readSheet(f, sheetToRead)
	if err != nil {
		log.Println("Error reading monthly data. ERROR:", err)
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

func (c *Base) InsertRowData(rowIdentifier interface{}, rowData interface{}, tableName string) {
	param := helpers.InsertParam{
		TableName: tableName,
		Data:      rowData,
	}

	err := helpers.Insert(param)
	if err != nil {
		log.Fatal("Error inserting row "+toolkit.ToString(rowIdentifier)+", ERROR:", err.Error())
	} else {
		log.Println("Row", rowIdentifier, "inserted.")
	}
}

func (c *Base) MoveToArchive(filePath string) {
	log.Println("Moving file to archive...")

	archivePath := filepath.Join(filepath.Dir(filePath), "archive")
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		os.Mkdir(archivePath, 0755)
	}

	err := os.Rename(filePath, filepath.Join(archivePath, filepath.Base(filePath)))
	if err != nil {
		log.Fatal(err)
	}
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

type readSheet func(f *excelize.File, sheetName string) error
