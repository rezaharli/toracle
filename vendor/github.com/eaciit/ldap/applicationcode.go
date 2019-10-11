package ldap

import (
	"github.com/eaciit/asn1-ber"
)

// LDAP Application Codes
type ApplicationCode ber.Tag

// go:generate stringer -type=ApplicationCode

const (
	ApplicationBindRequest           ApplicationCode = 0
	ApplicationBindResponse          ApplicationCode = 1
	ApplicationUnbindRequest         ApplicationCode = 2
	ApplicationSearchRequest         ApplicationCode = 3
	ApplicationSearchResultEntry     ApplicationCode = 4
	ApplicationSearchResultDone      ApplicationCode = 5
	ApplicationModifyRequest         ApplicationCode = 6
	ApplicationModifyResponse        ApplicationCode = 7
	ApplicationAddRequest            ApplicationCode = 8
	ApplicationAddResponse           ApplicationCode = 9
	ApplicationDelRequest            ApplicationCode = 10
	ApplicationDelResponse           ApplicationCode = 11
	ApplicationModifyDNRequest       ApplicationCode = 12
	ApplicationModifyDNResponse      ApplicationCode = 13
	ApplicationCompareRequest        ApplicationCode = 14
	ApplicationCompareResponse       ApplicationCode = 15
	ApplicationAbandonRequest        ApplicationCode = 16
	ApplicationSearchResultReference ApplicationCode = 19
	ApplicationExtendedRequest       ApplicationCode = 23
	ApplicationExtendedResponse      ApplicationCode = 24
)
