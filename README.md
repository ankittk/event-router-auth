# event-router-auth

A secure internal event delivery gateway written in Go.  
It accepts and routes incoming events from multiple trusted sources (webhooks, CLI, CI/CD pipelines) 
and authenticates each source using the appropriate strategy.

---

## Features

- Accepts **incoming webhook events** from trusted external sources.
- Routes events to **internal microservices** after authentication.
- Implements **multiple authentication strategies**:
	- HMAC (for webhooks)
	- JWT with service accounts (for internal app-to-app)
	- OAuth 2.0 / PAT (for developer CLI usage)
	- OIDC tokens (for GitHub Actions)
	- Optional: mTLS, API keys, Basic Auth

---

## Use Cases & Authentication Strategies

| Client / Source            | Authentication Method                                | Description                                                           |
|----------------------------|------------------------------------------------------|-----------------------------------------------------------------------|
| **External Webhooks**      | HMAC Signature                                       | Verifies the event source using a pre-shared secret.                  |
| **Internal Services**      | JWT Token (signed via Service Account)               | App-to-App communication using private key–signed tokens.             |
| **Service Account**        | JSON Key → RSA Private Key for JWT signing           | Credentials are securely loaded from disk.                            |
| **Go CLI Tool**            | OAuth 2.0 Device Flow or Personal Access Token (PAT) | Developers can manually push events using a secure CLI.               |
| **GitHub Actions (CI/CD)** | GitHub OIDC Token + Audience Validation              | Validates tokens issued by GitHub Actions using OIDC trust policies.  |
| **mTLS**                   | Mutual TLS (client certificate authentication)       | Ensures both parties (client and server) are authenticated via certs. |
| **API Key**                | API key passed via headers                           | Simple key-based authentication for internal automation.              |
| **Basic Auth**             | Base64-encoded credentials over HTTPS                | For legacy systems or simple admin endpoints.                         |

