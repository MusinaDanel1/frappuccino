package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	d "frappuccino/internal/dal"
	h "frappuccino/internal/handler"
	s "frappuccino/internal/service"
	u "frappuccino/internal/utils"
)

func main() {
	flag.Parse()
	if *u.Help {
		u.PrintHelp()
		return
	}

	if !u.CheckPort(*u.Port) {
		log.Fatalf("Invalid port specified!")
	}
	// logFile := *u.Dir + "/log.log"
	logFile := "/log.log"

	dsn := "host=db port=5432 user=latte password=latte dbname=frappuccino sslmode=disable"

	log.Println("Starting inventory handler setup...")
	// Inventory handler
	invRepo, err := d.NewInventoryRepository(dsn)
	if err != nil {
		log.Fatalf("Error creating inventory repository: %v", err)
	}
	log.Println("Inventory repository created successfully.")
	invService := s.NewIngredientService(invRepo)
	invHandler, err := h.NewInventoryHandler(invService, logFile)
	if err != nil {
		log.Fatalf("Error creating inventory handler: %v", err)
	}
	log.Println("Inventory handler created successfully.")
	// Menu handler
	log.Println("Starting menu handler setup...")
	menuRepo, err := d.NewMenuRepository(dsn)
	if err != nil {
		log.Fatalf("Error creating menu repository: %v", err)
	}
	log.Println("Menu repository created successfully.")
	menuService := s.NewMenuItemService(menuRepo)
	menuHandler, err := h.NewMenuHandler(menuService, logFile)
	if err != nil {
		log.Fatalf("Error creating menu handler: %v", err)
	}
	log.Println("Menu handler created successfully.")
	// Order handler
	log.Println("Starting order handler setup...")
	orderRepo, err := d.NewOrderRepository(dsn)
	if err != nil {
		log.Fatalf("Error creating order repository: %v", err)
	}
	log.Println("Order repository created successfully.")
	orderService := s.NewOrderService(orderRepo, invRepo, menuRepo)
	orderHandler, err := h.NewOrderHandler(orderService, logFile)
	if err != nil {
		log.Fatalf("Error creating order Handler: %v", err)
	}
	log.Println("Order handler created successfully.")
	// Aggregations handler
	// reportsService := s.NewReportsService(orderRepo, menuRepo)
	// reportsHandler, err := h.NewReportsHandler(reportsService, logFile)
	// if err != nil {
	// 	log.Fatalf("Error creating reports handler: %v", err)
	// }

	mux := u.NewCustomMux()

	// Orders:
	mux.HandleFunc("POST /orders", orderHandler.CreateOrder) // Create a new order
	mux.HandleFunc("GET /orders", orderHandler.ListOrders)   // Retrieve all orders
	mux.HandleFunc("GET /orders/numberOfOrderedItems", orderHandler.GetOrderedItemsCount)
	mux.HandleFunc("GET /orders/{id}", orderHandler.GetOrder)          // Retrieve a specific order by ID.
	mux.HandleFunc("PUT /orders/{id}", orderHandler.UpdateOrder)       // Update an existing order
	mux.HandleFunc("DELETE /orders/{id}", orderHandler.DeleteOrder)    // Delete an order.
	mux.HandleFunc("POST /orders/{id}/close", orderHandler.CloseOrder) // Close an order.

	// Menu Items
	mux.HandleFunc("POST /menu", menuHandler.CreateMenuItem)        // Add a new menu item.
	mux.HandleFunc("GET /menu", menuHandler.LissMenu)               // Retrieve all menu items
	mux.HandleFunc("GET /menu/{id}", menuHandler.GetMenuItem)       // Retrieve a specific menu item
	mux.HandleFunc("PUT /menu/{id}", menuHandler.UpdateMenuItem)    // Update a menu item.
	mux.HandleFunc("DELETE /menu/{id}", menuHandler.DeleteMenuItem) // Delete a menu item.

	// Inventory
	mux.HandleFunc("POST /inventory", invHandler.CreateIngredient)        // Add a new inventory item.
	mux.HandleFunc("GET /inventory", invHandler.ListInventory)            // Retrieve all inventory items.
	mux.HandleFunc("GET /inventory/{id}", invHandler.GetIngredient)       // Retrieve a specific inventory item.
	mux.HandleFunc("PUT /inventory/{id}", invHandler.UpdateIngredient)    // Update an inventory item.
	mux.HandleFunc("DELETE /inventory/{id}", invHandler.DeleteIngredient) // Delete an inventory item.

	// // Aggregations
	// mux.HandleFunc("GET /reports/total-sales", reportsHandler.GetTotalSales)     // Get the total sales amount.
	// mux.HandleFunc("GET /reports/popular-items", reportsHandler.GetPopularItems) // Get a list of popular menu items.

	address := fmt.Sprintf(":%s", *u.Port)
	fmt.Printf(u.PINK+"Server is starting on: \nhttp://localhost:%s\n"+u.RESET, *u.Port)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatal(err)
	}
}
