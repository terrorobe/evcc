# evcc Remote Access Reverse Proxy

### 1. Connection Protocol
**DECISION**: WebSocket tunnel for evcc → Proxy, SNI-based TCP forwarding for Mobile → evcc
- evcc opens WebSocket tunnel to proxy (outbound, firewall-friendly)
- Mobile app uses standard HTTPS to subdomain (e.g., abc123.evcc.example)
- Proxy routes based on SNI in TLS handshake, forwards raw TCP
- Enables standard HTTP clients while maintaining end-to-end TLS

**WebSocket Connection Architecture:**
- **Control Connection**: Separate WebSocket for enrollment, status, ACME challenges (`/register`, `/connect`)
- **Tunnel Connection**: Single WebSocket for multiplexed client traffic forwarding
- **Connection Multiplexing**: Multiple concurrent mobile client connections share one WebSocket tunnel using connection IDs
- **Message Format**: `{connection_id, type: "connect|data|close", data: bytes}` for tunnel multiplexing

### 2. Encryption Implementation
**DECISION**: End-to-end TLS with Let's Encrypt via challenge forwarding
- evcc instance initiates Let's Encrypt certificate requests for its assigned subdomain
- Proxy forwards ACME challenges from Let's Encrypt to evcc via WebSocket tunnel
- evcc handles challenge response and receives certificate + private key locally
- Mobile app connects via HTTPS to abc123.evcc.example with trusted Let's Encrypt cert
- Proxy forwards raw TCP after SNI routing - cannot decrypt traffic
- All sensitive certificate data remains on evcc instance

### 3. Sponsor Token Validation
Sponsor token validation via gRPC to `sponsor.evcc.io:8080`
- gRPC call: `Auth.IsAuthorized(token)` returns `{authorized, subject, expires_at}`
- Proxy reuses same validation endpoint during instance registration

