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
	clit.LoadConfigFromFlag("config", "hc", filepath.Join(clit.ExeDir(), "config", "hc.json"))
	clit.LoadConfigFromFlag("config", "petikemas", filepath.Join(clit.ExeDir(), "config", "petikemas.json"))
	clit.LoadConfigFromFlag("config", "ftw", filepath.Join(clit.ExeDir(), "config", "ftw.json"))
	clit.LoadConfigFromFlag("config", "marketshare", filepath.Join(clit.ExeDir(), "config", "marketshare.json"))
	clit.LoadConfigFromFlag("config", "induksi", filepath.Join(clit.ExeDir(), "config", "induksi.json"))
	clit.LoadConfigFromFlag("config", "kinerja", filepath.Join(clit.ExeDir(), "config", "kinerja.json"))
	clit.LoadConfigFromFlag("config", "pemenuhansdm", filepath.Join(clit.ExeDir(), "config", "pemenuhansdm.json"))
	clit.LoadConfigFromFlag("config", "rkap", filepath.Join(clit.ExeDir(), "config", "rkap.json"))
	clit.LoadConfigFromFlag("config", "lb", filepath.Join(clit.ExeDir(), "config", "lb.json"))
	clit.LoadConfigFromFlag("config", "realisasiAnggaran", filepath.Join(clit.ExeDir(), "config", "realisasiAnggaran.json"))
	clit.LoadConfigFromFlag("config", "rekapKonsol", filepath.Join(clit.ExeDir(), "config", "rekapKonsol.json"))
	clit.LoadConfigFromFlag("config", "rekapKonsol2", filepath.Join(clit.ExeDir(), "config", "rekapKonsol2.json"))
	clit.LoadConfigFromFlag("config", "rekapLegi", filepath.Join(clit.ExeDir(), "config", "rekapLegi.json"))
	clit.LoadConfigFromFlag("config", "rekapLegi2", filepath.Join(clit.ExeDir(), "config", "rekapLegi2.json"))
	clit.LoadConfigFromFlag("config", "rekapLegi3", filepath.Join(clit.ExeDir(), "config", "rekapLegi3.json"))
	clit.LoadConfigFromFlag("config", "rekapTTL", filepath.Join(clit.ExeDir(), "config", "rekapTTL.json"))
	clit.LoadConfigFromFlag("config", "rekapTTL2", filepath.Join(clit.ExeDir(), "config", "rekapTTL2.json"))
	clit.LoadConfigFromFlag("config", "rekapTTL3", filepath.Join(clit.ExeDir(), "config", "rekapTTL3.json"))
	clit.LoadConfigFromFlag("config", "RUPS", filepath.Join(clit.ExeDir(), "config", "RUPS.json"))
	clit.LoadConfigFromFlag("config", "kinerjaTerminal", filepath.Join(clit.ExeDir(), "config", "kinerjaTerminal.json"))

	firstTimer := clit.Config("default", "fetchApiFromFirstTime", false).(bool)

	if err := clit.Commit(); err != nil {
		log.Fatal("Error reading config file, ERROR:", err.Error())
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
				var err error

				(&c.Base{Controller: &c.AscController{}}).Extract()
				(&c.Base{Controller: &c.QhsseController{}}).Extract()
				(&c.Base{Controller: &c.CorsecController{}}).Extract()
				(&c.Base{Controller: &c.KeluhanController{}}).Extract()
				(&c.Base{Controller: &c.EtlController{}}).Extract()
				(&c.Base{Controller: &c.CttController{}}).Extract()
				(&c.Base{Controller: &c.ReadinessController{}}).Extract()
				(&c.Base{Controller: &c.InvestmentController{}}).Extract()
				(&c.Base{Controller: &c.EquipmentPerformance10STSController{}}).Extract()
				(&c.Base{Controller: &c.EquipmentPerformance5SCController{}}).Extract()
				(&c.Base{Controller: &c.RekapPetikemasController{}}).Extract()
				(&c.Base{Controller: &c.InduksiController{}}).Extract()
				(&c.Base{Controller: &c.KinerjaController{}}).Extract()
				(&c.Base{Controller: &c.FTWController{}}).Extract()

				(&c.Base{Controller: &c.RUPSController{}}).Extract()
				(&c.Base{Controller: &c.KinerjaTerminalController{}}).Extract()

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

				// Read Realisasi Files
				err = c.NewRealisasiController().ReadExcels()
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
