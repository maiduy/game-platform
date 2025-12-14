# Game Platform API - Postman Collections

This directory contains Postman collections and environment files for testing the Game Platform APIs with Bearer token authentication.

## Available Collections

- **MasterConfig API** - Character configuration and game data management
- **Ability Effects API** - Game ability and effect system management

## Files

- `MasterConfig_API.postman_collection.json` - MasterConfig service endpoints
- `Ability_Effects_API.postman_collection.json` - Ability & Effect service endpoints
- `MasterConfig_Dev.postman_environment.json` - Development environment configuration (shared)
- `test_token_generation.js` - Node.js script to test token generation
- `README.md` - This documentation file

## Quick Start

### 1. Import Collections and Environment

1. Open Postman
2. Click **Import** button (top left)
3. Drag and drop all JSON files:
   - `MasterConfig_API.postman_collection.json`
   - `Ability_Effects_API.postman_collection.json`
   - `MasterConfig_Dev.postman_environment.json`
4. Click **Import**

### 2. Select Environment

1. Click the environment dropdown (top right)
2. Select **MasterConfig - Dev Environment**

### 3. Start the Game Platform Service

```bash
cd /Users/duy.mai/Data/6.SourceCode/1.VNG/Game/DuAn-02/game-platform
export MASTERCONFIG_PORT=4002
go run cmd/game/main.go
```

### 4. Test the APIs

The collections are now ready to use! The token authentication is handled automatically by the pre-request script.

**Recommended first tests**:
- **MasterConfig**: Navigate to `Characters` → `Get Character by ID` and click **Send**
- **Ability Effects**: Navigate to `Effects` → `Create Effect - Fire DoT` and click **Send**

## Authentication Strategy

### Bearer Token Authentication

This API uses **Bearer token authentication** (RFC 6750) compatible with the Go implementation in `internal/platform/auth/auth.go:verifyGameToken`.

**Authentication Headers:**
1. **Authorization** (`Authorization: Bearer {token}`) - Bearer token for authentication
2. **X-API-Key** (`X-API-Key: {api_key}`) - API key for caller identification

### How It Works

The collection includes a **global pre-request script** that automatically:

1. Generates a token payload with required fields:
   - `uid` - User ID
   - `rid` - Request ID (auto-generated with timestamp)
   - `app_id` - Application ID (e.g., "3X")
   - `server_id` - Server ID (e.g., "server_01")
   - `ts` - Current Unix timestamp (in seconds)

2. Converts the payload to compact JSON (no whitespace)

3. Base64URL encodes the JSON payload (RFC 4648)

4. Computes HMAC-SHA256 signature on the JSON payload using the secret key

5. Encodes the signature as hexadecimal string

6. Combines into token format: `{base64url_payload}.{hex_signature}`

7. Stores the token in the `bearer_token` environment variable

8. Attaches the token to every request via the `Authorization: Bearer {token}` header

### Token Structure

```javascript
// Step 1: Token Payload (GameTokenPayload struct)
{
  "uid": "test_user_001",
  "rid": "req_1703001600000",
  "app_id": "3X",
  "server_id": "server_01",
  "ts": 1703001600
}

// Step 2: JSON String (compact, no whitespace)
'{"uid":"test_user_001","rid":"req_1703001600000","app_id":"3X","server_id":"server_01","ts":1703001600}'

// Step 3: Base64URL Encode
"eyJ1aWQiOiJ0ZXN0X3VzZXJfMDAxIiwicmlkIjoicmVxXzE3MDMwMDE2MDAwMDAiLCJhcHBfaWQiOiIzWCIsInNlcnZlcl9pZCI6InNlcnZlcl8wMSIsInRzIjoxNzAzMDAxNjAwfQ"

// Step 4: Compute HMAC-SHA256 Signature (hex)
"f5af08a2e54e489e6ee4b26bf25f73477900b1bd48802b0f3bcbef2168aa430f"

// Step 5: Final Bearer Token
"eyJ1aWQi...fQ.f5af08a2e54e489e6ee4b26bf25f73477900b1bd48802b0f3bcbef2168aa430f"

// Step 6: HTTP Header
"Authorization: Bearer eyJ1aWQi...fQ.f5af08a2e54e489e6ee4b26bf25f73477900b1bd48802b0f3bcbef2168aa430f"
```

