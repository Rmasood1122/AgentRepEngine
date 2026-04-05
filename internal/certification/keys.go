package certification

// CertificationKeyManager manages the ARE certification signing key hierarchy
// Root key: offline (HSM in production, file in development)
// Operational key: rotated annually, signs individual certification reports
// Key hierarchy:
//   ARE Root Certification Key (offline)
//     └── ARE Operational Certification Key (rotated annually)
//           └── Per-report RS256 signature

// TODO: Implement LoadOperationalKey()
// TODO: Implement SignReport(report *CertificationReport) error
// TODO: Implement PublishJWKS() — endpoint /.well-known/certification-jwks.json
// TODO: Implement RotateKey() — annual rotation with continuity proof
