// The ListPermissions API can return more than the single grant this CR
// manages: entries derived from LF-Tag policies or RAM resource shares show
// up as "effective" permissions alongside explicitly granted ones, and the
// server-side Resource/Principal filter above narrows but does not
// guarantee an exact match. Re-scan the raw response here rather than
// trusting whichever entry the generated code above happened to pick first.
matched := false
for _, elem := range resp.PrincipalResourcePermissions {
	if elem.AdditionalDetails != nil && len(elem.AdditionalDetails.ResourceShare) > 0 {
		// Share-derived, not a grant this CR made.
		continue
	}
	if elem.Principal == nil || elem.Principal.DataLakePrincipalIdentifier == nil {
		continue
	}
	if r.ko.Spec.Principal == nil || r.ko.Spec.Principal.DataLakePrincipalIdentifier == nil {
		continue
	}
	if *elem.Principal.DataLakePrincipalIdentifier != *r.ko.Spec.Principal.DataLakePrincipalIdentifier {
		continue
	}

	permissions := make([]*string, 0, len(elem.Permissions))
	for _, p := range elem.Permissions {
		permissions = append(permissions, aws.String(string(p)))
	}
	ko.Spec.Permissions = permissions

	permissionsWithGrantOption := make([]*string, 0, len(elem.PermissionsWithGrantOption))
	for _, p := range elem.PermissionsWithGrantOption {
		permissionsWithGrantOption = append(permissionsWithGrantOption, aws.String(string(p)))
	}
	ko.Spec.PermissionsWithGrantOption = permissionsWithGrantOption

	matched = true
	break
}
if !matched {
	return nil, ackerr.NotFound
}
