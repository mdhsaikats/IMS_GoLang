package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Inventory struct {
	ID         string
	Name       string
	Quantity   int
	Category   string
	Price      float64
	ExpiryDate string
}

type User struct {
	Username string
	Password string
	Role     string // Mork, Manager, Staff
}

var inventory []Inventory
var lastOperation string // Tracks the last operation for undo
const fileName = "inventory.txt"
const lowStockThreshold = 10

var currentUser *User

// Define hardcoded users for demo purposes
var users = []User{
	{"mork", "29112003", "Mork"},
	{"manager", "manager123", "Manager"},
	{"staff", "staff123", "Staff"},
}

func login() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	fmt.Print("Enter password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Authenticate user
	for _, user := range users {
		if user.Username == username && user.Password == password {
			currentUser = &user
			fmt.Printf("Welcome %s! You are logged in as %s.\n", username, user.Role)
			return
		}
	}
	fmt.Println("Invalid username or password.")
	login() // Recurse if authentication fails
}

func hasPermission(requiredRole string) bool {
	if currentUser == nil {
		fmt.Println("You must be logged in to perform this action.")
		return false
	}
	roles := map[string]int{
		"Mork":    3,
		"Manager": 2,
		"Staff":   1,
	}

	// If the current user's role has the required access level, return true
	if roles[currentUser.Role] >= roles[requiredRole] {
		return true
	}
	fmt.Println("You don't have permission to perform this action.")
	return false
}

func toLowerCase(str string) string {
	return strings.ToLower(str)
}

func push(name string, quantity int, category string, price float64, expiryDate string) {
	if !hasPermission("Manager") {
		return
	}
	name = toLowerCase(name)
	for i, item := range inventory {
		if toLowerCase(item.Name) == name {
			inventory[i].Quantity += quantity
			lastOperation = "add"
			saveToFile()
			return
		}
	}
	id := generateID()
	inventory = append(inventory, Inventory{ID: id, Name: name, Quantity: quantity, Category: category, Price: price, ExpiryDate: expiryDate})
	lastOperation = "add"
	saveToFile()
}

func showInventory() {
	if !hasPermission("Staff") {
		return
	}
	if len(inventory) == 0 {
		fmt.Println("There is no product in the inventory")
		return
	}
	fmt.Println("List of products:")
	for _, item := range inventory {
		fmt.Printf("Product ID: %s, Name: %s, Quantity: %d, Category: %s, Price: %.2f, Expiry Date: %s\n", item.ID, item.Name, item.Quantity, item.Category, item.Price, item.ExpiryDate)
	}
}

func deadStock() {
	if !hasPermission("Staff") {
		return
	}
	deadStockFound := false
	fmt.Println("Dead Stock:")
	for _, item := range inventory {
		if item.Quantity == 0 {
			fmt.Printf("Product name: %s\n", item.Name)
			deadStockFound = true
		}
	}
	if !deadStockFound {
		fmt.Println("No dead stock found.")
	}
}

func sell(name string, quantity int) {
	if !hasPermission("Staff") {
		return
	}
	name = toLowerCase(name)
	for i, item := range inventory {
		if toLowerCase(item.Name) == name {
			if inventory[i].Quantity >= quantity {
				inventory[i].Quantity -= quantity
				fmt.Printf("Sold %d units of %s\n", quantity, name)
				lastOperation = "sell"
				saveToFile()
				return
			} else {
				fmt.Printf("Not enough quantity of %s in stock.\n", name)
				return
			}
		}
	}
	fmt.Println("Product not found in inventory.")
}

func deleteProduct(name string) {
	if !hasPermission("Mork") {
		return
	}
	name = toLowerCase(name)
	for i, item := range inventory {
		if toLowerCase(item.Name) == name {
			inventory = append(inventory[:i], inventory[i+1:]...)
			fmt.Printf("Product %s deleted from inventory.\n", name)
			lastOperation = "delete"
			saveToFile()
			return
		}
	}
	fmt.Println("Product not found in inventory.")
}

func lowStockAlert() {
	if !hasPermission("Staff") {
		return
	}
	fmt.Println("Low Stock Alert:")
	for _, item := range inventory {
		if item.Quantity < lowStockThreshold {
			fmt.Printf("Product %s has low stock (Quantity: %d)\n", item.Name, item.Quantity)
		}
	}
}

func search(query string) {
	if !hasPermission("Staff") {
		return
	}
	query = toLowerCase(query)
	matchesFound := false

	for _, item := range inventory {
		if toLowerCase(item.Name) == query || item.ID == query {
			matchesFound = true
			fmt.Printf("Found product - ID: %s, Name: %s, Quantity: %d\n", item.ID, item.Name, item.Quantity)
		}
	}

	if !matchesFound {
		fmt.Println("No products found matching the search query.")
	}
}

