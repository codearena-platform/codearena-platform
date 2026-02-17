package user

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
)

// TestRemoveRoleFromUserObj tests the removeRoleFromUserObj helper function
func TestRemoveRoleFromUserObj(t *testing.T) {
	t.Run("success - remove role and its resources", func(t *testing.T) {
		// Create test user with roles and resources
		user := &sdk.User{
			Id: "user-123",
			Roles: map[string]sdk.UserRole{
				"role-1": {Id: "role-1", Name: "Role One"},
				"role-2": {Id: "role-2", Name: "Role Two"},
			},
			Resources: map[string]sdk.UserResource{
				"resource-1": {
					RoleIds:   map[string]bool{"role-1": true, "role-2": true},
					PolicyIds: map[string]bool{},
					Key:       "resource-1",
					Name:      "Resource One",
				},
				"resource-2": {
					RoleIds:   map[string]bool{"role-1": true},
					PolicyIds: map[string]bool{},
					Key:       "resource-2",
					Name:      "Resource Two",
				},
			},
		}

		roleToRemove := sdk.Role{
			Id:   "role-1",
			Name: "Role One",
			Resources: map[string]sdk.Resources{
				"resource-1": {Key: "resource-1", Name: "Resource One"},
				"resource-2": {Key: "resource-2", Name: "Resource Two"},
			},
		}

		// Remove the role
		removeRoleFromUserObj(user, roleToRemove)

		// Verify role is removed
		assert.NotContains(t, user.Roles, "role-1")
		assert.Contains(t, user.Roles, "role-2")

		// Verify resource-1 still exists (because role-2 still needs it)
		assert.Contains(t, user.Resources, "resource-1")
		assert.NotContains(t, user.Resources["resource-1"].RoleIds, "role-1")
		assert.Contains(t, user.Resources["resource-1"].RoleIds, "role-2")

		// Verify resource-2 is removed (no roles or policies need it)
		assert.NotContains(t, user.Resources, "resource-2")
	})

	t.Run("success - remove role and keep resource with policies", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Roles: map[string]sdk.UserRole{
				"role-1": {Id: "role-1", Name: "Role One"},
			},
			Resources: map[string]sdk.UserResource{
				"resource-1": {
					RoleIds:   map[string]bool{"role-1": true},
					PolicyIds: map[string]bool{"policy-1": true},
					Key:       "resource-1",
					Name:      "Resource One",
				},
			},
		}

		roleToRemove := sdk.Role{
			Id:   "role-1",
			Name: "Role One",
			Resources: map[string]sdk.Resources{
				"resource-1": {Key: "resource-1", Name: "Resource One"},
			},
		}

		removeRoleFromUserObj(user, roleToRemove)

		// Verify role is removed
		assert.NotContains(t, user.Roles, "role-1")

		// Verify resource-1 still exists (because policy-1 still needs it)
		assert.Contains(t, user.Resources, "resource-1")
		assert.NotContains(t, user.Resources["resource-1"].RoleIds, "role-1")
		assert.Contains(t, user.Resources["resource-1"].PolicyIds, "policy-1")
	})

	t.Run("success - remove role with no matching resources", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Roles: map[string]sdk.UserRole{
				"role-1": {Id: "role-1", Name: "Role One"},
			},
			Resources: map[string]sdk.UserResource{
				"resource-other": {
					RoleIds:   map[string]bool{"role-other": true},
					PolicyIds: map[string]bool{},
					Key:       "resource-other",
					Name:      "Other Resource",
				},
			},
		}

		roleToRemove := sdk.Role{
			Id:   "role-1",
			Name: "Role One",
			Resources: map[string]sdk.Resources{
				"resource-nonexistent": {Key: "resource-nonexistent", Name: "Non-existent Resource"},
			},
		}

		removeRoleFromUserObj(user, roleToRemove)

		// Verify role is removed
		assert.NotContains(t, user.Roles, "role-1")

		// Verify other resources are unchanged
		assert.Contains(t, user.Resources, "resource-other")
		assert.Contains(t, user.Resources["resource-other"].RoleIds, "role-other")
	})

	t.Run("success - handle nil user fields", func(t *testing.T) {
		user := &sdk.User{
			Id:        "user-123",
			Roles:     nil,
			Resources: nil,
		}

		roleToRemove := sdk.Role{
			Id:   "role-1",
			Name: "Role One",
			Resources: map[string]sdk.Resources{
				"resource-1": {Key: "resource-1", Name: "Resource One"},
			},
		}

		// Should not panic
		assert.NotPanics(t, func() {
			removeRoleFromUserObj(user, roleToRemove)
		})

		// Verify fields are initialized
		assert.NotNil(t, user.Roles)
		assert.NotNil(t, user.Resources)
	})

	t.Run("success - remove role with empty resources", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Roles: map[string]sdk.UserRole{
				"role-1": {Id: "role-1", Name: "Role One"},
			},
			Resources: map[string]sdk.UserResource{},
		}

		roleToRemove := sdk.Role{
			Id:        "role-1",
			Name:      "Role One",
			Resources: map[string]sdk.Resources{}, // Empty resources
		}

		removeRoleFromUserObj(user, roleToRemove)

		// Verify role is removed
		assert.NotContains(t, user.Roles, "role-1")
		assert.Empty(t, user.Resources)
	})
}

