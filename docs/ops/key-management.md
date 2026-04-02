# Key Management — AgentRepEngine

## Key Location
RS256 key pair stored at `keys/private_key.pem` and `keys/public_key.pem`.
Directory is gitignored — keys are never committed to version control.

## What the Key Signs
- All agent JWTs issued by ARE
- All audit record hashes in the hash chain

## Key Rotation Procedure
1. Generate new key pair: `go run cmd/keygen/main.go` (or openssl)
2. Replace `keys/private_key.pem` and `keys/public_key.pem`
3. Restart scoring service
4. Historical audit records remain valid — they were signed under the previous key
5. New records are signed under the new key
6. Document rotation date in this file

## Rotation Log
| Date | Reason | Who |
|------|--------|-----|
| 2026-03-31 | Initial generation | Rehan Rana |

## On Compromise
If private key is suspected compromised:
1. Generate new key pair immediately
2. Restart scoring service
3. All future JWTs issued under new key
4. Historical audit records signed under old key remain intact
5. Notify pilot customer security contact within 24 hours

## Key Generation Command
```bash
openssl genrsa -out keys/private_key.pem 2048
openssl rsa -in keys/private_key.pem -pubout -out keys/public_key.pem
```