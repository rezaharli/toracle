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

type StatusController struct {
	*Base
}

func NewStatusController() *StatusController {
	return new(StatusController)
}

func (c *StatusController) ReadAPI() error {
	log.Println("\n--------------------------------------\nReading Status API")

	results, err := c.GetStatus()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = c.InsertStatusDatas(results, "status")
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return err
}

func (c *StatusController) GetStatus() ([]toolkit.M, error) {
	log.Println("Get Status")
	config := clit.Config("hc", "status", map[string]interface{}{}).(map[string]interface{})

	payload := []byte(strings.TrimSpace(`
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
   <soapenv:Header/>
   <soapenv:Body>
      <urn:ZFM_HC_102>
         <!--You may enter the following 5 items in any order-->
         <!--Optional:-->
         <!--Fiscal Period-->
            <MONTH xmlns="">11</MONTH>
            <!--Year for which levy is to be carried out-->
            <YEAR xmlns="">2019</YEAR>
            <!--HC Dashboard-->
            <ZHRST02 xmlns="">
                <item></item>
            </ZHRST02>
      </urn:ZFM_HC_102>
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

	r := &models.StatusResponse{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.ZHRST02.Item {
		result, _ := toolkit.ToM(value)
		results = append(results, result)
	}

	return results, err
}

func (c *StatusController) InsertStatusDatas(results []toolkit.M, jsonconf string) error {
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
		}

		// toolkit.Println(rowData)
		param := helpers.InsertParam{
			TableName: "F_HC_STATUS",
			Data:      rowData,
		}

		log.Println("Inserting data Status")
		err := helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		}
	}

	return err
}
