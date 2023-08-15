package main

//for WAP, App&API Protector
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/comcast/go-edgegrid/edgegrid"
	"github.com/kataras/golog"
)

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
	ID                string `json:"id"`
	ProductionVersion string `json:"productionVersion"`
}

var AkamaiHost string = "https://" + os.Getenv("AKAMAI_EDGEGRID_HOST")
var configID string = os.Getenv("AKAMAI_CONFIGID") //92484
var version = os.Getenv("AKAMAI_CONFIG_VERSION")   //1
var policyID = os.Getenv("AKAMAI_POLICYID")        //os.Getenv("AKAMAI_POLICYID")
var mode = "append"

func main() {

	//get configuration ID, version and policyID for WAP product
	//function calls...
	configJson := GetConfig(AkamaiHost)
	config := new(securityConfig)
	err := json.Unmarshal(configJson, config)
	if err != nil {
		golog.Fatal("Error!", err)
		return
	}
	golog.Info("Security config data: ")
	golog.Info(config)
	configID = config.ID
	version = config.ProductionVersion
	//Check for new hostnames

	//result := ListSelected(AkamaiHost, configID, version, policyID)

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

	h.Mode = mode
	newSet, err := json.Marshal(h)
	if err != nil {
		golog.Fatal("Error!", err)
		return
	}
	golog.Info(string(newSet))

	//clone latest config

	cloneData := cloneConfig{CreateFromVersion: version, RuleUpdate: false}
	cloneJson, err := json.Marshal(cloneData)
	if err != nil {
		golog.Fatal("Error!", err)
		return
	}
	clone := CloneConfig(AkamaiHost, configID, version, cloneJson)

	fmt.Printf("%+v", clone)

	//sent update request
	//Modify(AkamaiHost, configID, version, policyID, mode, newSet)

}

func Send(method, url string, data []byte) []byte {
	client := &http.Client{}
	payload := bytes.NewReader(data)
	golog.Info("Ready to send data " + string(data) + "\n to URL " + url)

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
	//golog.Info("Received data: " + string(contents))
	if resp.StatusCode > 201 {
		golog.Fatal("Response code is not OK!")
	}
	return contents
}

// only for KSD and Advanced AAP
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

func ModifySelectedHostnames(hostname, configID, version, policyID, mode string, data []byte) {
	golog.Info("Starting Modify func")
	url := hostname + "/appsec/v1/configs/" + configID + "/versions/" + version + "/security-policies/" + policyID + "/selected-hostnames"

	Send("PUT", url, data)
}
