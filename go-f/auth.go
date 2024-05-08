package main

import (
	"fmt"
	"net/http"
)

func registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// Handle user registration
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Here you can register the user in your database
	// For simplicity, let's just print the username and password
	fmt.Printf("Registered user: %s, Password: %s\n", username, password)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
