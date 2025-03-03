package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	d "frappuccino/internal/dal"
	h "frappuccino/internal/handler"
	s "frappuccino/internal/service"
	u "frappuccino/internal/utils"
)

func initRepository(dsn string) (*d.InventoryRepository, *d.MenuRepository, *d.OrderRepository, *d.ReportRepository, *sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("error connecting to database: %v", err)
	}

	invRepo, err := d.NewInventoryRepository(db)
	if err != nil {
		return nil, nil, nil, nil, db, fmt.Errorf("error creating inventory repository: %v", err)
	}

	menuRepo, err := d.NewMenuRepository(db)
	if err != nil {
		return nil, nil, nil, nil, db, fmt.Errorf("error creating menu repository: %v", err)
	}

	orderRepo, err := d.NewOrderRepository(db)
	if err != nil {
		return nil, nil, nil, nil, db, fmt.Errorf("error creating order repository: %v", err)
	}

	reportRepo, err := d.NewReportRepository(db)
	if err != nil {
		return nil, nil, nil, nil, db, fmt.Errorf("error creating report repository: %v", err)
	}

	return invRepo, menuRepo, orderRepo, reportRepo, db, nil
}

func main() {
	logFile := "/log.log"
	dsn := "host=db port=5432 user=latte password=latte dbname=frappuccino sslmode=disable"

	log.Println("Starting application setup...")

	invRepo, menuRepo, orderRepo, reportRepo, db, err := initRepository(dsn)
	if err != nil {
		log.Fatalf("Initialization error: %v", err)
	}
	defer db.Close()

	// create services
	invService := s.NewIngredientService(invRepo)
	menuService := s.NewMenuItemService(menuRepo)
	orderService := s.NewOrderService(orderRepo, invRepo, menuRepo)
	reportsService := s.NewReportService(reportRepo)

	// create handlers
	invHandler, err := h.NewInventoryHandler(invService, logFile)
	if err != nil {
		log.Fatalf("Error creating inventory handler: %v", err)
	}

	menuHandler, err := h.NewMenuHandler(menuService, logFile)
	if err != nil {
		log.Fatalf("Error creating menu handler: %v", err)
	}

	orderHandler, err := h.NewOrderHandler(orderService, logFile)
	if err != nil {
		log.Fatalf("Error creating order handler: %v", err)
	}

	reportsHandler, err := h.NewReportsHandler(reportsService, logFile)
	if err != nil {
		log.Fatalf("Error creating reports handler: %v", err)
	}

	mux := u.NewCustomMux()

	// Orders:
	mux.HandleFunc("POST /orders/batch-process", orderHandler.BatchProcessOrders)
	mux.HandleFunc("POST /orders", orderHandler.CreateOrder) // Create a new order
	mux.HandleFunc("GET /orders", orderHandler.ListOrders)   //Retrieve all orders
	mux.HandleFunc("GET /orders/numberOfOrderedItems", orderHandler.GetOrderedItemsCount)
	mux.HandleFunc("GET /orders/{id}", orderHandler.GetOrder)          // Retrieve a specific order by ID
	mux.HandleFunc("PUT /orders/{id}", orderHandler.UpdateOrder)       // Update an existing order
	mux.HandleFunc("DELETE /orders/{id}", orderHandler.DeleteOrder)    // Delete an order
	mux.HandleFunc("POST /orders/{id}/close", orderHandler.CloseOrder) //Close an order

	// Menu:
	mux.HandleFunc("POST /menu", menuHandler.CreateMenuItem)        // Add a new menu item
	mux.HandleFunc("GET /menu", menuHandler.LissMenu)               // Retrieve all menu items
	mux.HandleFunc("GET /menu/{id}", menuHandler.GetMenuItem)       // Retrieve a specific menu item
	mux.HandleFunc("PUT /menu/{id}", menuHandler.UpdateMenuItem)    // Update a menu item
	mux.HandleFunc("DELETE /menu/{id}", menuHandler.DeleteMenuItem) // Delete a menu item

	// Inventory:
	mux.HandleFunc("POST /inventory", invHandler.CreateIngredient) // Add a new inventory item
	mux.HandleFunc("GET /inventory/getLeftOvers", invHandler.GetLeftOvers)
	mux.HandleFunc("GET /inventory", invHandler.ListInventory)            // Retrieve all inventory items
	mux.HandleFunc("GET /inventory/{id}", invHandler.GetIngredient)       // Retrieve a specific inventory item
	mux.HandleFunc("PUT /inventory/{id}", invHandler.UpdateIngredient)    // Update an inventory item
	mux.HandleFunc("DELETE /inventory/{id}", invHandler.DeleteIngredient) // Delete an inventory item

	// Reports:
	mux.HandleFunc("GET /reports/search", reportsHandler.HandleSearch)
	mux.HandleFunc("GET /reports/total-sales", reportsHandler.GetTotalSales)     // Get the total sales amount
	mux.HandleFunc("GET /reports/popular-items", reportsHandler.GetPopularItems) // Get a list of popular menu items
	mux.HandleFunc("GET /reports/orderedItemsByPeriod", reportsHandler.GetOrderedItemsByPeriod)

	address := fmt.Sprintf(":%s", *u.Port)
	fmt.Printf("Server is starting on: \nhttp://localhost:%s\n", *u.Port)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatal(err)
	}
}
