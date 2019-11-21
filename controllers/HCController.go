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

type HcController struct {
	*Base
}

func NewHcController() *HcController {
	return new(HcController)
}

func (c *HcController) ReadAPI() error {
	log.Println("\n--------------------------------------\nReading HC API")

	results, err := c.FetchTraining()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = c.InsertTrainingDatas(results)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return err
}

func (c *HcController) FetchTraining() ([]toolkit.M, error) {
	log.Println("FetchTraining")
	config := clit.Config("hc", "training", map[string]interface{}{}).(map[string]interface{})

	payload := []byte(strings.TrimSpace(`
		<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
			<soapenv:Body>
				<urn:ZFM_HC_006>
					<!--You may enter the following 5 items in any order-->
					<!--Optional:-->
					<!--Fiscal Period-->
						<DATE_FROM xmlns="">20190101</DATE_FROM>
						<!--Year for which levy is to be carried out-->
						<DATE_TO xmlns="">20191031</DATE_TO>
						<!--HC Dashboard-->
						<ZHCDT003 xmlns="">
							<item></item>
						</ZHCDT003>
				</urn:ZFM_HC_006>
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

	r := &models.Response{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.ZHCDT003.Item {
		result, _ := toolkit.ToM(value)
		results = append(results, result)
	}

	// res2B, _ := json.MarshalIndent(results, "", "		")
	// fmt.Println(string(res2B))

	return results, err
}

func (c *HcController) InsertTrainingDatas(results []toolkit.M) error {
	log.Println("inserting data....")
	var err error

	config := clit.Config("hc", "training", nil).(map[string]interface{})
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
			TableName: "F_HC_TRAINING",
			Data:      rowData,
		}

		log.Println("Inserting data training")
		err := helpers.Insert(param)
		if err != nil {
			log.Fatal("Error inserting data, ERROR:", err.Error())
		}
	}

	return err
}
