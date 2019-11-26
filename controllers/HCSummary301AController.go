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

type HcSummary301AController struct {
	*Base
}

func NewHcSummary301AController() *HcSummary301AController {
	return new(HcSummary301AController)
}

func (c *HcSummary301AController) ReadAPI() error {
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

func (c *HcSummary301AController) FetchTraining() ([]toolkit.M, error) {
	log.Println("FetchTrainingSummary301A")
	config := clit.Config("hc", "summary301A", map[string]interface{}{}).(map[string]interface{})

	payload := []byte(strings.TrimSpace(`
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
		<soapenv:Header/>
		<soapenv:Body>
			<urn:ZFM_HC_301A>
				<!--You may enter the following 5 items in any order-->
				<!--Optional:-->
				<!--Fiscal Period-->
					<MONTH xmlns="">10</MONTH>
					<!--Year for which levy is to be carried out-->
					<YEAR xmlns="">2019</YEAR>
					<!--HC Dashboard-->
					<ZHRST11 xmlns="">
						<item></item>
					</ZHRST11>
			</urn:ZFM_HC_301A>
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

	r := &models.SummaryResponse301A{}
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

func (c *HcSummary301AController) InsertTrainingDatas(results []toolkit.M) error {
	log.Println("inserting data....")
	var err error

	config := clit.Config("hc", "summary301A", nil).(map[string]interface{})
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
			rowData.Set("SOURCE", "301A")
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
