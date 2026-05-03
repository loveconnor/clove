package repos

type Repo struct {
	ID            string `json:"id"`
	OwnerType     string `json:"owner_type"`
	OwnerID       string `json:"owner_id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Visibility    string `json:"visibility"`
	DefaultBranch string `json:"default_branch"`
	GitPath       string `json:"git_path"`
}