// TestAddRoleToUserObj tests the addRoleToUserObj helper function
func TestAddRoleToUserObj(t *testing.T) {
	t.Run("success - add role to user with nil fields", func(t *testing.T) {
		user := &sdk.User{
			Id:        "user-123",
			Roles:     nil,
			Resources: nil,
		}

		roleToAdd := sdk.Role{
			Id:   "role-1",
			Name: "Role One",
			Resources: map[string]sdk.Resources{
				"resource-1": {Key: "resource-1", Name: "Resource One"},
				"resource-2": {Key: "resource-2", Name: "Resource Two"},
			},
		}

		addRoleToUserObj(user, roleToAdd)

		// Verify fields are initialized and role is added
		assert.NotNil(t, user.Roles)
		assert.NotNil(t, user.Resources)
		assert.Contains(t, user.Roles, "role-1")
		assert.Equal(t, "role-1", user.Roles["role-1"].Id)
		assert.Equal(t, "Role One", user.Roles["role-1"].Name)

		// Verify resources are added
		assert.Len(t, user.Resources, 2)
		assert.Contains(t, user.Resources, "resource-1")
		assert.Contains(t, user.Resources, "resource-2")

		// Verify resource-1 details
		res1 := user.Resources["resource-1"]
		assert.Equal(t, "resource-1", res1.Key)
		assert.Equal(t, "Resource One", res1.Name)
		assert.Contains(t, res1.RoleIds, "role-1")
		assert.True(t, res1.RoleIds["role-1"])

		// Verify resource-2 details
		res2 := user.Resources["resource-2"]
		assert.Equal(t, "resource-2", res2.Key)
		assert.Equal(t, "Resource Two", res2.Name)
		assert.Contains(t, res2.RoleIds, "role-1")
		assert.True(t, res2.RoleIds["role-1"])
	})

	t.Run("success - add role to user with existing roles and resources", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Roles: map[string]sdk.UserRole{
				"existing-role": {Id: "existing-role", Name: "Existing Role"},
			},
			Resources: map[string]sdk.UserResource{
				"existing-resource": {
					RoleIds:   map[string]bool{"existing-role": true},
					PolicyIds: map[string]bool{"policy-1": true},
					Key:       "existing-resource",
					Name:      "Existing Resource",
				},
			},
		}

		roleToAdd := sdk.Role{
			Id:   "role-1",
			Name: "Role One",
			Resources: map[string]sdk.Resources{
				"resource-1": {Key: "resource-1", Name: "Resource One"},
			},
		}

		addRoleToUserObj(user, roleToAdd)

		// Verify existing role is preserved
		assert.Contains(t, user.Roles, "existing-role")
		assert.Contains(t, user.Roles, "role-1")
		assert.Len(t, user.Roles, 2)

		// Verify new role is added
		assert.Equal(t, "role-1", user.Roles["role-1"].Id)
		assert.Equal(t, "Role One", user.Roles["role-1"].Name)

		// Verify existing resource is preserved
		assert.Contains(t, user.Resources, "existing-resource")
		assert.Contains(t, user.Resources, "resource-1")
		assert.Len(t, user.Resources, 2)

		// Verify new resource is added
		res1 := user.Resources["resource-1"]
		assert.Equal(t, "resource-1", res1.Key)
		assert.Equal(t, "Resource One", res1.Name)
		assert.Contains(t, res1.RoleIds, "role-1")
	})

	t.Run("success - add role with existing resource (merge)", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Roles: map[string]sdk.UserRole{
				"existing-role": {Id: "existing-role", Name: "Existing Role"},
			},
			Resources: map[string]sdk.UserResource{
				"shared-resource": {
					RoleIds:   map[string]bool{"existing-role": true},
					PolicyIds: map[string]bool{"policy-1": true},
					Key:       "shared-resource",
					Name:      "Shared Resource",
				},
			},
		}

		roleToAdd := sdk.Role{
			Id:   "role-1",
			Name: "Role One",
			Resources: map[string]sdk.Resources{
				"shared-resource": {Key: "shared-resource", Name: "Shared Resource"},
			},
		}

		addRoleToUserObj(user, roleToAdd)

		// Verify role is added
		assert.Contains(t, user.Roles, "role-1")
		assert.Len(t, user.Roles, 2)

		// Verify shared resource is merged, not replaced
		assert.Len(t, user.Resources, 1)
		sharedRes := user.Resources["shared-resource"]
		assert.Equal(t, "shared-resource", sharedRes.Key)
		assert.Equal(t, "Shared Resource", sharedRes.Name)

		// Should contain both role IDs
		assert.Contains(t, sharedRes.RoleIds, "existing-role")
		assert.Contains(t, sharedRes.RoleIds, "role-1")
		assert.True(t, sharedRes.RoleIds["existing-role"])
		assert.True(t, sharedRes.RoleIds["role-1"])

		// Should preserve existing policy IDs
		assert.Contains(t, sharedRes.PolicyIds, "policy-1")
	})

	t.Run("success - add role with resource having empty RoleIds", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Roles: map[string]sdk.UserRole{
				"existing-role": {Id: "existing-role", Name: "Existing Role"},
			},
			Resources: map[string]sdk.UserResource{
				"resource-1": {
					RoleIds:   map[string]bool{}, // Empty role IDs
					PolicyIds: map[string]bool{"policy-1": true},
					Key:       "resource-1",
					Name:      "Resource One",
				},
			},
		}

		roleToAdd := sdk.Role{
			Id:   "role-1",
			Name: "Role One",
			Resources: map[string]sdk.Resources{
				"resource-1": {Key: "resource-1", Name: "Resource One"},
			},
		}

		addRoleToUserObj(user, roleToAdd)

		// Verify role is added
		assert.Contains(t, user.Roles, "role-1")

		// Verify resource role IDs are properly initialized and role is added
		res1 := user.Resources["resource-1"]
		assert.NotNil(t, res1.RoleIds)
		assert.Contains(t, res1.RoleIds, "role-1")
		assert.True(t, res1.RoleIds["role-1"])

		// Preserve existing policy IDs
		assert.Contains(t, res1.PolicyIds, "policy-1")
	})

	t.Run("success - add role with no resources", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Roles: map[string]sdk.UserRole{
				"existing-role": {Id: "existing-role", Name: "Existing Role"},
			},
			Resources: map[string]sdk.UserResource{
				"existing-resource": {
					RoleIds: map[string]bool{"existing-role": true},
					Key:     "existing-resource",
					Name:    "Existing Resource",
				},
			},
		}

		roleToAdd := sdk.Role{
			Id:        "role-1",
			Name:      "Role One",
			Resources: map[string]sdk.Resources{}, // No resources
		}

		addRoleToUserObj(user, roleToAdd)

		// Verify role is added
		assert.Contains(t, user.Roles, "role-1")
		assert.Equal(t, "role-1", user.Roles["role-1"].Id)
		assert.Equal(t, "Role One", user.Roles["role-1"].Name)

		// Verify existing resources are unchanged
		assert.Len(t, user.Resources, 1)
		assert.Contains(t, user.Resources, "existing-resource")
		existingRes := user.Resources["existing-resource"]
		assert.Contains(t, existingRes.RoleIds, "existing-role")
		assert.NotContains(t, existingRes.RoleIds, "role-1") // Should not be added
	})

	t.Run("success - add role with nil resource RoleIds", func(t *testing.T) {
		user := &sdk.User{
			Id:    "user-123",
			Roles: map[string]sdk.UserRole{},
			Resources: map[string]sdk.UserResource{
				"resource-1": {
					RoleIds:   nil, // Nil role IDs
					PolicyIds: map[string]bool{"policy-1": true},
					Key:       "resource-1",
					Name:      "Resource One",
				},
			},
		}

		roleToAdd := sdk.Role{
			Id:   "role-1",
			Name: "Role One",
			Resources: map[string]sdk.Resources{
				"resource-1": {Key: "resource-1", Name: "Resource One"},
			},
		}

		addRoleToUserObj(user, roleToAdd)

		// Verify role is added
		assert.Contains(t, user.Roles, "role-1")

		// Verify nil RoleIds are properly initialized
		res1 := user.Resources["resource-1"]
		assert.NotNil(t, res1.RoleIds)
		assert.Contains(t, res1.RoleIds, "role-1")
		assert.True(t, res1.RoleIds["role-1"])
	})
}

