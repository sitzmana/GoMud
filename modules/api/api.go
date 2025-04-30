package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs" // for config access
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/web"
)

func init() {
	cfg := configs.GetConfig() // assume this returns the loaded config struct
	if cfg.EnableAPI != true {
		return // API is disabled; do not register routes
	}
	registerAPIRoutes()

}

func registerAPIRoutes() {
	// Players endpoints
	http.Handle("GET /api/players", web.RunWithMUDLocked(
		web.DoBasicAuth(http.HandlerFunc(getPlayersHandler)),
	))
	http.Handle("POST /api/players", web.RunWithMUDLocked(
		web.DoBasicAuth(http.HandlerFunc(createPlayerHandler)),
	))

	// Items endpoints
	http.Handle("GET /api/items", web.RunWithMUDLocked(
		web.DoBasicAuth(http.HandlerFunc(listItemsHandler)),
	))
	http.Handle("POST /api/items", web.RunWithMUDLocked(
		web.DoBasicAuth(http.HandlerFunc(createItemHandler)),
	))
	http.Handle("GET /api/items/", web.RunWithMUDLocked(
		web.DoBasicAuth(http.HandlerFunc(getItemByIDHandler)),
	))
	http.Handle("PATCH /api/items/", web.RunWithMUDLocked(
		web.DoBasicAuth(http.HandlerFunc(updateItemByIDHandler)),
	))

	// Admin commands endpoint
	http.Handle("POST /api/commands", web.RunWithMUDLocked(
		web.DoBasicAuth(http.HandlerFunc(executeAdminCommandHandler)),
	))
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
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	id := parts[3]
	// TODO: implement logic to fetch item by id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id": "%s", "item": {}}`, id)))
}

// updateItemByIDHandler handles PATCH /api/items/{id}
func updateItemByIDHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	id := parts[3]
	// TODO: implement logic to update item by id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id": "%s", "message": "Item updated"}`, id)))
}

// executeAdminCommandHandler handles POST /api/commands to run an admin command
func executeAdminCommandHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: add admin verification
	// Parse JSON body: {"command": "<admin command string>"}
	var req struct {
		Command string `json:"command"`
		RoomID  *int   `json:"room_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Determine a user context for executing commands (e.g., first active user)
	activeUsers := users.GetAllActiveUsers()
	var adminUser *users.UserRecord
	if len(activeUsers) > 0 {
		adminUser = activeUsers[0]
	}

	// Determine a room context based on JSON or default to room 0
	var cmdRoom *rooms.Room
	if req.RoomID != nil {
		cmdRoom = rooms.LoadRoom(*req.RoomID)
		if cmdRoom == nil {
			http.Error(w, "Invalid room_id", http.StatusBadRequest)
			return
		}
	} else {
		cmdRoom = rooms.LoadRoom(0)
	}

	// Execute the admin command string
	handled, err := usercommands.Command(req.Command, adminUser, cmdRoom, events.EventFlag(0))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !handled {
		http.Error(w, "Command not handled", http.StatusBadRequest)
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Executed command: %q", req.Command),
	})
}