### Go Implementation Compatibility

The token format matches the Go verification logic in `auth.go`:

```go
// From: internal/platform/auth/auth.go:387-443
func verifyGameToken(tokenString string) (*GameTokenPayload, error) {
    // Split token into payload and signature
    parts := strings.Split(tokenString, ".")  // Expects 2 parts
    payloadEncoded := parts[0]  // Base64URL encoded JSON
    signatureHex := parts[1]    // Hex-encoded HMAC-SHA256

    // Decode base64url payload
    payloadJSON, _ := base64.RawURLEncoding.DecodeString(payloadEncoded)

    // Verify HMAC-SHA256 signature
    expectedSignature := computeHMAC(payloadJSON, secret)
    // Signature match confirms authenticity
}
```

### Token Expiration

- **Dev Environment**: 300 seconds (5 minutes)
- **Clock Skew Tolerance**: 30 seconds
- **Future Time Allowance**: 10 seconds

The pre-request script generates a fresh token for every request, so you don't need to worry about expiration.

## Environment Variables

### Required Variables

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `base_url` | Base URL of the API | `http://localhost:4002` |
| `api_key` | API key for authentication | `game#123` |
| `app_id` | Application ID | `3X` |
| `secret` | Secret key for signing tokens | `Q!w2e3r4t5` |

### Optional Variables

| Variable | Description | Default Value |
|----------|-------------|---------------|
| `uid` | User ID | `test_user_001` |
| `server_id` | Server ID | `server_01` |
| `rid` | Request ID | Auto-generated |
| `rarity` | Character rarity filter | `SSR` |
| `element` | Character element filter | `Fire` |

### Auto-Generated Variables

| Variable | Description |
|----------|-------------|
| `auth_token` | Generated authentication token |
| `last_created_character_id` | ID of last created character |

## Available Endpoints

### MasterConfig - Characters

#### 1. Get Character by ID
```
GET /api/v1/masterconfig/characters/:character_id
```

**Description**: Retrieve a specific character by ID

**Path Parameters**:
- `character_id` (string) - Character unique identifier

**Example**:
```
GET /api/v1/masterconfig/characters/char_001
```

**Headers**:
- `X-API-Key`: {{api_key}}
- `X-Auth-Token`: {{auth_token}}
- `X-API-Version`: v1

---

#### 2. List Characters
```
GET /api/v1/masterconfig/characters
```

**Description**: Get paginated list of characters with filtering

**Query Parameters**:
- `page` (int) - Page number (default: 1)
- `page_size` (int) - Items per page (default: 20, max: 100)
- `rarity` (string) - Filter by rarity: R, SR, SSR
- `element` (string) - Filter by element: Fire, Ice, Light, Dark, Nature, Lightning, Water, Earth
- `gender` (string) - Filter by gender: Male, Female, Other
- `sort_by` (string) - Sort field (e.g., "created_at")
- `sort_order` (string) - Sort order: asc, desc

**Example**:
```
GET /api/v1/masterconfig/characters?page=1&page_size=20&rarity=SSR&element=Fire
```

---

#### 3. Create Character
```
POST /api/v1/masterconfig/characters
```

**Description**: Create a new character

