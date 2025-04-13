# ByteKitchen - A Restaurant Management Backend System

## Project Overview

This is a RESTful backend for a Hotel Management System built using Go, MongoDB, JWT Authentication, and Gorilla Mux. It includes features for managing users, tables, menus, food, orders, order items, and invoices.

## Project Structure Overview

```
Restaurant_Management_Backend/
│
├── main.go                  # App entry point
├── config/                  # DB Configuration
├── routes/                  # API route registrations
├── middlewares/             # Auth and other middlewares
├── controllers/             # Business logic handlers
├── models/                  # MongoDB schemas & structs
├── helpers/                 # Utility/helper functions
├── docs                     # Postman Collection
├── .env                     # Environment variables
├── go.mod                   # Go module dependencies
├── go.sum                   # Version checksum for modules
```

---

## Prerequisites

Ensure you have the following installed:

- [Go 1.20+](https://golang.org/dl/)
- [MongoDB Atlas](https://www.mongodb.com/cloud/atlas) or local MongoDB instance
- Git

---

## Setup Instructions

### 1. Clone the repository

```bash
git clone https://github.com/02priyeshraj/ByteKitchen_Restaurant_Management_Backend_System.git
cd Restaurant_Management_Backend
```

### 2. Set up environment variables

Create a `.env` file in the project root:

```bash
touch .env
```

Then add the following content:

```env
export PORT=8080
export DB= mongo_db_connection_string
export JWT_SECRET=your_secret_key
```

---

### 3. Install Go modules

Make sure you are inside the project directory, then run:

```bash
go mod tidy
```

This will install all required packages listed in `go.mod`.

---

### 4. Run the server

```bash
go run main.go
```

The server will start on the specified port (default is `8080`):

---

Here's the updated **Routes Overview** table based on the routes you provided:

---

## Routes Overview

| Route Type                        | Description                                | Auth Required |
| --------------------------------- | ------------------------------------------ | ------------- |
| `/users/signup`<br>`/users/login` | User registration and login                | ❌            |
| `/users/...`                      | User management & logout                   | ✅            |
| `/tables/...`                     | Table management (CRUD + reserve)          | ✅            |
| `/menus/...`                      | Menu management (CRUD)                     | ✅            |
| `/foods/...`                      | Food items CRUD + filter by menu           | ✅            |
| `/orders/...`                     | Order management (CRUD, status)            | ✅            |
| `/orderitems/...`                 | Order item control & filtering             | ✅            |
| `/invoices/...`                   | Invoice CRUD + filter by user/order/status | ✅            |

> See `routes/` and `controllers/` folders for detailed route logic.

---

## 📬 API Testing – Postman Collection

You can test all API endpoints using the Postman Collection below:

[Download Postman Collection](docs/Hotel_Management_Golang.postman_collection.json)

> Import the collection into Postman and set your environment variables for authentication.

---

## Have Suggestions or Issues?

Feel free to open an issue or submit a pull request on GitHub. Let's make Bytekitchen even better together!
