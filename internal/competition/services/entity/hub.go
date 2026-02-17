package entity

type HubType string
type HubVisibility string

const (
	HubUniversity HubType = "UNIVERSITY"
	HubCorporate  HubType = "CORPORATE"
	HubCommunity  HubType = "COMMUNITY"

	HubPublic  HubVisibility = "PUBLIC"
	HubPrivate HubVisibility = "PRIVATE"
)

type Hub struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Type        HubType       `json:"type"`
	Visibility  HubVisibility `json:"visibility"`
	OwnerID     string        `json:"owner_id"`
	Description string        `json:"description"`
	CreatedAt   int64         `json:"created_at"`
}

type HubMembership struct {
	HubID    string `json:"hub_id"`
	UserID   string `json:"user_id"`
	Role     string `json:"role"` // OWNER, ADMIN, MEMBER
	JoinedAt int64  `json:"joined_at"`
}