**Request Body**:
```json
{
  "character_id": "char_new_001",
  "meta": {
    "nameKey": "character.warrior.name",
    "descriptionKey": "character.warrior.desc",
    "icon": "assets/icons/warrior.png",
    "model": "assets/models/warrior",
    "rarity": "SSR",
    "element": "Fire",
    "roles": ["Tank", "DPS"],
    "factions": ["Kingdom"],
    "gender": "Male",
    "tags": ["melee", "fire"]
  },
  "base_stats": {
    "level": 1,
    "maxLevel": 100,
    "hp": 1200,
    "atk": 180,
    "def": 95,
    "spd": 55,
    "crit": 0.18,
    "critDmg": 2.2,
    "accuracy": 0.96,
    "evasion": 0.12,
    "effectRes": 0.25
  },
  "growths": {
    "base": "growth_curve_ssr_1",
    "assension": "ascension_curve_ssr_1"
  },
  "assets": {
    "prefab": "assets/prefabs/warrior",
    "animations": "assets/animations/warrior",
    "portrait": "assets/portraits/warrior.png",
    "skill_vfx": "assets/vfx/skills"
  },
  "skill_set": ["skill_101", "skill_102"]
}
```

---

#### 4. Update Character
```
PUT /api/v1/masterconfig/characters/:character_id
```

**Description**: Update an existing character (partial update supported)

**Path Parameters**:
- `character_id` (string) - Character to update

**Request Body** (all fields optional):
```json
{
  "meta": {
    "rarity": "SSR",
    "tags": ["updated"]
  },
  "base_stats": {
    "hp": 1300,
    "atk": 190
  }
}
```

---

### Ability & Effect System - Effects

#### 1. Create Effect
```
POST /api/v1/ability/effects
```

**Description**: Create a new effect configuration

**Request Body**:
```json
{
  "effectId": "effect_fireball_dot",
  "category": "DamageOverTime",
  "name": "Fireball Burn",
  "order": 100,
  "effectType": "GenericEffect",
  "durationPolicy": 1,
  "duration": 5.0,
  "interval": 1.0,
  "stacks": 3,
  "tags": ["fire", "dot", "magical"],
  "modifiers": [
    {
      "name": "Attack",
      "flat": 50.0,
      "percent": 0.0
    }
  ],
  "metadata": {
    "vfx": "vfx_burning_loop",
    "sfx": "sfx_fire_burn",
    "tagsToApply": ["Burn"]
  }
}
```

**Effect Categories**:
- `Buff`, `Debuff`, `DamageOverTime`, `HealOverTime`, `Shield`, `StatusEffect`, `NegativeEffect`, `Passive`

**Duration Policies**:
- `0`: Instant (applied once, no duration)
- `1`: Duration (active for fixed time)
- `2`: Persistent (active until removed)

**Validation Rules**:
- `effectId`: Required, unique, max 50 chars, pattern `^[a-zA-Z0-9_]+$`
- `duration`, `interval`, `stacks`: Must be >= 0
- `order`: Must be 0-100

---

#### 2. List Effects
```
GET /api/v1/ability/effects
```

**Description**: Get paginated list of effects with filtering

**Query Parameters**:
- `page` (int) - Page number (default: 1)
- `limit` (int) - Items per page (default: 50, max: 100)
- `ids` (string) - Filter by effect IDs (comma-separated): `?ids=effect_fireball_dot,effect_ice_slow`
- `category` (string) - Filter by category: `?category=Buff`
- `tags` (string) - Filter by tags (comma-separated): `?tags=fire,dot`

**Example**:
```
GET /api/v1/ability/effects?category=Buff&page=1&limit=20
GET /api/v1/ability/effects?tags=fire,dot
GET /api/v1/ability/effects?ids=effect_fireball_dot,effect_ice_slow,effect_stun
```

---

#### 3. Get Effect by ID
```
GET /api/v1/ability/effects/:id
```

**Description**: Retrieve a specific effect configuration by ID

**Path Parameters**:
- `id` (string) - Effect ID (e.g., `effect_fireball_dot`)

**Example**:
```
GET /api/v1/ability/effects/effect_fireball_dot
```

---

#### 4. Update Effect
```
PUT /api/v1/ability/effects/:id
```

**Description**: Update an existing effect (partial update supported)

**Path Parameters**:
- `id` (string) - Effect ID to update

