# Noir-barretenberg-with-signature

A high-performance zero-knowledge proof (ZKP) generation backend built with Go and the Noir proving system. This project provides REST APIs for generating and verifying zk proofs with ECDSA signature verification.

## Overview

This application serves as a backend infrastructure for zero-knowledge proof generation. It combines:

- **Go**: For high-performance backend implementation
- **Noir**: A domain-specific language for creating zero-knowledge circuits
- **ECDSA Signing**: For cryptographic message authentication
- **Worker Pool**: For concurrent proof generation handling

## Features

- **Zero-Knowledge Proof Generation**: Generate zk proofs for custom circuits
- **Proof Verification**: Verify generated proofs
- **ECDSA Signing**: Sign messages using secp256k1 curve
- **Concurrent Processing**: Worker pool for handling multiple proof generation requests
- **Docker Support**: Pre-configured Dockerfile for containerized deployment
- **Performance Optimized**: Efficiently manages proof generation workflows

## Project Structure

```
.
├── main.go                    # Entry point and server setup
├── handler.go                 # HTTP request handlers
├── routes.go                  # API route definitions
├── bbBackend.go               # Backend proof system integration
├── signer.go                  # ECDSA signing functionality
├── worker.go                  # Worker pool implementation
├── types.go                   # Type definitions
├── utils.go                   # Utility functions
├── Dockerfile                 # Container configuration
├── go.mod                      # Go module dependencies
├── proofsystem/               # Noir proof system
│   ├── Nargo.toml            # Noir package configuration
│   ├── Prover.toml           # Prover settings
│   └── src/
│       └── main.nr           # Main Noir circuit definition
└── tmp/                        # Temporary working directories
```

## Prerequisites

- **Go 1.22.0** or higher
- **Nargo**: Noir package manager and CLI tool
- **Barretenberg (`bb`)**
- **Docker** (required if you plan to build or run the service in a container)
  - _Docker Compose_ is optional but recommended for multi‑container setups

> The first thing you should do when working with this project is ensure that Docker is installed on your machine. The included `Dockerfile` and any `docker-compose` configuration depend on it.

### Installation

1. **Install Go (Linux)**:

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

2. **Install Nargo (Noir CLI)**:

   ```bash
   curl -L https://raw.githubusercontent.com/noir-lang/noirup/main/install | bash
   noirup
   ```

3. **Install Barretenberg (bb)**:

   The project uses the native Barretenberg backend for proof generation and verification.

   ```bash
   curl -L https://raw.githubusercontent.com/AztecProtocol/aztec-packages/refs/heads/next/barretenberg/bbup/install | bash
   bbup
   ```

4. **Verify installations**:

   ```bash
   go version
   nargo --version
   bb --version
   ```

## Build

1. **Clone/Navigate to the project**:

   ```bash
   cd noir-barrentenberg-with-signature
   ```

2. **Install Go dependencies**:

   ```bash
   go mod download
   go mod tidy
   ```

3. **Install Noir dependencies**:
   ```bash
   cd proofsystem
   nargo check
   cd ..
   ```

## API Endpoints

- ### Generate Proof

**Endpoint**: `POST /proof/prove`

**Request Body**:

```json
{
  "circuitName": "main",
  "data": [1, 2, 3, 4, 5],
  "constraints": [
    {
      "operation": 0,
      "value": 10,
      "attribute_type": 1
    },
    {
      "operation": 1,
      "value": 20,
      "attribute_type": 2
    },
    // more constraints
  ]
}
```

**Response**:

```json
{
  "success": true,
  "proof": "base64_encoded_proof",
  "vk": "base64_encoded_verification_key",
  "publicInputs": "base64_encoded_public_inputs"
}
```

- ### Verify Proof

**Endpoint**: `POST /proof/verify`

**Request Body**:

```json
{
  "proof": "base64_encoded_proof",
  "vk": "base64_encoded_verification_key",
  "publicInputs": "base64_encoded_public_inputs"
}
```

**Response**:

```json
{
  "valid": true
}
```

