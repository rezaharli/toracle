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
)

type HCAbsencesController struct {
	*Base
}

func NewHCAbsencesController() *HCAbsencesController {
	return new(HCAbsencesController)
}

func (c *HCAbsencesController) SetParamBody(startdate string, enddate string) []byte {
	payload := []byte(strings.TrimSpace(`
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
	<soapenv:Header/>
	<soapenv:Body>
		<urn:ZFM_HC_400>
			<DATE_FROM xmlns="">` + startdate + `</DATE_FROM>
			<DATE_TO xmlns="">` + enddate + `</DATE_TO>
			<T_DATA xmlns="">
				<item></item>
			</T_DATA>
		</urn:ZFM_HC_400>
	</soapenv:Body>
	</soapenv:Envelope>`,
	))

	return payload
}

func (c *HCAbsencesController) GetAPIDatas(payload []byte) ([]toolkit.M, error) {
	log.Println("Get Absences Data")
	config := clit.Config("hc", "absence", map[string]interface{}{}).(map[string]interface{})

	// payload := c.SetParamBody("2019", "12")
	username := clit.Config("hc", "username", nil).(string)
	password := clit.Config("hc", "password", nil).(string)

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

	r := &models.HCAbsence{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.T_DATA.Item {
		result, _ := toolkit.ToM(value)
		results = append(results, result)
	}

	return results, err
}

func (c *HCAbsencesController) InsertAPIDatas(results []toolkit.M) error {
	log.Println("inserting data....")
	var err error

	config := clit.Config("hc", "absence", nil).(map[string]interface{})
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
			if header.DBFieldName == "EVENT_DATE" {
				dateString := result[header.Column].(string)
				t, err := time.Parse("2006-01-02", dateString)
				if err != nil {
					helpers.HandleError(err)
				}

				rowData.Set(header.DBFieldName, t)
			} else {
				rowData.Set(header.DBFieldName, result[header.Column])
			}
		}
		rowData.Set("UPDATE_DATE", time.Now())
		// toolkit.Println(rowData)
		if rowData.GetString("PERSONNEL_NUMBER") != "00000000" {
			param := helpers.InsertParam{
				TableName: "HC_ABSENCE",
				Data:      rowData,
			}

			log.Println("Inserting data API")
			err := helpers.Insert(param)
			if err != nil {
				helpers.HandleError(err)
			}
		}

	}

	return err
}

func (c *HCAbsencesController) ExecProcess(payload []byte) error {
	results, err := c.GetAPIDatas(payload)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = c.InsertAPIDatas(results)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return err
}

func (c *HCAbsencesController) CreateParamToday() string {
	thisYear := time.Now().Year()
	thisMonth := int(time.Now().Month())
	thisDay := time.Now().Day() - 1

	stryear := strconv.Itoa(thisYear)
	strmonth := ""
	strday := ""
	if thisMonth < 10 {
		strmonth = "0" + strconv.Itoa(thisMonth)
	} else {
		strmonth = strconv.Itoa(thisMonth)
	}
	if thisDay < 10 {
		strday = "0" + strconv.Itoa(thisDay)
	} else {
		strday = strconv.Itoa(thisDay)
	}

	return stryear + strmonth + strday
}

func (c *HCAbsencesController) ReadAPI() error {
	log.Println("\n---------------------------------\nReading HC Absences API")
	var err error

	isFirstTimer := clit.Config("hc", "firsttimer", nil).(bool)
	payload := []byte{}
	if isFirstTimer {
		startdate := clit.Config("hc", "startdate", nil).(string)
		enddate := clit.Config("hc", "enddate", nil).(string)
		payload = c.SetParamBody(startdate, enddate)
	} else {
		strparampayload := c.CreateParamToday()
		payload = c.SetParamBody(strparampayload, strparampayload)
	}

	err = c.ExecProcess(payload)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return err
}
