package ldap

import (
	"errors"
	"fmt"
	"github.com/eaciit/asn1-ber"
	"log"
)

// Control Interface
type Control interface {
	Encode() (*ber.Packet, error)
	GetControlType() ControlType
	String() string
}

type ControlString struct {
	ControlType  ControlType
	Criticality  bool
	ControlValue string
}

func NewControlStringFromPacket(p *ber.Packet) (Control, error) {
	controlType, criticality, valuePacket := decodeControlTypeAndCrit(p)
	c := new(ControlString)
	c.ControlType = controlType
	c.Criticality = criticality

	// FIXME: this is hacky, but like the original implementation in the asn1-ber packet previously used
	switch t := valuePacket.Value.(type) {
	case string:
		c.ControlValue = t
	case []byte:
		c.ControlValue = string(t)
	default:
		c.ControlValue = ""
	}

	return c, nil
}

func (c *ControlString) GetControlType() ControlType {
	return c.ControlType
}

func (c *ControlString) Encode() (p *ber.Packet, err error) {
	p = ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, string(c.ControlType), fmt.Sprintf("Control Type (%v)", c.ControlType)))
	if c.Criticality {
		p.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, c.Criticality, "Criticality"))
	}
	if len(c.ControlValue) != 0 {
		p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, c.ControlValue, "Control Value"))
	}
	return p, nil
}

func (c *ControlString) String() string {
	return fmt.Sprintf("Control Type: %s (%q)  Criticality: %t  Control Value: %s", c.ControlType.String(), string(c.ControlType), c.Criticality, c.ControlValue)
}

type ControlPaging struct {
	PagingSize uint32
	Cookie     []byte
}

func NewControlPaging(PagingSize uint32) *ControlPaging {
	return &ControlPaging{PagingSize: PagingSize}
}

func NewControlPagingFromPacket(p *ber.Packet) (Control, error) {
	_, _, value := decodeControlTypeAndCrit(p)
	value.Description += " (Paging)"
	c := new(ControlPaging)

	if value.Value != nil {
		value_children := ber.DecodePacket(value.Data.Bytes())
		value.Data.Truncate(0)
		value.Value = nil
		value.AppendChild(value_children)
	}
	value = value.Children[0]
	value.Description = "Search Control Value"
	value.Children[0].Description = "Paging Size"
	value.Children[1].Description = "Cookie"
	pagingSize, ok := value.Children[0].Value.(uint64)
	if !ok {
		return c, errors.New(fmt.Sprintf("type assertion uint64 for %v failed!\n", p.Children[0].Value))
	}
	c.PagingSize = uint32(pagingSize)
	c.Cookie = value.Children[1].Data.Bytes()
	value.Children[1].Value = c.Cookie
	return c, nil
}

func (c *ControlPaging) GetControlType() ControlType {
	return ControlTypePaging
}

func (c *ControlPaging) Encode() (p *ber.Packet, err error) {
	p = ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, string(ControlTypePaging), fmt.Sprintf("Control Type (%v)", ControlTypePaging)))

	p2 := ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, nil, "Control Value (Paging)")
	seq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Search Control Value")
	seq.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, uint64(c.PagingSize), "Paging Size"))
	cookie := ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, nil, "Cookie")
	cookie.Value = c.Cookie
	cookie.Data.Write(c.Cookie)
	seq.AppendChild(cookie)
	p2.AppendChild(seq)

	p.AppendChild(p2)
	return p, nil
}

func (c *ControlPaging) String() string {
	return fmt.Sprintf(
		"Control Type: %s (%q)  Criticality: %t  PagingSize: %d  Cookie: %q",
		ControlTypePaging.String(),
		string(ControlTypePaging),
		false,
		c.PagingSize,
		c.Cookie)
}

func (c *ControlPaging) SetCookie(Cookie []byte) {
	c.Cookie = Cookie
}

func FindControl(controls []Control, controlType ControlType) (position int, control Control) {
	for pos, c := range controls {
		if c.GetControlType() == controlType {
			return pos, c
		}
	}
	return -1, nil
}

func ReplaceControl(controls []Control, control Control) (oldControl Control) {
	ControlType := control.GetControlType()
	pos, c := FindControl(controls, ControlType)
	if c != nil {
		controls[pos] = control
		return c
	}
	controls = append(controls, control)
	return control
}

