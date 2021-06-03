package game

type CreateUserBody struct {
	CustomSubdomain string `json:"customSubdomain" binding:"hostname_rfc1123"`
}
