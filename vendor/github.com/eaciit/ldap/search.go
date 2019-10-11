package ldap

import (
	"errors"
	"fmt"
	"github.com/eaciit/asn1-ber"
	"log"
)

type SearchResult struct {
	Entries   []*Entry
	Referrals []string
	Controls  []Control
}

type DiscreteSearchResult struct {
	SearchResultType SearchResultType
	Entry            *Entry
	Referrals        []string
	Controls         []Control
}

type ConnectionInfo struct {
	Conn      *Connection
	MessageID int64
}

type SearchResultHandler interface {
	ProcessDiscreteResult(*DiscreteSearchResult, *ConnectionInfo) (bool, error)
}

// SearchRequest passed to Search functions.
type SearchRequest struct {
	BaseDN       string
	Scope        Scope
	DerefAliases Deref
	SizeLimit    int
	TimeLimit    int
	TypesOnly    bool
	Filter       string
	Attributes   []string
	Controls     []Control
}

//NewSimpleSearchRequest only requires four parameters and defaults the
//other returned SearchRequest values to typical values...
//
//	DerefAliases: NeverDerefAliases
//	SizeLimit:    0
//	TimeLimit:    0
//	TypesOnly:    false
//	Controls:     nil
func NewSimpleSearchRequest(BaseDN string, Scope Scope, Filter string, Attributes []string) *SearchRequest {
	return &SearchRequest{
		BaseDN:       BaseDN,
		Scope:        Scope,
		DerefAliases: NeverDerefAliases,
		SizeLimit:    0,
		TimeLimit:    0,
		TypesOnly:    false,
		Filter:       Filter,
		Attributes:   Attributes,
		Controls:     nil,
	}
}

func NewSearchRequest(BaseDN string, Scope Scope, DerefAliases Deref, SizeLimit, TimeLimit int, TypesOnly bool, Filter string, Attributes []string, Controls []Control) *SearchRequest {
	return &SearchRequest{
		BaseDN:       BaseDN,
		Scope:        Scope,
		DerefAliases: DerefAliases,
		SizeLimit:    SizeLimit,
		TimeLimit:    TimeLimit,
		TypesOnly:    TypesOnly,
		Filter:       Filter,
		Attributes:   Attributes,
		Controls:     Controls,
	}
}

//SearchWithPaging adds a paging control to the the searchRequest, with a size of pagingSize.
//It combines all the paged results into the returned SearchResult. It is a helper function for
//use with servers that require paging for certain result sizes (AD?).
//
//It is NOT an efficent way to process huge result sets i.e. it doesn't process on a pageSize
//number of entries, it returns the combined result.
func (l *Connection) SearchWithPaging(searchRequest *SearchRequest, pagingSize uint32) (*SearchResult, error) {
	pagingControl := NewControlPaging(pagingSize)
	searchRequest.AddControl(pagingControl)
	allResults := new(SearchResult)

	for i := 0; ; i++ {
		searchResult := new(SearchResult)
		err := l.SearchWithHandler(searchRequest, searchResult, nil)
		if err != nil {
			return allResults, err
		}

		allResults.Entries = append(allResults.Entries, searchResult.Entries...)
		allResults.Referrals = append(allResults.Referrals, searchResult.Referrals...)
		allResults.Controls = append(allResults.Controls, searchResult.Controls...)

		_, pagingResponsePacket := FindControl(searchResult.Controls, ControlTypePaging)
		// If initial result and no paging control then server doesn't support paging
		if pagingResponsePacket == nil && i == 0 {
			if l.Debug {
				fmt.Println("Requested paging but no control returned, control unsupported.")
			}
			return allResults, nil
		} else if pagingResponsePacket == nil {
			return allResults, newError(ErrorMissingControl, "Expected paging Control, it was not found.")
		}
		control, ok := pagingResponsePacket.(*ControlPaging)
		if !ok {
			return allResults, errors.New(fmt.Sprintf("type assertion *ControlPaging for %v failed!", pagingResponsePacket))
		}
		pagingControl.SetCookie(control.Cookie)
		if len(pagingControl.Cookie) == 0 {
			break
		}
	}
	return allResults, nil
}

