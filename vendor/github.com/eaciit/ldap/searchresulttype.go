package ldap

type SearchResultType uint8

const (
	SearchResultEntry     SearchResultType = SearchResultType(ApplicationSearchResultEntry)
	SearchResultReference SearchResultType = SearchResultType(ApplicationSearchResultReference)
	SearchResultDone      SearchResultType = SearchResultType(ApplicationSearchResultDone)
)
