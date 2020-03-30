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

type MasterController struct {
	*Base
}

func NewMasterController() *MasterController {
	return new(MasterController)
}

func (c *MasterController) ReadAPI() error {
	var err error
	compcode := []string{"PTTL", "ZTOS"}

	for _, cc := range compcode {

		log.Println("\n--------------------------------------\nReading Master Vendor API", cc)

		resultsVendor, err := c.FetchVendor(cc)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		for _, vendor := range resultsVendor {
			vendor.Set("COMP_CODE", cc)
		}

		log.Println("\n--------------------------------------\nInserting", len(resultsVendor), "Rows Master Vendor API")

		err = c.InsertData(resultsVendor, "vendor")
		if err != nil {
			log.Println(err.Error())
			return err
		}

		log.Println("\n--------------------------------------\nReading Master Customer API", cc)

		resultsCustomer, err := c.FetchCustomer(cc)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		for _, cust := range resultsCustomer {
			cust.Set("COMP_CODE", cc)
		}

		log.Println("\n--------------------------------------\nInserting", len(resultsCustomer), " Master Customer API")

		err = c.InsertData(resultsCustomer, "customer")
		if err != nil {
			log.Println(err.Error())
			return err
		}

	}

	return err
}

func (c *MasterController) FetchCustomer(compcode string) ([]toolkit.M, error) {
	log.Println("Fetch Customers")

	config := clit.Config("master", "customer", map[string]interface{}{}).(map[string]interface{})
	username := clit.Config("master", "username", nil).(string)
	password := clit.Config("master", "password", nil).(string)
	// compcode := clit.Config("master", "compcode", nil).(string)

	parambody := `<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
					<Body>
						<ZFMFI_CUSTOMER_GETLIST xmlns="urn:sap-com:document:sap:rfc:functions">
							<!--Company Code-->
							<COMP_CODE xmlns="">` + compcode + `</COMP_CODE>
							<!--List of Customers-->
							<!-- Optional -->
							<CUSTOMER xmlns="">
								<item></item>
							</CUSTOMER>
						</ZFMFI_CUSTOMER_GETLIST>
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

	r := &models.MasterCustomer{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.CUSTOMER.Item {
		result, _ := toolkit.ToM(value)
		results = append(results, result)
	}

	return results, err
}

func (c *MasterController) FetchVendor(compcode string) ([]toolkit.M, error) {
	log.Println("Fetch Vendor")

	config := clit.Config("master", "vendor", map[string]interface{}{}).(map[string]interface{})
	username := clit.Config("master", "username", nil).(string)
	password := clit.Config("master", "password", nil).(string)
	// compcode := clit.Config("master", "compcode", nil).(string)

	parambody := `<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
					<Body>
						<ZFMFI_VENDOR_GETLIST xmlns="urn:sap-com:document:sap:rfc:functions">
							<!--Company Code-->
							<COMP_CODE xmlns="">` + compcode + `</COMP_CODE>
							<!--List of Vendors-->
							<!-- Optional -->
							<VENDOR xmlns="">
								<item></item>
							</VENDOR>
						</ZFMFI_VENDOR_GETLIST>
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

	r := &models.MasterVendor{}
	err = xml.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	results := make([]toolkit.M, 0)
	for _, value := range r.Body.Urn.VENDOR.Item {
		result, _ := toolkit.ToM(value)
		results = append(results, result)
	}

	return results, err
}

func (c *MasterController) InsertData(results []toolkit.M, configname string) error {
	var err error

	config := clit.Config("master", configname, nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})
	// compcode := clit.Config("master", "compcode", nil).(string)
	tablename := ""
	wherecol := ""
	if configname == "vendor" {
		tablename = "SAP_VENDOR"
		wherecol = "VENDOR_NO"
	} else if configname == "customer" {
		tablename = "SAP_CUSTOMER"
		wherecol = "CUST_NO"
	}

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
			if header.DBFieldName == "VENDOR_NAME" || header.DBFieldName == "CUST_NAME" {
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
		// rowData.Set("COMP_CODE", compcode)

		sql := "DELETE FROM " + tablename + " WHERE " + wherecol + " = '" + rowData.GetString(wherecol) + "'"
		conn := helpers.Database()
		query, err := conn.Prepare(dbflex.From(tablename).SQL(sql))
		if err != nil {
			log.Println(err)
		}

		_, err = query.Execute(toolkit.M{}.Set("data", toolkit.M{}))
		if err != nil {
			log.Println(err)
		}

		// log.Println(tablename, rowData)
		param := helpers.InsertParam{
			TableName: tablename,
			Data:      rowData,
		}

		// log.Println("Inserting Data:", configname)
		err = helpers.Insert(param)
		if err != nil {
			helpers.HandleError(err)
		}
	}

	return err
}
