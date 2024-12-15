package ds

type ErrorResp struct {
	Description string `json:"description"`
	Status      string `json:"status"`
}

type OkResp struct {
	Message string `json:"message"`
}
