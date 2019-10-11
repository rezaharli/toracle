package ldap

import (
	"fmt"
	"github.com/eaciit/asn1-ber"
)

type Mod struct {
	ModOperation ModificationCode
	Modification EntryAttribute
}

// LDAP modify request [https://tools.ietf.org/html/rfc4511#section-4.6]
type ModifyRequest struct {
	// DN of entry that is modified
	DN string

	// Changes
	Mods []Mod

	// Server controls
	Controls []Control
}

func (l *Connection) Modify(modReq *ModifyRequest) error {
	messageID, ok := l.nextMessageID()
	if !ok {
		return newError(ErrorClosing, "MessageID channel is closed.")
	}
	encodedModify := encodeModifyRequest(modReq)

	packet, err := requestBuildPacket(messageID, encodedModify, modReq.Controls)
	if err != nil {
		return err
	}

	return l.sendReqRespPacket(messageID, packet)
}

func (req *ModifyRequest) Bytes() []byte {
	return encodeModifyRequest(req).Bytes()
}

func encodeModifyRequest(req *ModifyRequest) (p *ber.Packet) {
	modpacket := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ber.Tag(ApplicationModifyRequest), nil, ApplicationModifyRequest.String())
	modpacket.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, req.DN, "LDAP DN"))
	seqOfChanges := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Changes")
	for _, mod := range req.Mods {
		modification := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Modification")
		op := ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, uint64(mod.ModOperation), "Modify Op")
		modification.AppendChild(op)
		partAttr := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "PartialAttribute")

		partAttr.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, mod.Modification.Name, "AttributeDescription"))
		valuesSet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSet, nil, "Attribute Value Set")
		for _, val := range mod.Modification.Values {
			value := ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, val, "AttributeValue")
			valuesSet.AppendChild(value)
		}
		partAttr.AppendChild(valuesSet)
		modification.AppendChild(partAttr)
		seqOfChanges.AppendChild(modification)
	}
	modpacket.AppendChild(seqOfChanges)

	return modpacket
}

func NewModifyRequest(dn string) (req *ModifyRequest) {
	req = &ModifyRequest{
		DN:       dn,
		Mods:     make([]Mod, 0),
		Controls: make([]Control, 0),
	}
	return
}

// Basic LDIF dump, no formating, etc
func (req *ModifyRequest) String() (dump string) {
	dump = fmt.Sprintf("dn: %s\n", req.DN)
	dump += fmt.Sprintf("changetype: modify\n")
	for _, mod := range req.Mods {
		dump += mod.DumpMod()
	}
	return
}

// Basic LDIF dump, no formating, etc
func (mod *Mod) DumpMod() (dump string) {
	dump += fmt.Sprintf("%s: %s\n", mod.ModOperation.String(), mod.Modification.Name)
	for _, val := range mod.Modification.Values {
		dump += fmt.Sprintf("%s: %s\n", mod.Modification.Name, val)
	}
	dump += "-\n"
	return dump
}

func NewMod(modType ModificationCode, attr string, values []string) (mod *Mod) {
	if values == nil {
		values = []string{}
	}
	partEntryAttr := EntryAttribute{Name: attr, Values: values}
	mod = &Mod{ModOperation: modType, Modification: partEntryAttr}
	return
}

func (req *ModifyRequest) AddMod(mod *Mod) {
	req.Mods = append(req.Mods, *mod)
}

func (req *ModifyRequest) AddMods(mods []Mod) {
	req.Mods = append(req.Mods, mods...)
}

func (req *ModifyRequest) AddControl(control Control) {
	if req.Controls == nil {
		req.Controls = make([]Control, 0)
	}
	req.Controls = append(req.Controls, control)
}
