// A Create call for a DatabasePermission that already has ACK resource
// metadata in its Status is not a first-time grant: it means the CR
// previously synced successfully but sdkFind just returned NotFound, so the
// generic reconciler is self-healing by recreating it. The grant may have
// vanished because another DatabasePermission CR (or something outside ACK)
// revoked it - surface that possibility as an advisory condition rather than
// silently recreating it, since there's no cluster-wide guarantee today that
// two DatabasePermission CRs can't target the same principal/database.
if desired.ko.Status.ACKResourceMetadata != nil {
	ackcondition.SetAdvisory(
		&resource{ko},
		corev1.ConditionTrue,
		aws.String("This grant was not found in AWS and has been recreated. If more than one DatabasePermission resource targets the same principal and database, one of them may be revoking permissions the other manages."),
		aws.String("SELF_HEALED_RECREATE"),
	)
}
