package main

// ARE Certification Authority Portal v1
// Domain: are-certified.io
// Two pages only — no dashboard, no login, no management console.
//
// Page 1: Verify a certification
//   - Enter certification hash
//   - Returns verification result from /v1/certify/verify
//   - Human-readable format for DORA examiners
//
// Page 2: Certification standard
//   - Agentic Behavioral Certification Standard v0.1
//   - Links to Zenodo publication
//   - Regulatory anchors
//
// Stack: Go HTTP server + static HTML
// No framework. Minimal dependencies. Fast to audit.

// TODO: Implement main() with two routes
// TODO: Route /verify — serves Page 1 (static HTML form + JS fetch to /v1/certify/verify)
// TODO: Route /standard — serves Page 2 (rendered certification standard)
// TODO: No auth required. Public endpoints only.

func main() {
	// TODO
}
