package holland_vless

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func GetKeyByConfig(clientEmail string) bool {
	client := &http.Client{Timeout: 10 * time.Second}

	loginBody := map[string]string{"username": username, "password": password}
	loginBodyBytes, err := json.Marshal(loginBody)
	if err != nil {
		fmt.Printf("Error marshaling login body: %v\n", err)
		return false
	}

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(loginBodyBytes))
	if err != nil {
		fmt.Printf("Failed to create login request: %v\n", err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Login request failed: %v\n", err)
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Login failed, unexpected status code: %d\n", resp.StatusCode)
		return false
	}

	rawCookies := resp.Header["Set-Cookie"]
	if len(rawCookies) < 2 {
		fmt.Println("Not enough cookies in login response")
		return false
	}

	parts := strings.Split(rawCookies[1], ";")
	if len(parts) == 0 || !strings.HasPrefix(parts[0], "3x-ui=") {
		fmt.Println("Cookie not found in the second entry")
		return false
	}
	cookie := parts[0]

	req, err = http.NewRequest("GET", fmt.Sprintf("%s/inbounds/list", apiBaseURL), nil)
	if err != nil {
		fmt.Printf("Failed to create request for clients list: %v\n", err)
		return false
	}
	req.Header.Set("Cookie", cookie)

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("Failed to fetch clients list: %v\n", err)
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("failed to fetch inbounds: %s", string(body))
		return false
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
		fmt.Printf("failed to decode inbounds response: %v", err)
		return false
	}

	if !response.Success {
		fmt.Println("Failed to fetch inbounds: API response unsuccessful")
		return false
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
		fmt.Println("Target inbound not found")
		return false
	}

	var settings struct {
		Clients []struct {
			Email string `json:"email"`
		} `json:"clients"`
	}
	if err := json.Unmarshal([]byte(targetInbound.Settings), &settings); err != nil {
		fmt.Printf("failed to parse inbound settings: %v", err)
		return false
	}

	for _, clients := range settings.Clients {
		if clients.Email == clientEmail {
			return true
		}
	}
	return false
}
