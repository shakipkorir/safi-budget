package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"safi-budget/internal/app"
)

func main() {
	root := app.ResolveProjectRoot(".")
	budgetApp, err := app.NewAppFromProjectRoot(root)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/signup", budgetApp.HandleSignup)
	mux.HandleFunc("/login", budgetApp.HandleLogin)
	mux.HandleFunc("/dashboard", budgetApp.HandleDashboard)
	mux.HandleFunc("/update-revenue", budgetApp.HandleUpdateRevenue)
	mux.HandleFunc("/deduct-expense", budgetApp.HandleDeductExpense)
	mux.HandleFunc("/profile-update", budgetApp.HandleProfileUpdate)
	mux.HandleFunc("/reset", budgetApp.HandleReset)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Safi Budget Engine running smoothly on :%s...\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
