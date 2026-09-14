// Command hello-world evaluates a "hello world" DMN model against a
// kogito-jit (JIT DMN executor) service running in Docker.
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

//go:embed hello.dmn
var helloDMN string

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
	name := flag.String("name", "World", "value bound to the DMN input data 'Name'")
	flag.Parse()

	endpoint := strings.TrimRight(*server, "/") + "/jitdmn/dmnresult"

	payload := jitDMNPayload{
		Model:   helloDMN,
		Context: map[string]any{"Name": *name},
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
