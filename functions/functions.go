package functions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// FunctionListEntry represents a function returned from /system/functions
type FunctionListEntry struct {
	// The name of the function
	Name string `json:"name"`
	// The fully qualified docker image name of the function
	Image string `json:"image"`
	// The amount of invocations for the specified function
	InvocationCount float32 `json:"invocationCount"`
	// The current minimal ammount of replicas
	Replicas float32 `json:"replicas"`
	// The current available amount of replicas
	AvailableReplicas float32 `json:"availableReplicas"`
	// Process for watchdog to fork
	EnvProcess string `json:"envProcess"`
	// A map of labels for making scheduling or routing decisions
	Labels map[string]string `json:"labels"`
	// A map of annotations for management, orchestration, events and build tasks
	Annotations map[string]string `json:"annotations,omitempty"`
}

var gateway string

func getConf(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("Error finding %s", key)
	}
	return val, nil
}
func init() {
	var err error
	gateway, err = getConf("FAAS_GATEWAY")
	if err != nil {
		fmt.Println(err)
	}
}

// List the available functions in FAAS
func List() ([]FunctionListEntry, error) {
	resp, err := http.Get(fmt.Sprintf("%s/system/functions", gateway))
	if err != nil {
		return nil, err
	}
	if resp.Header.Get("status") != "200" {
		return nil, fmt.Errorf("Can't connect to %s", gateway)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var list []FunctionListEntry
	err = json.Unmarshal(body, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
