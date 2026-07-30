# Integration Test Results

Manual integration testing was performed to verify the service works as intended.

## Test Environment
- Service started successfully on port 8080
- Health check endpoint: `GET /health` → "OK" ✅

## Test Results

### 1. POST /v1/transfers - Success ✅
```bash
curl -X POST http://localhost:8080/v1/transfers \
  -H "Idempotency-Key: test-key-1" \
  -H "Content-Type: application/json" \
  -d '{"from_account":"treasury-01","to_account":"investor-42","asset_id":"gold-lot-2026-001","quantity_units":250}'
```
**Result**: 201 Created with transfer record ✅
```json
{
  "id":"transfer-1785333696584833000",
  "idempotency_key":"test-key-1",
  "request":{"from_account":"treasury-01","to_account":"investor-42","asset_id":"gold-lot-2026-001","quantity_units":250},
  "status":"submitted",
  "gateway_ref":"gw-ref-1",
  "created_at":"2026-07-29T17:01:36.584835+03:00",
  "updated_at":"2026-07-29T17:01:36.584873+03:00"
}
```

### 2. GET /v1/transfers/{id} - Success ✅
```bash
curl -X GET http://localhost:8080/v1/transfers/transfer-1785333696584833000
```
**Result**: 200 OK with transfer record ✅

### 3. Idempotency - Same Key, Same Payload ✅
```bash
curl -X POST http://localhost:8080/v1/transfers \
  -H "Idempotency-Key: test-key-1" \
  -H "Content-Type: application/json" \
  -d '{"from_account":"treasury-01","to_account":"investor-42","asset_id":"gold-lot-2026-001","quantity_units":250}'
```
**Result**: 201 Created with same transfer ID (cached result) ✅

### 4. Idempotency - Same Key, Different Payload ✅
```bash
curl -X POST http://localhost:8080/v1/transfers \
  -H "Idempotency-Key: test-key-1" \
  -H "Content-Type: application/json" \
  -d '{"from_account":"treasury-01","to_account":"investor-99","asset_id":"gold-lot-2026-001","quantity_units":250}'
```
**Result**: 409 Conflict with error code "PAYLOAD_CONFLICT" ✅

### 5. Error Cases ✅

**Missing Idempotency Key:**
```bash
curl -X POST http://localhost:8080/v1/transfers \
  -H "Content-Type: application/json" \
  -d '{"from_account":"treasury-01","to_account":"investor-42","asset_id":"gold-lot-2026-001","quantity_units":250}'
```
**Result**: Error code "IDEMPOTENCY_KEY_MISSING" ✅

**Same Source and Destination:**
```bash
curl -X POST http://localhost:8080/v1/transfers \
  -H "Idempotency-Key: test-key-2" \
  -H "Content-Type: application/json" \
  -d '{"from_account":"treasury-01","to_account":"treasury-01","asset_id":"gold-lot-2026-001","quantity_units":250}'
```
**Result**: Error code "INVALID_REQUEST" ✅

**Negative Quantity:**
```bash
curl -X POST http://localhost:8080/v1/transfers \
  -H "Idempotency-Key: test-key-3" \
  -H "Content-Type: application/json" \
  -d '{"from_account":"treasury-01","to_account":"investor-42","asset_id":"gold-lot-2026-001","quantity_units":-5}'
```
**Result**: Error code "INVALID_REQUEST" ✅

**Missing From Account:**
```bash
curl -X POST http://localhost:8080/v1/transfers \
  -H "Idempotency-Key: test-key-4" \
  -H "Content-Type: application/json" \
  -d '{"from_account":"","to_account":"investor-42","asset_id":"gold-lot-2026-001","quantity_units":250}'
```
**Result**: Error code "INVALID_REQUEST" ✅

**Unknown Fields:**
```bash
curl -X POST http://localhost:8080/v1/transfers \
  -H "Idempotency-Key: test-key-5" \
  -H "Content-Type: application/json" \
  -d '{"from_account":"treasury-01","to_account":"investor-42","asset_id":"gold-lot-2026-001","quantity_units":250,"extra_field":"value"}'
```
**Result**: Error "Invalid JSON body or unknown fields" ✅

**Non-existent Transfer:**
```bash
curl -X GET http://localhost:8080/v1/transfers/non-existent-id
```
**Result**: Error code "TRANSFER_NOT_FOUND" ✅

## Summary

All integration tests passed successfully:
- ✅ Service starts and health check works
- ✅ POST /v1/transfers creates transfers successfully
- ✅ GET /v1/transfers/{id} retrieves transfers successfully
- ✅ Idempotency works correctly (same key/payload returns cached result)
- ✅ Payload conflict detection works (same key/different payload returns 409)
- ✅ All validation errors work correctly
- ✅ Unknown fields are rejected
- ✅ Non-existent transfers return 404

The service is working as intended and ready for submission.
