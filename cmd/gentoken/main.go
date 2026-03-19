package main

import (
	"fmt"
	"time"

	"github.com/agentrepengine/are/internal/identity"
)

func main() {
	keys, err := identity.LoadOrGenerateKeys()
	if err != nil {
		panic(err)
	}
	claims := identity.NewAgentClaims(
		"did:jwt:finserv-demo:trading-agent:001",
		"inst-001",
		identity.ComputeLineageHash("root", time.Now()),
		"11111111-1111-1111-1111-111111111111",
		0,
	)
	token, err := identity.SignToken(claims, keys)
	if err != nil {
		panic(err)
	}
	fmt.Println(token)
}
