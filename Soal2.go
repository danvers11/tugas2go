// File: main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

const (
	dbUser = "root"
	dbPass = "password"
	dbName = "orders_by"
)

// Define the structure for an order
type Order struct {
	OrderId      int    `json:"order_id"`
	CustomerName string `json:"customer_name"`
	OrderedAt    string `json:"ordered_at"`
	Items        []Item `json:"items"`
}

// Define the structure for an item
type Item struct {
	ItemId      int    `json:"item_id"`
	ItemCode    string `json:"item_code"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	OrderId     int    `json:"order_id"`
}

// SQL statement to create tables
var createTablesSQL = `
CREATE TABLE IF NOT EXISTS orders (
    order_id INT AUTO_INCREMENT PRIMARY KEY,
    customer_name VARCHAR(255),
    ordered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS items (
    item_id INT AUTO_INCREMENT PRIMARY KEY,
    item_code VARCHAR(255),
    description VARCHAR(255),
    quantity INT,
    order_id INT,
    FOREIGN KEY (order_id) REFERENCES orders(order_id)
);
`

func main() {
	// Connect to MySQL database
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s", dbUser, dbPass, dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Run SQL statement to create tables
	_, err = db.Exec(createTablesSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Create Gin router
	router := gin.Default()

	// Endpoint for creating a new order
	router.POST("/orders", func(c *gin.Context) {
		var order Order
		if err := c.ShouldBindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Insert the order into the database
		result, err := db.Exec("INSERT INTO orders (customer_name, ordered_at) VALUES (?, ?)", order.CustomerName, order.OrderedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get the last inserted order ID
		orderId, err := result.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Insert items into the database
		for _, item := range order.Items {
			_, err := db.Exec("INSERT INTO items (item_code, description, quantity, order_id) VALUES (?, ?, ?, ?)", item.ItemCode, item.Description, item.Quantity, orderId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Order created successfully"})
	})

	// Endpoint for getting all orders with items
	router.GET("/orders", func(c *gin.Context) {
		rows, err := db.Query("SELECT * FROM orders o JOIN items i ON o.order_id = i.order_id")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var orders []Order
		for rows.Next() {
			var order Order
			var item Item
			err := rows.Scan(&order.OrderId, &order.CustomerName, &order.OrderedAt, &item.ItemId, &item.ItemCode, &item.Description, &item.Quantity)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			order.Items = append(order.Items, item)
		}

		c.JSON(http.StatusOK, orders)
	})

	// Endpoint for updating an order
	router.PUT("/orders/:orderId", func(c *gin.Context) {
		orderId := c.Param("orderId")

		var order Order
		if err := c.ShouldBindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Update the order details
		_, err = tx.Exec("UPDATE orders SET customer_name = ?, ordered_at = ? WHERE order_id = ?", order.CustomerName, order.OrderedAt, orderId)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Delete existing items for the order
		_, err = tx.Exec("DELETE FROM items WHERE order_id = ?", orderId)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Insert new items for the order
		for _, item := range order.Items {
			_, err := tx.Exec("INSERT INTO items (item_code, description, quantity, order_id) VALUES (?, ?, ?, ?)", item.ItemCode, item.Description, item.Quantity, orderId)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		// Commit the transaction
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "Order updated successfully"})
	})

	// Endpoint for deleting an order
	router.DELETE("/orders/:orderId", func(c *gin.Context) {
		orderId := c.Param("orderId")

		// Start a transaction
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Delete items for the order
		_, err = tx.Exec("DELETE FROM items WHERE order_id = ?", orderId)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Delete the order
		_, err = tx.Exec("DELETE FROM orders WHERE order_id = ?", orderId)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Commit the transaction
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
	})

	// Run the server on port 8080
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
