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

type EducationResponse struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName  xml.Name
			ZHCDT002 struct {
				Item []struct {
					PERNR  string
					MASSN  string
					BEGDA  string
					BTRTL  string
					PERSG  string
					SLART  string
					ZBEGDA string
					ZENDDA string
				} `xml:"item"`
			} `xml:"ZHCDT002"`
			// Return  []interface{}
		} `xml:"ZFM_HC_004.Response"`
	}
}

type StatusResponse struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			ZHRST02 struct {
				Item []struct {
					ZMONTH      string
					ZYEAR       string
					ZDEPT       string
					ZCOUNT_PEL  string
					ZCOUNT_TTL  string
					ZCOUNT_PKWT string
				} `xml:"item"`
			} `xml:"ZHRST02"`
			// Return  []interface{}
		} `xml:"ZFM_HC_102.Response"`
	}
}

type ProductivityResponse struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName  xml.Name
			ZHCDT004 struct {
				Item []struct {
					PERNR      string
					MASSN      string
					BEGDA      string
					ENDDA      string
					BTRTL      string
					PLANS      string
					ZPLAN_TIME string
					ZREAL_TIME string
					ZDEPT      string
					ZCOUNT_DAY string
				} `xml:"item"`
			} `xml:"ZHCDT004"`
			// Return  []interface{}
		} `xml:"ZFM_HC_007.Response"`
	}
}

type AttendanceResponse struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName  xml.Name
			ZHCDT005 struct {
				Item []struct {
					PERNR string
					MASSN string
					BEGDA string
					ENDDA string
					BTRTL string
					AWART string
					KALTG string
					ZDEPT string
				} `xml:"item"`
			} `xml:"ZHCDT005"`
			// Return  []interface{}
		} `xml:"ZFM_HC_008.Response"`
	}
}

type LB1Response struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			T_DATA  struct {
				Item []struct {
					NO      string
					NOREK   string
					POSNER  string
					TGLA    string
					TGLB    string
					NOREK2  string
					POSNER2 string
					TGL2A   string
					TGL2B   string
				} `xml:"item"`
			} `xml:"T_DATA"`
			// Return  []interface{}
		} `xml:"ZFM_FI_11.Response"`
	}
}

type LB2Response struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			T_DATA  struct {
				Item []struct {
					RACCT           string
					KOSTL           string
					TXT50           string
					BUDGET          string
					BLN_INI         string
					AKUM_BLN_INI    string
					AKUM_BLN_LALU   string
					AKUM_THN_LALU   string
					DEVIASI_A       string
					DEVIASI_B       string
					DEVIASI_C       string
					PERSEN_ANGGARAN string
					SISA_ANGGARAN   string
					WAERS           string
				} `xml:"item"`
			} `xml:"T_DATA"`
			// Return  []interface{}
		} `xml:"ZFM_FI_12.Response"`
	}
}

type LB4Response struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			T_DATA  struct {
				Item []struct {
					HKONT             string
					JENIS             string
					PREV_MONTH_IN     string
					PREV_MONTH_OUT    string
					CURRENT_MONTH_IN  string
					CURRENT_MONTH_OUT string
					AKUM_MONTH_IN     string
					AKUM_MONTH_OUT    string
					TEXT              string
					WAERS             string
					FLAG_DIS_IN       string
					FLAG_DIS_OUT      string
				} `xml:"item"`
			} `xml:"T_DATA"`
			// Return  []interface{}
		} `xml:"ZFM_FI_14.Response"`
	}
}

type LB5Response struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			T_DATA  struct {
				Item []struct {
					NO     string
					KETER  string
					FORM   string
					HASIL  string
					HASILD string
				} `xml:"item"`
			} `xml:"T_DATA"`
			// Return  []interface{}
		} `xml:"ZFM_FI_15.Response"`
	}
}

type LB13Response struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			T_DATA  struct {
				Item []struct {
					RACCT           string
					KOSTL           string
					TXT50           string
					BUDGET          string
					BLN_INI         string
					AKUM_BLN_INI    string
					AKUM_BLN_LALU   string
					AKUM_THN_LALU   string
					DEVIASI_A       string
					DEVIASI_B       string
					DEVIASI_C       string
					PERSEN_ANGGARAN string
					SISA_ANGGARAN   string
					WAERS           string
				} `xml:"item"`
			} `xml:"T_DATA"`
			// Return  []interface{}
		} `xml:"ZFM_FI_23.Response"`
	}
}

type LB10Response struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			T_DATA  struct {
				Item []struct {
					RACCT           string
					KOSTL           string
					TXT50           string
					BUDGET          string
					BLN_INI         string
					AKUM_BLN_INI    string
					AKUM_BLN_LALU   string
					AKUM_THN_LALU   string
					DEVIASI_A       string
					DEVIASI_B       string
					DEVIASI_C       string
					PERSEN_ANGGARAN string
					SISA_ANGGARAN   string
					WAERS           string
				} `xml:"item"`
			} `xml:"T_DATA"`
			// Return  []interface{}
		} `xml:"ZFM_FI_20.Response"`
	}
}

type LB11Response struct {
	XMLName xml.Name
	Body    struct {
		XMLName xml.Name
		Urn     struct {
			XMLName xml.Name
			T_DATA  struct {
				Item []struct {
					RACCT           string
					KOSTL           string
					TXT50           string
					BUDGET          string
					BLN_INI         string
					AKUM_BLN_INI    string
					AKUM_BLN_LALU   string
					AKUM_THN_LALU   string
					DEVIASI_A       string
					DEVIASI_B       string
					DEVIASI_C       string
					PERSEN_ANGGARAN string
					SISA_ANGGARAN   string
					WAERS           string
				} `xml:"item"`
			} `xml:"T_DATA"`
			// Return  []interface{}
		} `xml:"ZFM_FI_21.Response"`
	}
}
