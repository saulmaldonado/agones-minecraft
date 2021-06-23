package mc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	appHttp "agones-minecraft/services/http"
)

const (
	UserEndpoint = "https://api.ashcon.app/mojang/v2/user/"
)

type ErrMcUserNotFound struct {
	mcUsername string
}

type ErrUnmarshalingMCAccountJSON struct{ error }

func (e ErrMcUserNotFound) Error() string {
	return fmt.Sprintf("minecraft username %s not found", e.mcUsername)
}

type McUser struct {
	UUID            uuid.UUID `json:"uuid"`
	Username        string    `json:"username"`
	UsernameHistory []struct {
		Username  string         `json:"username"`
		ChangedAt *McAccountDate `json:"changed_at,omitempty"`
	} `json:"username_history"`
	Textures struct {
		Custom bool `json:"custom"`
		Slim   bool `json:"slim"`
		Skin   struct {
			URL string `json:"url"`
		} `json:"skin"`
	} `json:"textures"`
	CreatedAt *McAccountDate `json:"created_at"`
}

// Custom time format for MC account creation dates
type McAccountDate time.Time

// Custom unmarshaler for mc account creation dates
func (d *McAccountDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}

	*d = McAccountDate(t)
	return nil
}

func GetUser(mcUsername string) (*McUser, error) {
	req, err := http.NewRequest("GET", UserEndpoint+mcUsername, nil)
	if err != nil {
		return nil, err
	}

	res, err := appHttp.Client().Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, &ErrMcUserNotFound{mcUsername}
	}

	var user McUser
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return nil, &ErrUnmarshalingMCAccountJSON{err}
	}

	return &user, nil
}
