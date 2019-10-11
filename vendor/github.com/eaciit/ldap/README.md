# LDAPv3 client package in pure Go

## Implemented functionality
- Connecting and binding to a LDAP server
- Search / Modify / Add / Delete requests
- Password modify request (RFC3062)
- Compare request
- Search filter compiling
- Request Controls (MatchedValuesRequest, PermissiveModifyRequest, ManageDsaITRequest, SubtreeDeleteRequest, Paging, ServerSideSort)

## Plans
- Real tests against a LDAP server
- I still have to decide what to do with things I will supposedly never touch or use, like the ldif writing/reading functionality
- More cleaning
- Modify DN request
- Own type for DNs with methods for modification and escaping (like the escape_dn_chars function of the python ldap module)
- Binary Attributes (there is another fork which implemented this, I think)

## Licence
The licence used before this fork was copied from the Go sources.
As I am not with Google, I added a 2-clause BSD license for this fork.