func exportToCSV() {
	if !hasPermission("Manager") {
		return
	}
	file, err := os.Create("inventory_export.csv")
	if err != nil {
		fmt.Println("Error creating CSV file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Writing header
	writer.Write([]string{"ID", "Name", "Quantity", "Category", "Price", "ExpiryDate"})

	// Writing data
	for _, item := range inventory {
		record := []string{item.ID, item.Name, strconv.Itoa(item.Quantity), item.Category, fmt.Sprintf("%.2f", item.Price), item.ExpiryDate}
		writer.Write(record)
	}

	fmt.Println("Inventory exported to inventory_export.csv")
}

func undoLastOperation() {
	if !hasPermission("Mork") {
		return
	}
	// Based on the last operation, undo it
	if lastOperation == "add" {
		// Remove the last added product
		inventory = inventory[:len(inventory)-1]
		fmt.Println("Last add operation undone.")
	} else if lastOperation == "sell" {
		// Revert the last sale by adding back the quantity
		// Here we assume we are tracking the quantity and name properly
		// You'll need to implement a way to track which product was sold
		fmt.Println("Undoing sell operation is not yet implemented.")
	} else if lastOperation == "delete" {
		// Restore the last deleted product
		// You'll need to track deleted products to restore them
		fmt.Println("Undo delete operation is not yet implemented.")
	} else {
		fmt.Println("No operation to undo.")
	}
}

func loadingScreen(duration time.Duration, steps int) {
	fmt.Print("         Loading\n [")
	sleepDuration := duration / time.Duration(steps)
	for i := 0; i < steps; i++ {
		fmt.Print("#")
		time.Sleep(sleepDuration)
	}
	fmt.Println("]")
}

func printBoxedText(text string) {
	length := len(text)
	fmt.Print("+")
	fmt.Println(strings.Repeat("-", length+2) + "+")
	fmt.Printf("| %s |\n", text)
	fmt.Print("+")
	fmt.Println(strings.Repeat("-", length+2) + "+")
}

func start() {
	login()
	text := "Welcome to IMS"
	printBoxedText(text)
	fmt.Println("Press Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	loadingScreen(2*time.Second, 20)
	fmt.Println("Proceeding with the rest of the program...")
	fmt.Println()
	time.Sleep(time.Second)
}

func loadFromFile() {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("No inventory data found.")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 6 {
			quantity, _ := strconv.Atoi(fields[2])
			price, _ := strconv.ParseFloat(fields[4], 64)
			inventory = append(inventory, Inventory{ID: fields[0], Name: fields[1], Quantity: quantity, Category: fields[3], Price: price, ExpiryDate: fields[5]})
		}
	}
}

func saveToFile() {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error opening file for writing.")
		return
	}
	defer file.Close()

	for _, item := range inventory {
		fmt.Fprintf(file, "%s %s %d %s %.2f %s\n", item.ID, item.Name, item.Quantity, item.Category, item.Price, item.ExpiryDate)
	}
}

func generateID() string {
	timestamp := time.Now().Unix()
	randomNum := rand.Intn(10000) // Using math/rand here
	return fmt.Sprintf("%d-%d", timestamp, randomNum)
}

func main() {
	loadFromFile()
	start()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("  |**Enter your option**|\n1: Add your product\n2: Show the list\n3: Dead stock\n4: Sell\n5: Delete product\n6: Exit\n7: Search for a product\n8: Low Stock Alerts\n9: Export to CSV\n10: Undo Last Operation")
		fmt.Print("Enter your option: ")
		optionInput, _ := reader.ReadString('\n')
		optionInput = strings.TrimSpace(optionInput)
		option, _ := strconv.Atoi(optionInput)

		switch option {
		case 1:
			fmt.Print("Enter product name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			fmt.Print("Enter product quantity: ")
			quantityInput, _ := reader.ReadString('\n')
			quantityInput = strings.TrimSpace(quantityInput)
			quantity, _ := strconv.Atoi(quantityInput)
			fmt.Print("Enter product category: ")
			category, _ := reader.ReadString('\n')
			category = strings.TrimSpace(category)
			fmt.Print("Enter product price: ")
			priceInput, _ := reader.ReadString('\n')
			priceInput = strings.TrimSpace(priceInput)
			price, _ := strconv.ParseFloat(priceInput, 64)
			fmt.Print("Enter expiry date (YYYY-MM-DD): ")
			expiryDate, _ := reader.ReadString('\n')
			expiryDate = strings.TrimSpace(expiryDate)
			push(name, quantity, category, price, expiryDate)
		case 2:
			showInventory()
		case 3:
			deadStock()
		case 4:
			fmt.Print("Enter product name to sell: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			fmt.Print("Enter quantity to sell: ")
			quantityInput, _ := reader.ReadString('\n')
			quantityInput = strings.TrimSpace(quantityInput)
			quantity, _ := strconv.Atoi(quantityInput)
			sell(name, quantity)
		case 5:
			fmt.Print("Enter product name to delete: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			deleteProduct(name)
		case 6:
			fmt.Println("|**Saved**|")
			time.Sleep(500 * time.Millisecond)
			fmt.Println("|**Exiting**|")
			saveToFile()
			return
		case 7:
			fmt.Print("Enter product name or ID to search: ")
			searchQuery, _ := reader.ReadString('\n')
			searchQuery = strings.TrimSpace(searchQuery)
			search(searchQuery)
		case 8:
			lowStockAlert()
		case 9:
			exportToCSV()
		case 10:
			undoLastOperation()
		default:
			fmt.Println("Invalid option.")
		}
	}
}
