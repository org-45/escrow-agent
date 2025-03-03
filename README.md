### General purpose Escrow agent written in Go, PostgreSQL following OESD methodology.

Question: What's OESD ? \
Answer: Over Engineered Software Development.

![coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/org-45/escrow-agent/main/badge.json)

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

#### Running the Application

1.  **Start the Application with Docker Compose:**

    ```bash
    docker compose up --build -d
    ```

    This command builds the Docker images and starts the application in detached mode.

2.  **Access Swagger UI:**

    Once the application is running, you can access the Swagger UI at `http://localhost:8081`.  This provides a visual interface for exploring and interacting with the API endpoints.

#### Testing

This project incorporates several testing strategies to ensure code quality and application reliability.

1.  **Unit Tests:**

    Unit tests verify the functionality of individual components or functions in isolation.  To run the unit tests, execute the following command:

    ```bash
    go test ./...
    ```

    This command will run all tests in the project directory and its subdirectories. Code coverage is reported to ensure the majority of code paths are being tested.

2.  **Scenario Tests (Integration Tests):**

    Scenario tests, also known as integration tests, validate the interactions between different components of the system. They ensure that the components work together correctly to achieve specific use cases.  These tests typically involve setting up a test environment (e.g., a test database), executing a series of actions, and verifying the expected results.

    The following integration test scenarios are defined:

    *   **`buyer_cancel`:** Tests the scenario where a buyer cancels a transaction before the seller fulfills it.
    *   **`dispute_buyer_wins`:** Tests the scenario where a buyer raises a dispute and the dispute is resolved in their favor.
    *   **`dispute_seller_wins`:** Tests the scenario where a buyer raises a dispute and the dispute is resolved in the seller's favor.
    *   **`escrow_extended`:** Tests the scenario where the escrow period is extended.
    *   **`high_value_approval`:** Tests the scenario where a high-value transaction requires administrative approval.
    *   **`seller_cancel`:** Tests the scenario where a seller cancels a transaction before fulfilling it.
    *   **`simple_purchase`:** Tests the basic purchase flow (buyer creates transaction, seller fulfills, buyer confirms).

    To run scenario tests:

    *   **Prerequisites:** Ensure a test PostgreSQL database is running and accessible. Configure the database connection details (host, port, username, password, database name) in the appropriate test configuration file.
    *   **Execute the tests:**

        To run *all* integration tests:

        ```bash
        go test -tags=integration ./...
        ```

        To run a *specific* integration test (e.g., `simple_purchase`):

        ```bash
        go test -tags=integration,simple_purchase ./...
        ```

        To run *multiple specific* integration tests (e.g., `simple_purchase` and `buyer_cancel`):

         ```bash
        go test -tags=integration,simple_purchase,buyer_cancel ./...
        ```

        The `-tags` flag tells the `go test` command to only run tests that are tagged with the specified build tags. This allows you to selectively run integration tests.  If you specify multiple tags, the test must have *all* of those tags to be run.

3.  **Endpoint Testing with Swagger UI:**

    The Swagger UI at `http://localhost:8081` can be used to manually test the API endpoints. This is useful for verifying that the API is working as expected and for exploring the API's functionality.

    *   **Start the Application:** Ensure the application is running using `docker compose up --build -d`.
    *   **Open Swagger UI:** Navigate to `http://localhost:8081` in your web browser.
    *   **Interact with Endpoints:** Use the Swagger UI to send requests to the API endpoints and verify the responses. You can expand each endpoint to see its parameters, request body, and response schemas.

#### Next Steps (For Contributors)

*   **Explore the Codebase:** Familiarize yourself with the project's structure and code.
*   **Add Payment Gateway Integrations:** Contribute integrations for additional payment gateways.
*   **Improve Testing:** Expand the existing test suite to cover more code paths and scenarios.  Pay special attention to adding tests for the conflict resolution process.
*   **Enhance Conflict Resolution:** Implement more sophisticated conflict resolution mechanisms.

This project is tested with BrowserStack