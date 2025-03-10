package gcs

import "cloud.google.com/go/storage"

// ACLRule represents a grant for a role to an entity (user, group or team) for a
// Google Cloud Storage object or bucket.
type ACLRule struct {
	Entity      storage.ACLEntity `json:"entity,omitempty"`
	EntityID    string            `json:"entity_id,omitempty"`
	Role        storage.ACLRole   `json:"role,omitempty"`
	Domain      string            `json:"domain,omitempty"`
	Email       string            `json:"email,omitempty"`
	ProjectTeam *ProjectTeam      `json:"project_team,omitempty"`
}

// ProjectTeam is the project team associated with the entity, if any.
type ProjectTeam struct {
	ProjectNumber string `json:"project_number,omitempty"`
	Team          string `json:"team,omitempty"`
}

func makeACLRule(acl storage.ACLRule) ACLRule {
	rule := ACLRule{
		Entity:   acl.Entity,
		EntityID: acl.EntityID,
		Role:     acl.Role,
		Domain:   acl.Domain,
		Email:    acl.Email,
	}

	if acl.ProjectTeam != nil {
		rule.ProjectTeam = &ProjectTeam{
			ProjectNumber: acl.ProjectTeam.ProjectNumber,
			Team:          acl.ProjectTeam.Team,
		}
	}

	return rule
}
