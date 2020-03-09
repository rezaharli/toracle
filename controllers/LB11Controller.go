package controllers

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/rezaharli/toracle/models"
	"git.eaciitapp.com/sebar/dbflex"
)

type LB11Controller struct {
	*Base
}

func NewLB11Controller() *LB11Controller {
	return new(LB11Controller)
}

func (c *LB11Controller) ReadAPI() error {
	log.Println("\n--------------------------------------\nReading LB11 API")
	var err error

	lastmonth, lastyear := 0, 0
	paramstartyear := clit.Config("lb", "year", nil).(string)
	startyear, err := strconv.Atoi(paramstartyear)
	thisMonth := int(time.Now().Month())
	thisYear := time.Now().Year()
	monthcounter := 1
	latest, err := c.GetLatestData()

	if len(latest) == 0 {
		toolkit.Println("FETCHING DATA FROM START PERIOD")
		for {
			if thisMonth+1 == monthcounter && thisYear == startyear {
				break
			}
			if monthcounter <= 12 {
				toolkit.Println(startyear, monthcounter)
				monthparam := strconv.Itoa(monthcounter)
				yearparam := strconv.Itoa(startyear)
				err = c.ExecProcess(yearparam, monthparam)

				if monthcounter == 12 {
					if startyear < thisYear {
						startyear++
					}
					monthcounter = 1
				} else {
					monthcounter++
				}
			}
		}
	} else if len(latest) > 3 {
		toolkit.Println("DELETING THIS PERIOD")
		for i := 0; i < 3; i++ {
			toolkit.Println(latest[i].GetInt("TAHUN"), latest[i].GetInt("BULAN"))
			lastmonth = latest[i].GetInt("BULAN")
			lastyear = latest[i].GetInt("TAHUN")
			err = c.DeleteSelectedData(lastyear, lastmonth)
		}
		toolkit.Println("FETCHING THIS PERIOD")
		for {
			if thisMonth+1 == lastmonth && thisYear == lastyear {
				break
			}
			if lastmonth <= 12 {
				toolkit.Println(lastyear, lastmonth)
				monthparam := strconv.Itoa(monthcounter)
				yearparam := strconv.Itoa(startyear)
				err = c.ExecProcess(yearparam, monthparam)

				if lastmonth == 12 {
					if lastyear < thisYear {
						lastyear++
					}
					lastmonth = 1
				} else {
					lastmonth++
				}
			}
		}
	}

	return err
}

func (c *LB11Controller) GetAPIDatas(payload []byte, month string, year string) ([]toolkit.M, error) {
	log.Println("Get LB11")
	config := clit.Config("lb", "lb11", map[string]interface{}{}).(map[string]interface{})

	// payload := c.SetParamBody("2019", "12")
	username := clit.Config("lb", "username", nil).(string)
	password := clit.Config("lb", "password", nil).(string)

	request, err := http.NewRequest("POST", config["url"].(string), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/xml")
	request.SetBasicAuth(username, password)

	client := http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := &models.LB11Response{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.T_DATA.Item {
		result, _ := toolkit.ToM(value)
		result.Set("TAHUN", year)
		result.Set("BULAN", month)
		results = append(results, result)
	}

	return results, err
}

func (c *LB11Controller) InsertAPIDatas(results []toolkit.M, jsonconf string) error {
	log.Println("inserting data....")
	var err error

	config := clit.Config("lb", jsonconf, nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	var headers []Header
	for dbFieldName, attributeName := range columnsMapping {
		header := Header{
			DBFieldName: dbFieldName,
			Column:      attributeName.(string),
		}

		headers = append(headers, header)
	}

	for _, result := range results {
		rowData := toolkit.M{}
		for _, header := range headers {
			rowData.Set(header.DBFieldName, result[header.Column])
			rowData.Set("TAHUN", result["TAHUN"])
			rowData.Set("BULAN", result["BULAN"])
		}

		toolkit.Println(rowData)
		param := helpers.InsertParam{
			TableName: "F_FA_IKHTISAR_BIAYA_PERJENIS",
			Data:      rowData,
		}

		log.Println("Inserting data API")
		err := helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		}
	}

	return err
}

func (c *LB11Controller) SetParamBody(year string, month string) []byte {
	payload := []byte(strings.TrimSpace(`
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
	<soapenv:Header/>
	<soapenv:Body>
		<urn:ZFM_FI_21>
				<P_BUKRS xmlns="">PTTL</P_BUKRS>
				<P_GJAHR xmlns="">` + year + `</P_GJAHR>
				<P_MONAT xmlns="">` + month + `</P_MONAT>
				<P_PDF xmlns=""></P_PDF>
				<P_RLDNR xmlns=""></P_RLDNR>
				<P_VERSI xmlns=""></P_VERSI>
				<T_DATA xmlns="">
					<item></item>
				</T_DATA>
		</urn:ZFM_FI_21>
	</soapenv:Body>
	</soapenv:Envelope>`,
	))

	return payload
}

func (c *LB11Controller) GetLatestData() ([]toolkit.M, error) {
	sqlQuery := "SELECT DISTINCT TAHUN,BULAN,TAHUN||''||BULAN AS TAHUNBULAN FROM F_FA_IKHTISAR_BIAYA_PERJENIS ORDER BY TAHUN DESC,BULAN DESC"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From("D_Item").SQL(sqlQuery), nil)
	defer cursor.Close()

	res := []toolkit.M{}
	err := cursor.Fetchs(&res, 0)

	return res, err
}

func (c *LB11Controller) DeleteSelectedData(year int, month int) error {
	conn := helpers.Database()
	var res interface{}
	res, err := conn.Execute(
		dbflex.
			From("F_FA_IKHTISAR_BIAYA_PERJENIS").
			Where(dbflex.And(dbflex.Eq("BULAN", month), dbflex.Eq("TAHUN", year))).
			Delete(), nil)
	toolkit.Println(res)
	return err
}

func (c *LB11Controller) ExecProcess(year string, month string) error {
	payload := c.SetParamBody(year, month)
	results, err := c.GetAPIDatas(payload, month, year)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = c.InsertAPIDatas(results, "lb11")
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return err
}
