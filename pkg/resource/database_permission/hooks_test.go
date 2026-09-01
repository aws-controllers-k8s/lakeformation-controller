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
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func strPtrSlice(vals ...string) []*string {
	out := make([]*string, 0, len(vals))
	for _, v := range vals {
		out = append(out, aws.String(v))
	}
	return out
}

func assertSameSet(t *testing.T, got []*string, want []string) {
	t.Helper()
	gotStrs := make([]string, 0, len(got))
	for _, p := range got {
		gotStrs = append(gotStrs, *p)
	}
	sort.Strings(gotStrs)
	wantSorted := append([]string(nil), want...)
	sort.Strings(wantSorted)
	if len(gotStrs) != len(wantSorted) {
		t.Fatalf("got %v, want %v", gotStrs, wantSorted)
	}
	for i := range gotStrs {
		if gotStrs[i] != wantSorted[i] {
			t.Fatalf("got %v, want %v", gotStrs, wantSorted)
		}
	}
}

func TestPermissionsDiff_PureAdd(t *testing.T) {
	desired := strPtrSlice("DESCRIBE", "ALTER")
	latest := strPtrSlice("DESCRIBE")

	added, removed := permissionsDiff(desired, latest)

	assertSameSet(t, added, []string{"ALTER"})
	assertSameSet(t, removed, nil)
}

func TestPermissionsDiff_PureRemove(t *testing.T) {
	desired := strPtrSlice("DESCRIBE")
	latest := strPtrSlice("DESCRIBE", "ALTER")

	added, removed := permissionsDiff(desired, latest)

	assertSameSet(t, added, nil)
	assertSameSet(t, removed, []string{"ALTER"})
}

func TestPermissionsDiff_MixedAddAndRemove(t *testing.T) {
	desired := strPtrSlice("DESCRIBE", "DROP")
	latest := strPtrSlice("DESCRIBE", "ALTER")

	added, removed := permissionsDiff(desired, latest)

	assertSameSet(t, added, []string{"DROP"})
	assertSameSet(t, removed, []string{"ALTER"})
}

func TestPermissionsDiff_NoOp(t *testing.T) {
	desired := strPtrSlice("DESCRIBE", "ALTER")
	latest := strPtrSlice("ALTER", "DESCRIBE")

	added, removed := permissionsDiff(desired, latest)

	assertSameSet(t, added, nil)
	assertSameSet(t, removed, nil)
}

func TestPermissionsDiff_BothEmpty(t *testing.T) {
	added, removed := permissionsDiff(nil, nil)

	assertSameSet(t, added, nil)
	assertSameSet(t, removed, nil)
}

func TestPermissionsDiff_IdentityChangedStyle_FullRevokeAndGrant(t *testing.T) {
	// Simulates the identity-changed path: everything in "latest" should be
	// revoked, everything in "desired" should be granted, with no overlap
	// assumed since it's a different AWS-level grant record.
	desired := strPtrSlice("SELECT", "DESCRIBE")
	latest := strPtrSlice("ALTER", "DROP")

	added, removed := permissionsDiff(desired, latest)

	assertSameSet(t, added, []string{"SELECT", "DESCRIBE"})
	assertSameSet(t, removed, []string{"ALTER", "DROP"})
}
