# SAM-BCADA on Hyperledger Fabric (Go)

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
