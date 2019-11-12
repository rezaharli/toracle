package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	c "git.eaciitapp.com/rezaharli/toracle/controllers"
	"git.eaciitapp.com/rezaharli/toracle/helpers"
)

func main() {
	clit.LoadConfigFromFlag("config", "default", filepath.Join(clit.ExeDir(), "config", "app.json"))
	clit.LoadConfigFromFlag("config", "asc", filepath.Join(clit.ExeDir(), "config", "asc.json"))
	clit.LoadConfigFromFlag("config", "qhsse", filepath.Join(clit.ExeDir(), "config", "qhsse.json"))
	clit.LoadConfigFromFlag("config", "corsec", filepath.Join(clit.ExeDir(), "config", "corsec.json"))
	clit.LoadConfigFromFlag("config", "keluhan", filepath.Join(clit.ExeDir(), "config", "keluhan.json"))
	clit.LoadConfigFromFlag("config", "etl", filepath.Join(clit.ExeDir(), "config", "etl.json"))
	clit.LoadConfigFromFlag("config", "ctt", filepath.Join(clit.ExeDir(), "config", "ctt.json"))
	clit.LoadConfigFromFlag("config", "readiness", filepath.Join(clit.ExeDir(), "config", "readiness.json"))
	clit.LoadConfigFromFlag("config", "investment", filepath.Join(clit.ExeDir(), "config", "investment.json"))
	clit.LoadConfigFromFlag("config", "proc", filepath.Join(clit.ExeDir(), "config", "proc.json"))
	clit.LoadConfigFromFlag("config", "equipmentPerformance", filepath.Join(clit.ExeDir(), "config", "equipmentPerformance.json"))

	if err := clit.Commit(); err != nil {
		toolkit.Println("Error reading config file, ERROR:", err.Error())
	}

	defer clit.Close()

	conn := helpers.Database()
	if conn != nil {
		var ticker *time.Ticker = nil

		var totalRunInADay int
		if ticker == nil {
			loopInterval := clit.Config("default", "interval", 1).(float64)
			durationInterval := time.Duration(int(loopInterval)) * time.Minute

			dailyInterval := time.Duration(24) * time.Hour
			totalRunInADay = int(dailyInterval.Hours() / durationInterval.Hours())

			ticker = time.NewTicker(durationInterval)
		}

		// do the loop
		firstTimer := true
		i := 0
		for {
			go func() {
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

				// READ CTT FILES
				err = c.NewCttController().ReadExcels()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ Readiness FILES
				err = c.NewReadinessController().ReadExcels()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ Investment FILES
				err = c.NewInvestmentController().ReadExcels()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ Equipment FILES
				err = c.NewEquipmentPerformance10STSController().ReadExcels()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ Equipment FILES
				err = c.NewEquipmentPerformance5SCController().ReadExcels()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ Proc API DAILY
				if i%totalRunInADay == 0 {
					procController := c.NewProcController()
					procController.FirstTimer = firstTimer
					err = procController.ReadAPI()
					if err != nil {
						log.Fatal(err.Error())
					}

					firstTimer = false
				}

				i++
				toolkit.Println()
				log.Println("Waiting...")
				toolkit.Println()
			}()

			<-ticker.C
		}
		//loop ends
	}
}
