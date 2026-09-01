if r.ko.Spec.Resource != nil && r.ko.Spec.Resource.Database != nil {
	input.Resource = &svcsdktypes.Resource{
		Database: &svcsdktypes.DatabaseResource{
			CatalogId: r.ko.Spec.Resource.Database.CatalogID,
			Name:      r.ko.Spec.Resource.Database.Name,
		},
	}
}
if r.ko.Spec.Principal != nil {
	input.Principal = &svcsdktypes.DataLakePrincipal{
		DataLakePrincipalIdentifier: r.ko.Spec.Principal.DataLakePrincipalIdentifier,
	}
}
