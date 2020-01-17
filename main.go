package main

import (
	"log"
	"path/filepath"
	"time"

	c "git.eaciitapp.com/rezaharli/toracle/controllers"
	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"
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
	clit.LoadConfigFromFlag("config", "hc", filepath.Join(clit.ExeDir(), "config", "hc.json"))
	clit.LoadConfigFromFlag("config", "petikemas", filepath.Join(clit.ExeDir(), "config", "petikemas.json"))
	clit.LoadConfigFromFlag("config", "ftw", filepath.Join(clit.ExeDir(), "config", "ftw.json"))
	clit.LoadConfigFromFlag("config", "marketshare", filepath.Join(clit.ExeDir(), "config", "marketshare.json"))
	clit.LoadConfigFromFlag("config", "induksi", filepath.Join(clit.ExeDir(), "config", "induksi.json"))
	clit.LoadConfigFromFlag("config", "kinerja", filepath.Join(clit.ExeDir(), "config", "kinerja.json"))
	clit.LoadConfigFromFlag("config", "pemenuhansdm", filepath.Join(clit.ExeDir(), "config", "pemenuhansdm.json"))
	clit.LoadConfigFromFlag("config", "rkap", filepath.Join(clit.ExeDir(), "config", "rkap.json"))
	clit.LoadConfigFromFlag("config", "lb", filepath.Join(clit.ExeDir(), "config", "lb.json"))
	clit.LoadConfigFromFlag("config", "rekapKonsol", filepath.Join(clit.ExeDir(), "config", "rekapKonsol.json"))
	clit.LoadConfigFromFlag("config", "rekapKonsol2", filepath.Join(clit.ExeDir(), "config", "rekapKonsol2.json"))
	clit.LoadConfigFromFlag("config", "rekapLegi2", filepath.Join(clit.ExeDir(), "config", "rekapLegi2.json"))
	clit.LoadConfigFromFlag("config", "rekapLegi3", filepath.Join(clit.ExeDir(), "config", "rekapLegi3.json"))

	firstTimer := clit.Config("default", "fetchApiFromFirstTime", false).(bool)

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
		isExecute := true

		// do the loop
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

				// READ Petikemas FILES
				err = c.NewRekapPetikemasController().ReadExcels()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ Induksi FILES
				err = c.NewInduksiController().ReadExcels()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ Kinerja FILES
				err = c.NewKinerjaController().ReadExcels()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ FTW FILES
				err = c.NewFTWController().ReadExcels()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ Pemenuhan SDM FILES
				err = c.NewPemenuhanSDMController().ReadExcels()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ RKAP FILES
				err = c.NewRKAPController().ReadExcels()
				if err != nil {
					log.Fatal(err.Error())
				}

				// Read Market Share Files
				err = c.NewMarketShareController().ReadExcels()
				if err != nil {
					log.Fatal(err.Error())
				}

				// Read Pencapaian Files
				err = c.NewPencapaianController().ReadExcels()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ Proc API DAILY
				if i%totalRunInADay == 0 {
					procController := c.NewProcController()
					procController.FirstTimer = firstTimer
					// err = procController.ReadAPI()
					if err != nil {
						log.Fatal(err.Error())
					}

					firstTimer = false
				}

				// READ Hc API
				hcController := c.NewHcController()
				err = hcController.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ Hc API Summary 201
				hcsumController := c.NewHcSummaryController()
				err = hcsumController.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ Hc API Summary 301A
				hcsum301AController := c.NewHcSummary301AController()
				err = hcsum301AController.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				// READ Hc API Summary 301B
				hcsum301BController := c.NewHcSummary301BController()
				err = hcsum301BController.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				genderController := c.NewGenderController()
				err = genderController.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				educationController := c.NewEducationController()
				err = educationController.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				attendanceController := c.NewAttendanceController()
				err = attendanceController.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				statusController := c.NewStatusController()
				err = statusController.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				productivityController := c.NewProductivityController()
				err = productivityController.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				lb1Controller := c.NewLB1Controller()
				err = lb1Controller.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				lb2Controller := c.NewLB2Controller()
				err = lb2Controller.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				lb4Controller := c.NewLB4Controller()
				err = lb4Controller.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				if isExecute {
					isExecute = false
					lb13Controller := c.NewLB13Controller()
					err = lb13Controller.ReadAPI()
					if err != nil {
						log.Fatal(err.Error())
					}

				}

				lb5Controller := c.NewLB5Controller()
				err = lb5Controller.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				lb10Controller := c.NewLB10Controller()
				err = lb10Controller.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
				}

				lb11Controller := c.NewLB11Controller()
				err = lb11Controller.ReadAPI()
				if err != nil {
					log.Fatal(err.Error())
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
