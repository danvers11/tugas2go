package main

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql" // Driver MySQL
)

const (
	dbUser = "root"
	dbPass = "password"
	dbName = "orders_by"
)

func main() {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s", dbUser, dbPass, dbName))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	router := gin.Default()

	// Buat endpoint untuk POST /orders
	router.POST("/orders", func(c *gin.Context) {
		var order Order
		err := c.BindJSON(&order)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		tx, err := db.Begin()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		stmt, err := tx.Prepare("INSERT INTO orders (customer_name, ordered_at) VALUES (?, ?)")
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer stmt.Close()

		result, err := stmt.Exec(order.CustomerName, order.OrderedAt)
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		orderId, err := result.LastInsertId()
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		for _, item := range order.Items {
			stmt, err := tx.Prepare("INSERT INTO items (item_code, description, quantity, order_id) VALUES (?, ?, ?, ?)")
			if err != nil {
				tx.Rollback()
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			defer stmt.Close()

			_, err = stmt.Exec(item.ItemCode, item.Description, item.Quantity, orderId)
			if err != nil {
				tx.Rollback()
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
		}

		tx.Commit()
		c.JSON(201, gin.H{"message": "Order created successfully"})
	})

	// Buat endpoint untuk GET /orders
	router.GET("/orders", func(c *gin.Context) {
		rows, err := db.Query("SELECT * FROM orders o JOIN items i ON o.order_id = i.order_id")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var orders []Order
		for rows.Next() {
			var order Order
			var item Item
			err := rows.Scan(&order.OrderId, &order.CustomerName, &order.OrderedAt, &item.ItemId, &item.ItemCode, &item.Description, &item.Quantity)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			order.Items = append(order.Items, item)
		}

		c.JSON(200, orders)
	})

	// Buat endpoint untuk PUT /orders/:orderId
	router.PUT("/orders/:orderId", func(c *gin.Context) {
		orderId := c.Param("orderId")

		var order Order
		err := c.BindJSON(&order)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		tx, err := db.Begin()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		stmt, err := tx.Prepare("UPDATE orders SET customer_name = ?, ordered_at = ? WHERE order_id = ?")
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(order.CustomerName, order.OrderedAt, orderId)
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		stmt, err = tx.Prepare("DELETE FROM items WHERE order_id = ?")
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(orderId)
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		for _, item := range order.Items {
			stmt, err := tx.Prepare("INSERT INTO items (item_code, description, quantity, order_id) VALUES (?, ?, ?, ?)")
			if err != nil {
				tx.Rollback()
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			defer stmt.Close()

			_, err = stmt.Exec(item.ItemCode, item.Description, item.Quantity, orderId)
			if err != nil {
				tx.Rollback()
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
		}

		tx.Commit()
		c.JSON(200, gin.H{"message": "Order updated successfully"})
	})

	// Buat endpoint untuk DELETE /orders/:orderId
	router.DELETE("/orders/:orderId", func(c *gin.Context) {
		orderId := c.Param("orderId")

		stmt, err := db.Prepare("DELETE FROM orders WHERE order_id = ?")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(orderId)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Order deleted successfully"})
	})

	router.Run(":8080")
}
