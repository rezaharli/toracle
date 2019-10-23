package controllers

type Base struct {
}

type SqlQueryParam struct {
	ItemName string
	Results  interface{}
}

type Header struct {
	DBFieldName string
	HeaderName  string

	Column       string
	ColumnNumber int
	Row          string

	Value string
}
