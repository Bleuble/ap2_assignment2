# Order & Payment Platform (Microservices)

This project implements a resilient two-service platform (Order & Payment) using **Go**, **Clean Architecture**, and **PostgreSQL**.

## 🏗 Architecture Overview

Each service is designed following **Clean Architecture** principles to ensure separation of concerns and testability.

### Service Boundaries & Flow
```mermaid
graph TD
    Client[HTTP Client / cURL] -->|REST| OS[Order Service :8080]
    subgraph "Order Context"
        OS -->|Usecase| OUC[Order UseCase]
        OUC -->|Repository| ORP[Postgres Repo]
        ORP --> ODB[(order_db)]
    end
    OUC -->|HTTP Client (Timeout 2s)| PS[Payment Service :8081]
    subgraph "Payment Context"
        PS -->|Usecase| PUC[Payment UseCase]
        PUC -->|Repository| PRP[Postgres Repo]
        PRP --> PDB[(payment_db)]
    end
```

## 🛠 Features Implemented
- **Clean Architecture:** Strict separation into Domain, Use Case, Repository, and Transport layers.
- **Service Decomposition:** Two bounded contexts with **Database-per-Service** (no shared storage).
- **Resilient Communication:** Order Service uses a custom HTTP client with a **2-second timeout**.
- **Idempotency (Bonus +10%):** Prevents duplicate orders/payments using the `Idempotency-Key` header.
- **Search Filter:** Search orders by amount range with strict validation.
- **Fault Tolerance:** Returns 503 Service Unavailable if the payment service is down.

## 🚀 How to Run

### 1. Database Setup (PostgreSQL)
Ensure you have PostgreSQL running. Run the following to create the databases and tables:
```sql
-- Join as superuser
CREATE DATABASE order_db;
CREATE DATABASE payment_db;

-- In order_db:
CREATE TABLE orders (id VARCHAR(50) PRIMARY KEY, customer_id VARCHAR(50) NOT NULL, item_name VARCHAR(100) NOT NULL, amount BIGINT NOT NULL, status VARCHAR(20) NOT NULL, idempotency_key VARCHAR(100), created_at TIMESTAMP NOT NULL);

-- In payment_db:
CREATE TABLE payments (id VARCHAR(50) PRIMARY KEY, order_id VARCHAR(50) NOT NULL, transaction_id VARCHAR(50) NOT NULL, amount BIGINT NOT NULL, status VARCHAR(20) NOT NULL);
```

### 2. Start Services
Open two terminal tabs:
```bash
# Tab 1: Payment Service
cd payment-service && go run cmd/payment-service/main.go

# Tab 2: Order Service
cd order-service && go run cmd/order-service/main.go
```

## 🧪 API Examples

**Create Order (POST):**
```bash
curl -X POST http://localhost:8080/orders \
-H "Content-Type: application/json" \
-H "Idempotency-Key: unique-123" \
-d '{"customer_id": "user_01", "item_name": "Laptop", "amount": 50000}'
```

**Filter Orders by Range (GET):**
```bash
curl "http://localhost:8080/orders/filter?min=1000&max=60000"
```

## 🛡 Business Rules Logic
1. **Financial Accuracy:** All monetary values use `int64` (cents).
2. **Payment Limits:** Amounts > 100,000 cents ($1,000) are automatically **Declined**.
3. **Immutability:** "Paid" orders cannot be cancelled.
4. **Resilience:** If Payment Service is down, orders are marked as **Failed** and a 503 error is returned.
