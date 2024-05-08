package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var registeredMap = make(map[string]string)

var (
	menu     = make(map[string]float64)
	menuLock sync.Mutex
)

type MenuItem struct {
	Name  string
	Price float64
}

var uname = ""

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/menu", menuHandler)
	http.HandleFunc("/place_order", placeOrderHandler)
	http.HandleFunc("/bill", billHandler)
	http.HandleFunc("/add_menu", addMenuHandler)
	http.HandleFunc("/delete_menu", deleteMenuHandler)
	http.HandleFunc("/add_menu_item", addMenuItemHandler)
	http.HandleFunc("/delete_menu_item", deleteMenuItemHandler)

	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/home.html"))
	tmpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		storedPassword, ok := registeredMap[username]
		if !ok || storedPassword != password {
			http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
			return
		}

		redirect := r.URL.Query().Get("redirect")
		if redirect != "" {
			http.Redirect(w, r, redirect, http.StatusSeeOther)
			return
		}

		uname = username

		http.Redirect(w, r, "/menu", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/login.html"))
	tmpl.Execute(w, nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		if password != confirmPassword {
			http.Redirect(w, r, "/register?error=password", http.StatusSeeOther)
			return
		}

		if _, ok := registeredMap[username]; ok {
			http.Redirect(w, r, "/register?error=exists", http.StatusSeeOther)
			return
		}

		uname = username

		registeredMap[username] = password
		http.Redirect(w, r, "/menu", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/register.html"))
	tmpl.Execute(w, nil)
}

func menuHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/menu.html"))

	menuLock.Lock()
	defer menuLock.Unlock()

	menuItems := make([]MenuItem, 0, len(menu))
	for name, price := range menu {
		menuItems = append(menuItems, MenuItem{Name: name, Price: price})
	}

	fmt.Println("Menu Items:", menuItems) // Debugging output

	err := tmpl.Execute(w, menuItems)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// func placeOrderHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	err := r.ParseForm()
// 	if err != nil {
// 		http.Error(w, "Error parsing form", http.StatusInternalServerError)
// 		return
// 	}

// 	orderedItems := r.Form["item"]
// 	fmt.Println("Ordered items:", orderedItems)

// 	if len(orderedItems) == 0 {
// 		http.Error(w, "No items ordered", http.StatusBadRequest)
// 		return
// 	}
// 	redirectURL := "/bill?items=" + strings.Join(orderedItems, ", ")
// 	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
// }

func placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	orderedItems := r.Form["item"]
	fmt.Println("Ordered items:", orderedItems)
	var validOrderedItems []string
	menuLock.Lock()
	defer menuLock.Unlock()
	for _, item := range orderedItems {
		if _, exists := menu[item]; exists {
			validOrderedItems = append(validOrderedItems, item)
		}
	}

	if len(validOrderedItems) == 0 {
		http.Error(w, "No valid items ordered", http.StatusBadRequest)
		return
	}

	redirectURL := "/bill?items=" + strings.Join(validOrderedItems, ", ")
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func billHandler(w http.ResponseWriter, r *http.Request) {
	items := r.FormValue("items")

	orderedItems := strings.Split(items, ",")

	totalPrice := calculateTotalPrice(orderedItems)

	tmpl := template.Must(template.ParseFiles("templates/bill.html"))
	tmpl.Execute(w, struct {
		Username     string
		OrderedItems string
		Total        float64
	}{
		Username:     uname,
		OrderedItems: items,
		Total:        totalPrice,
	})
}

func calculateTotalPrice(items []string) float64 {
	itemPrices := map[string]float64{
		"Doner":  1200.00,
		"Pide":   1200.00,
		"Burger": 990.00,
		"Coke":   320.00,
		"Piko":   300.00,
		"Sprite": 310.00,
		"Airan":  120.00,
		"Water":  200.00,
	}

	var totalPrice float64
	for _, item := range items {
		totalPrice += itemPrices[item]
	}
	return totalPrice
}

func addMenuHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/add_menu.html"))
	tmpl.Execute(w, nil)
}

func addMenuItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	name := r.FormValue("name")
	priceStr := r.FormValue("price")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		http.Error(w, "Invalid price", http.StatusBadRequest)
		return
	}

	menuLock.Lock()
	defer menuLock.Unlock()

	if _, exists := menu[name]; exists {
		http.Error(w, "Item already exists", http.StatusBadRequest)
		return
	}

	menu[name] = price

	http.Redirect(w, r, "/menu", http.StatusSeeOther)
}

func deleteMenuHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/delete.html"))
	tmpl.Execute(w, nil)
}

func deleteMenuItemHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/delete.html"))

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	item := r.FormValue("item")

	menuLock.Lock()
	defer menuLock.Unlock()

	if _, exists := menu[item]; !exists {
		http.Error(w, "Item not found", http.StatusBadRequest)
		return
	}

	menuItems := make([]MenuItem, 0, len(menu))
	for name, price := range menu {
		menuItems = append(menuItems, MenuItem{Name: name, Price: price})
	}

	fmt.Println("Menu Items:", menuItems) // Debugging output

	err := tmpl.Execute(w, menuItems)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	delete(menu, item)

	http.Redirect(w, r, "/menu", http.StatusSeeOther)
}
