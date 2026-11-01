package booking

import (
	"testing"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
)

func TestCancellationProvenanceUsesActorRoleAtCancellation(t *testing.T) {
	tests := []struct {
		role  identity.Role
		want  CancellationSource
		valid bool
	}{
		{identity.RoleMember, CancellationSourceMember, true},
		{identity.RolePT, CancellationSourcePT, true},
		{identity.RoleAdmin, CancellationSourceFacility, true},
		{identity.RoleManager, CancellationSourceFacility, true},
		{identity.RoleReceptionist, CancellationSourceFacility, true},
		{identity.Role("UNKNOWN"), "", false},
	}
	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			got, valid := CancellationProvenance(tt.role)
			if got != tt.want || valid != tt.valid {
				t.Fatalf("CancellationProvenance(%q) = (%q, %t), want (%q, %t)", tt.role, got, valid, tt.want, tt.valid)
			}
		})
	}
}