## Usage

### Running Locally

1. **Start the server**:

   ```bash
   go run .
   ```

   The server will start on port `4000` (configurable via `PORT` environment variable):

   ```bash
   PORT=8080 go run .
   ```

2. **Test the prove endpoint**:
   ```bash
   curl -X POST http://localhost:4000/proof/prove \
     -H "Content-Type: application/json" \
     -d '{
       "circuitName": "main",
       "data": [1, 2, 3, 4, 5],
       "constraints": [
         {"operation": 0, "value": 10, "attribute_type": 1}
       ]
     }'
   ```

### Running with Docker

> **Note:** Docker must be installed before proceeding with any of the following steps. If you haven't installed it yet, download it from https://www.docker.com/get-started and verify with `docker --version`.

1. **Build the Docker image**:

   ```bash
   docker build -t noir-barrentenberg-with-signature:latest .
   ```

2. **Run the container**:

   ```bash
   docker run -p 4000:4000 \
     -e SIGNING_KEY=<your_private_signing_key> \
     noir-barrentenberg-with-signature:latest
   ```

   > **Note**: Provide your own secure 32-byte hexadecimal ECDSA private key.

## Configuration

### Environment Variables

- `PORT`: Server port (default: `4000`)

### Worker Pool

The worker pool size is dynamically configured based on the number of available CPU cores for optimal performance.

Update the worker initialization in `main.go`:

```go
import "runtime"
// choose n slightly less than available cores to leave capacity for the OS
cores := runtime.NumCPU()
n := cores - 1
if n < 1 {
   n = 1
}
StartWorkers(n) // n < CPU cores
```

## Architecture

### Components

1. **HTTP Server (Gin Framework)**
   - RESTful API for proof operations
   - Request validation and response formatting

2. **Signer Module**
   - ECDSA signing with secp256k1 curve
   - Message hashing and signature verification

3. **Worker Pool**
   - Concurrent job processing
   - Queue-based proof generation requests

4. **Backend Proof System**
   - Integration with Noir compiler
   - Witness generation and proof creation
   - Verification key extraction

5. **Noir Circuit**
   - Custom constraint validation logic
   - Constraint operations (>0, <1, ==2, !=3, >=4, <=5)
   - ECDSA signature verification within the circuit

## Circuit Details

The main Noir circuit (`proofsystem/src/main.nr`) supports:

- **Private inputs**: 5-element array of u32 values
- **Public inputs**: Constraint array, message hash, and signature
- **Constraint operations**:
  - 0: Greater than (>)
  - 1: Less than (<)
  - 2: Equal (==)
  - 3: Not equal (!=)
  - 4: Greater than or equal (>=)
  - 5: Less than or equal (<=)

## Building and Compiling

### Build Go Binary

```bash
go build -o server .
./server
```

### Compile Noir Circuit

```bash
cd proofsystem
nargo compile
cd ..
```

### View Circuit Information

```bash
cd proofsystem
nargo info
cd ..
```

## Performance Considerations

- The worker pool size should match your CPU core count for optimal performance
- Proof generation time depends on constraint complexity
- Nargo execution is the primary performance bottleneck
- Consider caching compiled circuits for repeated proof patterns

## Development

### Modifying the Circuit

1. Edit `proofsystem/src/main.nr`
2. Compile: `cd proofsystem && nargo compile`
3. Restart the server

### Debugging

Enable verbose output by examining logs in the console when running the server.

## Error Handling

Common error messages:

- **"error": "invalid signature"**: Signature validation failed in the circuit
- **"error": "constraint violated"**: One or more constraints failed validation
- **"error": "nargo command not found"**: Noir is not installed or not in PATH

## Security Considerations

- Messages are signed using ECDSA secp256k1
- Responses include base64-encoded proofs and keys
- Validate all inputs on the client side before submission
- Use HTTPS in production environments
- Implement rate limiting for proof endpoints

## Support

For issues and questions:

- Check existing documentation
- Review Noir documentation: https://docs.noir-lang.org
- Examine logs for detailed error messages
