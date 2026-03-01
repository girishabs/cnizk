# Noir-Barretenberg-Without-Signature

- A Go + Noir based Zero-Knowledge Proof backend using the Native Barretenberg proving system. This project generates and verifies zero-knowledge proofs for constraint validation circuits.

---

## Overview

This backend provides:

- Zero-Knowledge proof generation
- Proof verification using native Barretenberg (`bb`)
- Constraint-based validation logic
- Concurrent worker pool execution
- Optional backend-level attestation (not part of circuit proof)

---

## Key Difference From Signature Version

In this version:

- The Noir circuit does NOT verify ECDSA signatures.
- Proof generation is purely constraint-based.
- Signature logic exists only at the API layer (for response attestation).
- Signature data is NOT included inside the zero-knowledge proof.
- Signature verification is not part of the zero-knowledge circuit and is used only for backend attestation.

---

## Architecture Flow

```
Client Request
      ↓
Gin Router
      ↓
Handler (Prepare Inputs)
      ↓
Worker Pool
      ↓
Nargo (Witness Generation)
      ↓
Barretenberg Native (Proof Generation)
      ↓
Proof + VK + Public Inputs
```

---

## Project Structure

```
.
├── main.go
├── handler.go
├── routes.go
├── worker.go
├── bbBackend.go
├── utils.go
├── types.go
├── signer.go               
├── signature_verify.go     
├── Dockerfile
├── go.mod
├── proofsystem/
│   ├── Nargo.toml
│   ├── Prover.toml
│   └── src/
│       └── main.nr
└── tmp/
```

---

## Noir Circuit

The circuit validates:

- A private array of 5 values
- A public array of 5 constraints

Supported operations:

- 0 → Equal (==)
- 1 → Not Equal (!=)
- 2 → Less Than (<)
- 3 → Less Than or Equal (<=)
- 4 → Greater Than (>)
- 5 → Greater Than or Equal (>=)

The circuit DOES NOT perform signature verification.

---

## API Endpoints

### Generate Proof

POST `/proof/prove`

Request:

```json
{
  "data": [1, 2, 3, 4, 5],
  "constraints": [
    {
      "operation": 0,
      "value": 10,
      "attribute_type": 1
    }
  ]
}
```

Response includes:

- proof
- verification key
- public inputs
- optional message signature (API-level only)

---

### Verify Proof

POST `/proof/verify`

Verifies:

1. Optional backend signature (API level)
2. Zero-knowledge proof using `bb verify`

---

## Prerequisites

- Go 1.22+
- Nargo (Noir CLI)
- Barretenberg (`bb`)

**Install Go (Linux)**:

Follow the official Go installation guide: https://go.dev/doc/install

```bash
# download Go archive
wget https://go.dev/dl/go<version>.linux-amd64.tar.gz

# remove previous installation (if any)
sudo rm -rf /usr/local/go

# extract archive
sudo tar -C /usr/local -xzf go<version>.linux-amd64.tar.gz

# add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# verify installation
go version
```

### Install Noir

```bash
curl -L https://raw.githubusercontent.com/noir-lang/noirup/main/install | bash
noirup
```

### Install Barretenberg

```bash
curl -L https://raw.githubusercontent.com/AztecProtocol/aztec-packages/refs/heads/next/barretenberg/bbup/install | bash
bbup
```

Verify:

```bash
go version
nargo --version
bb --version
```

---

## Running Locally

Set environment variable:

```bash
export SIGNING_KEY=<hex_private_key>
```

Start server:

```bash
go run .
```

Default port: 4000

---

## Docker

Build:

```bash
docker build -t noir-barrentenberg-without-signature .
```

Run:

```bash
docker run -d \
-p 4000:4000 \
-e PORT=4000 \
-e SIGNING_KEY=<hex_private_key> \
noir-barrentenberg-without-signature
```

---

## Performance Logging

The system logs:

- Witness generation time
- Proof generation time
- Proof verification time
- Optional signing time

---

## Summary

This project demonstrates:

- Constraint-based Zero-Knowledge proofs using Noir
- Native Barretenberg proving backend
- Parallel worker execution
- Clean separation between circuit logic and API-level attestation
