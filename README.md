# Frappuccino – Coffee Shop Management System

## Overview  
Frappuccino is a **coffee shop management system** designed to handle orders, inventory, employees, and analytics efficiently. The system is built in **Go** with a **PostgreSQL** database and runs inside **Docker** containers for easy deployment.  

It provides a **REST API** to manage key operations such as **placing orders, tracking inventory, generating sales reports, and managing employees**.  

---

## Technologies Used  
- **Go** – Backend development  
- **PostgreSQL** – Database  
- **Docker & Docker Compose** – Containerization and orchestration  
- **gofumpt** – Code formatting  
- **REST API** – Communication between services  

---

## System Architecture  
The project follows a **modular structure** with separate handlers for different functionalities. Data is stored in **PostgreSQL**, and all business logic is handled in the **Go backend**.

### Entity-Relationship Diagram (ERD)  
Your ERD diagram outlines key database entities such as:
- **Orders** (tracking customer purchases)
- **Products** (menu items)
- **Inventory** (stock management)
- **Employees** (staff details)
- **Transactions** (sales records)
- **Customers** (optional, for tracking loyalty programs)  

---

## Key Features  
✅ **Order Management** – Create, update, and track orders  
✅ **Inventory Tracking** – Monitor stock levels, update ingredients  
✅ **Employee Management** – Add staff, assign roles  
✅ **Sales Reports & Analytics** – Aggregate daily/weekly/monthly sales  
✅ **Database Migration** – JSON to PostgreSQL  

---

## API Endpoints  
| Method | Endpoint | Description |
|--------|---------|-------------|
| **GET** | `/orders` | Get all orders |
| **POST** | `/orders` | Create a new order |
| **GET** | `/orders/{id}` | Get order details |
| **PUT** | `/orders/{id}` | Update an order |
| **DELETE** | `/orders/{id}` | Cancel an order |
| **GET** | `/inventory` | Get inventory status |
| **POST** | `/inventory` | Add new stock |
| **PUT** | `/inventory/{id}` | Update stock details |
| **GET** | `/employees` | Get employee list |
| **POST** | `/employees` | Add new employee |
| **GET** | `/sales/reports` | Generate sales report |
| **GET** | `/analytics/top-products` | Get best-selling products |

---

## How It Works  
1. The **admin** sets up the coffee shop's menu and inventory.  
2. A **barista** takes customer orders, which are recorded in the system.  
3. Inventory levels automatically **update** after an order is placed.  
4. The **system** tracks employee shifts and performance.  
5. Business **owners** can generate reports on sales, top products, and revenue trends.  

---

## Setup & Deployment  
### Run with Docker  
1. Clone the repository  
   ```sh
   git clone https://github.com/your-username/frappuccino.git  
   cd frappuccino  
   ```
2. Start the services  
   ```sh
   docker compose up  
   ```
3. API will be available at `http://localhost:8080`  


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
