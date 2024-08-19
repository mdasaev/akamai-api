package main

//for WAP, App&API Protector
import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/comcast/go-edgegrid/edgegrid"
	"github.com/kataras/golog"
)

type edgerc struct {
	Host         string `json:"hostname"`
	Access_token string `json:"access_token"`
	Client_token string `json:"client_token"`
	Secret       string `json:"secret"`
}
type hostnameList struct {
	Items []hostname `json:"hostnameList"`
	Mode  string     `json:"mode"`
}

type hostname struct {
	Item string `json:"hostname"`
}

type hostRecord struct {
	ActiveInProduction bool `json:"activeInProduction"`
	ActiveInStaging    bool `json:"activeInStaging"`
	hostname
}

type selectableHostnames struct {
	AvailableSet []hostRecord `json:"availableSet"`
	//SelectedSet []hostRecord
}

type cloneConfig struct {
	CreateFromVersion string `json:"createFromVersion"`
	RuleUpdate        bool   `json:"ruleUpdate"`
}

type securityConfig struct {
	Items []config `json:"configurations"`
}
type config struct {
	ID                int `json:"id"`
	ProductionVersion int `json:"productionVersion"`
}
type clonedConfig struct {
	ConfigID int `json:"configId"`
	Version  int `json:"version"`
}

type activationConfig struct {
	ConfigId      int `json:"configId"`
	ConfigVersion int `json:"configVersion"`
}

type activation struct {
	Action             string             `json:"action"`
	ActivationConfigs  []activationConfig `json:"activationConfigs"`
	Network            string             `json:"network"`
	Note               string             `json:"note"`
	NotificationEmails []string           `json:"notificationEmails"`
}

var AkamaiHost string = "https://" + os.Getenv("EDGERC")

var configID string
var version string
var mode = "append"
var action = "ACTIVATE"
var network = "STAGING" //Update network for PRODUCTION to use in prod
var note = "Updated by Manage Hostname List script"
var notificationEmails = []string{"marat@globaldots.com"} //Set emails for notifications

func main() {

	dat, er1 := os.ReadFile(".edgerc")
	if er1 != nil {
		golog.Fatal(er1)
	}
	golog.Info(string(dat))

	//get configuration ID, version and policyID for WAP product
	golog.Info(AkamaiHost)
	configJson := GetConfig(AkamaiHost)
	config := new(securityConfig)
	err := json.Unmarshal(configJson, config)
	if err != nil {
		golog.Fatal("Error! ", err)
		return
	}
	//At least one security configuration  is expected - if its not there - script will fail

	configID = strconv.Itoa(config.Items[0].ID)
	version = strconv.Itoa(config.Items[0].ProductionVersion)

	//Check for new hostnames

	configHostnames := ListSelectableOnConfig(AkamaiHost, configID, version)

	policy := new(selectableHostnames)

	err = json.Unmarshal(configHostnames, policy)
	if err != nil {
		golog.Fatal("Error!", err)
		return
	}
	//grab available hostnames
	h := hostnameList{}
	for _, v := range policy.AvailableSet {
		h.Items = append(h.Items, v.hostname)
	}
	//If no new hostnames available - exit
	if len(h.Items) < 1 {
		golog.Warn("Available hostnames set is empty: No new hostnames present.\n Exiting...")
		return
	}
	//If there are, then create new hostname list to append
	golog.Info("Hostnames to be added: ")
	golog.Info(h)
	h.Mode = mode
	newSet, err := json.Marshal(h)
	if err != nil {
		golog.Fatal("Error!", err)
		return
	}
	golog.Info(string(newSet))

	//clone latest config

	cloneData := cloneConfig{CreateFromVersion: version, RuleUpdate: false}
	cloneJSON, err := json.Marshal(cloneData)
	if err != nil {
		golog.Fatal("Error!", err)
		return
	}
	clone := CloneConfig(AkamaiHost, configID, version, cloneJSON)
	v := new(clonedConfig)
	err = json.Unmarshal(clone, v)
	if err != nil {
		golog.Fatal("Error!", err)
		return
	}

	//sent update selectedSet request
	version = strconv.Itoa(v.Version)
	ModifySelectedHostnamesOnConfig(AkamaiHost, configID, version, mode, newSet)

	//Activate new version on "Network"
	//prepare POST data

	cid, err := strconv.Atoi(configID)
	if err != nil {
		golog.Fatal("Error convertation!", err)
		return
	}
	ver, err := strconv.Atoi(version)
	if err != nil {
		golog.Fatal("Error convertation!", err)
		return
	}

	activationData := new(activation)
	activationData.Action = action
	activationData.ActivationConfigs = make([]activationConfig, 0)
	aC := activationConfig{ConfigId: cid, ConfigVersion: ver}
	activationData.ActivationConfigs = append(activationData.ActivationConfigs, aC)
	activationData.Network = network
	activationData.Note = note
	activationData.NotificationEmails = notificationEmails

	activationJSON, err := json.Marshal(activationData)
	if err != nil {
		golog.Fatal("Error!", err)
		return
	}

	//ActivateConfiguration(AkamaiHost, activationJSON)
	golog.Info(activationJSON)

}