// TestAddResourceToUserObj tests the addResourceToUserObj helper function
func TestAddResourceToUserObj(t *testing.T) {
	t.Run("success - add resource to user with nil resources", func(t *testing.T) {
		user := &sdk.User{
			Id:        "user-123",
			Resources: nil,
		}

		resourceRequest := sdk.AddUserResourceRequest{
			Key:      "resource-1",
			Name:     "Resource One",
			PolicyId: "policy-1",
		}

		addResourceToUserObj(user, resourceRequest)

		// Verify resources field is initialized and resource is added
		assert.NotNil(t, user.Resources)
		assert.Len(t, user.Resources, 1)
		assert.Contains(t, user.Resources, "resource-1")

		res := user.Resources["resource-1"]
		assert.Equal(t, "resource-1", res.Key)
		assert.Equal(t, "Resource One", res.Name)
		assert.Contains(t, res.PolicyIds, "policy-1")
		assert.True(t, res.PolicyIds["policy-1"])
	})

	t.Run("success - add resource to existing resources", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Resources: map[string]sdk.UserResource{
				"existing-resource": {
					Key:       "existing-resource",
					Name:      "Existing Resource",
					PolicyIds: map[string]bool{"existing-policy": true},
					RoleIds:   map[string]bool{"existing-role": true},
				},
			},
		}

		resourceRequest := sdk.AddUserResourceRequest{
			Key:      "resource-1",
			Name:     "Resource One",
			PolicyId: "policy-1",
		}

		addResourceToUserObj(user, resourceRequest)

		// Verify existing resource is preserved
		assert.Len(t, user.Resources, 2)
		assert.Contains(t, user.Resources, "existing-resource")
		assert.Contains(t, user.Resources, "resource-1")

		// Verify new resource is added correctly
		res := user.Resources["resource-1"]
		assert.Equal(t, "resource-1", res.Key)
		assert.Equal(t, "Resource One", res.Name)
		assert.Contains(t, res.PolicyIds, "policy-1")
	})

	t.Run("success - add policy to existing resource", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Resources: map[string]sdk.UserResource{
				"resource-1": {
					Key:       "resource-1",
					Name:      "Resource One",
					PolicyIds: map[string]bool{"existing-policy": true},
					RoleIds:   map[string]bool{"role-1": true},
				},
			},
		}

		resourceRequest := sdk.AddUserResourceRequest{
			Key:      "resource-1",
			Name:     "Resource One Updated", // Name in request (but won't be used for existing resource)
			PolicyId: "policy-2",
		}

		addResourceToUserObj(user, resourceRequest)

		// Verify resource exists with both policies
		assert.Len(t, user.Resources, 1)
		res := user.Resources["resource-1"]
		assert.Equal(t, "resource-1", res.Key)
		assert.Equal(t, "Resource One", res.Name) // Name remains unchanged for existing resource
		assert.Contains(t, res.PolicyIds, "existing-policy")
		assert.Contains(t, res.PolicyIds, "policy-2")
		assert.True(t, res.PolicyIds["existing-policy"])
		assert.True(t, res.PolicyIds["policy-2"])

		// Verify existing role IDs are preserved
		assert.Contains(t, res.RoleIds, "role-1")
	})

	t.Run("success - add policy to resource with empty PolicyIds", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Resources: map[string]sdk.UserResource{
				"resource-1": {
					Key:       "resource-1",
					Name:      "Resource One",
					PolicyIds: map[string]bool{}, // Empty policy IDs
					RoleIds:   map[string]bool{"role-1": true},
				},
			},
		}

		resourceRequest := sdk.AddUserResourceRequest{
			Key:      "resource-1",
			Name:     "Resource One",
			PolicyId: "policy-1",
		}

		addResourceToUserObj(user, resourceRequest)

		// Verify policy is added to empty PolicyIds
		res := user.Resources["resource-1"]
		assert.NotNil(t, res.PolicyIds)
		assert.Contains(t, res.PolicyIds, "policy-1")
		assert.True(t, res.PolicyIds["policy-1"])
	})
}

