package main

//for WAP, App&API Protector
import (
	"bytes"
	"encoding/json"
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

var AkamaiHost string = "https://" + os.Getenv("AKAMAI_EDGEGRID_HOST")
var configID string = os.Getenv("AKAMAI_CONFIGID") //92484
var version = os.Getenv("AKAMAI_CONFIG_VERSION")   //1
var policyID = "doc1_213348"                       //os.Getenv("AKAMAI_POLICYID")
var mode = "replace"

func main() {
	result := ListSelected(AkamaiHost, configID, version, policyID)
	//ListSelected(AkamaiHost, configID, version, policyID)

	policy := new(selectableHostnames)

	err := json.Unmarshal(result, policy)
	if err != nil {
		golog.Fatal("Error!", err)
		return
	}

	h := hostnameList{}
	for _, v := range policy.AvailableSet {
		h.Items = append(h.Items, v.hostname)
	}
	h.Mode = mode
	newSet, err := json.Marshal(h)
	if err != nil {
		golog.Fatal("Error!", err)
		return
	}
	golog.Info(string(newSet))
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
	golog.Info("Received data: " + string(contents))
	if resp.StatusCode != 200 {
		golog.Error("Request not successfull")
	}
	return contents
}

/*
func ListSelectable(hostname, configID, version, policyID string) []byte {
	golog.Info("Starting ListSelectable func")
	url := hostname + "/appsec/v1/configs/" + configID + "/version/" + version + "/security-policies/" + policyID + "/selectable-hostnames"
	return Send("GET", url, []byte{})

}*/

func ListSelected(hostname, configID, version, policyID string) []byte {
	golog.Info("Starting ListSelected func")
	url := hostname + "/appsec/v1/configs/" + configID + "/version/" + version + "/security-policies/" + policyID + "/selected-hostnames"
	return Send("GET", url, []byte{})
}

func Modify(hostname, configID, version, policyID, mode string, data []byte) {
	golog.Info("Starting Modify func")
	url := hostname + "/appsec/v1/configs/" + configID + "/version/" + version + "/security-policies/" + policyID + "/selected-hostnames"

	Send("PUT", url, data)
}
