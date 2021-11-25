package mullCtrl

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Control Mullvad VPN via CLI commands
// - Connection check
// - Rotate IP addresses

func NewMullControlClient() MullControl {
	rand.Seed(time.Now().UnixNano())

	return MullControl{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		},
	}
}

func removeEmpty(slice []string) []string {
	var ret []string
	for _, s := range slice {
		if strings.TrimSpace(s) != "" {
			ret = append(ret, s)
		}
	}
	return ret
}

func getVpnType(s string) string {
	if strings.Contains(strings.ToLower(s), "openvpn") {
		return "openvpn"
	} else {
		return "wireguard"
	}
}

func getTextFirstParentheses(s string) string {
	resCountry := parenthesisExtractor.FindAllString(s, -1)
	if len(resCountry) >= 1 {
		return strings.TrimSpace(strings.Replace(strings.Replace(resCountry[0], ")", "", -1), "(", "", -1))
	}
	return s
}

func runWithoutOutput(args []string) (string, error) {
	cmd := exec.Command("mullvad", args...)
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	cmd.Dir = dir

	err = cmd.Run()
	if err != nil {
		return "", err
	}
	fmt.Println("ran", cmd.String())
	time.Sleep(time.Second * 2)
	return "", nil
}
func runCommand(args []string) (string, error) {
	cmd := exec.Command("mullvad", args...)
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	fmt.Println("ran", cmd.String())
	time.Sleep(time.Second * 2)
	return string(output), nil
}