// TestAddPoliciesToUserObj tests the addPoliciesToUserObj helper function
func TestAddPoliciesToUserObj(t *testing.T) {
	t.Run("success - add policies to user with nil policies", func(t *testing.T) {
		user := &sdk.User{
			Id:       "user-123",
			Policies: nil,
		}

		policiesToAdd := map[string]sdk.UserPolicy{
			"policy-1": {
				Name: "Policy One",
				Mapping: sdk.UserPolicyMapping{
					Arguments: map[string]sdk.UserPolicyMappingValue{
						"arg1": {Static: "value1"},
					},
				},
			},
			"policy-2": {
				Name: "Policy Two",
				Mapping: sdk.UserPolicyMapping{
					Arguments: map[string]sdk.UserPolicyMappingValue{
						"arg2": {Static: "value2"},
					},
				},
			},
		}

		addPoliciesToUserObj(user, policiesToAdd)

		// Verify policies field is initialized and policies are added
		assert.NotNil(t, user.Policies)
		assert.Len(t, user.Policies, 2)
		assert.Contains(t, user.Policies, "policy-1")
		assert.Contains(t, user.Policies, "policy-2")

		// Verify policy-1 details
		policy1 := user.Policies["policy-1"]
		assert.Equal(t, "Policy One", policy1.Name)
		assert.Contains(t, policy1.Mapping.Arguments, "arg1")
		assert.Equal(t, "value1", policy1.Mapping.Arguments["arg1"].Static)

		// Verify policy-2 details
		policy2 := user.Policies["policy-2"]
		assert.Equal(t, "Policy Two", policy2.Name)
		assert.Contains(t, policy2.Mapping.Arguments, "arg2")
		assert.Equal(t, "value2", policy2.Mapping.Arguments["arg2"].Static)
	})

	t.Run("success - add policies to existing policies", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Policies: map[string]sdk.UserPolicy{
				"existing-policy": {
					Name: "Existing Policy",
					Mapping: sdk.UserPolicyMapping{
						Arguments: map[string]sdk.UserPolicyMappingValue{
							"existing-arg": {Static: "existing-value"},
						},
					},
				},
			},
		}

		policiesToAdd := map[string]sdk.UserPolicy{
			"policy-1": {
				Name: "Policy One",
				Mapping: sdk.UserPolicyMapping{
					Arguments: map[string]sdk.UserPolicyMappingValue{
						"arg1": {Static: "value1"},
					},
				},
			},
		}

		addPoliciesToUserObj(user, policiesToAdd)

		// Verify existing policy is preserved and new policy is added
		assert.Len(t, user.Policies, 2)
		assert.Contains(t, user.Policies, "existing-policy")
		assert.Contains(t, user.Policies, "policy-1")

		// Verify existing policy is unchanged
		existingPolicy := user.Policies["existing-policy"]
		assert.Equal(t, "Existing Policy", existingPolicy.Name)
		assert.Contains(t, existingPolicy.Mapping.Arguments, "existing-arg")

		// Verify new policy is added correctly
		policy1 := user.Policies["policy-1"]
		assert.Equal(t, "Policy One", policy1.Name)
		assert.Contains(t, policy1.Mapping.Arguments, "arg1")
	})

	t.Run("success - overwrite existing policy", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Policies: map[string]sdk.UserPolicy{
				"policy-1": {
					Name: "Original Policy",
					Mapping: sdk.UserPolicyMapping{
						Arguments: map[string]sdk.UserPolicyMappingValue{
							"original-arg": {Static: "original-value"},
						},
					},
				},
			},
		}

		policiesToAdd := map[string]sdk.UserPolicy{
			"policy-1": {
				Name: "Updated Policy",
				Mapping: sdk.UserPolicyMapping{
					Arguments: map[string]sdk.UserPolicyMappingValue{
						"updated-arg": {Static: "updated-value"},
					},
				},
			},
		}

		addPoliciesToUserObj(user, policiesToAdd)

		// Verify policy is overwritten
		assert.Len(t, user.Policies, 1)
		policy1 := user.Policies["policy-1"]
		assert.Equal(t, "Updated Policy", policy1.Name)
		assert.Contains(t, policy1.Mapping.Arguments, "updated-arg")
		assert.NotContains(t, policy1.Mapping.Arguments, "original-arg")
	})

	t.Run("success - add empty policies map", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Policies: map[string]sdk.UserPolicy{
				"existing-policy": {Name: "Existing Policy"},
			},
		}

		policiesToAdd := map[string]sdk.UserPolicy{}

		addPoliciesToUserObj(user, policiesToAdd)

		// Verify existing policies are unchanged
		assert.Len(t, user.Policies, 1)
		assert.Contains(t, user.Policies, "existing-policy")
	})
}

