# Restaurant Order Processing System (Go + RabbitMQ + PostgreSQL)

A fully distributed, microservices-based restaurant order management system built in Go.

This project demonstrates how modern delivery apps (like pizza trackers) use message queues and independent services to handle high traffic, maintain reliability, and scale smoothly.

What looks like a simple "Your order is cooking..." update is powered by microservices, message queues, worker pools, and fanout notifications. This project implements all of that — from order creation to kitchen processing to real-time status updates.

## 🚀 Project Overview

This system models a real restaurant backend with four independent services communicating through RabbitMQ and storing data in PostgreSQL:

### 1. Order Service (HTTP API)
**Handles:**
- Receiving customer orders
- Validating input
- Saving orders to PostgreSQL
- Publishing tasks to RabbitMQ

### 2. Kitchen Worker
**Simulates kitchen staff:**
- Consumes order messages
- Updates order status (received → cooking → ready)
- Logs all status changes
- Supports specialization (dine_in / takeout / delivery)
- Auto-requeues failed jobs
- Supports graceful shutdown

### 3. Tracking Service (HTTP API)
**Provides:**
- Order status API
- Full history of each order
- Real-time worker status

### 4. Notification Subscriber
**Fanout subscriber** that listens for all status updates and prints notifications.

## 🧱 Architectural Diagram
                                +------------------------+
                                |     PostgreSQL DB      |
                                |   (orders, logs, etc.) |
                                +-----------+------------+
                                            ^
                                            |
+-------------+         +-------------------+-----------------+
| HTTP Client | ----->  |   Order Service (API + Publisher)   |
+-------------+         +-------------------+-----------------+
                                    |
                                    | Publishes new orders
                                    v
                        +-----------------------------+
                        |     RabbitMQ Broker         |
                        | (orders_topic, fanout, etc.)|
                        +------+---------------+------+
                               |               |
                               |               |
                     Consumes  |               | Status updates
                               v               v
                     +----------------+   +-------------------------+
                     | Kitchen Worker |   | Notification Subscriber |
                     +----------------+   +-------------------------+

                               ^
                               |
                     +------------------+
                     | Tracking Service |
                     |    (API)         |
                     +------------------+
Below is a **clean, professional, GitHub-ready README.md** fully based on your project description.
It includes: overview, architecture, features, setup, usage, technologies, diagrams, logs, and more.

---

# **Restaurant Order Processing System (Go + RabbitMQ + PostgreSQL)**

A fully distributed, microservices-based restaurant order management system built in Go.
This project demonstrates how modern delivery apps (like pizza trackers) use message queues and independent services to handle high traffic, maintain reliability, and scale smoothly.

What looks like a simple “Your order is cooking…” update is powered by **microservices**, **message queues**, **worker pools**, and **fanout notifications**.
This project implements all of that — from order creation to kitchen processing to real-time status updates.

---

# 🚀 **Project Overview**

This system models a real restaurant backend with four independent services communicating through **RabbitMQ** and storing data in **PostgreSQL**:

### **1. Order Service (HTTP API)**

Handles:

* receiving customer orders
* validating input
* saving orders to PostgreSQL
* publishing tasks to RabbitMQ

### **2. Kitchen Worker**

Simulates kitchen staff:

* consumes order messages
* updates order status (`received → cooking → ready`)
* logs all status changes
* supports specialization (dine_in / takeout / delivery)
* auto-requeues failed jobs
* supports graceful shutdown

### **3. Tracking Service (HTTP API)**

Provides:

* order status API
* full history of each order
* real-time worker status

### **4. Notification Subscriber**

Fanout subscriber that listens for all status updates and prints notifications.

---

# 🧱 **Architectural Diagram**

