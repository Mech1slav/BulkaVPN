package holland_vless

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func DeleteKeyByConfig(clientEmail string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	loginBody := map[string]string{"username": username, "password": password}
	loginBodyBytes, _ := json.Marshal(loginBody)

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(loginBodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return errors.New("login failed, unexpected status code")
	}

	rawCookies := resp.Header["Set-Cookie"]
	if len(rawCookies) < 2 {
		return errors.New("not enough cookies in login response")
	}

	parts := strings.Split(rawCookies[1], ";")
	if len(parts) == 0 || !strings.HasPrefix(parts[0], "3x-ui=") {
		return errors.New("cookie not found in the second entry")
	}
	cookie := parts[0]

	req, err = http.NewRequest("GET", fmt.Sprintf("%s/inbounds/list", apiBaseURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request for client list: %v", err)
	}
	req.Header.Set("Cookie", cookie)

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch client list: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to fetch inbounds: %s", string(body))
	}

	var response struct {
		Success bool `json:"success"`
		Obj     []struct {
			ID       int    `json:"id"`
			Remark   string `json:"remark"`
			Settings string `json:"settings"`
		} `json:"obj"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode inbounds response: %v", err)
	}

	if !response.Success {
		return errors.New("failed to fetch inbounds: API response unsuccessful")
	}

	var targetInbound *struct {
		ID       int    `json:"id"`
		Remark   string `json:"remark"`
		Settings string `json:"settings"`
	}
	for _, inbound := range response.Obj {
		if inbound.Remark == inboundRemark {
			targetInbound = &inbound
			break
		}
	}
	if targetInbound == nil {
		return errors.New("target inbound not found")
	}

	var settings struct {
		Clients []struct {
			Email string `json:"email"`
			ID    string `json:"id"`
		} `json:"clients"`
	}
	if err := json.Unmarshal([]byte(targetInbound.Settings), &settings); err != nil {
		return fmt.Errorf("failed to parse inbound settings: %v", err)
	}

	var clientID string
	for _, clients := range settings.Clients {
		if clients.Email == clientEmail {
			clientID = clients.ID
			break
		}
	}
	if clientID == "" {
		return fmt.Errorf("client with email %s not found", clientEmail)
	}

	deleteClientURL := fmt.Sprintf("%s/inbounds/1/delClient/%s", apiBaseURL, clientID)
	req, err = http.NewRequest("POST", deleteClientURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete client request: %v", err)
	}
	req.Header.Set("Cookie", cookie)

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("delete client request failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete client: %s", string(body))
	}

	return nil
}