// TestRemovePoliciesFromUserObj tests the removePoliciesFromUserObj helper function
func TestRemovePoliciesFromUserObj(t *testing.T) {
	t.Run("success - remove policies from user with nil policies", func(t *testing.T) {
		user := &sdk.User{
			Id:       "user-123",
			Policies: nil,
		}

		policyIds := []string{"policy-1", "policy-2"}

		removePoliciesFromUserObj(user, policyIds)

		// Verify policies field is initialized
		assert.NotNil(t, user.Policies)
		assert.Empty(t, user.Policies)
	})

	t.Run("success - remove existing policies", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Policies: map[string]sdk.UserPolicy{
				"policy-1": {Name: "Policy One"},
				"policy-2": {Name: "Policy Two"},
				"policy-3": {Name: "Policy Three"},
			},
		}

		policyIds := []string{"policy-1", "policy-3"}

		removePoliciesFromUserObj(user, policyIds)

		// Verify specified policies are removed and others remain
		assert.Len(t, user.Policies, 1)
		assert.NotContains(t, user.Policies, "policy-1")
		assert.Contains(t, user.Policies, "policy-2")
		assert.NotContains(t, user.Policies, "policy-3")

		// Verify remaining policy is unchanged
		policy2 := user.Policies["policy-2"]
		assert.Equal(t, "Policy Two", policy2.Name)
	})

	t.Run("success - remove non-existent policies", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Policies: map[string]sdk.UserPolicy{
				"policy-1": {Name: "Policy One"},
				"policy-2": {Name: "Policy Two"},
			},
		}

		policyIds := []string{"policy-999", "policy-888"}

		removePoliciesFromUserObj(user, policyIds)

		// Verify existing policies are unchanged
		assert.Len(t, user.Policies, 2)
		assert.Contains(t, user.Policies, "policy-1")
		assert.Contains(t, user.Policies, "policy-2")
	})

	t.Run("success - remove some existing and some non-existent policies", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Policies: map[string]sdk.UserPolicy{
				"policy-1": {Name: "Policy One"},
				"policy-2": {Name: "Policy Two"},
				"policy-3": {Name: "Policy Three"},
			},
		}

		policyIds := []string{"policy-1", "policy-999", "policy-3"}

		removePoliciesFromUserObj(user, policyIds)

		// Verify only existing policies are removed
		assert.Len(t, user.Policies, 1)
		assert.NotContains(t, user.Policies, "policy-1")
		assert.Contains(t, user.Policies, "policy-2")
		assert.NotContains(t, user.Policies, "policy-3")
	})

	t.Run("success - remove policies with empty list", func(t *testing.T) {
		user := &sdk.User{
			Id: "user-123",
			Policies: map[string]sdk.UserPolicy{
				"policy-1": {Name: "Policy One"},
				"policy-2": {Name: "Policy Two"},
			},
		}

		policyIds := []string{}

		removePoliciesFromUserObj(user, policyIds)

		// Verify no policies are removed
		assert.Len(t, user.Policies, 2)
		assert.Contains(t, user.Policies, "policy-1")
		assert.Contains(t, user.Policies, "policy-2")
	})
}
