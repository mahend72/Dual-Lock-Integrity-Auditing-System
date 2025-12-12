package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ===== Data models =====

type CryptoAnchor struct {
	UID           string `json:"uid"`
	DFPEnc        string `json:"dfpEnc"`        // base64 of encrypted DFP
	Sign          string `json:"sign"`         // base64 signature over token(UID || DFP)
	Manufacturer  string `json:"manufacturer"` // manufacturer id
	VerifierPKPem string `json:"verifierPk"`   // verifier public key PEM
	CreatedAt     string `json:"createdAt"`
}

type VerificationRecord struct {
	ID         string `json:"id"`
	UID        string `json:"uid"`
	VerifierID string `json:"verifierId"`
	Status     string `json:"status"`  // AUTHENTIC / COUNTERFEIT
	Comment    string `json:"comment"` // optional description
	Timestamp  string `json:"timestamp"`
}

type TagRecord struct {
	ID         string `json:"id"`
	UID        string `json:"uid"`
	FileID     string `json:"fileId"`
	BlockIndex int    `json:"blockIndex"`
	TagValue   string `json:"tagValue"` // hex or base64 of g^d_i mod N
	CreatedAt  string `json:"createdAt"`
}

type AuditRecord struct {
	ID        string `json:"id"`
	FileID    string `json:"fileId"`
	UID       string `json:"uid"`
	ProofHash string `json:"proofHash"` // hash(proof)
	Mu        string `json:"mu"`        // μ
	Status    string `json:"status"`    // SUCCESS / MALICIOUS
	Timestamp string `json:"timestamp"`
}

type DownloadRecord struct {
	ID          string `json:"id"`
	FileID      string `json:"fileId"`
	UID         string `json:"uid"`
	UserID      string `json:"userId"`
	Allowed     bool   `json:"allowed"`
	RequestHash string `json:"requestHash"` // hash of Enc({FID, allowed, U})
	Timestamp   string `json:"timestamp"`
}

// ===== Chaincode struct =====

type SmartContract struct {
	contractapi.Contract
}

// ===== Helpers =====

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func assetExists(ctx contractapi.TransactionContextInterface, key string) (bool, error) {
	b, err := ctx.GetStub().GetState(key)
	if err != nil {
		return false, err
	}
	return b != nil && len(b) > 0, nil
}

// ===== 1. Crypto Anchor =====

// CreateCryptoAnchor stores [DFP_enc, Sign, UID] into the ledger.
func (s *SmartContract) CreateCryptoAnchor(
	ctx contractapi.TransactionContextInterface,
	uid string,
	dfpEncBase64 string,
	signBase64 string,
	manufacturerID string,
	verifierPKPem string,
) error {

	key := "ANCHOR_" + uid
	exists, err := assetExists(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("anchor already exists for UID %s", uid)
	}

	anchor := CryptoAnchor{
		UID:           uid,
		DFPEnc:        dfpEncBase64,
		Sign:          signBase64,
		Manufacturer:  manufacturerID,
		VerifierPKPem: verifierPKPem,
		CreatedAt:     nowISO(),
	}

	bytes, err := json.Marshal(anchor)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(key, bytes)
}

// GetCryptoAnchor returns the anchor as JSON string.
func (s *SmartContract) GetCryptoAnchor(
	ctx contractapi.TransactionContextInterface,
	uid string,
) (string, error) {

	key := "ANCHOR_" + uid
	b, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", err
	}
	if b == nil || len(b) == 0 {
		return "", fmt.Errorf("anchor not found for UID %s", uid)
	}
	return string(b), nil
}

// ===== 2. Verification logs =====

// RecordVerificationResult logs the outcome of a verification.
func (s *SmartContract) RecordVerificationResult(
	ctx contractapi.TransactionContextInterface,
	uid string,
	verifierID string,
	status string,
	comment string,
) error {

	if status != "AUTHENTIC" && status != "COUNTERFEIT" {
		return fmt.Errorf("status must be AUTHENTIC or COUNTERFEIT")
	}

	anchorKey := "ANCHOR_" + uid
	exists, err := assetExists(ctx, anchorKey)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no anchor exists for UID %s", uid)
	}

	ts := nowISO()
	id := fmt.Sprintf("VER_%s_%s", uid, ts)

	rec := VerificationRecord{
		ID:         id,
		UID:        uid,
		VerifierID: verifierID,
		Status:     status,
		Comment:    comment,
		Timestamp:  ts,
	}

	bytes, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, bytes)
}

