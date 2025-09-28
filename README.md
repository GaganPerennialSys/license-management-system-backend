# License Management System

A comprehensive license management system with admin and customer portals, plus SDK integration for mobile and desktop applications.

## Features

### Core Components

- **User Management & Authentication**: Role-based access control for admin and customer users
- **Subscription Pack Management**: Create and manage subscription plans with pricing and validity
- **Customer Management**: Full customer lifecycle with profile management
- **Subscription Lifecycle**: Request, approve, assign, and manage subscriptions
- **App SDK Integration**: API key-based authentication for mobile/desktop applications

### Business Rules

- Only one active subscription per customer at any time
- Subscription lifecycle: `requested` → `approved` → `active` → `inactive`/`expired`
- Customer requests require admin approval before activation
- Soft delete for customers and subscription packs
- Automatic expiry handling based on validity periods

## Quick Start

### Prerequisites

- Go 1.21 or higher
- SQLite (included with the application)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd cursor-ai-backend
```

2. Install dependencies:
```bash
go mod download
```

3. Run the application:
```bash
go run main.go
```

The API will be available at `http://localhost:8080`

### Default Admin Account

- **Email**: `admin@example.com`
- **Password**: `admin123`

## API Documentation

### Interactive Documentation

Visit `http://localhost:8080/swagger/index.html` for interactive API documentation.

### Authentication

The API supports two authentication methods:

#### Frontend APIs (JWT)
- **Login**: `POST /api/admin/login` or `POST /api/customer/login`
- **Usage**: Include `Authorization: Bearer <jwt_token>` in request headers
- **Expiration**: 24 hours

#### SDK APIs (API Key)
- **Login**: `POST /sdk/auth/login`
- **Usage**: Include `X-API-Key: <api_key>` in request headers
- **Generation**: API keys are generated automatically on first SDK login

### API Endpoints

#### Frontend APIs (`/api/`)

**Authentication (No auth required)**
- `POST /api/admin/login` - Admin login
- `POST /api/customer/login` - Customer login
- `POST /api/customer/signup` - Customer registration

**Admin Management (JWT + Admin role required)**
- `GET /api/v1/admin/customers` - List customers
- `POST /api/v1/admin/customers` - Create customer
- `GET /api/v1/admin/customers/{id}` - Get customer
- `PUT /api/v1/admin/customers/{id}` - Update customer
- `DELETE /api/v1/admin/customers/{id}` - Delete customer

- `GET /api/v1/admin/packs` - List subscription packs
- `POST /api/v1/admin/packs` - Create subscription pack
- `GET /api/v1/admin/packs/{id}` - Get subscription pack
- `PUT /api/v1/admin/packs/{id}` - Update subscription pack
- `DELETE /api/v1/admin/packs/{id}` - Delete subscription pack

- `GET /api/v1/admin/subscriptions` - List subscriptions
- `POST /api/v1/admin/subscriptions` - Create subscription
- `GET /api/v1/admin/subscriptions/{id}` - Get subscription
- `PUT /api/v1/admin/subscriptions/{id}/approve` - Approve subscription
- `PUT /api/v1/admin/subscriptions/{id}/assign` - Assign subscription
- `PUT /api/v1/admin/subscriptions/{id}/unassign` - Unassign subscription
- `DELETE /api/v1/admin/subscriptions/{id}` - Delete subscription

**Customer Management (JWT + Customer role required)**
- `GET /api/v1/customer/profile` - Get profile
- `PUT /api/v1/customer/profile` - Update profile
- `GET /api/v1/customer/subscription` - Get current subscription
- `POST /api/v1/customer/subscription/request` - Request subscription
- `PUT /api/v1/customer/subscription/deactivate` - Deactivate subscription
- `GET /api/v1/customer/subscription/history` - Get subscription history

#### SDK APIs (`/sdk/`)

**Authentication (No auth required)**
- `POST /sdk/auth/login` - SDK login (generates API key)

