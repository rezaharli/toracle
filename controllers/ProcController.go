package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/eaciit/clit"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/rezaharli/toracle/helpers"
)

type ProcController struct {
	*Base
	FirstTimer bool
}

func NewProcController() *ProcController {
	return new(ProcController)
}

func (c *ProcController) ReadAPI() error {
	log.Println("\n--------------------------------------\nReading PROC API, fromFirst: " + toolkit.ToString(c.FirstTimer))

	var start time.Time
	var err error
	if c.FirstTimer == true {
		start, err = time.Parse("2006-01-02", "2019-02-01")
		if err != nil {
			return err
		}
	} else {
		start = time.Now()
	}

	today := time.Now()
	for d := start; d.Before(today) || d.Equal(today); d = d.AddDate(0, 0, 1) {
		summaries, err := c.Fetch(d)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		err = c.InsertDatas(summaries)
		if err != nil {
			return err
		}
	}

	return err
}

func (c *ProcController) Fetch(t time.Time) ([]interface{}, error) {
	stringTimeNow := fmt.Sprintf("%d%02d%02d", t.Year(), t.Month(), t.Day())

	toolkit.Println()
	log.Println("fetching data at tanggal_update:", stringTimeNow)

	var param = url.Values{}
	param.Set("tanggal_update", stringTimeNow)

	var payload = bytes.NewBufferString(param.Encode())

	request, err := http.NewRequest("POST", clit.Config("proc", "url", nil).(string), payload)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	username := clit.Config("proc", "username", nil).(string)
	password := clit.Config("proc", "password", nil).(string)
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

	result := make([]interface{}, 0)
	err = json.Unmarshal([]byte(string(body)), &result)
	if err != nil {
		return nil, err
	}

	return result, err
}

func (c *ProcController) InsertDatas(summaries []interface{}) error {
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

func (c *ProcController) InsertSummary(summary map[string]interface{}) error {
	config := clit.Config("proc", "summary", nil).(map[string]interface{})
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
		TableName: "F_PROC_SUMMARY",
		Data:      rowData,
	}

	log.Println("Inserting data summary kode paket", rowData.GetString("KODE_PAKET"))
	err := helpers.Insert(param)
	if err != nil {
		helpers.HandleError(err)
	}

	return nil
}

func (c *ProcController) InsertDetail(detail map[string]interface{}) error {
	config := clit.Config("proc", "detail", nil).(map[string]interface{})
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
		TableName: "F_PROC_DETAIL",
		Data:      rowData,
	}

	log.Println("Inserting data detail kode paket", rowData.GetString("KODE_PAKET"))
	err := helpers.Insert(param)
	if err != nil {
		helpers.HandleError(err)
	}

	return nil
}
