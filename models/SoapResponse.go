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

type GenderResponse struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			ZHRST03 struct {
				Item []struct {
					ZMONTH     string
					ZYEAR      string
					ZEMP_GROUP string
					ZGEN_MAN   string
					ZGEN_WMN   string
					ZSER_DIR   string
					ZSER_IND   string
					ZSER_SUP   string
					ZEDU_AB    string
					ZEDU_05    string
					ZEDU_06    string
					ZEDU_07    string
					ZEDU_08    string
					ZAGE_30    string
					ZAGE_3135  string
					ZAGE_3640  string
					ZAGE_4145  string
					ZAGE_4650  string
					ZAGE_50    string
				} `xml:"item"`
			} `xml:"ZHRST03"`
			// Return  []interface{}
		} `xml:"ZFM_HC_103.Response"`
	}
}

type GenderDashboardAResponse struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			ZHRST13 struct {
				Item []struct {
					ZMONTH     string
					ZYEAR      string
					ZEMP_GROUP string
					ZGEN_MAN   string
					ZGEN_WMN   string
					ZSER_DIR   string
					ZSER_IND   string
					ZSER_SUP   string
					ZEDU_AB    string
					ZEDU_05    string
					ZEDU_06    string
					ZEDU_07    string
					ZEDU_08    string
					ZAGE_30    string
					ZAGE_3135  string
					ZAGE_3640  string
					ZAGE_4145  string
					ZAGE_4650  string
					ZAGE_50    string
				} `xml:"item"`
			} `xml:"ZHRST13"`
			// Return  []interface{}
		} `xml:"ZFM_HC_303A.Response"`
	}
}

type GenderDashboardBResponse struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			ZHRST13 struct {
				Item []struct {
					ZMONTH     string
					ZYEAR      string
					ZEMP_GROUP string
					ZGEN_MAN   string
					ZGEN_WMN   string
					ZSER_DIR   string
					ZSER_IND   string
					ZSER_SUP   string
					ZEDU_AB    string
					ZEDU_05    string
					ZEDU_06    string
					ZEDU_07    string
					ZEDU_08    string
					ZAGE_30    string
					ZAGE_3135  string
					ZAGE_3640  string
					ZAGE_4145  string
					ZAGE_4650  string
					ZAGE_50    string
				} `xml:"item"`
			} `xml:"ZHRST13"`
			// Return  []interface{}
		} `xml:"ZFM_HC_303B.Response"`
	}
}
