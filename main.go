package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
)

func publicIP() (string, error) {
	resp, err := http.Get("https://1.1.1.1/cdn-cgi/trace")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(string(body), "\n") {
		if ip, ok := strings.CutPrefix(line, "ip="); ok {
			return ip, nil
		}
	}
	return "", errors.New("ip not found in trace")
}

func cloudflare(method, path string, body io.Reader, result any) error {
	req, err := http.NewRequest(method, "https://api.cloudflare.com/client/v4/zones/"+os.Getenv("CLOUDFLARE_ZONE_ID")+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CLOUDFLARE_API_TOKEN"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data := struct {
		Success bool
		Errors  []struct {
			Code    int
			Message string
		}
		Result any
	}{Result: result}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}
	if !data.Success {
		return fmt.Errorf("%s %s: %v", method, path, data.Errors)
	}
	return nil
}

func main() {
	if os.Getenv("CLOUDFLARE_API_TOKEN") == "" || os.Getenv("CLOUDFLARE_ZONE_ID") == "" || os.Getenv("DOMAINS") == "" {
		log.Fatal("CLOUDFLARE_API_TOKEN, CLOUDFLARE_ZONE_ID and DOMAINS must be set")
	}
	domains := strings.Split(os.Getenv("DOMAINS"), ",")

	ip, err := publicIP()
	if err != nil {
		log.Fatal(err)
	}

	var records []struct{ ID, Name, Content string }
	if err := cloudflare(http.MethodGet, "/dns_records?type=A", nil, &records); err != nil {
		log.Fatal(err)
	}
	for _, r := range records {
		if !slices.Contains(domains, r.Name) {
			continue
		}
		if r.Content == ip {
			log.Printf("%s: already %s", r.Name, ip)
			continue
		}
		if err := cloudflare(http.MethodPatch, "/dns_records/"+r.ID, strings.NewReader(fmt.Sprintf(`{"content":%q}`, ip)), nil); err != nil {
			log.Fatal(err)
		}
		log.Printf("%s: %s -> %s", r.Name, r.Content, ip)
	}
}
