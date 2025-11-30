🚀 URL Shortener Service
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

📁 Folder Structure

url-shortener/
  internal/
    analytics/
    auth/
    cache/
    config/
    database/
    dto/
    error/
    http/
    logger/
    model/
    urlshortner/
  migrations/
  main.go
  Dockerfile
  docker-compose.yml
  .env.example
  README.md


⚙️ Setup & Run Instructions
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


🔗 REST API Documentation
Health Check
curl http://localhost:8080/health

1) Shorten URL

POST /api/v1/urls

curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://google.com",
    "custom_alias": ""
  }'


Response:

{
  "short_code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "original_url": "https://google.com"
}

2) Redirect Short URL

GET /:code

curl -v http://localhost:8080/abc123


Redirects with 302.

3) Admin Login (JWT)

POST /api/v1/admin/login

curl -X POST http://localhost:8080/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'


Response:

{
  "access_token": "<jwt>",
  "token_type": "Bearer"
}


Use this token in:

Authorization: Bearer <token>

4) List URLs (Admin)

GET /api/v1/admin/urls?page=1&page_size=20

curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/admin/urls?page=1&page_size=20"

5) Daily Analytics

GET /api/v1/admin/analytics/clicks?from=&to=

curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/admin/analytics/clicks?from=2025-01-01&to=2025-01-31"


📘 Postman Collection

Import the file:

URL-Shortener-Service.postman_collection.json

Includes:

Shorten URL

Redirect

Admin login

List URLs

Analytics

🧠 Architectural Overview
                 +-----------------------+
                 |      REST API     |
                 |    (Gin server)   |
                 +-----------+-----------+
                             |
                             v
                 +-----------------------+
                 |     Service Layer     |
                 |  (Business Logic)     |
                 +-----------+-----------+
                             |
               +-------------+--------------+
               |                            |
               v                            v
     +------------------+           +------------------+
     |   PostgreSQL     |           |      Redis       |
     | (Persistent DB)  | <--cache--| (Fast Redirects) |
     +------------------+           +------------------+

🧩 Design Decisions & Trade-offs

1. PostgreSQL as Source of Truth

Strong consistency

Relational structure supports analytics

Durable (5+ year requirement)

Trade-off: Slower than Redis → solved via caching.


2. Redis as Read-Through Cache

Redirects are extremely frequent

Redis reduces DB load by 80–95%

TTL ensures periodic refresh

Trade-off: Cache invalidation needed when deleting URLs (not required here).


3. Deterministic Short Codes (Idempotent)

Using SHA256(originalURL) → base62 → first 10 chars

Same URL always returns same code

Very low collision probability

Trade-off: Not sequential (Snowflake style).


4. Rate Limiting (Per-IP Token Bucket)

Protects public API from abuse

Prevents brute-force crawling

Lightweight (in-memory)

Trade-off: Not distributed across multiple instances — but can be extended using Redis.


5. JWT Authentication for Admin

Stateless, minimal overhead

Works well with load balancers

Easy to integrate with admin dashboards

Trade-off: Requires token rotation support for high security (optional).