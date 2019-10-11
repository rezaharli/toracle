package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/eaciit/toolkit"

	"github.com/eaciit/clit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
)

var wg sync.WaitGroup

func main() {
	clit.LoadConfigFromFlag("", "", filepath.Join(clit.ExeDir(), "config", "app.json"))

	if err := clit.Commit(); err != nil {
		helpers.KillApp(err)
	}
	defer clit.Close()

	conn := helpers.Database()
	if conn != nil {
		// do the loop
		var ticker *time.Ticker = nil

		loopInterval := clit.Config("default", "interval", 1).(float64)

		if ticker == nil {
			ticker = time.NewTicker(time.Duration(int(loopInterval)) * time.Minute)
		}

		for {
			fmt.Println("Reading Files.")

			files := helpers.FetchFilePathsWithExt(".xlsx")

			for _, file := range files {
				err := helpers.ReadExcel(file)
				if err == nil {
					// move file if succeeded
					toolkit.Println("Moving file to archive...")
					archivePath := filepath.Join(clit.ExeDir(), "resource", "archive")
					if _, err := os.Stat(archivePath); os.IsNotExist(err) {
						os.Mkdir(archivePath, 0755)
					}

					err := os.Rename(file, filepath.Join(clit.ExeDir(), "resource", "archive", filepath.Base(file)))
					if err != nil {
						log.Fatal(err)
					}
				}
			}

			<-ticker.C
		}
		//loop ends

		wg.Add(1)
		// do normal task here
		wg.Wait()
	}
}
