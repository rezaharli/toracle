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

type HcSummaryController struct {
	*Base
}

func NewHcSummaryController() *HcSummaryController {
	return new(HcSummaryController)
}

func (c *HcSummaryController) ReadAPI() error {
	log.Println("\n--------------------------------------\nReading HC Summary API")

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

func (c *HcSummaryController) FetchTraining() ([]toolkit.M, error) {
	log.Println("FetchTrainingSummary201")
	config := clit.Config("hc", "summary201", map[string]interface{}{}).(map[string]interface{})

	payload := []byte(strings.TrimSpace(`
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
		<soapenv:Header/>
		<soapenv:Body>
			<urn:ZFM_HC_201>
				<!--You may enter the following 5 items in any order-->
				<!--Optional:-->
				<!--Fiscal Period-->
					<DATE_FROM xmlns="">20190101</DATE_FROM>
					<!--Year for which levy is to be carried out-->
					<DATE_TO xmlns="">20191031</DATE_TO>
					<!--HC Dashboard-->
					<ZHRST11 xmlns="">
						<item></item>
					</ZHRST11>
			</urn:ZFM_HC_201>
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

	r := &models.SummaryResponse201{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.ZHRST11.Item {
		result, _ := toolkit.ToM(value)
		results = append(results, result)
	}

	// res2B, _ := json.MarshalIndent(results, "", "		")
	// fmt.Println(string(res2B))

	return results, err
}

func (c *HcSummaryController) InsertTrainingDatas(results []toolkit.M) error {
	log.Println("inserting data....")
	var err error

	config := clit.Config("hc", "summary201", nil).(map[string]interface{})
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
			rowData.Set("SOURCE", "201")
		}

		// toolkit.Println(rowData)
		param := helpers.InsertParam{
			TableName: "F_HC_TRAINING_SUMMARY",
			Data:      rowData,
		}

		log.Println("Inserting data training summary")
		err := helpers.Insert(param)
		if err != nil {
			log.Fatal("Error inserting data, ERROR:", err.Error())
		}
	}

	return err
}
