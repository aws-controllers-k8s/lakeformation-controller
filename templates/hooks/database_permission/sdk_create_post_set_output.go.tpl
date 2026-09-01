// ACKResourceMetadata already set means this Create is a self-heal
// (sdkFind returned NotFound for a previously-synced CR), not a first grant.
if desired.ko.Status.ACKResourceMetadata != nil {
	rm.setSelfHealAdvisory(ko)
}
