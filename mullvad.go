package mullCtrl

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/showwin/speedtest-go/speedtest"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
"github.com/showwin/speedtest-go"
	"time"
)

var parenthesisExtractor = regexp.MustCompile(`(?m)\(([^)]+)\)`)
var useragents = []string{
	"curl/7.74.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36",
	"ozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15",
}

type MullControl struct {
	httpClient    *http.Client
	serverList    []Server
	iterationList []Server
	connectionMap map[string]bool
}

func (m *MullControl) FastestServerInCountry(country string) (err error) {
	return errors.New("Not implemented")

	if m.serverList == nil {
		_, err = m.GetServers()
		if err != nil {
			return err
		}
		m.iterationList = make([]Server, 0)
		user, _ := speedtest.FetchUserInfo()

		// Only use servers from the chosen country
		for _, server := range m.serverList {
			if server.CountryShort == country {
				serverList, _ := speedtest.FetchServerList(user)
				targets, _ := serverList.FindServer([]int{})

				for _, s := range targets {
					s.PingTest()
					s.DownloadTest(false)
					s.UploadTest(false)

					fmt.Printf("Latency: %s, Download: %f, Upload: %f\n", s.Latency, s.DLSpeed, s.ULSpeed)
				}
			}
		}
	}

}

// IterateCountryRandom connects on each call to a randomly chosen server from the selected country.
// Each server is chosen exactly once.
func (m *MullControl) IterateCountryRandom(country string) (err error) {
	// Cleanup this function
	if country == "" {
		return errors.New("country is empty")
	}

	// Get server list and connection map if we don't have it
	if m.serverList == nil {
		_, err = m.GetServers()
		if err != nil {
			return err
		}
		m.iterationList = make([]Server, 0)

		// Only use servers from the chosen country
		for _, server := range m.serverList {
			if server.CountryShort == country {
				m.iterationList = append(m.iterationList, server)
			}
		}
	}

	if len(m.iterationList) == 0 {
		return errors.New("All servers have been selected.")
	}

	// Pick one
	pick := rand.Intn(len(m.iterationList))

	err = m.ConnectToServer(m.iterationList[pick])
	if err != nil {
		return err
	}

	// Remove it
	m.iterationList[pick] = m.iterationList[len(m.iterationList)-1]
	m.iterationList[len(m.iterationList)-1] = Server{}
	m.iterationList = m.iterationList[:len(m.iterationList)-1]

	return
}

// IterateAllRandom connects on each call to a randomly chosen server.
// Each server is chosen exactly once.
func (m *MullControl) IterateAllRandom() (err error) {
	// Get server list and connection map if we don't have it
	if m.serverList == nil {
		_, err = m.GetServers()
		if err != nil {
			return err
		}
		m.iterationList = make([]Server, len(m.serverList))
		copy(m.iterationList, m.serverList)
	}

	if len(m.iterationList) == 0 {
		return errors.New("All servers have been selected.")
	}

	// Pick one
	pick := rand.Intn(len(m.iterationList))

	err = m.ConnectToServer(m.iterationList[pick])
	if err != nil {
		return err
	}

	// Remove it
	m.iterationList[pick] = m.iterationList[len(m.iterationList)-1]
	m.iterationList[len(m.iterationList)-1] = Server{}
	m.iterationList = m.iterationList[:len(m.iterationList)-1]

	return
}

// ResetIterationList resets the list of servers to be iterated through.
func (m *MullControl) ResetIteration() (err error) {
	if m.serverList == nil {
		_, err = m.GetServers()
		if err != nil {
			return err
		}
	}

	m.iterationList = make([]Server, len(m.serverList))
	copy(m.iterationList, m.serverList)
	return
}

// ConnectToServer connects to the given server.
func (m *MullControl) ConnectToServer(s Server) (err error) {
	_, err = runWithoutOutput([]string{"disconnect"})
	if err != nil {
		return err
	}
	_, err = runWithoutOutput([]string{"relay", "set", "tunnel-protocol", "any"})
	if err != nil {
		return err
	}
	_, err = runWithoutOutput([]string{"relay", "set", "tunnel", s.VpnType})
	if err != nil {
		return err
	}
	_, err = runWithoutOutput([]string{"relay", "set", "location", s.CountryShort, s.CityShort, s.ServerString})
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 2)
	_, err = runWithoutOutput([]string{"connect"})
	if err != nil {
		return err
	}
	return
}

// GetAccount returns the account id and expiry date.
func (m *MullControl) GetAccount() (account, expiry string, err error) {
	var result string
	result, err = runCommand([]string{"account", "get"})
	if err != nil {
		return
	}

	lines := strings.Split(result, "\n")
	account = strings.TrimSpace(strings.Split(lines[0], ":")[1])
	expiry = strings.TrimSpace(strings.Join(strings.Split(lines[1], ":")[1:], ":"))

	return
}

// GetStatus returns a MullvadResponse containing the status of the VPN connection.
func (m *MullControl) GetStatus() (MullvadResponse, error) {
	var response MullvadResponse

	req, _ := http.NewRequest("GET", "https://am.i.mullvad.net/json", nil)
	req.Header.Set("user-agent", useragents[rand.Intn(len(useragents))])
	req.Header.Set("accept", "*/*")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// IsConnected returns true if the VPN is connected.
func (m *MullControl) IsConnected() bool {
	status, err := m.GetStatus()
	if err != nil {
		return false
	}

	return status.MullvadExitIP
}

// GetServers returns the full list of available servers.
func (m *MullControl) GetServers() ([]Server, error) {
	result, err := runCommand([]string{"relay", "list"})
	if err != nil {
		return nil, err
	}

	lines := removeEmpty(strings.Split(result, "\n"))

	var servers = make([]Server, 0, len(lines)/3)
	var currentCountry, currentCity string

	for _, line := range lines {
		if strings.HasPrefix(line, "\t\t") {
			servers = append(servers, Server{
				Country:      strings.TrimSpace(strings.Split(currentCountry, "(")[0]),
				CountryShort: getTextFirstParentheses(currentCountry),
				City:         strings.TrimSpace(strings.Split(currentCity, "(")[0]),
				CityShort:    getTextFirstParentheses(currentCity),
				ServerString: strings.TrimSpace(strings.Split(strings.TrimSpace(line), "(")[0]),
				IP:           getTextFirstParentheses(strings.TrimSpace(line)),
				VpnType:      getVpnType(line),
			})

		} else if strings.HasPrefix(line, "\t") {
			currentCity = strings.TrimSpace(line)
		} else {
			currentCountry = strings.TrimSpace(line)
		}
	}

	m.serverList = servers

	return servers, nil

}

type MullvadResponse struct {
	IP                    string  `json:"ip"`
	Country               string  `json:"country"`
	City                  string  `json:"city"`
	Longitude             float64 `json:"longitude"`
	Latitude              float64 `json:"latitude"`
	MullvadExitIP         bool    `json:"mullvad_exit_ip"`
	MullvadExitIPHostname string  `json:"mullvad_exit_ip_hostname"`
	MullvadServerType     string  `json:"mullvad_server_type"`
	Blacklisted           struct {
		Blacklisted bool `json:"blacklisted"`
		Results     []struct {
			Name        string `json:"name"`
			Link        string `json:"link"`
			Blacklisted bool   `json:"blacklisted"`
		} `json:"results"`
	} `json:"blacklisted"`
	Organization string `json:"organization"`
}

type Server struct {
	Country      string
	CountryShort string
	City         string
	CityShort    string
	ServerString string
	IP           string
	VpnType      string
}
