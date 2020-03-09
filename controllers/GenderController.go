package controllers

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/rezaharli/toracle/models"
)

type GenderController struct {
	*Base
}

func NewGenderController() *GenderController {
	return new(GenderController)
}

func (c *GenderController) ReadAPI() error {
	log.Println("\n--------------------------------------\nReading Gender API")

	results, err := c.GetGenderDashboard()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = c.InsertGenderDatas(results, "genderdashboard")
	if err != nil {
		log.Println(err.Error())
		return err
	}

	resultsA, err := c.GetGenderDashboardA()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = c.InsertGenderDatas(resultsA, "genderdashboarda")
	if err != nil {
		log.Println(err.Error())
		return err
	}

	resultsB, err := c.GetGenderDashboardB()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = c.InsertGenderDatas(resultsB, "genderdashboardb")
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return err
}

func (c *GenderController) GetGenderDashboard() ([]toolkit.M, error) {
	log.Println("Get Gender Dashboard")
	config := clit.Config("hc", "genderdashboard", map[string]interface{}{}).(map[string]interface{})

	payload := []byte(strings.TrimSpace(`
		<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
			<soapenv:Header/>
			<soapenv:Body>
				<urn:ZFM_HC_103>
					<!--You may enter the following 5 items in any order-->
					<!--Optional:-->
					<!--Fiscal Period-->
						<MONTH xmlns="">11</MONTH>
						<!--Year for which levy is to be carried out-->
						<YEAR xmlns="">2019</YEAR>
						<!--HC Dashboard-->
						<ZHRST03 xmlns="">
							<item></item>
						</ZHRST03>
				</urn:ZFM_HC_103>
			</soapenv:Body>
		</soapenv:Envelope>`,
	))

	// soapAction := "urn:listUsers"
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

	r := &models.GenderResponse{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.ZHRST03.Item {
		result, _ := toolkit.ToM(value)
		results = append(results, result)
	}

	return results, err
}

func (c *GenderController) GetGenderDashboardA() ([]toolkit.M, error) {
	log.Println("Get Gender Dashboard A")
	config := clit.Config("hc", "genderdashboarda", map[string]interface{}{}).(map[string]interface{})

	payload := []byte(strings.TrimSpace(`
		<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
			<soapenv:Header/>
			<soapenv:Body>
				<urn:ZFM_HC_303A>
					<!--You may enter the following 5 items in any order-->
					<!--Optional:-->
					<!--Fiscal Period-->
						<MONTH xmlns="">11</MONTH>
						<!--Year for which levy is to be carried out-->
						<YEAR xmlns="">2019</YEAR>
						<!--HC Dashboard-->
						<ZHRST13 xmlns="">
							<item></item>
						</ZHRST13>
				</urn:ZFM_HC_303A>
			</soapenv:Body>
		</soapenv:Envelope>`,
	))

	// soapAction := "urn:listUsers"
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

	r := &models.GenderDashboardAResponse{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.ZHRST13.Item {
		result, _ := toolkit.ToM(value)
		results = append(results, result)
	}

	return results, err
}

func (c *GenderController) GetGenderDashboardB() ([]toolkit.M, error) {
	log.Println("Get Gender Dashboard B")
	config := clit.Config("hc", "genderdashboardb", map[string]interface{}{}).(map[string]interface{})

	payload := []byte(strings.TrimSpace(`
		<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
			<soapenv:Header/>
			<soapenv:Body>
				<urn:ZFM_HC_303B>
					<!--You may enter the following 5 items in any order-->
					<!--Optional:-->
					<!--Fiscal Period-->
						<MONTH xmlns="">11</MONTH>
						<!--Year for which levy is to be carried out-->
						<YEAR xmlns="">2019</YEAR>
						<!--HC Dashboard-->
						<ZHRST13 xmlns="">
							<item></item>
						</ZHRST13>
				</urn:ZFM_HC_303B>
			</soapenv:Body>
		</soapenv:Envelope>`,
	))

	// soapAction := "urn:listUsers"
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

	r := &models.GenderDashboardBResponse{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.ZHRST13.Item {
		result, _ := toolkit.ToM(value)
		results = append(results, result)
	}

	return results, err
}

func (c *GenderController) InsertGenderDatas(results []toolkit.M, jsonconf string) error {
	log.Println("inserting data....")
	var err error

	config := clit.Config("hc", jsonconf, nil).(map[string]interface{})
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
			if jsonconf == "genderdashboard" {
				rowData.Set("SOURCE", "103")
			} else if jsonconf == "genderdashboarda" {
				rowData.Set("SOURCE", "303A")
			} else if jsonconf == "genderdashboardb" {
				rowData.Set("SOURCE", "303B")
			}

		}

		// toolkit.Println(rowData)
		param := helpers.InsertParam{
			TableName: "F_HC_GENDER",
			Data:      rowData,
		}

		log.Println("Inserting data gender")
		err := helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		}
	}

	return err
}
