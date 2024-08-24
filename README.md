### General Purpose Escrow Agent

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
```