func decodeControlTypeAndCrit(p *ber.Packet) (controlType ControlType, criticality bool, valuePacket *ber.Packet) {
	// FIXME: this is hacky, but like the original implementation in the asn1-ber packet previously used
	switch t := p.Children[0].Value.(type) {
	case string:
		controlType = ControlType(t)
	case []byte:
		controlType = ControlType(string(t))
	default:
		controlType = ControlType("")
	}

	p.Children[0].Description = fmt.Sprintf("Control Type (%v)", controlType)
	criticality = false
	if len(p.Children) == 3 {
		// at least guard against type assertion failure
		criticality, _ = p.Children[1].Value.(bool)
		p.Children[1].Description = "Criticality"
		valuePacket = p.Children[2]
	} else {
		valuePacket = p.Children[1]
	}
	valuePacket.Description = "Control Value"
	return
}

func NewControlString(ControlType ControlType, Criticality bool, ControlValue string) *ControlString {
	return &ControlString{
		ControlType:  ControlType,
		Criticality:  Criticality,
		ControlValue: ControlValue,
	}
}

func encodeControls(Controls []Control) (*ber.Packet, error) {
	p := ber.Encode(ber.ClassContext, ber.TypeConstructed, 0, nil, "Controls")
	for _, control := range Controls {
		pack, err := control.Encode()
		if err != nil {
			return nil, err
		}
		p.AppendChild(pack)
	}
	return p, nil
}

/************************/
/* MatchedValuesRequest */
/************************/

func NewControlPermissiveModifyRequest(criticality bool) *ControlString {
	return NewControlString(ControlTypePermissiveModifyRequest, criticality, "")
}

/***************/
/* ManageDsaITRequest */
/***************/

func NewControlManageDsaITRequest(criticality bool) *ControlString {
	return NewControlString(ControlTypeManageDsaITRequest, criticality, "")
}

/************************/
/* SubtreeDeleteRequest */
/************************/

func NewControlSubtreeDeleteRequest(criticality bool) *ControlString {
	return NewControlString(ControlTypeSubtreeDeleteRequest, criticality, "")
}

/***************/
/* NoOpRequest */
/***************/

func NewControlNoOpRequest() *ControlString {
	return NewControlString(ControlTypeNoOpRequest, true, "")
}

/************************/
/* MatchedValuesRequest */
/************************/

type ControlMatchedValuesRequest struct {
	Criticality bool
	Filter      string
}

func NewControlMatchedValuesRequest(criticality bool, filter string) *ControlMatchedValuesRequest {
	return &ControlMatchedValuesRequest{criticality, filter}
}

func (c *ControlMatchedValuesRequest) Decode(p *ber.Packet) (*Control, error) {
	return nil, newError(ErrorDecoding, "Decode of Control unsupported.")
}

func (c *ControlMatchedValuesRequest) GetControlType() ControlType {
	return ControlTypeMatchedValuesRequest
}

func (c *ControlMatchedValuesRequest) Encode() (p *ber.Packet, err error) {
	p = ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "ControlMatchedValuesRequest")
	p.AppendChild(
		ber.NewString(ber.ClassUniversal, ber.TypePrimitive,
			ber.TagOctetString, string(ControlTypeMatchedValuesRequest),
			fmt.Sprintf("Control Type (%v)", ControlTypeMatchedValuesRequest)))
	if c.Criticality {
		p.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, c.Criticality, "Criticality"))
	}
	octetString := ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, nil, "Octet String")
	simpleFilterSeq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "SimpleFilterItem")
	filterPacket, err := filterParse(c.Filter)
	if err != nil {
		return nil, err
	}
	simpleFilterSeq.AppendChild(filterPacket)
	octetString.AppendChild(simpleFilterSeq)
	p.AppendChild(octetString)
	return p, nil
}

func (c *ControlMatchedValuesRequest) String() string {
	return fmt.Sprintf(
		"Control Type: %s (%q)  Criticality: %t  Filter: %s",
		ControlTypeMatchedValuesRequest.String(),
		string(ControlTypeMatchedValuesRequest),
		c.Criticality,
		c.Filter,
	)
}

/*************************/
/* ServerSideSortRequest */
/*************************/

/*
SortKeyList ::= SEQUENCE OF SEQUENCE {
                 attributeType   AttributeDescription,
                 orderingRule    [0] MatchingRuleId OPTIONAL,
                 reverseOrder    [1] BOOLEAN DEFAULT FALSE }

*/

