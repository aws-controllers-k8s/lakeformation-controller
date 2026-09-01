// Resource-only filter above can return rows for other principals/shares;
// match the right one here instead of trusting the first row.
principalARN := ""
if r.ko.Spec.Principal != nil && r.ko.Spec.Principal.DataLakePrincipalIdentifier != nil {
	principalARN = *r.ko.Spec.Principal.DataLakePrincipalIdentifier
}
if !rm.matchAndApplyPermissions(ko, r.ko.Spec.Principal, r.ko.Spec.Resource, resp.PrincipalResourcePermissions, principalARN) {
	return nil, ackerr.NotFound
}
