package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	rules := Rules{}
	err := json.NewDecoder(os.Stdin).Decode(&rules)
	if err != nil {
		fmt.Println(err)
		os.Exit(8)
	}
	fmt.Printf("%-7s %-45s %-40s %-60s\n", "Order", "Host condition", "Path Condition", "Target Host Header")
	fmt.Printf("%-7s %-45s %-40s %-60s\n", "======", "=============================================", "========================================", "=================================================================")
	for _, rule := range rules {
		if rule.ActionSet.RequestHeaderConfigurations != nil && rule.ActionSet.RequestHeaderConfigurations[0].HeaderName == "Host" {
			printLine := fmt.Sprintf("%-7d %-45s", rule.RuleSequence, rule.Conditions[0].Pattern)
			if len(rule.Conditions) > 1 {
				printLine += fmt.Sprintf(" %-40s", rule.Conditions[1].Pattern)
			} else {
				printLine += fmt.Sprintf(" %-40s", "")
			}
			printLine += fmt.Sprintf(" %-60s", rule.ActionSet.RequestHeaderConfigurations[0].HeaderValue)
			fmt.Println(printLine)
		}
	}
}

type Rules []struct {
	ActionSet struct {
		RequestHeaderConfigurations []struct {
			HeaderName  string `json:"headerName"`
			HeaderValue string `json:"headerValue"`
		} `json:"requestHeaderConfigurations"`
		ResponseHeaderConfigurations []any `json:"responseHeaderConfigurations"`
	} `json:"actionSet"`
	Conditions   []Condition `json:"conditions"`
	RuleSequence int         `json:"ruleSequence"`
}

type Condition struct {
	Pattern  string `json:"pattern"`
	Variable string `json:"variable"`
}
