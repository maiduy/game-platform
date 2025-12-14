/**
 * Test script to verify Bearer token generation matches Go implementation
 * Run with: node test_token_generation.js
 *
 * This script generates a token using the same algorithm as the Postman pre-request script
 * and verifies it matches the format expected by auth.go:verifyGameToken
 */

const crypto = require('crypto');

function base64UrlEncode(str) {
    return Buffer.from(str, 'utf8')
        .toString('base64')
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=/g, '');
}

function generateBearerToken(appId, secret, uid, serverId) {
    // Create timestamp (current Unix time in seconds)
    const ts = Math.floor(Date.now() / 1000);
    const rid = `req_${Date.now()}`;

    // Build token payload (must match GameTokenPayload struct)
    const payload = {
        uid: uid,
        rid: rid,
        app_id: appId,
        server_id: serverId,
        ts: ts
    };

    // Convert payload to JSON string (compact, no whitespace)
    const payloadJson = JSON.stringify(payload);

    // Base64URL encode the payload (RFC 4648)
    const payloadBase64Url = base64UrlEncode(payloadJson);

    // Compute HMAC-SHA256 signature on the JSON payload
    const hmac = crypto.createHmac('sha256', secret);
    hmac.update(payloadJson);
    const signatureHex = hmac.digest('hex');

    // Combine into final token format: {base64url_payload}.{hex_signature}
    const token = `${payloadBase64Url}.${signatureHex}`;

    console.log('=== Bearer Token Generation Test ===');
    console.log('Payload:', payload);
    console.log('Payload JSON:', payloadJson);
    console.log('Base64URL Payload:', payloadBase64Url);
    console.log('HMAC-SHA256 Signature:', signatureHex);
    console.log('Final Token:', token);
    console.log('Timestamp:', ts, '(' + new Date(ts * 1000).toISOString() + ')');
    console.log('===================================');

    return token;
}

// Test with dev environment credentials
const appId = '3X';
const secret = 'Q!w2e3r4t5';
const uid = 'test_user_001';
const serverId = 'server_01';

const token = generateBearerToken(appId, secret, uid, serverId);

console.log('\n✓ Token generated successfully!');
console.log('Copy this token to test manually:\n');
console.log(`Authorization: Bearer ${token}\n`);

// Verify token format
const parts = token.split('.');
if (parts.length === 2) {
    console.log('✓ Token format is correct (2 parts: payload.signature)');

    // Decode and verify payload
    try {
        const decodedPayload = Buffer.from(parts[0], 'base64url').toString('utf8');
        const parsed = JSON.parse(decodedPayload);
        console.log('✓ Payload is valid JSON');
        console.log('✓ Decoded payload:', parsed);

        // Check required fields
        const requiredFields = ['uid', 'rid', 'app_id', 'server_id', 'ts'];
        const missingFields = requiredFields.filter(field => !parsed[field]);
        if (missingFields.length === 0) {
            console.log('✓ All required fields present:', requiredFields.join(', '));
        } else {
            console.log('✗ Missing fields:', missingFields.join(', '));
        }

        // Verify signature is hex
        if (/^[a-f0-9]{64}$/.test(parts[1])) {
            console.log('✓ Signature is valid hex (64 chars)');
        } else {
            console.log('✗ Signature format invalid');
        }
    } catch (e) {
        console.log('✗ Failed to decode payload:', e.message);
    }
} else {
    console.log('✗ Token format is incorrect (expected 2 parts, got ' + parts.length + ')');
}
