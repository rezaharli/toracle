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
	"git.eaciitapp.com/sebar/dbflex"
)

type EducationController struct {
	*Base
}

func NewEducationController() *EducationController {
	return new(EducationController)
}

func (c *EducationController) ReadAPI() error {
	log.Println("\n--------------------------------------\nReading Education API")

	results, err := c.GetEducation()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = c.InsertEducationDatas(results, "education")
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return err
}

func (c *EducationController) GetEducation() ([]toolkit.M, error) {
	log.Println("Get Education")
	config := clit.Config("hc", "education", map[string]interface{}{}).(map[string]interface{})

	payload := []byte(strings.TrimSpace(`
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
   <soapenv:Header/>
   <soapenv:Body>
      <urn:ZFM_HC_004>
         <!--You may enter the following 5 items in any order-->
         <!--Optional:-->
         <!--Fiscal Period-->
            <BULAN xmlns="">10</BULAN>
            <!--Field of type DATS-->
            <DATE_FROM xmlns="">20191001</DATE_FROM>
            <!--Field of type DATS-->
            <DATE_TO xmlns="">20191031</DATE_TO>
            <!--Year for which levy is to be carried out-->
            <TAHUN xmlns="">2019</TAHUN>
            <!--HC Dashboard-->
            <ZHCDT002 xmlns="">
                <item></item>
            </ZHCDT002>
      </urn:ZFM_HC_004>
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

	r := &models.EducationResponse{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.ZHCDT002.Item {
		result, _ := toolkit.ToM(value)
		results = append(results, result)
	}

	return results, err
}

func (c *EducationController) InsertEducationDatas(results []toolkit.M, jsonconf string) error {
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
			if header.DBFieldName == "START_DATE" || header.DBFieldName == "END_DATE" || header.DBFieldName == "JOIN_DATE" {
				dateString := result[header.Column].(string)
				if dateString != "0000-00-00" {
					t, err := time.Parse("2006-01-02", dateString)
					if err != nil {
						helpers.HandleError(err)
					}
					rowData.Set(header.DBFieldName, t)
				} else {
					rowData.Set(header.DBFieldName, time.Time{})
				}
			} else {
				rowData.Set(header.DBFieldName, result[header.Column])
			}
		}

		// toolkit.Println(rowData)

		sql := "DELETE FROM F_HC_EDUCATION WHERE PERSONNEL_NUMBER = " + rowData.GetString("PERSONNEL_NUMBER")
		conn := helpers.Database()
		query, err := conn.Prepare(dbflex.From("F_HC_EDUCATION").SQL(sql))
		if err != nil {
			log.Println(err)
		}

		_, err = query.Execute(toolkit.M{}.Set("data", toolkit.M{}))
		if err != nil {
			log.Println(err)
		}

		param := helpers.InsertParam{
			TableName: "F_HC_EDUCATION",
			Data:      rowData,
		}

		log.Println("Inserting data Education")
		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		}
	}

	return err
}
