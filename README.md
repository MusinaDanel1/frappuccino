Frappuccino - Coffee Shop Management System

Frappuccino is a coffee shop management system designed to streamline operations, manage orders, track sales, and enhance customer experience.

Features

Order Management

Create, edit, and cancel orders.

Track order status (processing, preparing, completed, etc.).

Support for pre-orders and delivery.

Customer Management

Store customer details (name, contact info, preferences).

Implement a loyalty program (reward points, discounts, personalized offers).

Menu & Product Management

Manage a catalog of drinks and food items.

Track ingredient availability.

Set pricing and promotions.

Financial Reports & Analytics

Generate sales reports and product popularity statistics.

Track revenue and expenses.

Integration with payment systems.

Administration

Manage user roles and access permissions.

Configure coffee shop settings (working hours, locations, etc.).

Integrations

Connect with POS systems and delivery services.

CRM and marketing tool integrations.

Tech Stack

Backend: Go + PostgreSQL

Containerization: Docker (docker compose up)

Code Formatting: gofumpt

API: SQL-based handlers for database interactions

Reports & Aggregation: Specialized endpoints for analytics

Getting Started

Prerequisites

Docker

Go

PostgreSQL

Installation

# Clone the repository
git clone https://github.com/your-username/frappuccino.git
cd frappuccino

# Start the services
docker compose up --build

Configuration

Update .env with your database credentials and other settings.

# Entity-Relationship Diagram (ERD)

```mermaid
erDiagram
    ORDERS {
        SERIAL order_id PK
        VARCHAR customer_name 
        JSONB special_instructions
        NUMERIC total_amount 
        ENUM status 
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
    ORDER_ITEMS {
        SERIAL order_item_id PK
        INT order_id FK
        INT menu_item_id FK
        INT quantity 
        NUMERIC price_at_order 
        JSONB customization_options
    }
    MENU_ITEMS {
        SERIAL menu_item_id PK
        VARCHAR name 
        TEXT description
        NUMERIC price 
        TEXT[] categories 
        TEXT[] allergens
        JSONB metadata
    }
    MENU_ITEM_INGREDIENTS {
        SERIAL menu_item_ingredient_id PK
        INT menu_item_id FK
        INT ingredient_id FK
        NUMERIC quantity 
    }
    INVENTORY {
        SERIAL ingredient_id PK
        VARCHAR name 
        NUMERIC quantity 
        VARCHAR unit
        TIMESTAMPTZ last_updated
    }
    ORDER_STATUS_HISTORY {
        SERIAL status_id PK
        INT order_id FK
        ENUM status 
        TIMESTAMPTZ changed_at
    }
    PRICE_HISTORY {
        SERIAL price_id PK
        INT menu_item_id FK
        NUMERIC old_price
        NUMERIC new_price
        TIMESTAMPTZ changed_at
    }
    INVENTORY_TRANSACTIONS {
        SERIAL transaction_id PK
        INT ingredient_id FK
        NUMERIC quantity_change
        ENUM transaction_type 
        TIMESTAMPTZ created_at
    }

    ORDERS ||--o{ ORDER_ITEMS : "has"
    MENU_ITEMS ||--o{ ORDER_ITEMS : "contains"
    MENU_ITEMS ||--o{ MENU_ITEM_INGREDIENTS : "includes"
    INVENTORY ||--o{ MENU_ITEM_INGREDIENTS : "contains"
    ORDERS ||--o{ ORDER_STATUS_HISTORY : "tracks"
    MENU_ITEMS ||--o{ PRICE_HISTORY : "has"
    INVENTORY ||--o{ INVENTORY_TRANSACTIONS : "logs"
