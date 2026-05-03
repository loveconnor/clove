package pulls

type Pull struct {
	ID         string `json:"id"`
	RepoID     string `json:"repo_id"`
	ExternalID string `json:"external_id"`
	Number     int    `json:"number"`
	Title      string `json:"title"`
	State      string `json:"state"`
	Author     string `json:"author"`
}
