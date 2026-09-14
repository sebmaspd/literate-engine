// Command haze-hourly evaluates a "haze-hourly" DMN model against a
// kogito-jit (JIT DMN executor) service running in Docker. The model maps a
// 1-hour PM2.5 concentration reading to NEA's Haze Band descriptor (Normal,
// Elevated, High, Very High), the same bands published alongside the PSI
// Haze Lookup Table during haze episodes.
package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

//go:embed haze-hourly.dmn
var hazeHourlyDMN string

// jitDMNPayload mirrors the JITDMNPayload schema exposed by kogito-jit's
// /jitdmn/dmnresult endpoint.
type jitDMNPayload struct {
	Model   string         `json:"model"`
	Context map[string]any `json:"context"`
}

// dmnResult mirrors the DMNResult JSON returned by /jitdmn/dmnresult.
type dmnResult struct {
	Namespace       string           `json:"namespace"`
	ModelName       string           `json:"modelName"`
	DMNContext      map[string]any   `json:"dmnContext"`
	Messages        []map[string]any `json:"messages"`
	DecisionResults []struct {
		DecisionID       string `json:"decisionId"`
		DecisionName     string `json:"decisionName"`
		Result           any    `json:"result"`
		EvaluationStatus string `json:"evaluationStatus"`
	} `json:"decisionResults"`
}

func main() {
	server := flag.String("server", "http://localhost:8080", "kogito-jit base URL")
	pm25 := flag.Float64("pm25", 0, "1-hr PM2.5 concentration in µg/m3, bound to the DMN input data 'PM25'")
	flag.Parse()

	endpoint := strings.TrimRight(*server, "/") + "/jitdmn/dmnresult"

	payload := jitDMNPayload{
		Model:   hazeHourlyDMN,
		Context: map[string]any{"PM25": *pm25},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("marshal request: %v", err)
	}

	resp, err := http.Post(endpoint, "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.Fatalf("call kogito-jit at %s: %v", endpoint, err)
	}
	defer resp.Body.Close()

	var result dmnResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("decode response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("kogito-jit returned %s: %+v", resp.Status, result)
	}

	for _, dr := range result.DecisionResults {
		fmt.Printf("%s (%s): %v\n", dr.DecisionName, dr.EvaluationStatus, dr.Result)
	}
}
