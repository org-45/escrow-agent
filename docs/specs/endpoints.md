| Method | Endpoint        | Description                                      |
|--------|-----------------|--------------------------------------------------|
| POST   | `/register`      | Register a new user (buyer, seller, or admin)    |
| POST   | `/login`         | Log in a user and return a JWT token             |
| GET    | `/profile`       | Get the logged-in user's profile                 |
| PUT    | `/profile`       | Update the logged-in user's profile (email, etc.)|




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
| GET    | `/notifications`                | Get a list of notifications for the logged-in user               |