type ServerSideSortAttrRuleOrder struct {
	AttributeName string
	OrderingRule  string
	ReverseOrder  bool
}

type ControlServerSideSortRequest struct {
	SortKeyList []ServerSideSortAttrRuleOrder
	Criticality bool
}

func NewControlServerSideSortRequest(sortKeyList []ServerSideSortAttrRuleOrder, criticality bool) *ControlServerSideSortRequest {
	return &ControlServerSideSortRequest{sortKeyList, criticality}
}

func (c *ControlServerSideSortRequest) Decode(p *ber.Packet) (*Control, error) {
	return nil, newError(ErrorDecoding, "Decode of Control unsupported.")
}

func (c *ControlServerSideSortRequest) GetControlType() ControlType {
	return ControlTypeServerSideSortRequest
}

func (c *ControlServerSideSortRequest) Encode() (p *ber.Packet, err error) {
	p = ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "ControlServerSideSortRequest")
	p.AppendChild(
		ber.NewString(ber.ClassUniversal, ber.TypePrimitive,
			ber.TagOctetString, string(ControlTypeServerSideSortRequest),
			fmt.Sprintf("Control Type (%v)", ControlTypeServerSideSortRequest)))
	if c.Criticality {
		p.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, c.Criticality, "Criticality"))
	}
	octetString := ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, nil, "Octet String")
	seqSortKeyLists := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "SortKeyLists")

	for _, sortKey := range c.SortKeyList {
		seqKey := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "SortKey")
		seqKey.AppendChild(
			ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, sortKey.AttributeName, "AttributeDescription"),
		)
		if len(sortKey.OrderingRule) > 0 {
			seqKey.AppendChild(
				ber.NewString(ber.ClassUniversal, ber.TypePrimitive, 0, sortKey.OrderingRule, "OrderingRule"),
			)
		}
		seqKey.AppendChild(
			ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, 1, sortKey.ReverseOrder, "ReverseOrder"),
		)
		seqSortKeyLists.AppendChild(seqKey)
	}
	octetString.AppendChild(seqSortKeyLists)
	p.AppendChild(octetString)
	return p, nil
}

func (c *ControlServerSideSortRequest) String() string {
	ctext := fmt.Sprintf(
		"Control Type: %s (%q)  Criticality: %t, SortKeys: ",
		ControlTypeServerSideSortRequest.String(),
		string(ControlTypeServerSideSortRequest),
		c.Criticality,
	)
	for _, sortKey := range c.SortKeyList {
		ctext += fmt.Sprintf("[%s,%s,%t]", sortKey.AttributeName, sortKey.OrderingRule, sortKey.ReverseOrder)
	}
	return ctext
}

/*************************/
/* VlvRequest */
/*************************/

var VlvDebug bool

type VlvOffSet struct {
	Offset       int32
	ContentCount int32
}

/*
  VirtualListViewRequest ::= SEQUENCE {
       beforeCount    INTEGER (0..maxInt),
       afterCount     INTEGER (0..maxInt),
       target       CHOICE {
                      byOffset        [0] SEQUENCE {
                           offset          INTEGER (1 .. maxInt),
                           contentCount    INTEGER (0 .. maxInt) },
                      greaterThanOrEqual [1] AssertionValue },
       contextID     OCTET STRING OPTIONAL }
*/
type ControlVlvRequest struct {
	Criticality        bool
	BeforeCount        int32
	AfterCount         int32
	ByOffset           *VlvOffSet
	GreaterThanOrEqual string
	ContextID          []byte
}

