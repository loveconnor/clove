package orgs

type Org struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	OwnerID     string `json:"owner_id"`
}