**Request Body** (all fields optional):
```json
{
  "duration": 6.0,
  "stacks": 5,
  "modifiers": [
    {
      "name": "Attack",
      "flat": 75.0,
      "percent": 0.1
    }
  ]
}
```

---

#### 5. Delete Effect
```
DELETE /api/v1/ability/effects/:id
```

**Description**: Delete an effect configuration

**Path Parameters**:
- `id` (string) - Effect ID to delete

**Example**:
```
DELETE /api/v1/ability/effects/effect_test_delete
```

---

### Health & Utility

#### Health Check
```
GET /health
```
No authentication required.

#### Metrics
```
GET /metrics
```
Prometheus metrics endpoint. No authentication required.

## Testing Workflow

### Test Scenario 1: Get Character by ID

1. Open `Characters` → `Get Character by ID`
2. Ensure `character_id` path variable is set (default: `char_001`)
3. Click **Send**
4. Verify response:
   - Status: 200 OK
   - Body contains character data
   - Token was automatically generated and attached

### Test Scenario 2: List All Characters

1. Open `Characters` → `List Characters`
2. Optionally modify query parameters (rarity, element, etc.)
3. Click **Send**
4. Verify pagination data in response

### Test Scenario 3: Create and Update Workflow

1. **Create**: Open `Characters` → `Create Character`
2. Modify the `character_id` in request body to ensure uniqueness
3. Click **Send**
4. Verify status: 201 Created
5. Note: `last_created_character_id` is automatically saved to environment
6. **Update**: Open `Characters` → `Update Character`
7. Update the path variable to use the created character ID
8. Modify request body as desired
9. Click **Send**
10. Verify status: 200 OK

## Troubleshooting

### "TypeError: Cannot read properties of undefined (reading 'stringify')"

This error occurs when using older versions of Postman or when `CryptoJS.enc.Base64url` is not available.

**Solution**: The collection has been updated to use a custom Base64URL encoding function that's compatible with all Postman versions. If you still see this error:

1. **Re-import the collection**: Delete the old collection and import the latest version
2. **Check Postman version**: Update to Postman v10.0 or later
3. **Verify CryptoJS library**: Ensure CryptoJS is available in the pre-request script sandbox

**Alternative**: Use the standalone Node.js test script:
```bash
cd postman
node test_token_generation.js
```

### "Missing required environment variables"

**Solution**: Ensure the environment is selected and contains `app_id` and `secret` values.

### "Token validation failed"

**Possible causes**:
1. **Clock skew**: Your system time differs from server time
2. **Wrong secret**: Verify `secret` matches server configuration
3. **Wrong app_id**: Verify `app_id` exists in `auth_config.json`

**Solution**:
```bash
# Check server logs for specific error
# Verify auth_config.json settings
cat configs/auth_config.json
```

### "Character not found"

**Solution**: Ensure the `character_id` exists in the database. Use `List Characters` to see available IDs.

### "API Key validation failed"

**Solution**: Verify `api_key` environment variable matches the configured key in `auth_config.json`:
```json
"apiKeys": [
  {
    "caller": "GAME_X3",
    "key": "game#123",
    "active": true
  }
]
```

### Response Time Issues

If responses are slow:
1. Check server logs
2. Verify database connectivity
3. Monitor server resources
4. Check network latency

## Configuration Files Reference

### auth_config.json
```json
{
  "environments": {
    "dev": {
      "apiKeys": [
        {
          "caller": "GAME_X3",
          "key": "game#123",
          "active": true
        }
      ],
      "token": {
        "enabled": true,
        "maxAgeSec": 300,
        "clockSkewSec": 30,
        "allowFutureSec": 10,
        "requiredFields": ["uid", "rid", "app_id", "server_id", "ts"],
        "secrets": [
          {
            "appId": "3X",
            "secret": "Q!w2e3r4t5",
            "active": true
          }
        ]
      }
    }
  }
}
```

