### General purpose Escrow agent written in Go, PostgreSQL following OESD methodology.

Question: What's OESD ? \
Answer: Over Engineered Software Development.


![coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/org-45/escrow-agent/main/badge.json)



| Method | Endpoint        | Description                                      |
|--------|-----------------|--------------------------------------------------|
| POST   | `/register`      | Register a new user (buyer, seller, or admin)    |
| POST   | `/login`         | Log in a user and return a JWT token             |
| GET    | `/profile`       | Get the logged-in user's profile                 |
| PUT    | `/profile`       | Update the logged-in user's profile (username, etc.)|




| Method | Endpoint                     | Description                                                       |
|--------|------------------------------|-------------------------------------------------------------------|
| POST   | `/transactions`               | Create a new transaction (by buyer)                               |
| GET    | `/transactions`               | Get a list of all transactions for the logged-in user (buyer/seller)|
| GET    | `/transactions/{id}`          | Get details of a specific transaction                             |
| PUT    | `/transactions/{id}/fulfill`  | Mark a transaction as fulfilled (by seller)                       |
| PUT    | `/transactions/{id}/confirm`  | Confirm the delivery of a product or service (by buyer)            |





| Method | Endpoint                        | Description                                                       |
|--------|----------------------------------|-------------------------------------------------------------------|
| POST   | `/escrow/{transaction_id}/deposit` | Deposit funds into escrow for a transaction (by buyer)            |
| PUT    | `/escrow/{transaction_id}/release` | Release funds from escrow to seller (by system or admin)          |
| PUT    | `/escrow/{transaction_id}/refund`  | Refund funds from escrow to buyer (by system or admin)            |
| GET    | `/escrow/{transaction_id}`         | Get details of the escrow account for a specific transaction       |




| Method | Endpoint                      | Description                                                      |
|--------|-------------------------------|------------------------------------------------------------------|
| POST   | `/transactions/{id}/dispute`   | Raise a dispute for a specific transaction (by buyer or seller)  |
| GET    | `/disputes`                    | Get a list of all disputes for the logged-in user (buyer/seller) |
| GET    | `/disputes/{id}`               | Get details of a specific dispute                                |
| PUT    | `/disputes/{id}/resolve`       | Resolve a dispute (by admin)                                     |



| Method | Endpoint                          | Description                                                     |
|--------|-----------------------------------|-----------------------------------------------------------------|
| GET    | `/admin/users`                    | Get a list of all users                                         |
| GET    | `/admin/users/{id}`               | Get details of a specific user                                  |
| GET    | `/admin/transactions`             | Get a list of all transactions                                  |
| GET    | `/admin/transactions/{id}`        | Get details of a specific transaction                           |
| PUT    | `/admin/transactions/{id}/release`| Manually release funds from escrow (by admin)                   |
| PUT    | `/admin/transactions/{id}/refund` | Manually refund funds to buyer (by admin)                       |
| GET    | `/admin/disputes`                 | Get a list of all disputes                                      |
| PUT    | `/admin/disputes/{id}/resolve`    | Resolve a dispute (by admin)                                    |


| Method | Endpoint                        | Description                                                     |
|--------|---------------------------------|-----------------------------------------------------------------|
| GET    | `/logs/{transaction_id}`         | Get a list of all logs for a specific transaction                |


#### Overview
This project is an open-source general-purpose escrow agent designed to facilitate secure transactions between buyers and sellers. The escrow agent ensures that funds are only released when the buyer confirms that the agreed-upon services or goods have been delivered. The platform is built to support multiple payment gateways with a flexible, plug-and-play architecture, making it easy for contributors to add their preferred gateways.

#### Key Components
1. **Escrow Software**  
   Manages the escrow process between buyers and sellers.  
   Holds and releases funds based on buyer confirmation.

2. **Payment Gateway Integration**  
   Supports multiple payment gateways with a plug-and-play architecture.  
   Contributors can add additional payment gateways as needed.

3. **Conflict Resolution Support**  
   Basic documentation and reporting to manage conflicts.  
   Stores requirement specifications, agreements, and related documents for conflict escalation.  
   Conflict resolution can be outsourced to a specialized operational department (outside the scope of this project).

#### Features
- **Escrow Flow**
  - Buyer agrees to purchase.
  - Escrow Agent holds funds.
  - Seller delivers goods/services.
  - Buyer confirms receipt, and Escrow Agent releases funds.

- **Conflict Management**
  - Collection of requirement specifications from the buyer.
  - Seller's agreement on the specifications.
  - Storage of agreement documents in S3 buckets.
  - In case of conflict, the service holds all relevant documents for escalation.

- **Extensibility**
  - The platform is designed for flexibility, allowing the integration of various payment gateways.
  - **Goal**: Build a comprehensive solution that contributors can extend, particularly for specific gateways.

#### How to run?

```
docker compose up --build -d

You will see swagger on localhost:8081
```

```
curl -X POST http://localhost:8080/signup \
-H "Content-Type: application/json" \
-d '{
      "username": "testuser",
      "password": "testpassword"
    }'
```

```
curl -X POST http://localhost:8080/login \
-H "Content-Type: application/json" \
-d '{
      "username": "testuser",
      "password": "testpassword"
    }'

```

```
curl -X POST http://localhost:8080/api/escrow \
-H "Content-Type: application/json" \
-H "Authorization: Bearer your-jwt-token" \
-d '{
      "buyer_id": "buyer123",
      "seller_id": "seller456",
      "amount": 500.0,
      "description": "Payment for services"
    }'
```

This project is tested with BrowserStack 