func (c *ControlVlvRequest) Encode() (*ber.Packet, error) {
	p := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "ControlVlvRequest")
	p.AppendChild(
		ber.NewString(ber.ClassUniversal, ber.TypePrimitive,
			ber.TagOctetString, string(ControlTypeVlvRequest),
			fmt.Sprintf("Control Type (%v)", ControlTypeVlvRequest)))
	if c.Criticality {
		p.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, c.Criticality, "Criticality"))
	}
	octetString := ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, nil, "Octet String")

	vlvSeq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "VirtualListViewRequest")
	beforeCount := ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, uint64(c.BeforeCount), "BeforeCount")
	afterCount := ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, uint64(c.AfterCount), "AfterCount")
	var target *ber.Packet
	switch {
	case c.ByOffset != nil:
		target = ber.Encode(ber.ClassContext, ber.TypeConstructed, 0, nil, "ByOffset")
		offset := ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, uint64(c.ByOffset.Offset), "Offset")
		contentCount := ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, uint64(c.ByOffset.ContentCount), "ContentCount")
		target.AppendChild(offset)
		target.AppendChild(contentCount)
	case len(c.GreaterThanOrEqual) > 0:
		// TODO incorrect for some values, binary?
		target = ber.NewString(ber.ClassContext, ber.TypePrimitive, 1, c.GreaterThanOrEqual, "GreaterThanOrEqual")
	}
	if target == nil {
		return nil, newError(ErrorEncoding, "VLV target equal to nil")
	}
	vlvSeq.AppendChild(beforeCount)
	vlvSeq.AppendChild(afterCount)
	vlvSeq.AppendChild(target)

	if len(c.ContextID) > 0 {
		contextID := ber.NewString(ber.ClassUniversal, ber.TypePrimitive,
			ber.TagOctetString, string(c.ContextID), "ContextID")
		vlvSeq.AppendChild(contextID)
	}

	octetString.AppendChild(vlvSeq)
	p.AppendChild(octetString)

	if VlvDebug {
		ber.PrintPacket(p)
	}

	return p, nil

}

func (c *ControlVlvRequest) GetControlType() string {
	return ControlTypeVlvRequest.String()
}

func (c *ControlVlvRequest) String() string {
	ctext := fmt.Sprintf(
		"Control Type: %s (%q)  Criticality: %t, BeforeCount: %d, AfterCount: %d"+
			", ByOffset.Offset: %d, ByOffset.ContentCount: %d, GreaterThanOrEqual: %s",
		ControlTypeVlvRequest.String(),
		string(ControlTypeVlvRequest),
		c.Criticality, c.BeforeCount, c.AfterCount, c.ByOffset.Offset,
		c.ByOffset.ContentCount, c.GreaterThanOrEqual,
	)
	return ctext
}

/***********************************/
/*      RESPONSE CONTROLS          */
/***********************************/

/**************************/
/* ServerSideSortResponse */
/**************************/

type ControlServerSideSortResponse struct {
	AttributeName string // Optional
	Criticality   bool
	Err           error
}

//SortResult ::= SEQUENCE {
//   sortResult  ENUMERATED {
//       success                   (0), -- results are sorted
//       operationsError           (1), -- server internal failure
//       timeLimitExceeded         (3), -- timelimit reached before
//                                      -- sorting was completed
//       strongAuthRequired        (8), -- refused to return sorted
//                                      -- results via insecure
//                                      -- protocol
//       adminLimitExceeded       (11), -- too many matching entries
//                                      -- for the server to sort
//       noSuchAttribute          (16), -- unrecognized attribute
//                                      -- type in sort key
//       inappropriateMatching    (18), -- unrecognized or
//                                      -- inappropriate matching
//                                      -- rule in sort key
//       insufficientAccessRights (50), -- refused to return sorted
//                                      -- results to this client
//       busy                     (51), -- too busy to process
//       unwillingToPerform       (53), -- unable to sort
//       other                    (80)
//       },
//   attributeType [0] AttributeDescription OPTIONAL }
func NewControlServerSideSortResponse(p *ber.Packet) (Control, error) {
	c := new(ControlServerSideSortResponse)
	_, criticality, value := decodeControlTypeAndCrit(p)
	c.Criticality = criticality

	if value.Value != nil {
		sortResult := ber.DecodePacket(value.Data.Bytes())
		value.Data.Truncate(0)
		value.Value = nil
		value.AppendChild(sortResult)
	}

	value = value.Children[0]
	value.Description = "ServerSideSortResponse Control Value"

	value.Children[0].Description = "SortResult"
	errNum, ok := value.Children[0].Value.(ResultCode)
	if !ok {
		return c, errors.New(fmt.Sprintf("type assertion ResultCode for %v failed!", p.Children[0].Value))
	}
	c.Err = newError(errNum, "")

	if len(value.Children) == 2 {
		value.Children[1].Description = "Attribute Name"

		// FIXME: this is hacky, but like the original implementation in the asn1-ber packet previously used
		switch t := value.Children[1].Value.(type) {
		case string:
			c.AttributeName = t
		case []byte:
			c.AttributeName = string(t)
		default:
			c.AttributeName = ""
		}

		value.Children[1].Value = c.AttributeName
	}
	return c, nil
}

