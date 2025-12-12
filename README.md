# Dual-lock integrity auditing for 3D-printed products on Hyperledger Fabric

> **Dual-lock integrity auditing system for 3D-printed products  
> implemented on Hyperledger Fabric using Go chaincode and Go microservices.**

This repository contains a reference implementation of the SAM-BCADA system on top of the official [`hyperledger/fabric-samples`](https://github.com/hyperledger/fabric-samples) test network.

SAM-BCADA combines:

- A **physical lock**: product-bound crypto anchor derived from a device fingerprint (DFP).
- A **digital lock**: encrypted design files stored off-chain with **homomorphic verifiable tags** (HVT) anchored on-chain and periodically audited.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Repository Layout](#repository-layout)
- [Prerequisites](#prerequisites)
- [Setup & Installation](#setup--installation)
  - [1. Clone `fabric-samples` and add this folder](#1-clone-fabric-samples-and-add-this-folder)
  - [2. Start the Fabric test network](#2-start-the-fabric-test-network)
  - [3. Deploy the SAM-BCADA chaincode](#3-deploy-the-sam-bcada-chaincode)
  - [4. Build and run the off-chain services](#4-build-and-run-the-off-chain-services)
- [Using the System](#using-the-system)
  - [Roles](#roles)
  - [Manufacturer Flow](#manufacturer-flow)
  - [Verifier Flow](#verifier-flow)
  - [Auditor Flow](#auditor-flow)
  - [Download / Access Flow](#download--access-flow)
- [Chaincode API](#chaincode-api)
- [Off-Chain Services](#off-chain-services)
  - [Coordinator Service](#coordinator-service)
  - [Storage Node Service](#storage-node-service)
- [Configuration](#configuration)
- [Development Notes](#development-notes)
- [Teardown](#teardown)
- [License](#license)

---

## Architecture Overview

High-level components:

- **Hyperledger Fabric Network**  
  - Standard `test-network` from `fabric-samples`  
  - Chaincode: `sambcada` (Go)

- **On-chain (digital lock)**  
  - Crypto anchors: encrypted device fingerprint + signature + metadata  
  - Block tags: RSA-based homomorphic verifiable tags for encrypted file blocks  
  - Audit results: integrity audit outcomes  
  - Download logs: accountability for file access

- **Off-chain**  
  - **Coordinator** (Go, REST):  
    - Receives design files from manufacturers  
    - Encrypts and chunks files  
    - Computes/verifies homomorphic tags  
    - Talks to Fabric chaincode and Storage Nodes  
    - Coordinates integrity audits and downloads
  - **Storage Node** (Go, REST):  
    - Stores encrypted file blocks  
    - Returns blocks for audits and downloads

Data & trust flow (simplified):

1. Manufacturer registers product → **crypto anchor** on Fabric.
2. Coordinator receives G-code / design → encrypts, chunks, tags → tags on Fabric, blocks on Storage Node.
3. Verifier checks live product fingerprint against anchor.
4. Auditor runs periodic HVT-based audits across Storage Node(s), logging results on Fabric.
5. Authorized users download encrypted file via Coordinator; each access is logged on-chain.

---

## Repository Layout

Assuming this project lives at:  

`fabric-samples/sam-bcada/`

```text
sam-bcada/
  chaincode-go/
    go.mod
    go.sum
    sambcada.go          # Go chaincode implementing the dual-lock logic

  app/
    common/
      crypto.go          # Crypto helpers (hashing, RSA/HVT helpers – extend as needed)

    coordinator/
      main.go            # Coordinator REST service (upload, audit, download – extend)
      Dockerfile

    storage-node/
      main.go            # Storage Node REST service (store/get blocks)
      Dockerfile

  docker-compose.sambcada.yaml
  README.md
```

## Prerequisites

Make sure you have:

- **Docker & Docker Compose**
- **Go ≥ 1.21**
- **Hyperledger Fabric binaries & images**  
  (as required by `fabric-samples`)
- **`fabric-samples` repository cloned**:

  ```bash
  git clone https://github.com/hyperledger/fabric-samples.git
  ```

## Setup & Installation

### 1. Clone fabric-samples and add this folder

From your workspace:
```bash
git clone https://github.com/hyperledger/fabric-samples.git
cd fabric-samples
```

Place this project folder as:

```bash
fabric-samples/sam-bcada/
```

So that the chaincode file is located at:
```bash
fabric-samples/sam-bcada/chaincode-go/sambcada.go
```

### 2. Start the Fabric test network

From fabric-samples/test-network:
```bash
cd test-network

# Add Fabric binaries to PATH (if not already)
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config

# Start the test network and create "mychannel"
./network.sh up createChannel
```

This brings up:
  - Orderer
  - Two orgs (Org1, Org2)
  - Peer nodes

Channel mychannel

### 3. Deploy the SAM-BCADA chaincode

Still in fabric-samples/test-network:

```bash  

./network.sh deployCC \
  -ccn sambcada \
  -ccp ../sam-bcada/chaincode-go \
  -ccl go
```

This:

Packages the Go chaincode in `../sam-bcada/chaincode-go`

Deploys it as sambcada to mychannel

### 4. Build and run the off-chain services

From fabric-samples/sam-bcada:
```bash 
cd ../sam-bcada
```

# Build Docker images for coordinator & storage node
docker-compose -f docker-compose.sambcada.yaml build

# Start the services in the background
docker-compose -f docker-compose.sambcada.yaml up -d


After it’s up:

```bash 
Coordinator: http://localhost:8080

Storage Node: http://localhost:4000
```

The `docker-compose.sambcada.yaml` file re-uses the test Docker network created by test-network.
