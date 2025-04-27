package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/configs" // for config access
	"github.com/GoMudEngine/GoMud/internal/web"
	"github.com/gorilla/mux" // for LockMud
)

func init() {
	cfg := configs.GetConfig() // assume this returns the loaded config struct
	if cfg.EnableAPI != true {
		return // API is disabled; do not register routes
	}
	registerAPIRoutes()

}

func registerAPIRoutes() {
	// Build a subrouter for /api
	router := mux.NewRouter().PathPrefix("/api").Subrouter()

	// Players endpoints
	router.HandleFunc("/players", getPlayersHandler).Methods(http.MethodGet)
	router.HandleFunc("/players", createPlayerHandler).Methods(http.MethodPost)

	// Items endpoints
	router.HandleFunc("/items", listItemsHandler).Methods(http.MethodGet)
	router.HandleFunc("/items", createItemHandler).Methods(http.MethodPost)
	router.HandleFunc("/items/{id}", getItemByIDHandler).Methods(http.MethodGet)
	router.HandleFunc("/items/{id}", updateItemByIDHandler).Methods(http.MethodPatch)

	// Admin commands endpoint
	router.HandleFunc("/commands", executeAdminCommandHandler).Methods(http.MethodPost)

	// Wrap the router with authentication and game lock middleware
	wrapped := web.RunWithMUDLocked(
		web.DoBasicAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			router.ServeHTTP(w, r)
		})),
	)
	http.Handle("/api/", wrapped)
}

// getPlayersHandler handles GET /api/players
func getPlayersHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement logic to list players
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"players": []}`))
}

// createPlayerHandler handles POST /api/players
func createPlayerHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement logic to create a new player
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Player created"}`))
}

// listItemsHandler handles GET /api/items
func listItemsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement logic to list items
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"items": []}`))
}

// createItemHandler handles POST /api/items
func createItemHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement logic to create a new item
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Item created"}`))
}

// getItemByIDHandler handles GET /api/items/{id}
func getItemByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	// TODO: implement logic to fetch item by id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id": "%s", "item": {}}`, id)))
}

// updateItemByIDHandler handles PATCH /api/items/{id}
func updateItemByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	// TODO: implement logic to update item by id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id": "%s", "message": "Item updated"}`, id)))
}

// executeAdminCommandHandler handles POST /api/commands to run an admin command
func executeAdminCommandHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: authenticate request as admin
	// Parse JSON body: {"command": "<admin command string>"}
	var req struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	// TODO: execute the admin command, e.g. usercommands.TryCommand or appropriate function
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message":"Executed command: %q"}`, req.Command)))
}
