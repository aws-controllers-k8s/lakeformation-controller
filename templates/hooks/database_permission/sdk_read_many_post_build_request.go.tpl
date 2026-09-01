// Intentionally NOT setting input.Principal here. Per the ListPermissions
// API docs: "If both Principal and Resource parameters are provided, the
// response returns effective permissions rather than the explicitly granted
// permissions." Effective permissions fold in anything implied by the
// account's Lake Formation catalog settings - notably the default hybrid
// access mode grant of ALL to the IAM_ALLOWED_PRINCIPALS pseudo-group, which
// then shows up as if it were an explicit grant for whatever principal is
// named in the request, even one that was never explicitly granted
// anything. Filtering by Resource only returns just the real, explicitly
// granted rows; sdk_read_many_post_set_output does the Principal match
// itself against that real data.
if r.ko.Spec.Resource != nil && r.ko.Spec.Resource.Database != nil {
	input.Resource = &svcsdktypes.Resource{
		Database: &svcsdktypes.DatabaseResource{
			CatalogId: r.ko.Spec.Resource.Database.CatalogID,
			Name:      r.ko.Spec.Resource.Database.Name,
		},
	}
}
