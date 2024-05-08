package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

var registeredMap = make(map[string]string)

var uname = ""

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/menu", menuHandler)
	http.HandleFunc("/place_order", placeOrderHandler)
	http.HandleFunc("/bill", billHandler)

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

	type MenuItem struct {
		Name  string
		Price float64
	}

	menu := []MenuItem{
		{Name: "Doner", Price: 1200.00},
		{Name: "Pide", Price: 1200.00},
		{Name: "Burger", Price: 990.00},
		{Name: "Coke", Price: 320.00},
		{Name: "Piko", Price: 300.00},
		{Name: "Sprite", Price: 310.00},
		{Name: "Airan", Price: 120.00},
		{Name: "Water", Price: 200.00},
	}

	tmpl.Execute(w, menu)
}

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

	if len(orderedItems) == 0 {
		http.Error(w, "No items ordered", http.StatusBadRequest)
		return
	}
	redirectURL := "/bill?items=" + strings.Join(orderedItems, ", ")
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
