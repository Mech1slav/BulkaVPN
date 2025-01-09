package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func GenerateVPNKey(clientEmail string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	loginBody := map[string]string{"username": username, "password": password}
	loginBodyBytes, _ := json.Marshal(loginBody)

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(loginBodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("login failed")
	}

	rawCookies := resp.Header["Set-Cookie"]
	if len(rawCookies) < 2 {
		return "", errors.New("not enough cookies in response")
	}

	parts := strings.Split(rawCookies[1], ";")
	if len(parts) == 0 || !strings.HasPrefix(parts[0], "3x-ui=") {
		return "", errors.New("cookie not found in the second entry")
	}
	cookie := parts[0]

	clientID := uuid.NewString()
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 16)
	rand.NewSource(time.Now().UnixNano())
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	subID := string(b)

	clientData := map[string]interface{}{
		"clients": []map[string]interface{}{
			{
				"id":     clientID,
				"flow":   "xtls-rprx-vision",
				"email":  clientEmail,
				"enable": true,
				"subId":  subID,
			},
		},
	}
	clientSettingsBytes, _ := json.Marshal(clientData)

	addClientBody := map[string]interface{}{
		"id":       1,
		"settings": string(clientSettingsBytes),
	}
	addClientBodyBytes, _ := json.Marshal(addClientBody)

	req, err = http.NewRequest("POST", fmt.Sprintf("%s/inbounds/addClient", apiBaseURL), bytes.NewBuffer(addClientBodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)

	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to add client: %s", string(body))
	}

	req, err = http.NewRequest("GET", fmt.Sprintf("%s/inbounds/list", apiBaseURL), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Cookie", cookie)

	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch inbounds: %s", string(body))
	}

	var response struct {
		Success bool                     `json:"success"`
		Obj     []map[string]interface{} `json:"obj"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if !response.Success {
		return "", errors.New("failed to fetch inbounds: API response unsuccessful")
	}

	var targetInbound map[string]interface{}
	for _, inbound := range response.Obj {
		if inbound["remark"].(string) == inboundRemark {
			targetInbound = inbound
			break
		}
	}

	if targetInbound == nil {
		return "", errors.New("target inbound not found")
	}

	streamSettingsRaw, ok := targetInbound["streamSettings"].(string)
	if !ok {
		return "", fmt.Errorf("streamSettings missing or not a string")
	}

	var streamSettings struct {
		Network         string `json:"network"`
		Security        string `json:"security"`
		RealitySettings struct {
			Settings struct {
				PublicKey   string `json:"publicKey"`
				Fingerprint string `json:"fingerprint"`
			}
			ServerName []string `json:"serverNames"`
			ShortIds   []string `json:"shortIds"`
		} `json:"realitySettings"`
	}
	if err := json.Unmarshal([]byte(streamSettingsRaw), &streamSettings); err != nil {
		return "", fmt.Errorf("failed to parse streamSettings: %v", err)
	}

	settingsRaw, ok := targetInbound["settings"].(string)
	if !ok {
		return "", fmt.Errorf("settings missing or not a string")
	}

	var settings struct {
		Clients []struct {
			Email string `json:"email"`
			ID    string `json:"id"`
		} `json:"clients"`
	}
	if err := json.Unmarshal([]byte(settingsRaw), &settings); err != nil {
		return "", fmt.Errorf("failed to parse settings: %v", err)
	}

	var clientIDFromInbound string
	for _, client := range settings.Clients {
		if client.Email == clientEmail {
			clientIDFromInbound = client.ID
			break
		}
	}
	if clientIDFromInbound == "" {
		return "", fmt.Errorf("client with email %s not found", clientEmail)
	}

	return fmt.Sprintf(
		"vless://%s@%s:%d?type=%s&security=%s&pbk=%s&fp=%s&sni=%s&sid=%s&spx=%s&flow=xtls-rprx-vision#%s-%s",
		clientIDFromInbound,
		"138.124.55.26",
		int(targetInbound["port"].(float64)),
		streamSettings.Network,
		streamSettings.Security,
		streamSettings.RealitySettings.Settings.PublicKey,
		streamSettings.RealitySettings.Settings.Fingerprint,
		streamSettings.RealitySettings.ServerName[0],
		streamSettings.RealitySettings.ShortIds[0],
		"%2F",
		targetInbound["remark"].(string),
		clientEmail,
	), nil
}

func main() {
	//clientEmail := "example-client-5"
	//vlessVpnKey, err := GenerateVPNKey(clientEmail)
	//if err != nil {
	//	fmt.Println("Error creating VPN key:", err)
	//	return
	//}
	//
	//fmt.Println("Generated VPN Key:", vlessVpnKey)

	//clientEmail := "example-client-3"
	//deleteVPN := DeleteKeyByConfig(clientEmail)
	//
	//fmt.Println("Deleted VPN Key", deleteVPN)

	//clientEmail := "julia"
	//existEmail, err := GetKeyByConfig(clientEmail)
	//if err != nil {
	//	fmt.Println("Error VPN key:", err)
	//	return
	//}
	//
	//fmt.Println("VPN Key exist:", existEmail)
}
