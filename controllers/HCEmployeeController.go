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

type HcEmployeeController struct {
	*Base
}

func NewHcEmployeeController() *HcEmployeeController {
	return new(HcEmployeeController)
}

func (c *HcEmployeeController) ReadAPI() error {
	log.Println("\n--------------------------------------\nReading HC Employee API")

	results, err := c.FetchEmployee()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = c.InsertEmployeeDatas(results)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return err
}

func (c *HcEmployeeController) FetchEmployee() ([]toolkit.M, error) {
	log.Println("FetchEmployee")
	config := clit.Config("hc", "employee", map[string]interface{}{}).(map[string]interface{})

	payload := []byte(strings.TrimSpace(`
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
		<soapenv:Header/>
		<soapenv:Body>
			<urn:ZFM_VIEW_PEGAWAI>
				<!--You may enter the following 2 items in any order-->
				<!--Optional:-->
				<PERNR></PERNR>
				<ZHRS002>
					<!--Zero or more repetitions:-->
					<item>
					</item>
				</ZHRS002>
			</urn:ZFM_VIEW_PEGAWAI>
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

	r := &models.HCEmployee{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.ZHRS002.Item {
		result, _ := toolkit.ToM(value)
		results = append(results, result)
	}

	return results, err
}

func (c *HcEmployeeController) InsertEmployeeDatas(results []toolkit.M) error {
	log.Println("inserting data....")
	var err error

	config := clit.Config("hc", "employee", nil).(map[string]interface{})
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
			if header.DBFieldName == "DOB" {
				dateString := result[header.Column].(string)
				t, err := time.Parse("2006-01-02", dateString)
				if err != nil {
					helpers.HandleError(err)
				}

				rowData.Set(header.DBFieldName, t)
			} else if header.DBFieldName == "FULL_NAME" {
				if strings.Contains(result[header.Column].(string), ",") {
					rowData.Set(header.DBFieldName, strings.Replace(result[header.Column].(string), ",", "", -1))
				} else if strings.Contains(result[header.Column].(string), "'") {
					rowData.Set(header.DBFieldName, strings.Replace(result[header.Column].(string), "'", "", -1))
				} else {
					rowData.Set(header.DBFieldName, result[header.Column])
				}
			} else if header.DBFieldName == "PERSONNEL_NUMBER" ||
				header.DBFieldName == "STATUS" ||
				header.DBFieldName == "NIPP" ||
				header.DBFieldName == "ID_POSITION" ||
				header.DBFieldName == "POSITION_CLASS" ||
				header.DBFieldName == "GENDER" ||
				header.DBFieldName == "ID_SUBDIR" {
				if result[header.Column].(string) == "*" || result[header.Column].(string) == "" {
					rowData.Set(header.DBFieldName, "0")
				} else {
					rowData.Set(header.DBFieldName, result[header.Column])
				}
			} else {
				rowData.Set(header.DBFieldName, result[header.Column])
			}
		}
		rowData.Set("UPDATE_DATE", time.Now())
		// toolkit.Println(rowData)
		sql := "DELETE FROM HC_EMPLOYEE WHERE PERSONNEL_NUMBER = " + rowData.GetString("PERSONNEL_NUMBER")
		conn := helpers.Database()
		query, err := conn.Prepare(dbflex.From("HC_EMPLOYEE").SQL(sql))
		if err != nil {
			log.Println(err)
		}

		_, err = query.Execute(toolkit.M{}.Set("data", toolkit.M{}))
		if err != nil {
			log.Println(err)
		}

		// log.Println("Data deleted.")

		param := helpers.InsertParam{
			TableName: "HC_EMPLOYEE",
			Data:      rowData,
		}

		log.Println("Inserting data Employee")
		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		}

		defer conn.Close()
	}

	return err
}