//ProcessDiscreteResult handles an individual result from a server. Member of the
//SearchResultHandler interface. Results are placed into a SearchResult.
func (sr *SearchResult) ProcessDiscreteResult(dsr *DiscreteSearchResult, connInfo *ConnectionInfo) (stopProcessing bool, err error) {
	switch dsr.SearchResultType {
	case SearchResultEntry:
		sr.Entries = append(sr.Entries, dsr.Entry)
	case SearchResultDone:
		if dsr.Controls != nil {
			sr.Controls = append(sr.Controls, dsr.Controls...)
		}
	case SearchResultReference:
		sr.Referrals = append(sr.Referrals, dsr.Referrals...)
	}
	return false, nil
}

//Search is a blocking search. nil error on success.
func (l *Connection) Search(searchRequest *SearchRequest) (*SearchResult, error) {
	result := &SearchResult{
		Entries:   make([]*Entry, 0),
		Referrals: make([]string, 0),
		Controls:  make([]Control, 0)}

	err := l.SearchWithHandler(searchRequest, result, nil)
	if err != nil {
		return result, err
	}
	return result, nil
}

func encodeSearchRequest(req *SearchRequest) (*ber.Packet, error) {
	searchRequest := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ber.Tag(ApplicationSearchRequest), nil, "Search Request")
	searchRequest.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, req.BaseDN, "Base DN"))
	searchRequest.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, uint64(req.Scope), "Scope"))
	searchRequest.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, uint64(req.DerefAliases), "Deref Aliases"))
	searchRequest.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, uint64(req.SizeLimit), "Size Limit"))
	searchRequest.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, uint64(req.TimeLimit), "Time Limit"))
	searchRequest.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, req.TypesOnly, "Types Only"))
	filterPacket, err := CompileFilter(req.Filter)
	if err != nil {
		return nil, err
	}
	searchRequest.AppendChild(filterPacket)
	attributesPacket := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Attributes")
	for _, attribute := range req.Attributes {
		attributesPacket.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, attribute, "Attribute"))
	}
	searchRequest.AppendChild(attributesPacket)
	return searchRequest, nil
}

//AddControl adds the provided control to a SearchRequest
func (req *SearchRequest) AddControl(control Control) {
	if req.Controls == nil {
		req.Controls = make([]Control, 0)
	}
	req.Controls = append(req.Controls, control)
}

// SearchResult decoded to Entry,Controls,Referral
func decodeSearchResponse(packet *ber.Packet) (discreteSearchResult *DiscreteSearchResult, err error) {
	discreteSearchResult = new(DiscreteSearchResult)
	switch SearchResultType(packet.Children[1].Tag) {
	case SearchResultEntry:
		discreteSearchResult.SearchResultType = SearchResultEntry
		entry := new(Entry)

		var ok bool
		if entry.DN, ok = packet.Children[1].Children[0].Value.(string); !ok {
			return nil, NewValueMismatchError(packet.Children[1].Children[0].Value)
		}

		for _, child := range packet.Children[1].Children[1].Children {
			attr := new(EntryAttribute)

			if attr.Name, ok = child.Children[0].Value.(string); !ok {
				return nil, NewValueMismatchError(child.Children[0].Value)
			}

			for _, value := range child.Children[1].Children {
				if valueString, ok := value.Value.(string); ok {
					attr.Values = append(attr.Values, valueString)
				} else {
					return nil, NewValueMismatchError(value.Value)
				}
			}
			entry.Attributes = append(entry.Attributes, attr)
		}
		discreteSearchResult.Entry = entry
		return discreteSearchResult, nil
	case SearchResultDone:
		discreteSearchResult.SearchResultType = SearchResultDone
		resultCode, result_description := getResultCode(packet)
		if resultCode != 0 {
			return discreteSearchResult, newError(ResultCode(resultCode), result_description)
		}

		if len(packet.Children) == 3 {
			controls := make([]Control, 0)
			for _, child := range packet.Children[2].Children {
				controlOid, ok := child.Children[0].Value.(string)
				if !ok {
					return nil, NewValueMismatchError(child.Children[0].Value)
				}

				decodeFunc, err := ControlType(controlOid).function()

				if err != nil {
					log.Println("Couldn't decode Control : " + controlOid)
				} else {
					c, _ := decodeFunc(child)
					controls = append(controls, c)
				}
			}
			discreteSearchResult.Controls = controls
		}
		return discreteSearchResult, nil
	case SearchResultReference:
		discreteSearchResult.SearchResultType = SearchResultReference
		for ref := range packet.Children[1].Children {
			if refString, ok := packet.Children[1].Children[ref].Value.(string); ok {
				discreteSearchResult.Referrals = append(discreteSearchResult.Referrals, refString)
			} else {
				return nil, NewValueMismatchError(packet.Children[1].Children[ref].Value)
			}
		}
		return discreteSearchResult, nil
	}
	return nil, newError(ErrorDecoding, "Couldn't decode search result.")
}