// GetVerificationHistory returns all verification records for a UID.
func (s *SmartContract) GetVerificationHistory(
	ctx contractapi.TransactionContextInterface,
	uid string,
) (string, error) {

	iterator, err := ctx.GetStub().GetStateByRange("VER_", "VES_")
	if err != nil {
		return "", err
	}
	defer iterator.Close()

	var list []VerificationRecord

	for iterator.HasNext() {
		res, err := iterator.Next()
		if err != nil {
			return "", err
		}
		var rec VerificationRecord
		if err := json.Unmarshal(res.Value, &rec); err != nil {
			return "", err
		}
		if rec.UID == uid {
			list = append(list, rec)
		}
	}

	out, err := json.Marshal(list)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ===== 3. Tags for encrypted blocks =====

// tagsJson is an array of { "blockIndex": int, "tagValue": string }
type tagInput struct {
	BlockIndex int    `json:"blockIndex"`
	TagValue   string `json:"tagValue"`
}

func (s *SmartContract) StoreBlockTags(
	ctx contractapi.TransactionContextInterface,
	uid string,
	fileID string,
	tagsJson string,
) error {

	var tags []tagInput
	if err := json.Unmarshal([]byte(tagsJson), &tags); err != nil {
		return fmt.Errorf("invalid tagsJson: %v", err)
	}

	now := nowISO()
	for _, t := range tags {
		id := fmt.Sprintf("TAG_%s_%d", fileID, t.BlockIndex)
		tag := TagRecord{
			ID:         id,
			UID:        uid,
			FileID:     fileID,
			BlockIndex: t.BlockIndex,
			TagValue:   t.TagValue,
			CreatedAt:  now,
		}
		b, err := json.Marshal(tag)
		if err != nil {
			return err
		}
		if err := ctx.GetStub().PutState(id, b); err != nil {
			return err
		}
	}
	return nil
}

// GetTagsForFile returns all tags for a given fileId.
func (s *SmartContract) GetTagsForFile(
	ctx contractapi.TransactionContextInterface,
	fileID string,
) (string, error) {

	iter, err := ctx.GetStub().GetStateByRange("TAG_", "TAH_")
	if err != nil {
		return "", err
	}
	defer iter.Close()

	var tags []TagRecord
	for iter.HasNext() {
		res, err := iter.Next()
		if err != nil {
			return "", err
		}
		var tr TagRecord
		if err := json.Unmarshal(res.Value, &tr); err != nil {
			return "", err
		}
		if tr.FileID == fileID {
			tags = append(tags, tr)
		}
	}

	out, err := json.Marshal(tags)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ===== 4. Audit results =====

func (s *SmartContract) StoreAuditResult(
	ctx contractapi.TransactionContextInterface,
	uid string,
	fileID string,
	proofHashHex string,
	muHex string,
	status string,
) error {

	if status != "SUCCESS" && status != "MALICIOUS" {
		return fmt.Errorf("status must be SUCCESS or MALICIOUS")
	}

	ts := nowISO()
	id := fmt.Sprintf("AUDIT_%s_%s", fileID, ts)

	rec := AuditRecord{
		ID:        id,
		FileID:    fileID,
		UID:       uid,
		ProofHash: proofHashHex,
		Mu:        muHex,
		Status:    status,
		Timestamp: ts,
	}

	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, b)
}

func (s *SmartContract) GetLatestAuditStatus(
	ctx contractapi.TransactionContextInterface,
	fileID string,
) (string, error) {

	iter, err := ctx.GetStub().GetStateByRange("AUDIT_", "AUE_")
	if err != nil {
		return "", err
	}
	defer iter.Close()

	var latest *AuditRecord
	for iter.HasNext() {
		res, err := iter.Next()
		if err != nil {
			return "", err
		}
		var rec AuditRecord
		if err := json.Unmarshal(res.Value, &rec); err != nil {
			return "", err
		}
		if rec.FileID != fileID {
			continue
		}
		if latest == nil || latest.Timestamp < rec.Timestamp {
			tmp := rec
			latest = &tmp
		}
	}

	if latest == nil {
		return "", fmt.Errorf("no audit record for fileId %s", fileID)
	}

	out, err := json.Marshal(latest)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ===== 5. Download logs =====

func (s *SmartContract) LogDownload(
	ctx contractapi.TransactionContextInterface,
	uid string,
	fileID string,
	userID string,
	allowed bool,
	requestHashHex string,
) error {

	ts := nowISO()
	id := fmt.Sprintf("DL_%s_%s_%s", fileID, userID, ts)

	rec := DownloadRecord{
		ID:          id,
		FileID:      fileID,
		UID:         uid,
		UserID:      userID,
		Allowed:     allowed,
		RequestHash: requestHashHex,
		Timestamp:   ts,
	}

	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, b)
}

func (s *SmartContract) GetDownloadHistory(
	ctx contractapi.TransactionContextInterface,
	fileID string,
) (string, error) {

	iter, err := ctx.GetStub().GetStateByRange("DL_", "DM_")
	if err != nil {
		return "", err
	}
	defer iter.Close()

	var list []DownloadRecord
	for iter.HasNext() {
		res, err := iter.Next()
		if err != nil {
			return "", err
		}
		var rec DownloadRecord
		if err := json.Unmarshal(res.Value, &rec); err != nil {
			return "", err
		}
		if rec.FileID == fileID {
			list = append(list, rec)
		}
	}

	out, err := json.Marshal(list)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ===== main =====

func main() {
	cc, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating SAM-BCADA chaincode: %v\n", err)
		return
	}

	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting SAM-BCADA chaincode: %v\n", err)
	}
}
