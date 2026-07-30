# Deployment Guide

## Running the Service

### Prerequisites
- Go 1.21 or higher
- Make (optional, for build automation)

### Local Development

#### Run the Service
```bash
go run cmd/api/main.go
```

The service will start on port 8080 (configurable via PORT environment variable).

#### Set Custom Port
```bash
PORT=3000 go run cmd/api/main.go
```

### Build for Production

#### Build Binary
```bash
go build -o asset-transfer-app cmd/api/main.go
```

#### Run Binary
```bash
./asset-transfer-app
```

#### Build for Different Platforms
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o asset-transfer-app-linux cmd/api/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o asset-transfer-app-macos cmd/api/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o asset-transfer-app.exe cmd/api/main.go
```

### Docker Deployment

#### Build Docker Image
```bash
docker build -t asset-transfer-app .
```

#### Run Docker Container
```bash
docker run -p 8080:8080 asset-transfer-app
```

#### Run with Custom Port
```bash
docker run -p 3000:8080 -e PORT=8080 asset-transfer-app
```

### Environment Variables

- `PORT`: Server port (default: 8080)
- `LOG_LEVEL`: Logging level (debug, info, warn, error) - optional

### Health Check

The service provides a health check endpoint:

```bash
curl http://localhost:8080/health
```

Expected response: `OK`

### Production Considerations

#### Current Limitations
- In-memory storage (data lost on restart)
- No persistence layer
- Single instance only (no horizontal scaling)
- No authentication/authorization
- No rate limiting
- No request logging

#### Production Requirements
For production deployment, you would need to:

1. **Add Persistence Layer**
   - Replace in-memory storage with PostgreSQL
   - Implement proper database migrations
   - Add connection pooling

2. **Add Authentication**
   - Implement API key authentication
   - Add JWT token support
   - Use HTTPS/TLS

3. **Add Monitoring**
   - Implement structured logging
   - Add metrics collection (Prometheus)
   - Set up alerting

4. **Add Rate Limiting**
   - Implement rate limiting per client
   - Add request throttling

5. **Add Horizontal Scaling**
   - Use shared storage (PostgreSQL)
   - Implement distributed locking (Redis)
   - Add load balancer

6. **Add Security**
   - Input validation and sanitization
   - SQL injection prevention
   - XSS protection
   - CSRF protection

### Monitoring

#### Basic Health Monitoring
```bash
# Check if service is running
curl http://localhost:8080/health

# Check response time
time curl http://localhost:8080/health
```

#### Log Monitoring
Logs are currently written to stdout. For production:
- Use structured logging (JSON format)
- Centralize logs (ELK stack, CloudWatch, etc.)
- Set up log rotation

### Troubleshooting

#### Service Won't Start
- Check if port 8080 is already in use
- Verify Go installation
- Check for missing dependencies

#### Connection Refused
- Verify service is running
- Check firewall settings
- Verify port configuration

#### High Memory Usage
- Check for memory leaks
- Monitor in-memory storage size
- Consider implementing storage limits
