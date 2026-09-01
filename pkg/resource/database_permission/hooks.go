// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package database_permission

import (
	"context"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
)

// customUpdateDatabasePermission handles updates for DatabasePermission
// resources. Lake Formation has no native update API for grants -
// GrantPermissions/RevokePermissions are the only mutating operations, both
// scoped to a (Principal, Resource) tuple. Verified live against the AWS
// API: repeated Grant calls for the same tuple merge into the same
// underlying permission record rather than fragmenting into separate
// records, Revoke only removes the permissions named in that call and
// leaves the rest of the record untouched, and Permissions and
// PermissionsWithGrantOption can each be changed independently of the other
// in a single call (an empty list for one leaves it untouched).
func (rm *resourceManager) customUpdateDatabasePermission(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (*resource, error) {
	ko := desired.ko.DeepCopy()
	rm.setStatusDefaults(ko)

	identityChanged := delta.DifferentAt("Spec.Principal.DataLakePrincipalIdentifier") ||
		delta.DifferentAt("Spec.Resource.Database.CatalogID") ||
		delta.DifferentAt("Spec.Resource.Database.Name")

	if identityChanged {
		// A changed Principal or Database is a different AWS-level grant
		// record entirely, not a mutation of the existing one: revoke
		// everything under the old tuple, then grant everything under the
		// new tuple.
		if err := rm.revokeSubset(ctx, latest, latest.ko.Spec.Permissions, latest.ko.Spec.PermissionsWithGrantOption); err != nil {
			return nil, err
		}
		if err := rm.grantSubset(ctx, desired, desired.ko.Spec.Permissions, desired.ko.Spec.PermissionsWithGrantOption); err != nil {
			return nil, err
		}
		return &resource{ko}, nil
	}

	// Same tuple: diff each list independently and only touch what actually
	// changed. This avoids a window with no access during a purely additive
	// change, which a revoke-then-full-regrant would cause.
	addedPermissions, removedPermissions := permissionsDiff(desired.ko.Spec.Permissions, latest.ko.Spec.Permissions)
	addedGrantable, removedGrantable := permissionsDiff(desired.ko.Spec.PermissionsWithGrantOption, latest.ko.Spec.PermissionsWithGrantOption)

	if err := rm.revokeSubset(ctx, desired, removedPermissions, removedGrantable); err != nil {
		return nil, err
	}
	if err := rm.grantSubset(ctx, desired, addedPermissions, addedGrantable); err != nil {
		return nil, err
	}

	return &resource{ko}, nil
}

// grantSubset calls GrantPermissions for r's identity (Principal/Resource/
// CatalogID/Condition), scoped to just the given permission subsets. A nil
// or empty subset leaves that field untouched at the AWS API level. No-op
// if both subsets are empty.
func (rm *resourceManager) grantSubset(
	ctx context.Context,
	r *resource,
	permissions []*string,
	permissionsWithGrantOption []*string,
) error {
	if len(permissions) == 0 && len(permissionsWithGrantOption) == 0 {
		return nil
	}
	sub := r.ko.DeepCopy()
	sub.Spec.Permissions = permissions
	sub.Spec.PermissionsWithGrantOption = permissionsWithGrantOption
	input, err := rm.newCreateRequestPayload(ctx, &resource{sub})
	if err != nil {
		return err
	}
	_, err = rm.sdkapi.GrantPermissions(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "GrantPermissions", err)
	return err
}

// revokeSubset calls RevokePermissions for r's identity (Principal/Resource/
// CatalogID/Condition), scoped to just the given permission subsets. A nil
// or empty subset leaves that field untouched at the AWS API level. No-op
// if both subsets are empty.
func (rm *resourceManager) revokeSubset(
	ctx context.Context,
	r *resource,
	permissions []*string,
	permissionsWithGrantOption []*string,
) error {
	if len(permissions) == 0 && len(permissionsWithGrantOption) == 0 {
		return nil
	}
	sub := r.ko.DeepCopy()
	sub.Spec.Permissions = permissions
	sub.Spec.PermissionsWithGrantOption = permissionsWithGrantOption
	input, err := rm.newDeleteRequestPayload(&resource{sub})
	if err != nil {
		return err
	}
	_, err = rm.sdkapi.RevokePermissions(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "RevokePermissions", err)
	return err
}

// permissionsDiff compares two []*string permission lists as sets (order
// insensitive) and returns the entries present in desired but not latest
// (added) and present in latest but not desired (removed).
func permissionsDiff(desired, latest []*string) (added, removed []*string) {
	desiredSet := make(map[string]struct{}, len(desired))
	for _, p := range desired {
		if p != nil {
			desiredSet[*p] = struct{}{}
		}
	}
	latestSet := make(map[string]struct{}, len(latest))
	for _, p := range latest {
		if p != nil {
			latestSet[*p] = struct{}{}
		}
	}
	for _, p := range desired {
		if p == nil {
			continue
		}
		if _, ok := latestSet[*p]; !ok {
			added = append(added, p)
		}
	}
	for _, p := range latest {
		if p == nil {
			continue
		}
		if _, ok := desiredSet[*p]; !ok {
			removed = append(removed, p)
		}
	}
	return added, removed
}
