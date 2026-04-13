# Assignment 2 Report: Microservices Migration to gRPC

**Student Name**: [Your Name]  
**Group**: [Your Group]

## 1. Project Overview
In this assignment, I successfully migrated the internal communication between the **Order Service** and **Payment Service** from a REST-based approach to **gRPC**. I adopted a **Contract-First approach** by separating the Protocol Buffer definitions from the actual service implementation.

---

## 2. Contract Management (Remote Generation)
I managed the service contracts using two dedicated repositories:
- **Repository A (Protos)**: https://github.com/Bleuble/my-grpc-proto
- **Repository B (Generated)**: https://github.com/Bleuble/my-grpc-generated

Automation was achieved via **GitHub Actions**. Every push to the Proto repository triggers a workflow that generates Go code and pushes it to the Generated repository.

> **Screenshot Assignment**: Insert a screenshot of your GitHub Actions tab showing the "Green Checkmark" ✅ for the last run.

---

## 3. gRPC Implementation Details
### Service Definition
I defined two main services in Protobuf:
1. `PaymentService`: Handles `ProcessPayment`.
2. `OrderTrackingService`: Handles real-time server-side streaming via `SubscribeToOrderUpdates`.

### Middleware (Bonus)
I implemented a **gRPC Interceptor** in the Payment Service to log the duration and method name of every incoming request.

> **Screenshot Assignment**: Insert a screenshot of the Payment Service terminal logs showing the "gRPC Call - Method: ..." log entries.

---

## 4. Server-Side Streaming (Order Tracking)
One of the core requirements was to demonstrate server-side streaming. In my implementation:
1. A client subscribes to an Order ID.
2. When the status changes in the database (e.g., from `Pending` to `Paid`), the server immediately pushes a message to the stream.

> **Screenshot Assignment**: Insert a screenshot of your gRPC client (Postman or grpcurl) receiving real-time status updates for a specific order.

---

## 5. Environment Configuration
Hardcoding addresses is prohibited. I used environment variables to store:
- `GRPC_PORT`: Port for gRPC servers.
- `REST_PORT`: Port for REST servers.
- `DATABASE_URL`: Connection string for Postgres.
- `PAYMENT_SERVICE_ADDR`: Address for the Order Service to reach the Payment Service.

---

## 6. Conclusion
The migration to gRPC has improved the system by providing strict typing, better performance, and a clear, version-controlled contract for inter-service communication.
