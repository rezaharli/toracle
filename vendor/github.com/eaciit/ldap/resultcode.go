package ldap

// LDAP Result Codes
type ResultCode uint8

// go:generate stringer -type=ResultCode
const (
	ResultSuccess                      ResultCode = 0
	ResultOperationsError              ResultCode = 1
	ResultProtocolError                ResultCode = 2
	ResultTimeLimitExceeded            ResultCode = 3
	ResultSizeLimitExceeded            ResultCode = 4
	ResultCompareFalse                 ResultCode = 5
	ResultCompareTrue                  ResultCode = 6
	ResultAuthMethodNotSupported       ResultCode = 7
	ResultStrongAuthRequired           ResultCode = 8
	ResultReferral                     ResultCode = 10
	ResultAdminLimitExceeded           ResultCode = 11
	ResultUnavailableCriticalExtension ResultCode = 12
	ResultConfidentialityRequired      ResultCode = 13
	ResultSaslBindInProgress           ResultCode = 14
	ResultNoSuchAttribute              ResultCode = 16
	ResultUndefinedAttributeType       ResultCode = 17
	ResultInappropriateMatching        ResultCode = 18
	ResultConstraintViolation          ResultCode = 19
	ResultAttributeOrValueExists       ResultCode = 20
	ResultInvalidAttributeSyntax       ResultCode = 21
	ResultNoSuchObject                 ResultCode = 32
	ResultAliasProblem                 ResultCode = 33
	ResultInvalidDNSyntax              ResultCode = 34
	ResultAliasDereferencingProblem    ResultCode = 36
	ResultInappropriateAuthentication  ResultCode = 48
	ResultInvalidCredentials           ResultCode = 49
	ResultInsufficientAccessRights     ResultCode = 50
	ResultBusy                         ResultCode = 51
	ResultUnavailable                  ResultCode = 52
	ResultUnwillingToPerform           ResultCode = 53
	ResultLoopDetect                   ResultCode = 54
	ResultNamingViolation              ResultCode = 64
	ResultObjectClassViolation         ResultCode = 65
	ResultNotAllowedOnNonLeaf          ResultCode = 66
	ResultNotAllowedOnRDN              ResultCode = 67
	ResultEntryAlreadyExists           ResultCode = 68
	ResultObjectClassModsProhibited    ResultCode = 69
	ResultAffectsMultipleDSAs          ResultCode = 71
	ResultOther                        ResultCode = 80

	ErrorNetwork         = 201
	ErrorFilterCompile   = 202
	ErrorFilterDecompile = 203
	ErrorDebugging       = 204
	ErrorEncoding        = 205
	ErrorDecoding        = 206
	ErrorMissingControl  = 207
	ErrorInvalidArgument = 208
	ErrorLDIFRead        = 209
	ErrorLDIFWrite       = 210
	ErrorClosing         = 211
	ErrorUnknown         = 212
)
