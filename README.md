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

## Build Docker images for coordinator & storage node
```bash
docker-compose -f docker-compose.sambcada.yaml build
```

### Start the services in the background
```bash
docker-compose -f docker-compose.sambcada.yaml up -d
```

After it’s up:

```bash 
Coordinator: http://localhost:8080

Storage Node: http://localhost:4000
```

The `docker-compose.sambcada.yaml` file re-uses the test Docker network created by `test-network`.


## Related Publications

Below are selected publications related to SHIELD-CAN, cyber-physical security, additive manufacturing security, threat intelligence, and trusted AI systems. These give additional background and context around secure architectures, intrusion detection, self-healing, and risk assessment.

### Automotive, VANET & Cyber-Physical / ICS Security

- **[Secure and Anonymous Batch Authentication and Key Exchange Protocols for 6G Enabled VANETs](https://ieeexplore.ieee.org/document/10972137/)**  
  **Mahender Kumar** and Carsten Maple
  *IEEE Transactions on Intelligent Transportation Systems, 2025*

- **[ICSThreatQA: A Knowledge-Graph Enhanced Question Answering Model for Industrial Control System Threat Intelligence](https://www.sciencedirect.com/science/article/abs/pii/S0957417425037959)**  
  Ruby Rani, **Mahender Kumar**, Gregory Epiphaniou and Carsten Maple  
  *Expert Systems with Applications, 2025*.  

- **[Securing connected and autonomous vehicles: analysing attack methods, mitigation strategies, and the role of large language models](https://digital-library.theiet.org/doi/abs/10.1049/icp.2024.2534)**  
  **Mahender Kumar**, Ruby Rani, Gregory Epiphaniou, Carsten Maple
  *IET Conference Proceedings, 2024*.  

### Additive Manufacturing & Industrial Security

- **[Security of cyber-physical Additive Manufacturing supply chain: Survey, attack taxonomy and solutions](https://www.sciencedirect.com/science/article/pii/S0167404825002469)**  
  **Mahender Kumar**, Gregory Epiphaniou, Carsten Maple
  *Computers & Security, 2025: 104557*.  

- **[SPM-SeCTIS: Severity Pattern Matching for Secure Computable Threat Information Sharing in Intelligent Additive Manufacturing](id.elsevier.com/as/authorization.oauth2?platSite=SD%2Fscience&additionalPlatSites=GH%2Fgeneralhospital%2CMDY%2Fmendeley%2CSC%2Fscopus%2CRX%2Freaxys&scope=openid%20email%20profile%20els_auth_info%20els_idp_info%20els_idp_analytics_attrs%20urn%3Acom%3Aelsevier%3Aidp%3Apolicy%3Aproduct%3Ainst_assoc&response_type=code&redirect_uri=https%3A%2F%2Fwww.sciencedirect.com%2Fuser%2Fidentity%2Flanding&authType=SINGLE_SIGN_IN&prompt=none&client_id=SDFE-v4&state=retryCounter%3D0%26csrfToken%3De0f7da65-4312-406d-b8b1-5d0077a91e35%26idpPolicy%3Durn%253Acom%253Aelsevier%253Aidp%253Apolicy%253Aproduct%253Ainst_assoc%26returnUrl%3D%252Fscience%252Farticle%252Fpii%252FS2542660524002750%26prompt%3Dnone%26cid%3Darp-3211915e-c7b0-4f78-a79d-8fc3c10269d1)**
  **Mahender Kumar, Gregory Epiphaniou, Carsten Maple**  
  *Internet of Things, 28 (2024): 101334*.  

- **[Comprehensive threat analysis in additive manufacturing supply chain: a hybrid qualitative and quantitative risk assessment framework](https://link.springer.com/article/10.1007/s11740-024-01283-1)**  
  **Mahender Kumar**, Gregory Epiphaniou, Carsten Maple
  *Production Engineering, 18(6): 955–973, 2024*.  

- **[Leveraging Semantic Relationships to Prioritise Indicators of Compromise in Additive Manufacturing Systems](https://link.springer.com/chapter/10.1007/978-3-031-41181-6_18)**  
  **Mahender Kumar**, Gregory Epiphaniou, Carsten Maple
  *International Conference on Applied Cryptography and Network Security, 2023*. 



### Healthcare, IoMT & Blockchain Systems

- **[A Provable Secure and Lightweight Smart Healthcare Cyber-Physical System With Public Verifiability](https://ieeexplore.ieee.org/document/9624169)**  
  **Mahender Kumar**, Satish Chand
  *IEEE Systems Journal, 16(4): 5501–5508, 2022*.  

- **[A Secure and Efficient Cloud-Centric Internet-of-Medical-Things-Enabled Smart Healthcare System With Public Verifiability](https://ieeexplore.ieee.org/document/9131770)**  
  **Mahender Kumar**, Satish Chand
  *IEEE Internet of Things Journal, 7(10): 10457–10465, 2020*.  

- **[MedHypChain: A patient-centered interoperability hyperledger-based medical healthcare system: Regulation in COVID-19 pandemic](https://www.sciencedirect.com/science/article/pii/S1084804521000023)**  
  **Mahender Kumar**, Satish Chand
  *Journal of Network and Computer Applications, 179:102975, 2021*.  

- **[A Lightweight Cloud-Assisted Identity-Based Anonymous Authentication and Key Agreement Protocol for Secure Wireless Body Area Network](https://ieeexplore.ieee.org/document/9099043)**  
  **Mahender Kumar**, Satish Chand
  *IEEE Systems Journal, 15(2): 1646–1657, 2021*.  


### Cryptography, Ontologies & Threat Intelligence Foundations

- **[Science and Technology Ontology: A Taxonomy of Emerging Topics](https://arxiv.org/abs/2305.04055)**  
  **Mahender Kumar**, Ruby Rani, Mirko Botarelli, Gregory Epiphaniou, Carsten Maple
  *arXiv preprint arXiv:2305.04055, 2023*.  

- **[Pairing for Greenhorn: Survey and Future Perspective](https://arxiv.org/abs/2108.12392)**  
  **Mahender Kumar**, Satish Chand
  arXiv preprint, 2021.  

For a complete and up-to-date list of publications, please refer to my full publication list or Google Scholar profile. [Google scholar](https://scholar.google.com/citations?hl=en&user=Ppmct6EAAAAJ&view_op=list_works&sortby=pubdate)

## Citing

If you use this work, please cite the corresponding work:

```bash
@article{kumar2025securing,
  title={Securing additive manufacturing with blockchain-based cryptographic anchoring and dual-lock integrity auditing},
  author={Kumar, Mahender and Epiphaniou, Gregory and Maple, Carsten},
  journal={Computers in Industry},
  volume={173},
  pages={104395},
  year={2025},
  publisher={Elsevier}
}
```

## Contributor 
  - [Mahender Kumar](https://scholar.google.com/citations?user=Ppmct6EAAAAJ&hl=en)
  - [Gregory Epiphaniou](https://warwick.ac.uk/fac/sci/wmg/about/our-people/profile/?wmgid=2175)
  - [Carsten Maple](https://warwick.ac.uk/fac/sci/wmg/about/our-people/profile/?wmgid=1102)