### app_config.json (Security Section)
```json
{
  "security": {
    "routes": {
      "masterconfig": {
        "auth": {
          "enabled": true,
          "strategy": "token",
          "required_roles": ["access"]
        },
        "validation": {
          "enabled": true,
          "validate_signature": true,
          "validate_key": true,
          "require_base64_encoding": true
        }
      }
    }
  }
}
```

## Advanced Usage

### Testing Token Generation

Use the included Node.js script to test token generation:

```bash
cd postman
node test_token_generation.js
```

This will:
- Generate a Bearer token using the same algorithm as Postman
- Validate the token format
- Decode and verify the payload
- Check all required fields
- Verify the signature format

**Example output:**
```
✓ Token format is correct (2 parts: payload.signature)
✓ Payload is valid JSON
✓ All required fields present: uid, rid, app_id, server_id, ts
✓ Signature is valid hex (64 chars)
```

### Manual Token Generation

To manually generate a Bearer token for testing outside Postman:

```javascript
const crypto = require('crypto');

function base64UrlEncode(str) {
    return Buffer.from(str, 'utf8')
        .toString('base64')
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=/g, '');
}

function generateBearerToken(appId, secret, uid, serverId) {
    const ts = Math.floor(Date.now() / 1000);
    const rid = `req_${Date.now()}`;

    const payload = {
        uid: uid,
        rid: rid,
        app_id: appId,
        server_id: serverId,
        ts: ts
    };

    const payloadJson = JSON.stringify(payload);
    const payloadBase64Url = base64UrlEncode(payloadJson);

    const hmac = crypto.createHmac('sha256', secret);
    hmac.update(payloadJson);
    const signatureHex = hmac.digest('hex');

    return `${payloadBase64Url}.${signatureHex}`;
}

// Example usage
const token = generateBearerToken('3X', 'Q!w2e3r4t5', 'test_user_001', 'server_01');
console.log('Authorization: Bearer', token);
```

### Using cURL

```bash
# First, generate a Bearer token using the Node.js script:
cd postman
TOKEN=$(node -e "$(cat test_token_generation.js | grep -A 50 'function generateBearerToken'); console.log(generateBearerToken('3X', 'Q!w2e3r4t5', 'test_user_001', 'server_01'))")

# Then use it in your request:
curl -X GET "http://localhost:4002/api/v1/masterconfig/characters/char_001" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-API-Key: game#123" \
  -H "X-API-Version: v1" \
  -H "Content-Type: application/json"

# Or with a manually generated token:
curl -X GET "http://localhost:4002/api/v1/masterconfig/characters/char_001" \
  -H "Authorization: Bearer eyJ1aWQiOiJ0ZXN0X3VzZXJfMDAxIiwicmlkIjoicmVxXzE3MDMwMDE2MDAwMDAiLCJhcHBfaWQiOiIzWCIsInNlcnZlcl9pZCI6InNlcnZlcl8wMSIsInRzIjoxNzAzMDAxNjAwfQ.f5af08a2e54e489e6ee4b26bf25f73477900b1bd48802b0f3bcbef2168aa430f" \
  -H "X-API-Key: game#123" \
  -H "X-API-Version: v1" \
  -H "Content-Type: application/json"

# Note: The token will expire after 300 seconds (5 minutes) in dev environment
```

### Environment Switching

To test against different environments:

1. Create additional environment files:
   - `MasterConfig_Staging.postman_environment.json`
   - `MasterConfig_Production.postman_environment.json`

2. Update `base_url`, `api_key`, `secret`, etc. for each environment

3. Switch environments using the dropdown in Postman

## Support

For issues or questions:
1. Check server logs: `docker compose logs` or application logs
2. Verify configuration files: `configs/auth_config.json`, `configs/app_config.json`
3. Test health endpoint: `GET /health`
4. Review API documentation: `docs/api_response.md`

## Version History

- **v1.0.0** (2025-12-14)
  - Initial collection with token-based authentication
  - All character CRUD endpoints
  - Automatic token generation
  - Comprehensive test scripts
  - Dev environment configuration
