package controllers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
)

type HcController struct {
	*Base
	FirstTimer bool
}

func NewHcController() *HcController {
	return new(HcController)
}

func (c *HcController) ReadAPI() error {
	log.Println("\n--------------------------------------\nReading HC API, fromFirst: " + toolkit.ToString(c.FirstTimer))

	summaries, err := c.FetchTraining()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = c.InsertDatas(summaries)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return err
}

func (c *HcController) FetchTraining() ([]interface{}, error) {
	log.Println("FetchTraining")
	config := clit.Config("hc", "training", map[string]interface{}{}).(map[string]interface{})

	payload := []byte(strings.TrimSpace(`
		<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:sap-com:document:sap:rfc:functions">
			<soapenv:Header/>
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

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, 0)
	err = json.Unmarshal([]byte(string(body)), &result)
	if err != nil {
		return nil, err
	}

	return result, err
}

func (c *HcController) InsertDatas(summaries []interface{}) error {
	log.Println("inserting data....")

	var err error

	for _, summary := range summaries {
		summaryMap := summary.(map[string]interface{})

		err = c.InsertSummary(summaryMap)
		if err != nil {
			return err
		}

		for key, value := range summaryMap {
			if key == "detail" {
				details := value.([]interface{})

				for _, detail := range details {
					detailMap := detail.(map[string]interface{})

					err = c.InsertDetail(detailMap)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return err
}

func (c *HcController) InsertSummary(summary map[string]interface{}) error {
	config := clit.Config("hc", "summary", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	var headers []Header
	for dbFieldName, attributeName := range columnsMapping {
		header := Header{
			DBFieldName: dbFieldName,
			Column:      attributeName.(string),
		}

		headers = append(headers, header)
	}

	rowData := toolkit.M{}
	for _, header := range headers {
		if header.DBFieldName == "TANGGAL_PEMBUATAN" || header.DBFieldName == "UPDATE_DATE" {
			date := summary[header.Column].(string)

			var t time.Time
			var err error
			if date != "" {
				timeFormats := []string{"2006-01-02", "02-01-2006", "02-JAN-2006", "02-JAN-06"}
				for i, timeFormat := range timeFormats {
					t, err = time.Parse(timeFormat, date)

					if err == nil {
						break
					} else {
						if i == len(timeFormats)-1 {
							log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
						} else {
							continue
						}
					}
				}
			}

			rowData.Set(header.DBFieldName, t)
		} else if header.DBFieldName == "TAHUN_ANGGARAN" || header.DBFieldName == "HPS" {
			var number float64
			var err error

			if summary[header.Column] != nil {
				value := summary[header.Column].(string)

				number, err = strconv.ParseFloat(value, 64)
				if err != nil {
					log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
				}
			}

			rowData.Set(header.DBFieldName, number)
		} else {
			rowData.Set(header.DBFieldName, summary[header.Column])
		}
	}

	param := helpers.InsertParam{
		TableName: "F_HC_SUMMARY",
		Data:      rowData,
	}

	log.Println("Inserting data summary kode paket", rowData.GetString("KODE_PAKET"))
	err := helpers.Insert(param)
	if err != nil {
		log.Fatal("Error inserting data, ERROR:", err.Error())
	}

	return nil
}

func (c *HcController) InsertDetail(detail map[string]interface{}) error {
	config := clit.Config("hc", "detail", nil).(map[string]interface{})
	columnsMapping := config["columnsMapping"].(map[string]interface{})

	var headers []Header
	for dbFieldName, attributeName := range columnsMapping {
		header := Header{
			DBFieldName: dbFieldName,
			Column:      attributeName.(string),
		}

		headers = append(headers, header)
	}

	rowData := toolkit.M{}
	for _, header := range headers {
		if header.DBFieldName == "UPDATE_DATE" {
			date := detail[header.Column].(string)

			var t time.Time
			var err error
			if date != "" {
				timeFormats := []string{"2006-01-02", "02-01-2006", "02-JAN-2006", "02-JAN-06"}
				for i, timeFormat := range timeFormats {
					t, err = time.Parse(timeFormat, date)

					if err == nil {
						break
					} else {
						if i == len(timeFormats)-1 {
							log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
						} else {
							continue
						}
					}
				}
			}

			rowData.Set(header.DBFieldName, t)
		} else if header.DBFieldName == "START_DATE" {
			date := detail["tanggal_awal"].(string)

			var t time.Time
			var err error
			if detail["jam_awal"] != nil {
				date = date + "-" + detail["jam_awal"].(string)

				if date != "" {
					timeFormats := []string{"2006-01-02-15:04", "02-01-2006-15:04", "02-JAN-2006-15:04", "02-JAN-06-15:04"}
					for i, timeFormat := range timeFormats {
						t, err = time.Parse(timeFormat, date)

						if err == nil {
							break
						} else {
							if i == len(timeFormats)-1 {
								log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
							} else {
								continue
							}
						}
					}
				}
			} else {
				if date != "" {
					timeFormats := []string{"2006-01-02", "02-01-2006", "02-JAN-2006", "02-JAN-06"}
					for i, timeFormat := range timeFormats {
						t, err = time.Parse(timeFormat, date)

						if err == nil {
							break
						} else {
							if i == len(timeFormats)-1 {
								log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
							} else {
								continue
							}
						}
					}
				}
			}

			rowData.Set(header.DBFieldName, t)
		} else if header.DBFieldName == "END_DATE" {
			date := detail["tanggal_akhir"].(string)

			var t time.Time
			var err error
			if detail["jam_akhir"] != nil {
				date = date + "-" + detail["jam_akhir"].(string)

				if date != "" {
					timeFormats := []string{"2006-01-02-15:04", "02-01-2006-15:04", "02-JAN-2006-15:04", "02-JAN-06-15:04"}
					for i, timeFormat := range timeFormats {
						t, err = time.Parse(timeFormat, date)

						if err == nil {
							break
						} else {
							if i == len(timeFormats)-1 {
								log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
							} else {
								continue
							}
						}
					}
				}
			} else {
				if date != "" {
					timeFormats := []string{"2006-01-02", "02-01-2006", "02-JAN-2006", "02-JAN-06"}
					for i, timeFormat := range timeFormats {
						t, err = time.Parse(timeFormat, date)

						if err == nil {
							break
						} else {
							if i == len(timeFormats)-1 {
								log.Println("Error getting value for", header.DBFieldName, "ERROR:", err)
							} else {
								continue
							}
						}
					}
				}
			}

			rowData.Set(header.DBFieldName, t)
		} else {
			rowData.Set(header.DBFieldName, detail[header.Column])
		}
	}

	param := helpers.InsertParam{
		TableName: "F_HC_DETAIL",
		Data:      rowData,
	}

	log.Println("Inserting data detail kode paket", rowData.GetString("KODE_PAKET"))
	err := helpers.Insert(param)
	if err != nil {
		log.Fatal("Error inserting data, ERROR:", err.Error())
	}

	return nil
}