func sendError(errChannel chan<- error, err error) error {
	if errChannel != nil {
		go func() {
			errChannel <- err
		}()
	}
	return err
}

//SearchWithHandler is the workhorse. Sends requests, decodes results and passes
//on to SearchResultHandlers to process.
//	SearchResultHandler, an interface, implemeneted by SearchResult.
//	Handles the discreteSearchResults. Can provide own implemented to work on
//	a result by result basis.
//	errorChan - if nil then blocking, else error returned via channel upon completion.
//	returns error if blocking.
func (l *Connection) SearchWithHandler(
	searchRequest *SearchRequest, resultHandler SearchResultHandler, errorChan chan<- error,
) error {
	messageID, ok := l.nextMessageID()
	if !ok {
		err := newError(ErrorClosing, "MessageID channel is closed.")
		return sendError(errorChan, err)
	}

	searchPacket, err := encodeSearchRequest(searchRequest)

	if err != nil {
		return sendError(errorChan, err)
	}

	packet, err := requestBuildPacket(messageID, searchPacket, searchRequest.Controls)
	if err != nil {
		return sendError(errorChan, err)
	}

	if l.Debug {
		ber.PrintPacket(packet)
	}

	channel, err := l.sendMessage(packet)

	if err != nil {
		return sendError(errorChan, err)
	}
	if channel == nil {
		err = newError(ErrorNetwork, "Could not send message")
		return sendError(errorChan, err)
	}
	defer l.finishMessage(messageID)

	connectionInfo := &ConnectionInfo{
		Conn:      l,
		MessageID: messageID,
	}

	for {
		if l.Debug {
			fmt.Printf("%d: waiting for response\n", messageID)
		}
		packet, ok = <-channel

		if l.Debug {
			fmt.Printf("%d: got response %p, %v\n", messageID, packet, ok)
		}

		if !ok {
			return newError(ErrorClosing, "Response Channel Closed")
		}

		if packet == nil {
			err = newError(ErrorNetwork, "Could not retrieve message")
			return sendError(errorChan, err)
		}

		if l.Debug {
			if err := addLDAPDescriptions(packet); err != nil {
				return sendError(errorChan, err)
			}
			ber.PrintPacket(packet)
		}

		discreteSearchResult, err := decodeSearchResponse(packet)

		if err != nil {
			return sendError(errorChan, err)
		}

		stop, err := resultHandler.ProcessDiscreteResult(discreteSearchResult, connectionInfo)
		if err != nil {
			return sendError(errorChan, err)
		}

		if discreteSearchResult.SearchResultType == SearchResultDone || stop {
			break
		}
	}
	return sendError(errorChan, nil)
}

func (sr *SearchResult) String() (dump string) {
	for _, entry := range sr.Entries {
		dump = fmt.Sprint(entry)
	}
	return
}
