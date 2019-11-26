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

type SummaryResponse201 struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			ZHRST11 struct {
				Item []struct {
					ZMONTH       string
					ZYEAR        string
					ZDEPT        string
					ZCOUNT_MONTH string
					ZCOUNT_YTD   string
				} `xml:"item"`
			} `xml:"ZHRST11"`
			// Return  []interface{}
		} `xml:"ZFM_HC_201.Response"`
	}
}

type SummaryResponse301A struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			ZHRST11 struct {
				Item []struct {
					ZMONTH       string
					ZYEAR        string
					ZDEPT        string
					ZCOUNT_MONTH string
					ZCOUNT_YTD   string
				} `xml:"item"`
			} `xml:"ZHRST11"`
			// Return  []interface{}
		} `xml:"ZFM_HC_301A.Response"`
	}
}

type SummaryResponse301B struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			ZHRST11 struct {
				Item []struct {
					ZMONTH       string
					ZYEAR        string
					ZDEPT        string
					ZCOUNT_MONTH string
					ZCOUNT_YTD   string
				} `xml:"item"`
			} `xml:"ZHRST11"`
			// Return  []interface{}
		} `xml:"ZFM_HC_301B.Response"`
	}
}
