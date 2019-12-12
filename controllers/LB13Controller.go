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

type LB13Controller struct {
	*Base
}

func NewLB13Controller() *LB13Controller {
	return new(LB13Controller)
}

func (c *LB13Controller) ReadAPI() error {
	log.Println("\n--------------------------------------\nReading LB13 API")
	var err error

	year := clit.Config("lb", "year", nil).(string)
	thisMonth := int(time.Now().Month())

	if year == strconv.Itoa(time.Now().Year()) {
		for i := 1; i <= thisMonth; i++ {
			payload := c.SetParamBody(year, strconv.Itoa(i))
			results, err := c.GetAPIDatas(payload, strconv.Itoa(i), year)
			if err != nil {
				log.Println(err.Error())
				return err
			}

			err = c.InsertAPIDatas(results, "lb13")
			if err != nil {
				log.Println(err.Error())
				return err
			}
		}
	} else {
		for i := 1; i <= 12; i++ {
			payload := c.SetParamBody(year, strconv.Itoa(i))
			results, err := c.GetAPIDatas(payload, strconv.Itoa(i), year)
			if err != nil {
				log.Println(err.Error())
				return err
			}

			err = c.InsertAPIDatas(results, "lb13")
			if err != nil {
				log.Println(err.Error())
				return err
			}
		}
	}

	return err
}

func (c *LB13Controller) GetAPIDatas(payload []byte, month string, year string) ([]toolkit.M, error) {
	log.Println("Get LB13 Month :", month)
	config := clit.Config("lb", "lb13", map[string]interface{}{}).(map[string]interface{})

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

	r := &models.LB13Response{}
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

func (c *LB13Controller) InsertAPIDatas(results []toolkit.M, jsonconf string) error {
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

		// toolkit.Println(rowData)
		param := helpers.InsertParam{
			TableName: "F_FA_REALISASI_PUSAT_PERJENIS",
			Data:      rowData,
		}

		log.Println("Inserting data API")
		err := helpers.Insert(param)
		if err != nil {
			log.Fatal("Error inserting data, ERROR:", err.Error())
		}
	}

	return err
}

func (c *LB13Controller) SetParamBody(year string, month string) []byte {
	payload := []byte(strings.TrimSpace(`
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
	<soapenv:Header/>
	<soapenv:Body>
		<urn:ZFM_FI_23>
				<P_BUKRS xmlns="">PTTL</P_BUKRS>
				<P_GJAHR xmlns="">` + year + `</P_GJAHR>
				<P_MONAT xmlns="">` + month + `</P_MONAT>
				<P_PDF xmlns=""></P_PDF>
				<P_RLDNR xmlns=""></P_RLDNR>
				<P_VERSI xmlns=""></P_VERSI>
				<T_DATA xmlns="">
					<item></item>
				</T_DATA>
		</urn:ZFM_FI_23>
	</soapenv:Body>
	</soapenv:Envelope>`,
	))

	return payload
}
