package models

import "encoding/xml"

type Response struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName  xml.Name
			DAYSS    int `xml:"DAYSS"`
			ZHCDT003 struct {
				Item []struct {
					PLVAR        string
					OTYPE        string
					OBJID        string
					STEXT        string
					BEGDA        string
					ENDDA        string
					ZCOUNT_NORM  string
					ZCOST_CENTER string
					ZDEPT        string
				} `xml:"item"`
			} `xml:"ZHCDT003"`
			// Return  []interface{}
		} `xml:"ZFM_HC_006.Response"`
	}
}