func Send(method, url string, data []byte) []byte {
	client := &http.Client{}
	payload := bytes.NewReader(data)
	golog.Info("Sending request " + string(data) + "\n to URL " + url)

	req, _ := http.NewRequest(method, url, payload)
	accessToken := os.Getenv("AKAMAI_EDGEGRID_ACCESS_TOKEN")
	clientToken := os.Getenv("AKAMAI_EDGEGRID_CLIENT_TOKEN")
	clientSecret := os.Getenv("AKAMAI_EDGEGRID_CLIENT_SECRET")

	params := edgegrid.NewAuthParams(req, accessToken, clientToken, clientSecret)
	auth := edgegrid.Auth(params)

	req.Header.Add("Authorization", auth)
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	resp, _ := client.Do(req)
	contents, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		golog.Error("Error!")
		return make([]byte, 0)
	}
	golog.Info("Request status code :")
	golog.Info(resp.StatusCode)
	golog.Info("Received data: " + string(contents))
	if resp.StatusCode > 201 {
		golog.Fatal("Response code is not OK!")
	}
	return contents
}

func ListSelectableOnConfig(hostname, configID, version string) []byte {
	golog.Info("Starting ListSelectable func")
	url := hostname + "/appsec/v1/configs/" + configID + "/versions/" + version + "/selectable-hostnames"
	return Send("GET", url, []byte{})

}

func ListSelectedOnPolicy(hostname, configID, version, policyID string) []byte {
	golog.Info("Starting ListSelected func")
	url := hostname + "/appsec/v1/configs/" + configID + "/versions/" + version + "/security-policies/" + policyID + "/selected-hostnames"
	return Send("GET", url, []byte{})
}

func GetConfig(hostname string) []byte {
	golog.Info("Starting GetConfig func")
	url := hostname + "/appsec/v1/configs/"
	return Send("GET", url, []byte{})
}

func CloneConfig(hostname, configID, version string, data []byte) []byte {
	golog.Info("Starting CloneConfig func")
	url := hostname + "/appsec/v1/configs/" + configID + "/versions"
	return Send("POST", url, data)
}

func ModifySelectedHostnamesOnConfig(hostname, configID, version, mode string, data []byte) {
	golog.Info("Starting Modify func")
	url := hostname + "/appsec/v1/configs/" + configID + "/versions/" + version + "/selected-hostnames"
	Send("PUT", url, data)
}
func ActivateConfiguration(hostname string, data []byte) {
	golog.Info("Starting ActivateConfiguration func")
	url := hostname + "/appsec/v1/activations"
	Send("POST", url, data)

}
