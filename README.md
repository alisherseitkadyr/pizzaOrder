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
┌─────────────────┐ HTTP ┌──────────────────┐
│ Client App │───────────▶│ Order Service │
│ │ │ (HTTP API) │
└─────────────────┘ └─────────┬────────┘
│
│ PostgreSQL
│ (Orders DB)
│
▼
┌─────────────────┐ ┌──────────────────┐
│ Tracking │ │ RabbitMQ │
│ Service │ │ │
│ (HTTP API) │◀───────────│ Order Queue │
└─────────────────┘ │ Status Exchange│
│ (Fanout) │
│
┌─────────┴────────┐
│ │
┌─────────▼────────┐ │
│ Kitchen Worker │ │
│ (Consumer) │ │
└──────────────────┘ │
┌────────▼────────┐
│ Notification │
│ Subscriber │
│ (Fanout Consumer)│
└─────────────────┘
text


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

