package mc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	UserEndpoint = "https://api.ashcon.app/mojang/v2/user/"
)

type ErrMcUserNotFound struct {
	mcUsername string
}

func (e ErrMcUserNotFound) Error() string {
	return fmt.Sprintf("minecraft username %s not found", e.mcUsername)
}

type McUser struct {
	UUID            uuid.UUID `json:"uuid"`
	Username        string    `json:"username"`
	UsernameHistory []struct {
		Username  string     `json:"username"`
		ChangedAt *time.Time `json:"changed_at,omitempty"`
	} `json:"username_history"`
	Textures struct {
		Custom bool `json:"custom"`
		Slim   bool `json:"slim"`
		Skin   struct {
			URL  string `json:"url"`
			Data string `json:"data"`
		} `json:"skin"`
		Raw struct {
			Value     string `json:"value"`
			Signature string `json:"signature"`
		} `json:"raw"`
	} `json:"textures"`
	CreatedAt *time.Time `json:"created_at"`
}

func GetUser(mcUsername string) (*McUser, error) {
	res, err := http.Get(UserEndpoint + mcUsername)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, &ErrMcUserNotFound{mcUsername}
	}

	var user McUser
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
