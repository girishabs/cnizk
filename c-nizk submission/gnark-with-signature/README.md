# zk-SNARKs Implementation for Zero-Knowledge Authorization

This repository implements a zero-knowledge authorization system using zk-SNARKs (Zero-Knowledge Succinct Non-Interactive Arguments of Knowledge). It enables parties to prove possession of certain data that satisfies specific constraints without revealing the actual data values.

## Table of Contents

1. [Overview](#overview)
2. [System Architecture](#system-architecture)
3. [Components](#components)
4. [Circuit Design](#circuit-design)
5. [How It Works](#how-it-works)
6. [Running the System](#running-the-system)
7. [Usage Examples](#usage-examples)
8. [Security Considerations](#security-considerations)
9. [Limitations](#limitations)
10. [Future Improvements](#future-improvements)

## Overview

This implementation provides a zero-knowledge authorization mechanism where:

- A prover can demonstrate possession of data that satisfies specific constraints
- The verifier can validate this claim without learning the actual data
- Proofs embed a timestamp and a cryptographic signature
- All data remains private except for the constraints that are satisfied

## System Architecture

The system consists of three main components:

1. **Prover**: Generates the zero-knowledge proof
2. **Verifier**: Validates the proof
3. **Circuit**: Defines the constraints and logic for the zk-SNARK

## Components

### 1. Circuit (`circuit/`)

Contains the zk-SNARK circuit definition and proof generation logic.

### 2. Prover (`prover/`)

Handles the proof generation process, including:

- Data extraction and conversion
- Constraint processing
- Signature creation
- Proof generation using Groth16

### 3. Verifier (`verifier/`)

Handles proof verification, including:

- Receiving and processing proofs
- Validating the zk-SNARK proof against the verification key

### 4. Model (`model/`)

Defines data structures and circuit constraints:

- `NewCircuit`: Main circuit structure
- `Asset`: Asset data structure
- `Statement`: Constraint definition structure

### 5. Utilities (`utils/`)

Helper functions for:

- Data extraction from assets
- Constraint processing
- Signature and public key conversion

### 6. Constants (`constant/`)

Configuration constants for:

- Operations (equality, inequality, etc.)
- Network connections
- Attribute names
- Data types

## Circuit Design

The zk-SNARK circuit implements the following features:

### Data Structure

The system works with 5 asset attributes:

1. Owner (string)
2. Status (boolean)
3. Value (integer)
4. Type (string)
5. Version (string)

### Constraints

Each constraint supports the following operations:

- Equality (`eq`)
- Inequality (`ne`)
- Less than (`lt`)
- Greater than or equal to (`gte`)
- Greater than (`gt`)
- Less than or equal to (`lte`)

### Types

- Integer (`int`)
- String (`string`)
- Boolean (`bool`)

### Security Features

- **Authenticity**: Data and constraints are cryptographically signed
- **Privacy**: Only private data is hidden from the verifier
- **Timestamp**: Timestamp included in proof
- **Non-interactive**: Proofs can be verified anywhere without interaction

## How It Works

1. **Setup Phase**:
   - The prover creates an asset with sensitive data
   - Constraints that the data must satisfy are defined
   - The prover generates an EdDSA key and signs the message

2. **Proof Generation**:
   - The prover extracts data and converts it to field elements
   - Constraints are processed and converted to field elements
   - A Poseidon hash is computed over the data and constraints
   - The prover generates and signs the message
   - A Groth16 proof is generated using the circuit

3. **Verification**:
   - The verifier receives the proof and verification key
   - Signature is embedded into zk proof verification
   - The verifier validates the circuit constraints

## Running the System

### Prerequisites

- Go 1.19+
- Git

### Installation

```bash
git clone <repository-url>
cd zk-snarks
go mod tidy
```

### Running

```bash
go run main.go
```

This will start a single HTTP server with the following endpoints:

- `/prove` - Proof generation endpoint
- `/verify` - Proof verification endpoint

## Usage Examples

### Basic Usage

```go
// Define asset data
asset := model.Asset{
    Owner:   "abcd",
    Version: "1.0.0",
    Type:    "cbdc",
    Value:   300000,
    Status:  true,
}

// Define constraints
statement := []model.Statement{
    {
        Attribute: constant.OWNER_ATTRIBUTE,
        Type:      constant.STRING,
        Operation: constant.EQUAL,
        Value:     "abcd",
    },
    {
        Attribute: constant.STATUS_ATTRIBUTE,
        Type:      constant.BOOL,
        Operation: constant.EQUAL,
        Value:     true,
    },
    // ... more constraints
}
```

### Custom Constraints

```go
// Integer comparison
{
    Attribute: constant.VALUE_ATTRIBUTE,
    Type:      constant.INT,
    Operation: constant.GREATERTHAN,
    Value:     100000,
}

// String equality
{
    Attribute: constant.TYPE_ATTRIBUTE,
    Type:      constant.STRING,
    Operation: constant.EQUAL,
    Value:     "cbdc",
}
```

## Security Considerations

### Strengths

- **Data Privacy**: Only the data that satisfies constraints is revealed
- **Authenticity**: Signatures are embedded and verified via the zk proof
- **Timestamp**: Proofs include a timestamp but expiration is not enforced by verifier
- **Non-interactive**: Proofs can be verified anywhere

### Known Limitations

- **Field Size**: Assumes values fit within 66-bit field (may cause truncation)
- **Public Constraints**: Constraint structure is visible to verifiers
- **Performance**: Cryptographic operations are computationally intensive

## Limitations

1. **Fixed Data Structure**: Currently limited to 5 predefined asset attributes
2. **Constraint Visibility**: Constraint structure is public
3. **Performance**: Large numbers of constraints may exceed memory limits
4. **Static Circuit**: Requires recompilation for new layouts

## Future Improvements

1. **Dynamic Constraints**: Support for variable number of constraints
2. **Improved Performance**: Optimizations for EDDSA verification
3. **Better Error Handling**: More robust error handling and logging
4. **Extended Operations**: Additional constraint types and operations
5. **Multi-asset Support**: Support for multiple assets in a single proof
6. **Web API Interface**: RESTful API for easier integration

## Acknowledgments

This implementation uses the [gnark](https://github.com/Consensys/gnark) library for zk-SNARK operations.