func (c *ControlServerSideSortResponse) Encode() (p *ber.Packet, err error) {
	return nil, newError(ErrorEncoding, "Encode of Control unsupported.")
}

func (c *ControlServerSideSortResponse) GetControlType() ControlType {
	return ControlTypeServerSideSortResponse
}

func (c *ControlServerSideSortResponse) String() string {
	err, ok := c.Err.(*Error)
	if !ok {
		err = &Error{}
	}
	return fmt.Sprintf("Control Type: %s (%q)  Criticality: %t, AttributeName: %s, ErrorValue: %d",
		ControlTypeServerSideSortResponse.String(),
		string(ControlTypeServerSideSortResponse),
		c.Criticality,
		c.AttributeName,
		err.ResultCode,
	)
}

/***************/
/* VlvResponse */
/***************/

type ControlVlvResponse struct {
	Criticality    bool
	TargetPosition uint64
	ContentCount   uint64
	Err            error // VirtualListViewResult
	ContextID      string
}

/*
 VirtualListViewResponse ::= SEQUENCE {
       targetPosition    INTEGER (0 .. maxInt),
       contentCount     INTEGER (0 .. maxInt),
       virtualListViewResult ENUMERATED {
            success (0),
            operationsError (1),
            protocolError (3),
            unwillingToPerform (53),
            insufficientAccessRights (50),
            timeLimitExceeded (3),
            adminLimitExceeded (11),
            innapropriateMatching (18),
            sortControlMissing (60),
            offsetRangeError (61),
            other(80),
            ... },
       contextID     OCTET STRING OPTIONAL }
*/
func NewControlVlvResponse(p *ber.Packet) (Control, error) {
	c := new(ControlVlvResponse)
	_, criticality, value := decodeControlTypeAndCrit(p)
	c.Criticality = criticality

	if value.Value != nil {
		vlvResult := ber.DecodePacket(value.Data.Bytes())
		value.Data.Truncate(0)
		value.Value = nil
		value.AppendChild(vlvResult)
	}

	value = value.Children[0]
	value.Description = "VlvResponse Control Value"

	value.Children[0].Description = "TargetPosition"
	value.Children[1].Description = "ContentCount"
	value.Children[2].Description = "VirtualListViewResult/Err"

	var ok bool
	c.TargetPosition, ok = value.Children[0].Value.(uint64)
	if !ok {
		return c, errors.New(fmt.Sprintf("type assertion uint64 for %v failed!", p.Children[0].Value))
	}
	c.ContentCount, ok = value.Children[1].Value.(uint64)
	if !ok {
		return c, errors.New(fmt.Sprintf("type assertion uint64 for %v failed!", p.Children[1].Value))
	}

	errNum, ok := value.Children[2].Value.(ResultCode)
	if !ok {
		log.Println("type assertion failed in control.go")
		errNum = 212
	}
	c.Err = newError(errNum, "")

	if len(value.Children) == 4 {
		value.Children[3].Description = "ContextID"

		// FIXME: this is hacky, but like the original implementation in the asn1-ber packet previously used
		switch t := value.Children[3].Value.(type) {
		case string:
			c.ContextID = t
		case []byte:
			c.ContextID = string(t)
		default:
			c.ContextID = ""
		}
	}

	return c, nil
}

func (c *ControlVlvResponse) Encode() (p *ber.Packet, err error) {
	return nil, newError(ErrorEncoding, "Encode of Control unsupported.")
}

func (c *ControlVlvResponse) GetControlType() ControlType {
	return ControlTypeVlvResponse
}

func (c *ControlVlvResponse) String() string {
	err, ok := c.Err.(*Error)
	if !ok {
		err = &Error{}
	}
	return fmt.Sprintf("Control Type: %s (%q)  Criticality: %t, TargetPosition: %d, ContentCount: %d, ErrorValue: %d, ContextID: %s",
		ControlTypeVlvResponse,
		string(ControlTypeVlvResponse),
		c.Criticality,
		c.TargetPosition,
		c.ContentCount,
		err.ResultCode,
		c.ContextID,
	)
}
