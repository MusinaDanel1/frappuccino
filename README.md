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