package main

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	c "git.eaciitapp.com/rezaharli/toracle/controllers"
	"git.eaciitapp.com/rezaharli/toracle/helpers"
)

var wg sync.WaitGroup

func main() {
	clit.LoadConfigFromFlag("config", "default", filepath.Join(clit.ExeDir(), "config", "app.json"))
	clit.LoadConfigFromFlag("config", "asc", filepath.Join(clit.ExeDir(), "config", "asc.json"))
	clit.LoadConfigFromFlag("config", "qhsse", filepath.Join(clit.ExeDir(), "config", "qhsse.json"))
	clit.LoadConfigFromFlag("config", "corsec", filepath.Join(clit.ExeDir(), "config", "corsec.json"))
	clit.LoadConfigFromFlag("config", "keluhan", filepath.Join(clit.ExeDir(), "config", "keluhan.json"))
	clit.LoadConfigFromFlag("config", "etl", filepath.Join(clit.ExeDir(), "config", "etl.json"))

	if err := clit.Commit(); err != nil {
		toolkit.Println("Error reading config file, ERROR:", err.Error())
	}

	defer clit.Close()

	conn := helpers.Database()
	if conn != nil {
		var ticker *time.Ticker = nil

		if ticker == nil {
			loopInterval := clit.Config("default", "interval", 1).(float64)
			ticker = time.NewTicker(time.Duration(int(loopInterval)) * time.Minute)
		}

		// do the loop
		for {
			// READ ASC FILES
			err := c.NewAscController().ReadExcels()
			if err != nil {
				log.Fatal(err.Error())
			}

			// READ QHSSE FILES
			err = c.NewQhsseController().ReadExcels()
			if err != nil {
				log.Fatal(err.Error())
			}

			// READ CORSEC FILES
			err = c.NewCorsecController().ReadExcels()
			if err != nil {
				log.Fatal(err.Error())
			}

			// READ Keluhan FILES
			err = c.NewKeluhanController().ReadExcels()
			if err != nil {
				log.Fatal(err.Error())
			}

			// READ ETL FILES
			err = c.NewEtlController().ReadExcels()
			if err != nil {
				log.Fatal(err.Error())
			}

			<-ticker.C
		}
		//loop ends

		wg.Add(1)
		// do normal task here
		wg.Wait()
	}
}