**Subscription Management (API Key required)**
- `GET /sdk/v1/subscription` - Get current subscription
- `POST /sdk/v1/subscription/request` - Request subscription
- `PUT /sdk/v1/subscription/deactivate` - Deactivate subscription
- `GET /sdk/v1/subscription/history` - Get subscription history

## Database Schema

### Core Tables

#### Users
- `id` (Primary Key)
- `email` (Unique)
- `password_hash`
- `role` (admin/customer)
- `api_key` (for SDK authentication)
- `created_at`, `updated_at`

#### Customers
- `id` (Primary Key)
- `user_id` (Foreign Key to Users)
- `name`
- `phone`
- `created_at`, `updated_at`, `deleted_at` (soft delete)

#### Subscription Packs
- `id` (Primary Key)
- `name`
- `description`
- `sku` (Unique identifier)
- `price` (Decimal)
- `validity_months` (1-12)
- `created_at`, `updated_at`, `deleted_at` (soft delete)

#### Subscriptions
- `id` (Primary Key)
- `customer_id` (Foreign Key to Customers)
- `pack_id` (Foreign Key to Subscription Packs)
- `status` (requested/approved/active/inactive/expired)
- `requested_at`, `approved_at`, `assigned_at`, `expires_at`, `deactivated_at`
- `created_at`, `updated_at`

## Docker Deployment

### Using Docker Compose

1. Build and run with Docker Compose:
```bash
docker-compose up --build
```

2. Access the API at `http://localhost:8080`

### Using Docker

1. Build the Docker image:
```bash
docker build -t license-management-api .
```

2. Run the container:
```bash
docker run -p 8080:8080 -v $(pwd)/data:/data license-management-api
```

## Configuration

### Environment Variables

- `PORT`: Server port (default: 8080)
- `DATABASE_PATH`: SQLite database file path (default: ./license_management.db)
- `JWT_SECRET`: JWT signing secret (default: your-secret-key-change-in-production)

### Production Considerations

1. **Change JWT Secret**: Update `JWT_SECRET` environment variable
2. **Database**: Consider using PostgreSQL for production
3. **HTTPS**: Use reverse proxy (nginx) with SSL certificates
4. **Rate Limiting**: Implement rate limiting for API endpoints
5. **Logging**: Add structured logging
6. **Monitoring**: Add health checks and metrics

## SDK Integration

### Authentication Flow

1. **Login**: Call `POST /sdk/auth/login` with email/password
2. **Store API Key**: Save the returned `api_key` securely
3. **Make Requests**: Include `X-API-Key: <api_key>` in all subsequent requests

### Example SDK Usage

```bash
# Login and get API key
curl -X POST http://localhost:8080/sdk/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"customer@example.com","password":"password123"}'

# Use API key for subsequent requests
curl -X GET http://localhost:8080/sdk/v1/subscription \
  -H "X-API-Key: sk-sdk-1234567890abcdef"
```

## Development

### Project Structure

```
├── main.go                 # Application entry point
├── go.mod                  # Go module dependencies
├── openapi.yaml           # OpenAPI 3.0 specification
├── Dockerfile             # Docker configuration
├── docker-compose.yml     # Docker Compose configuration
└── internal/
    ├── config/            # Configuration management
    ├── database/          # Database connection and setup
    ├── handlers/          # HTTP request handlers
    ├── middleware/        # HTTP middleware (auth, CORS, etc.)
    └── models/            # Database models and business logic
```

### Adding New Features

1. **Models**: Add new models in `internal/models/`
2. **Handlers**: Create handlers in `internal/handlers/`
3. **Routes**: Register routes in `main.go`
4. **Documentation**: Update `openapi.yaml`
5. **Tests**: Add tests for new functionality

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## License

MIT License - see LICENSE file for details.

## Support

For support and questions, please contact the development team or create an issue in the repository.
