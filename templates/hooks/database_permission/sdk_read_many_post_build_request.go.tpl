// Intentionally NOT setting input.Principal: setting both Principal and
// Resource switches ListPermissions to "effective permissions" mode, which
// folds in implicit grants (e.g. hybrid access mode's default ALL to
// IAM_ALLOWED_PRINCIPALS) as if explicitly granted. Resource-only filtering
// returns real grants; sdk_read_many_post_set_output matches Principal itself.
if r.ko.Spec.Resource != nil && r.ko.Spec.Resource.Database != nil {
	input.Resource = &svcsdktypes.Resource{
		Database: &svcsdktypes.DatabaseResource{
			CatalogId: r.ko.Spec.Resource.Database.CatalogID,
			Name:      r.ko.Spec.Resource.Database.Name,
		},
	}
}
