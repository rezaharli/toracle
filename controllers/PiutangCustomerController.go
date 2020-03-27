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

type PiutangController struct {
	*Base
}

func NewPiutangController() *PiutangController {
	return new(PiutangController)
}

func (c *PiutangController) FetchPiutang(customer string) ([]toolkit.M, error) {
	log.Println("Fetch Piutang Customers")

	config := clit.Config("master", "piutang_customer", map[string]interface{}{}).(map[string]interface{})
	username := clit.Config("master", "username", nil).(string)
	password := clit.Config("master", "password", nil).(string)
	compcode := clit.Config("master", "compcode", nil).(string)
	keydate := clit.Config("master", "keydate", nil).(string)

	parambody := `<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
					<Body>
						<BAPI_AR_ACC_GETOPENITEMS xmlns="urn:sap-com:document:sap:rfc:functions">
							<!--Company code-->
							<COMPANYCODE xmlns="">` + compcode + `</COMPANYCODE>
							<!--Customer-->
							<CUSTOMER xmlns="">` + customer + `</CUSTOMER>
							<!--Key date-->
							<KEYDATE xmlns="">` + keydate + `</KEYDATE>
							<!--Noted items requested-->
							<NOTEDITEMS xmlns=""></NOTEDITEMS>
							<!--Secondary Indexes Required-->
							<SECINDEX xmlns=""></SECINDEX>
							<!--Line items-->
							<LINEITEMS xmlns="">
								<item></item>
							</LINEITEMS>
						</BAPI_AR_ACC_GETOPENITEMS>
					</Body>
				</Envelope>`
	payload := []byte(strings.TrimSpace(parambody))

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

	r := &models.PiutangCustomer{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.LINEITEMS.Item {
		result, _ := toolkit.ToM(value)
		results = append(results, result)
	}

	return results, err
}

func (c *PiutangController) InsertData(results []toolkit.M) error {
	var err error

	config := clit.Config("master", "piutang_customer", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})
	keydate := clit.Config("master", "keydate", nil).(string)

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
			if header.DBFieldName == "DOCUMENT_DATE" ||
				header.DBFieldName == "POSTING_DATE" ||
				header.DBFieldName == "BASELINE_PAYMENT_DATE" {
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
		key, err := time.Parse("20060102", keydate)
		if err != nil {
			helpers.HandleError(err)
		}
		rowData.Set("PERIODE", key)

		param := helpers.InsertParam{
			TableName: "PIUTANG_CUSTOMER",
			Data:      rowData,
		}

		log.Println("Inserting Data:piutang")
		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		}
	}

	return err
}

func (c *PiutangController) ReadAPI() error {
	log.Println("\n--------------------------------------\nReading Master Customer Data")

	sqlQuery := "SELECT * FROM SAP_CUSTOMER"

	conn := helpers.Database()
	cursor := conn.Cursor(dbflex.From("SAP_CUSTOMER").SQL(sqlQuery), nil)
	defer cursor.Close()

	res := []toolkit.M{}
	err := cursor.Fetchs(&res, 0)

	for _, cust := range res {
		cust_no := cust.GetString("CUST_NO")
		resultsPiutang, err := c.FetchPiutang(cust_no)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		if len(resultsPiutang) > 0 {
			err = c.InsertData(resultsPiutang)
			if err != nil {
				log.Println(err.Error())
				return err
			}
		}
	}

	return err
}