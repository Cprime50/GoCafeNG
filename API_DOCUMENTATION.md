# Go9jaJobs API Documentation

This document provides information on how to use the Go9jaJobs API.

## Authentication

The API uses a secure authentication mechanism with a single API key and request signing:

* **API Key**: Required for all requests (sent via the `X-API-Key` header)
* **Timestamp**: Current Unix timestamp in seconds (sent via the `X-Timestamp` header)
* **Signature**: HMAC-SHA256 signature to verify request authenticity (sent via the `X-Signature` header)

### Setting up Authentication

1. Copy `.env.example` to `.env`
2. Generate a secure random value for `API_KEY`
3. Configure `ALLOWED_ORIGINS` to include your Next.js application's domain

### Generating Request Signatures

For each API request, you need to:
1. Use your API key for the `X-API-Key` header
2. Generate a Unix timestamp (seconds since epoch) for the `X-Timestamp` header
3. Create an HMAC-SHA256 signature using:
   - Message = `{request_path}{timestamp}`
   - Key = Your API key
4. Include the generated signature in the `X-Signature` header

This prevents request tampering and replay attacks. The timestamp must be within 5 minutes of the server time.

## API Endpoints

### Status Check

**Endpoint:** `/api/status`  
**Method:** GET  
**Authentication:** Requires API Key, Timestamp, and Signature

Returns the current status of the API service.

**Example Response:**
```json
{
  "status": "ok",
  "timestamp": "2023-06-20T14:30:00Z",
  "message": "API is running"
}
```

### Get All Jobs

**Endpoint:** `/api/jobs`  
**Method:** GET  
**Authentication:** Requires API Key, Timestamp, and Signature

Returns all jobs from the database.

**Example Response:**

```json
{
  "success": true,
  "count": 120,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "job_id": "jd-123456",
      "title": "Senior Golang Developer",
      "company": "Tech Company",
      "is_remote": true,
      "source": "linkedin",
      "posted_at": "2023-06-15T09:30:00Z",
      "location": "Lagos, Nigeria",
      "description": "We are looking for a skilled Golang developer...",
      "url": "https://example.com/jobs/123",
      "salary": "$80,000 - $120,000",
      "job_type": "Full-time"
    },
    // more jobs...
  ]
}
```

## Next.js Integration Example

```typescript
// api.ts
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const API_KEY = process.env.API_KEY || 'your_api_key_here';

// Generate signature for API requests
function generateSignature(path: string, timestamp: string): string {
  // In a real app, you should implement this on the server-side for security
  // This is a client-side example using the crypto-js library
  const message = path + timestamp;
  const CryptoJS = require('crypto-js');
  const hash = CryptoJS.HmacSHA256(message, API_KEY);
  return CryptoJS.enc.Hex.stringify(hash);
}

// Fetch all jobs
export async function getAllJobs() {
  try {
    const path = '/api/jobs';
    const timestamp = Math.floor(Date.now() / 1000).toString();
    const signature = generateSignature(path, timestamp);
    
    const response = await fetch(`${API_URL}${path}`, {
      method: 'GET',
      headers: {
        'X-API-Key': API_KEY,
        'X-Timestamp': timestamp,
        'X-Signature': signature,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    
    if (data.success) {
      return data.data;
    } else {
      console.error('Failed to fetch jobs');
      return [];
    }
  } catch (error) {
    console.error('Error fetching jobs:', error);
    return [];
  }
}

// Check API status
export async function checkApiStatus() {
  try {
    const path = '/api/status';
    const timestamp = Math.floor(Date.now() / 1000).toString();
    const signature = generateSignature(path, timestamp);
    
    const response = await fetch(`${API_URL}${path}`, {
      method: 'GET',
      headers: {
        'X-API-Key': API_KEY,
        'X-Timestamp': timestamp,
        'X-Signature': signature,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    return data.status === 'ok';
  } catch (error) {
    console.error('Error checking API status:', error);
    return false;
  }
}
```

## Server-Side Next.js Implementation (Recommended)

For better security, implement the signature generation on the server side:

```typescript
// pages/api/proxy/jobs.ts
import type { NextApiRequest, NextApiResponse } from 'next';
import crypto from 'crypto';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  try {
    const API_URL = process.env.API_URL || 'http://localhost:8080';
    const API_KEY = process.env.API_KEY;
    
    if (!API_KEY) {
      return res.status(500).json({ error: 'API key not configured' });
    }
    
    const path = '/api/jobs';
    const timestamp = Math.floor(Date.now() / 1000).toString();
    
    // Generate signature using server-side crypto
    const message = path + timestamp;
    const signature = crypto
      .createHmac('sha256', API_KEY)
      .update(message)
      .digest('hex');
    
    // Make the request to the actual API
    const response = await fetch(`${API_URL}${path}`, {
      method: 'GET',
      headers: {
        'X-API-Key': API_KEY,
        'X-Timestamp': timestamp,
        'X-Signature': signature,
      },
    });
    
    const data = await response.json();
    return res.status(200).json(data);
  } catch (error) {
    console.error('API proxy error:', error);
    return res.status(500).json({ error: 'Failed to fetch from API' });
  }
}
```

## Testing in the Browser

To test the API in a browser, you'll need to generate valid headers. Use a tool like Postman which can help with request signing:

1. Set the `X-API-Key` header with your API key
2. Add the current Unix timestamp as `X-Timestamp`
3. Generate and add the HMAC-SHA256 signature as `X-Signature`

## Using with cURL

For testing with cURL, you can use a script to generate the proper headers:

```bash
#!/bin/bash
API_KEY="your_api_key_here"
ENDPOINT="/api/jobs"
TIMESTAMP=$(date +%s)
SIGNATURE=$(echo -n "${ENDPOINT}${TIMESTAMP}" | openssl dgst -sha256 -hmac "${API_KEY}" | cut -d ' ' -f2)

curl -H "X-API-Key: ${API_KEY}" \
     -H "X-Timestamp: ${TIMESTAMP}" \
     -H "X-Signature: ${SIGNATURE}" \
     "http://localhost:8080${ENDPOINT}"
```

## Local Development

For local development:

1. Configure your `.env` file with your API key and allowed origins
2. Start the API server:
```bash
go run cmd/server/main.go
```
3. The API will be available at `http://localhost:8080`
4. Make sure to include the required authentication headers in all requests 