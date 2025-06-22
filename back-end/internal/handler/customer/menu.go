
// internal/handlers/customer/menu.go
package customer

import (
	"net/http"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/utils"
	"strconv"

	"github.com/gorilla/mux"
)

type MenuHandler struct {
	menuRepo repository.MenuRepository
}

func NewMenuHandler(menuRepo repository.MenuRepository) *MenuHandler {
	return &MenuHandler{
		menuRepo: menuRepo,
	}
}

func (h *MenuHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.menuRepo.GetCategories(true) // Only active categories
	if err != nil {
		utils.ErrorResponse(w, "Failed to get categories", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Categories retrieved", categories)
}

func (h *MenuHandler) GetMenuItems(w http.ResponseWriter, r *http.Request) {
	// Get category ID from query params
	categoryIDStr := r.URL.Query().Get("category_id")
	var categoryID *int
	
	if categoryIDStr != "" {
		id, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			utils.ErrorResponse(w, "Invalid category ID", http.StatusBadRequest)
			return
		}
		categoryID = &id
	}
	
	// Get menu items
	items, err := h.menuRepo.GetMenuItems(categoryID, true) // Only available items
	if err != nil {
		utils.ErrorResponse(w, "Failed to get menu items", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Menu items retrieved", items)
}

func (h *MenuHandler) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid menu item ID", http.StatusBadRequest)
		return
	}
	
	item, err := h.menuRepo.GetMenuItemByID(id)
	if err != nil {
		utils.ErrorResponse(w, "Menu item not found", http.StatusNotFound)
		return
	}
	
	if !item.IsAvailable {
		utils.ErrorResponse(w, "Menu item is not available", http.StatusNotFound)
		return
	}
	
	// Get media files
	mediaRepo := repository.NewMediaRepository(h.menuRepo.(*repository.menuRepository).db)
	mediaFiles, _ := mediaRepo.GetByMenuItemID(id)
	item.MediaFiles = mediaFiles
	
	utils.SuccessResponse(w, "Menu item retrieved", item)
}

func (h *MenuHandler) GetMenu360View(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid menu item ID", http.StatusBadRequest)
		return
	}
	
	item, err := h.menuRepo.GetMenuItemByID(id)
	if err != nil {
		utils.ErrorResponse(w, "Menu item not found", http.StatusNotFound)
		return
	}
	
	if item.Image360URL == nil || *item.Image360URL == "" {
		utils.ErrorResponse(w, "360° view not available for this item", http.StatusNotFound)
		return
	}
	
	response := map[string]interface{}{
		"menu_item_id": item.ID,
		"name":         item.Name,
		"image_360_url": item.Image360URL,
		"video_url":    item.VideoURL,
	}
	
	utils.SuccessResponse(w, "360° view data retrieved", response)
}