```
                                +------------------------+
                                |     PostgreSQL DB      |
                                |   (orders, logs, etc.) |
                                +-----------+------------+
                                            ^
                                            |
+-------------+         +-------------------+-----------------+
| HTTP Client | ----->  |   Order Service (API + Publisher)   |
+-------------+         +-------------------+-----------------+
                                    |
                                    | Publishes new orders
                                    v
                        +-----------------------------+
                        |     RabbitMQ Broker         |
                        | (orders_topic, fanout, etc.)|
                        +------+---------------+------+
                               |               |
                               |               |
                     Consumes  |               | Status updates
                               v               v
                     +----------------+   +-------------------------+
                     | Kitchen Worker |   | Notification Subscriber |
                     +----------------+   +-------------------------+

                               ^
                               |
                     +------------------+
                     | Tracking Service |
                     |    (API)         |
                     +------------------+
```

---


# 🧩 **Service-by-Service Breakdown**

---

## **1. Order Service**

### **Responsibilities**

* HTTP endpoint `POST /orders`
* Validates request data
* Calculates total & priority
* Generates order number (`ORD_YYYYMMDD_NNN`)
* Writes to PostgreSQL (transaction)
* Publishes order to RabbitMQ (orders_topic)

### **Routing Key Format**

```
kitchen.{order_type}.{priority}
```

Example:

```
kitchen.delivery.10
```

### **Start**

```
./restaurant-system --mode=order-service --port=3000
```

---

## **2. Kitchen Worker**

### **Responsibilities**

* Consumes messages from RabbitMQ
* Optional specialization (`--order-types=dine_in,takeout`)
* Sets order status to `cooking` → waits → `ready`
* Logs status changes
* Sends fanout notifications
* Keeps heartbeat alive
* Graceful shutdown support

### **Start**

```
./restaurant-system --mode=kitchen-worker --worker-name=chef_anna
```

Specialized:

```
./restaurant-system --mode=kitchen-worker --worker-name=chef_mario --order-types=dine_in
```

---

## **3. Tracking Service**

### Provides API:

* `/orders/{order_number}/status`
* `/orders/{order_number}/history`
* `/workers/status`

### Start:

```
./restaurant-system --mode=tracking-service --port=3002
```

---

## **4. Notification Subscriber**

A simple RabbitMQ consumer that prints events:

```
./restaurant-system --mode=notification-subscriber
```

---

# 📦 **Installation & Setup**

### 1. Clone repo

```
git clone https://github.com/YOUR_USERNAME/restaurant-system.git
cd restaurant-system
```

### 2. Install Go dependencies

```
go mod tidy
```

### 3. Start PostgreSQL & RabbitMQ

If using Docker:

```yaml
version: '3'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: restaurant_user
      POSTGRES_PASSWORD: restaurant_pass
      POSTGRES_DB: restaurant_db
    ports:
      - "5432:5432"

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"
```

### 4. Apply migrations

```
psql -h localhost -U restaurant_user -d restaurant_db -f migrations.sql
```

### 5. Build the app

```
go build -o restaurant-system .
```

---

# 🧪 **Example Request**

### Place an order:

```
curl -X POST http://localhost:3000/orders \
  -H "Content-Type: application/json" \
  -d '{
        "customer_name": "Jane Doe",
        "order_type": "takeout",
        "items": [
          {"name": "Pizza", "quantity": 1, "price": 15.99}
        ]
      }'
```

---

# 📊 **Logging Format**

Every service logs structured JSON:

```json
{
  "timestamp": "2024-12-16T10:30:00Z",
  "level": "INFO",
  "service": "order-service",
  "action": "service_started",
  "message": "Order Service started",
  "hostname": "srv-01",
  "request_id": "startup-001"
}
```


### Data Flow:
1. **Order Creation** → Client → Order Service → PostgreSQL + RabbitMQ
2. **Kitchen Processing** → Kitchen Worker consumes → Updates status → Publishes to fanout
3. **Status Tracking** → Tracking Service provides order history via HTTP API
4. **Notifications** → Notification Subscriber listens to fanout for real-time updates

### Technology Stack:
- **Go** - Core programming language for all services
- **RabbitMQ** - Message broker for service communication
- **PostgreSQL** - Primary data store for orders and status history
- **HTTP APIs** - RESTful interfaces for order creation and tracking


# 📫 **Contact & Support**

Feel free to open issues, discussions, or reach out for improvements.

