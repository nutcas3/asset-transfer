# API Documentation

## POST /v1/transfers

Create a new transfer with idempotency protection.

### Headers
- `Idempotency-Key`: Required client-generated key

### Request Body
```json
{
  "from_account": "treasury-01",
  "to_account": "investor-42",
  "asset_id": "gold-lot-2026-001",
  "quantity_units": 250
}
```

### Request Validation
- `from_account`: Required string - Source account identifier
- `to_account`: Required string - Destination account identifier
- `asset_id`: Required string - Asset identifier
- `quantity_units`: Required integer - Positive number of units to transfer
- Source and destination accounts must be different

### Response Codes
- `201 Created`: Transfer submitted successfully
- `400 Bad Request`: Invalid request (missing fields, validation errors)
- `409 Conflict`: Idempotency key reused with different payload
- `422 Unprocessable Entity`: Gateway deterministically rejected the transfer
- `503 Service Unavailable`: Gateway unavailable, timeout, or circuit open

### Response Body (Success)
```json
{
  "id": "transfer-1234567890",
  "idempotency_key": "client-key-1",
  "request": {
    "from_account": "treasury-01",
    "to_account": "investor-42",
    "asset_id": "gold-lot-2026-001",
    "quantity_units": 250
  },
  "status": "submitted",
  "gateway_ref": "gw-ref-1",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:01Z"
}
```

### Response Body (Error)
```json
{
  "code": "IDEMPOTENCY_KEY_MISSING",
  "message": "Idempotency key is required",
  "layer": "service"
}
```

### Transfer Status Values
- `submitting`: Transfer is being processed
- `submitted`: Transfer successfully submitted to gateway
- `rejected`: Gateway deterministically rejected the transfer
- `retryable_failure`: Gateway unavailable, timeout, or circuit open (retryable)

### Idempotency Behavior
- First request with an idempotency key creates a new transfer
- Subsequent requests with the same key and payload return the cached result
- Requests with the same key but different payload return 409 Conflict
- Idempotency is enforced at the service level with singleflight pattern

## GET /v1/transfers/{id}

Retrieve a transfer by ID.

### URL Parameters
- `id`: Transfer identifier (path parameter)

### Response Codes
- `200 OK`: Transfer found
- `404 Not Found`: Transfer not found

### Response Body (Success)
```json
{
  "id": "transfer-1234567890",
  "idempotency_key": "client-key-1",
  "request": {
    "from_account": "treasury-01",
    "to_account": "investor-42",
    "asset_id": "gold-lot-2026-001",
    "quantity_units": 250
  },
  "status": "submitted",
  "gateway_ref": "gw-ref-1",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:01Z"
}
```

### Response Body (Error)
```json
{
  "code": "TRANSFER_NOT_FOUND",
  "message": "Transfer not found",
  "layer": "service"
}
```

## Error Codes

### Service Layer Errors
- `INVALID_REQUEST`: Invalid request parameters
- `IDEMPOTENCY_KEY_MISSING`: Idempotency key is required
- `PAYLOAD_CONFLICT`: Idempotency key already used with different payload
- `TRANSFER_NOT_FOUND`: Transfer not found

### Gateway Layer Errors
- `GATEWAY_TIMEOUT`: Gateway timeout
- `GATEWAY_UNAVAILABLE`: Gateway unavailable
- `GATEWAY_REJECTED`: Gateway rejected transfer
- `CIRCUIT_OPEN`: Circuit breaker is open

### Storage Layer Errors
- `STORAGE_ERROR`: Storage operation error

## Health Check

### GET /health

Simple health check endpoint.

### Response Codes
- `200 OK`: Service is healthy

### Response Body
```
OK
```
