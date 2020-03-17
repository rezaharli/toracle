package controllers

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
	"git.eaciitapp.com/rezaharli/toracle/models"
)

type HcFullTrainingController struct {
	*Base
}

func NewHcFullTrainingController() *HcFullTrainingController {
	return new(HcFullTrainingController)
}

func (c *HcFullTrainingController) ReadAPI() error {
	log.Println("\n--------------------------------------\nReading HC Employee API")

	results, err := c.FetchFullTraining()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = c.InsertFullTrainingDatas(results)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return err
}

func (c *HcFullTrainingController) FetchFullTraining() ([]toolkit.M, error) {
	log.Println("FetchFullTraining")
	config := clit.Config("hc", "full_training", map[string]interface{}{}).(map[string]interface{})

	payload := []byte(strings.TrimSpace(`
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
		<soapenv:Header/>
		<soapenv:Body>
			<urn:ZFM_VIEW_TRAINING>
				<!--You may enter the following 3 items in any order-->
				<!--Optional:-->
				<BEGDA></BEGDA>
				<!--Optional:-->
				<ENDDA></ENDDA>
				<ZHRS001>
					<!--Zero or more repetitions:-->
					<item>
					</item>
				</ZHRS001>
			</urn:ZFM_VIEW_TRAINING>
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

	r := &models.HCFullTraining{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.ZHRS001.Item {
		result, _ := toolkit.ToM(value)
		results = append(results, result)
	}

	return results, err
}

func (c *HcFullTrainingController) InsertFullTrainingDatas(results []toolkit.M) error {
	log.Println("inserting data....")
	var err error

	config := clit.Config("hc", "full_training", nil).(map[string]interface{})
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
			if header.DBFieldName == "BEGIN_DATE" || header.DBFieldName == "END_DATE" || header.DBFieldName == "CHANGED_ON" {
				if result[header.Column].(string) != "0000-00-00" {
					dateString := result[header.Column].(string)
					t, err := time.Parse("2006-01-02", dateString)
					if err != nil {
						helpers.HandleError(err)
					}

					rowData.Set(header.DBFieldName, t)
				}

			} else if header.DBFieldName == "PERSONNEL_NUMBER" ||
				header.DBFieldName == "OBJ_ID" ||
				header.DBFieldName == "POSITION" ||
				header.DBFieldName == "ID_SUBDIR" ||
				header.DBFieldName == "PERSONNEL_AREA" ||
				header.DBFieldName == "MIN_QTY" ||
				header.DBFieldName == "MAX_QTY" ||
				header.DBFieldName == "DAYS_QTY" ||
				header.DBFieldName == "VDAYS_QTY" {
				if result[header.Column].(string) == "*" {
					rowData.Set(header.DBFieldName, "0")
				} else {
					rowData.Set(header.DBFieldName, result[header.Column])
				}
			} else if header.DBFieldName == "EMPLOYEE_NAME" || header.DBFieldName == "OBJ_NAME" {
				if strings.Contains(result[header.Column].(string), ",") {
					rowData.Set(header.DBFieldName, strings.Replace(result[header.Column].(string), ",", "", -1))
				} else if strings.Contains(result[header.Column].(string), "'") {
					rowData.Set(header.DBFieldName, strings.Replace(result[header.Column].(string), "'", "", -1))
				} else {
					rowData.Set(header.DBFieldName, result[header.Column])
				}
			} else {
				rowData.Set(header.DBFieldName, result[header.Column])
			}
		}
		rowData.Set("UPDATE_DATE", time.Now())
		// toolkit.Println(rowData)
		param := helpers.InsertParam{
			TableName: "HC_FULL_TRAINING",
			Data:      rowData,
		}

		log.Println("Inserting data Full Training")
		err := helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		}
	}

	return err
}
