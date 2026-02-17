package user

import (
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
)

func removeRoleFromUserObj(user *sdk.User, role sdk.Role) {
	// Ensure user's fields are initialized
	if user.Roles == nil {
		user.Roles = make(map[string]sdk.UserRole)
	}
	if user.Resources == nil {
		user.Resources = make(map[string]sdk.UserResource)
	}

	// update user roles
	delete(user.Roles, role.Id)

	// Remove resources only if no other roles require them
	for _, res := range role.Resources {

		vl, exists := user.Resources[res.Key]
		if !exists {
			continue
		}
		delete(vl.RoleIds, role.Id)

		// there are no requirement of the resource as no one needs it
		if len(vl.RoleIds) == 0 && len(vl.PolicyIds) == 0 {
			delete(user.Resources, res.Key)
		} else {
			user.Resources[res.Key] = vl
		}
	}
}

func addRoleToUserObj(user *sdk.User, role sdk.Role) {
	// Initialize user's fields if nil
	if user.Roles == nil {
		user.Roles = make(map[string]sdk.UserRole)
	}
	if user.Resources == nil {
		user.Resources = make(map[string]sdk.UserResource)
	}

	// Add new role
	user.Roles[role.Id] = sdk.UserRole{
		Id:   role.Id,
		Name: role.Name,
	}

	// Add unique resources from role
	for _, res := range role.Resources {
		// other ran roleids policy ids cuold also exist that is why special treatment for resources
		existingResource, exists := user.Resources[res.Key]
		if !exists {
			existingResource = sdk.UserResource{
				RoleIds: map[string]bool{role.Id: true},
				Key:     res.Key,
				Name:    res.Name,
			}
		} else {
			if len(existingResource.RoleIds) == 0 {
				existingResource.RoleIds = map[string]bool{}
			}
			existingResource.RoleIds[role.Id] = true
		}
		user.Resources[res.Key] = existingResource
	}
}

func addResourceToUserObj(user *sdk.User, res sdk.AddUserResourceRequest) {
	// Initialize user's fields if nil
	if user.Resources == nil {
		user.Resources = make(map[string]sdk.UserResource)
	}

	// other ran roleids policy ids cuold also exist that is why special treatment for resources
	existingResource, exists := user.Resources[res.Key]
	if !exists {
		existingResource = sdk.UserResource{
			PolicyIds: map[string]bool{res.PolicyId: true},
			Key:       res.Key,
			Name:      res.Name,
		}
	} else {
		if len(existingResource.PolicyIds) == 0 {
			existingResource.PolicyIds = map[string]bool{}
		}
		existingResource.PolicyIds[res.PolicyId] = true
	}
	user.Resources[res.Key] = existingResource
}

func addPoliciesToUserObj(user *sdk.User, policies map[string]sdk.UserPolicy) {
	// Initialize user's fields if nil
	if user.Policies == nil {
		user.Policies = make(map[string]sdk.UserPolicy)
	}

	for key, policy := range policies {
		user.Policies[key] = policy
	}
}

func removePoliciesFromUserObj(user *sdk.User, policyIds []string) {
	// Initialize user's fields if nil
	if user.Policies == nil {
		user.Policies = make(map[string]sdk.UserPolicy)
	}

	for _, policyId := range policyIds {
		delete(user.Policies, policyId)
	}
}