### 4. Proxy State Storage
**Instance registry**: `{subdomain, sponsor_token, last_connected}` per registered instance
- Quota enforcement: count instances per sponsor_token
- Dead system pruning: expire if last_connected > 90 days (Let's Encrypt cert lifetime)
- Conflict prevention: unique subdomain constraint

**Runtime state**: Active routing table (subdomain → WebSocket connection)

### 5. evcc Configuration
**NEEDS DEFINITION**: Config structure for proxy settings

## Technical Components

### Central Proxy Service
- Subdomain routing (*.evcc.example -> instance connections)
- Sponsor token validation for enrollment
- ACME challenge forwarding for Let's Encrypt certificates
- SNI-based TCP forwarding to WebSocket tunnels with connection multiplexing
- Instance registry: `{subdomain, sponsor_token, last_connected}`
- Connection multiplexer: maps mobile client connections to WebSocket tunnel connection IDs

### evcc Changes
- Let's Encrypt certificate management for assigned subdomain (certificates stored in database)
- HTTPS endpoint support
- Operational mode requiring authentication for all interface/API requests
- Proxy instance enrollment
- Control WebSocket connection for enrollment and ACME challenges
- Tunnel WebSocket connection with multiplexing support for concurrent client connections
- Connection manager to handle multiplexed inbound connections over single WebSocket
- QR code generation for mobile client enrollment
- Proxy configuration in settings

### Mobile App Changes
- QR scanner for instance enrollment
- Standard HTTPS API calls to subdomain endpoints
- Local connection fallback

## Open Questions

1. **QR Code Format**: What exact data goes in the QR code? (endpoint URL + JWT token)
2. **Mobile Client Token Management**: How long should mobile app JWT tokens be valid and how to handle expiry/refresh? (current web UI uses 90 days)
3. **Let's Encrypt Rate Limits**: Key constraints - 50 certificates per registered domain (`*.evcc.example`) per 7 days, 5 certificates per exact identifier set per 7 days. How does this affect subdomain assignment and quota planning?
4. **Static Asset Protection**: How to serve/protect static assets for browser clients to limit DoS potential? Basic auth prompt that sets a cookie on success? Login form? Something else?
5. **Proxy Implementation Location**: Should the proxy implementation live directly in the evcc repository or in a separate repository?
6. **Quota Limits**: How many instances per sponsor token?
7. **Subdomain Assignment**: Auto-generate or user-specified?
8. **Dead System Cleanup**: What's the exact cleanup policy for expired instances? (should be >90 days to allow for certificate renewal)

## Core Flows

### 1. Initial Instance Enrollment

```
1. evcc opens WebSocket to wss://proxy.evcc.example/register
2. Sends: { "sponsor_token": "abc123", "requested_subdomain": "my-home" }
3. Proxy validates sponsor token via gRPC to sponsor.evcc.io:8080
4. Proxy checks subdomain availability and quota
5. Proxy responds: { "fqdn": "my-home-evcc.evcc.example", "status": "registered" }
6. Proxy creates temporary routing entry: my-home-evcc.evcc.example → registration WebSocket

7. evcc starts Let's Encrypt ACME process:
   a. GET https://acme-v02.api.letsencrypt.org/directory to get endpoint URLs
   b. POST /acme/new-account with termsOfServiceAgreed: true
   c. POST /acme/new-order with identifiers: [{"type": "dns", "value": "my-home-evcc.evcc.example"}]
   d. Receive order with authorization URL for the domain

7. evcc fetches authorization object:
   a. POST-as-GET to authorization URL
   b. Receives challenges including HTTP-01 challenge with token

8. evcc sends challenge response:
   a. POST to challenge URL with empty payload: "{}"
   b. Let's Encrypt validation server makes HTTP request to:
      GET https://my-home-evcc.evcc.example/.well-known/acme-challenge/<token>

9. Proxy receives ACME challenge HTTP request for my-home-evcc.evcc.example and forwards to evcc via the active registration WebSocket:
   {
     "type": "acme_challenge",
     "request_id": "req_12345",
     "token": "LoqXcYV8q5ONbJQxbmR7SCTNo3tiAXDfowyjxAjEuX0"
   }

10. evcc responds with key authorization via WebSocket:
    {
      "type": "acme_response",
      "request_id": "req_12345",
      "key_authorization": "LoqXcYV8q5ONbJQxbmR7SCTNo3tiAXDfowyjxAjEuX0.9jg46WB3rQiXaXtqL_2gCO4_SoKAjHPgUOy21mqTI"
    }

11. Proxy returns key authorization to Let's Encrypt validation server as HTTP response

12. Let's Encrypt validates response and marks authorization as "valid"

13. evcc finalizes order:
    a. POST CSR to order's finalize URL
    b. Let's Encrypt issues certificate
    c. evcc downloads certificate via POST-as-GET to certificate URL
    d. evcc stores certificate and private key in database

14. evcc closes registration WebSocket (enrollment complete)

15. Proxy removes temporary routing entry when WebSocket closes

16. Registration complete - instance must reconnect via /connect to establish permanent routing
```

### 2. Instance Reconnection

```
1. evcc opens WebSocket to wss://proxy.evcc.example/connect with client certificate
2. Uses its Let's Encrypt certificate for my-home-evcc.evcc.example as client cert
3. Proxy validates certificate and extracts subdomain from cert (my-home-evcc.evcc.example → my-home-evcc)
4. Proxy creates permanent routing table entry and updates last_connected timestamp
5. Connection established - mobile traffic can flow through permanent tunnel
```

### 3. Mobile Client Enrollment

```
1. User accesses evcc web UI via local network
2. User logs in with admin password
3. User generates QR code for mobile app (protected by auth middleware)
4. evcc generates JWT token with long expiry (e.g., 90 days)
5. QR code contains: { "endpoint": "https://my-home-evcc.evcc.example", "token": "jwt-token-xyz" }
6. Mobile app scans QR code, validates SSL certificate (Let's Encrypt)
7. Mobile app makes test API call with Authorization: Bearer jwt-token-xyz
8. evcc validates JWT token using standard auth middleware
9. Mobile app stores endpoint and token for future requests
```

### 4. API Client Authentication

```
1. API client makes HTTPS request to https://my-home-evcc.evcc.example/api/state
2. Request includes: Authorization: Bearer jwt-token-xyz
3. Proxy routes via SNI to correct WebSocket tunnel (raw TCP forwarding)
4. evcc receives request, ensureAuthHandler middleware validates JWT token
5. evcc processes request, returns response
```

### 5. Request Authentication

**Current evcc Authentication System:**
- **Admin Web UI**: JWT tokens in HTTP-only cookies (`auth` cookie name)
- **API Clients**: JWT tokens in `Authorization: Bearer <token>` header
- **Auth Middleware**: `ensureAuthHandler` validates JWT tokens for protected routes
- **JWT Secret**: Stored in settings, auto-generated on first use

**Proxy Authentication Handling:**
- Proxy routes raw TCP based on SNI - cannot see or modify HTTP content
- evcc handles all authentication exactly as it does today
- JWT tokens work unchanged through the tunnel
- No proxy modifications needed to authentication system
