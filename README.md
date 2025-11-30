## 🚀 URL Shortener Service
Go + Gin + GORM + PostgreSQL + Redis + JWT + gRPC

A production-ready URL Shortener backend built with scalable architecture, clean layers, and Dockerized deployment.

This service supports:

* Shortening URLs (≤10 characters)

* Idempotent shortening (same URL → same code)

* Fast redirects using Redis caching

* Metadata tracking (click_count, last_accessed_at, created_at)

* Admin JWT authentication

* Pagination for admin URL listing

* Daily click analytics

* Custom aliases

* REST APIs

* Rate limiting (per-IP)

* Full Docker setup (app + Postgres + Redis)

## 📁 Folder Structure

```text
url-shortener/
  internal/
    analytics/     # Click analytics services
    auth/          # JWT authentication
    cache/         # Redis layer
    config/        # Env config loader
    database/      # PostgreSQL initialization
    dto/           # Request/response DTOs
    error/         # Custom error definitions
    http/          # Gin handlers & middleware
    logger/        # Custom logger (if extended)
    model/         # GORM models (URL, ClickStats)
    urlshortner/   # Core service logic
  migrations/      # SQL migrations
  main.go
  Dockerfile
  docker-compose.yml
  .env.example
  README.md

```

## ⚙️ Setup & Run Instructions
```text

1️⃣ Clone the repository
git clone https://github.com/patwaaman/Url-Shortener-Service.git
cd Url-Shortener-Service

2️⃣ Create .env file
cp .env.example .env


3️⃣ Run with Docker Compose (recommended)
docker compose up --build

Services available:

Component	URL
REST API	http://localhost:8080
PostgreSQL	localhost:5432
Redis	localhost:6379

4️⃣ Run Tests
go test ./... -v

Unit tests include:
Rate limiting middleware
```


## 🔗 **REST API Documentation**

---

### 🩺 **Health Check**

```bash
curl http://localhost:8080/health
```

---

## 1️⃣ **Shorten URL**

### **POST `/api/v1/urls`**

```bash
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://google.com",
    "custom_alias": ""
  }'
```

**Response**

```json
{
  "short_code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "original_url": "https://google.com"
}
```

---

## 2️⃣ **Redirect Short URL**

### **GET `/:code`**

```bash
curl -v http://localhost:8080/abc123
```

Redirects with **302 Found**.

---

## 3️⃣ **Admin Login (JWT)**

### **POST `/api/v1/admin/login`**

```bash
curl -X POST http://localhost:8080/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

**Response**

```json
{
  "access_token": "<jwt>",
  "token_type": "Bearer"
}
```

Use token for authenticated routes:

```
Authorization: Bearer <jwt>
```

---

## 4️⃣ **List URLs (Admin Only)**

### **GET `/api/v1/admin/urls?page=1&page_size=20`**

```bash
curl -H "Authorization: Bearer <jwt>" \
  "http://localhost:8080/api/v1/admin/urls?page=1&page_size=20"
```

---

## 5️⃣ **Daily Analytics**

### **GET `/api/v1/admin/analytics/clicks?from=YYYY-MM-DD&to=YYYY-MM-DD`**

```bash
curl -H "Authorization: Bearer <jwt>" \
  "http://localhost:8080/api/v1/admin/analytics/clicks?from=2025-01-01&to=2025-01-31"
```

---

## 📘 **Postman Collection**

You can test all API endpoints using the included Postman collection.

### 👉 Import the file:

```
URL-Shortener-Service.postman_collection.json
```

### The collection includes:

- **Shorten URL**
- **Redirect Short URL**
- **Admin Login (JWT)**
- **List Shortened URLs (Admin)**
- **Daily Click Analytics**

Each request comes pre-configured with:

- Correct HTTP method  
- Sample request body  
- Auth header (for admin routes)  
---


## 🧠 Architectural Overview

```text
                 +-----------------------+
                 |       REST API        |
                 |      (Gin server)     |
                 +-----------+-----------+
                             |
                             v
                 +-----------------------+
                 |     Service Layer     |
                 |    (Business Logic)   |
                 +-----------+-----------+
                             |
               +-------------+--------------+
               |                            |
               v                            v
     +------------------+           +------------------+
     |   PostgreSQL     |           |      Redis       |
     | (Persistent DB)  | <--cache--| (Fast Redirects) |
     +------------------+           +------------------+
```

## 🧩 **Design Decisions & Trade-offs**

---

### **1. PostgreSQL as Source of Truth**

- **Strong consistency**
- **Relational model supports analytics**
- **Durable for long-term (5+ years)**

*Trade-off:* Slower than Redis → solved using caching.

---

### **2. Redis as Read-Through Cache**

- **Redirects are extremely frequent**
- **Redis reduces DB load by 80–95%**
- **TTL keeps cache fresh**

*Trade-off:* Requires cache invalidation when URLs are deleted (not needed in this assignment).

---

### **3. Deterministic Short Codes (Idempotent)**

- **SHA256(original URL) → base62 → first 10 chars**
- **Same URL always returns the same short code**
- **Extremely low collision probability**

*Trade-off:* Codes are not sequential (unlike Snowflake or ULID).

---

### **4. Rate Limiting (Per-IP Token Bucket)**

- **Protects API from abuse & DDoS**
- **Prevents brute-force crawling of URLs**
- **Simple, in-memory, and fast**

*Trade-off:* Not distributed across instances — but can be extended using Redis.

---

### **5. JWT Authentication for Admin Routes**

- **Stateless and lightweight**
- **Works well with load balancers**
- **Easy integration with admin dashboards**

*Trade-off:* Requires token rotation and secret management for high security (optional here).