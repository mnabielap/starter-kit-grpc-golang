# 🚀 Starter Kit gRPC - Go (Golang)

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/doc/devel/release.html)
[![gRPC](https://img.shields.io/badge/gRPC-High%20Performance-blue?logo=google)](https://grpc.io/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20%2F%20Standard-purple)](https://github.com/golang-standards/project-layout)

A high-performance, production-ready microservice starter kit built with **Go**, **gRPC**, and **gRPC-Gateway**.

This project demonstrates a modern **Dual-Protocol** architecture: it serves high-speed **gRPC (Protobuf)** on one port and a **RESTful JSON API** (via reverse proxy) on another. It comes fully equipped with Authentication, RBAC, Database interactions (GORM), and automated testing.

---

## ✨ Features

- **⚡ Dual Protocol**:
  - **gRPC (Port 50051)**: HTTP/2 + Protobuf for internal microservices.
  - **REST Gateway (Port 8080)**: HTTP/1.1 + JSON for frontend/legacy clients.
- **🏗 Standard Go Layout**: Clean separation of concerns (`handler`, `service`, `repository`).
- **🔐 Security**:
  - **JWT Authentication**: Access & Refresh Tokens.
  - **RBAC**: Role-Based Access Control (Admin vs User).
  - **Interceptors**: Middleware for Auth, Logging, Rate Limiting, and Recovery.
- **💾 Database Agnostic**:
  - **GORM**: Seamlessly switch between **SQLite** (Local Dev) and **PostgreSQL** (Docker/Prod).
- **📄 API Documentation**: Built-in **Swagger UI** for the REST Gateway.
- **🧪 Automated Testing**: Full suite of **Python scripts** to test gRPC endpoints directly.
- **🐳 Docker Ready**: Multi-stage builds, Persistence volumes, and custom networking.

---

## 📂 Project Structure

```text
starter-kit-grpc-golang/
├── api/
│   ├── proto/v1/          # Protocol Buffer Definitions (.proto)
│   ├── gen/v1/            # Generated Go Code (pb.go)
│   └── openapi/           # Generated Swagger JSON
├── cmd/server/            # Entry point (main.go)
├── config/                # Environment & Database config
├── internal/
│   ├── grpc_handler/      # Transport Layer (gRPC Controllers)
│   ├── service/           # Business Logic Layer (Usecase)
│   ├── repository/        # Data Access Layer (GORM)
│   ├── interceptor/       # Middleware (Auth, Log, RateLimit)
│   └── models/            # Database Structs
├── pkg/
│   ├── logger/            # Structured Logging (slog)
│   ├── swagger/           # Swagger UI Handler
│   └── utils/             # Helpers (JWT, Pagination)
├── api_tests/grpc/        # Python Automated Tests
├── deploy/                # Dockerfile & Entrypoint
├── .env.example           # Environment template
└── buf.gen.yaml           # Buf Generation Config
```

---

## 🏃‍♂️ Getting Started (Local Development)

We recommended running the project locally first using **SQLite** to understand the flow.

### Prerequisites
1.  **Go** (v1.22+)
2.  **Buf** (For generating Proto files): `go install github.com/bufbuild/buf/cmd/buf@latest`
3.  **Python 3.x** (For testing scripts)

### 1. Clone & Install Dependencies
```bash
git clone https://github.com/mnabielap/starter-kit-grpc-golang.git
cd starter-kit-grpc-golang

# Download Go modules
go mod tidy
```

### 2. Configure Environment
Copy the example environment file. By default, it uses **SQLite**, so no extra DB setup is needed.
```bash
cp .env.example .env
```

### 3. Generate Protobuf Code 🛠️
Before running, you must generate the Go code from the `.proto` definitions.
```bash
# Update dependencies (downloads googleapis)
buf dep update

# Generate Go code and Swagger JSON
buf generate
```
*Note: If you encounter errors, ensure `buf` is in your PATH.*

### 4. Run the Application 🚀
```bash
go run cmd/server/main.go
```
You should see output indicating both servers are running:
> ℹ️ gRPC Server listening port=50051
> ℹ️ HTTP Gateway listening port=8080

### 5. Access the API
- **Swagger UI**: [http://localhost:8080/swagger-ui](http://localhost:8080/swagger-ui)
- **Health Check (JSON)**: [http://localhost:8080/v1/health](http://localhost:8080/v1/health)

---

## 🧪 Automated Testing (Python)

Forget Postman! We use **Python scripts** to test the gRPC endpoints directly using the generated protobufs.

### Setup Python Environment
Navigate to the test directory and install dependencies.
```bash
cd api_tests/grpc

# Install gRPC tools
pip install grpcio grpcio-tools googleapis-common-protos

# Generate Python Proto Code (Required for tests)
# Run this from the ROOT directory:
python -m grpc_tools.protoc -I. -Ithird_party --python_out=api_tests/grpc --grpc_python_out=api_tests/grpc api/proto/v1/auth.proto api/proto/v1/user.proto api/proto/v1/health.proto
```

### Running Tests
The scripts handle token storage automatically (`secrets.json`). Run them in order:

**1. Authentication Flow:**
```bash
# Register a new user
python api_tests/grpc/A1.auth_register.py

# Login (Saves tokens to secrets.json)
python api_tests/grpc/A2.auth_login.py
```

**2. User Management (RBAC):**
*Note: Ensure you login as an **Admin** in step A2 to run these.*
```bash
# Create User
python api_tests/grpc/B1.user_create.py

# Get All Users (Pagination & Sorting)
python api_tests/grpc/B2.user_get_all.py
```

---

## 🐳 Docker Deployment (Production Style)

We use a **Manual Docker Setup** (without Compose) to simulate a real orchestrated environment. We will use **PostgreSQL**.

> **Note for Windows Users:** The commands below use `\` for line breaks. If using Command Prompt (CMD), replace `\` with `^`. If using PowerShell, use `` ` ``.

### 1. Create Network 🌐
Create a dedicated network for App-to-DB communication.
```bash
docker network create grpc_golang_network
```

### 2. Create Volumes 📦
Create persistent storage so data isn't lost when containers stop.
```bash
# Volume for Postgres Data
docker volume create grpc_golang_db_volume

# Volume for Media/Uploads
docker volume create grpc_golang_media_volume
```

### 3. Start Database (PostgreSQL) 🐘
Run the Postgres container attached to our network and volume.
```bash
docker run -d \
  --name grpc-golang-postgres \
  --network grpc_golang_network \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=mysecretpassword \
  -e POSTGRES_DB=starter_kit_grpc_db \
  -v grpc_golang_db_volume:/var/lib/postgresql/data \
  postgres:15-alpine
```

### 4. Setup Docker Environment 📝
Create a `.env.docker` file. **Crucial:** `DB_HOST` must match the Postgres container name.

```properties
GO_ENV=production
GRPC_PORT=50051
GATEWAY_PORT=8080

# Database (Matches container name above)
DB_DRIVER=postgres
DB_HOST=grpc-golang-postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=mysecretpassword
DB_NAME=starter_kit_grpc_db
DB_SSLMODE=disable

JWT_SECRET=production_secure_secret
```

### 5. Build the App Image 🏗️
```bash
docker build -f deploy/Dockerfile -t grpc-golang-app .
```

### 6. Run the App Container 🚀
Run the app, injecting the env file and mounting volumes.
```bash
docker run -d \
  --name grpc-golang-container \
  --network grpc_golang_network \
  --env-file .env.docker \
  -p 50051:50051 \
  -p 8080:8080 \
  -v grpc_golang_media_volume:/usr/src/app/media \
  grpc-golang-app
```

The API is now live at `http://localhost:8080` (HTTP) and `localhost:50051` (gRPC).

---

## 🛠 Docker Management Commands

Here is a cheat sheet for managing your containers.

#### 📜 View Logs
See what's happening inside the container in real-time.
```bash
docker logs -f grpc-golang-container
```

#### 🛑 Stop Container
Safely stop the running application.
```bash
docker stop grpc-golang-container
```

#### ▶️ Start Container
Start the container again (Data in Postgres is preserved via volumes).
```bash
docker start grpc-golang-container
```

#### 🗑 Remove Container
Removes the container instance (requires stopping first).
```bash
docker stop grpc-golang-container
docker rm grpc-golang-container
```

#### 📦 List Volumes
Check your persistent data volumes.
```bash
docker volume ls
```

#### ⚠️ Remove Volume
**WARNING:** This deletes your database data permanently!
```bash
docker volume rm grpc_golang_db_volume
```

---

## 🤝 Contributing

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📝 License

Distributed under the MIT License. See `LICENSE` for more information.