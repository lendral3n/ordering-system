// internal/handlers/staff/menu.go
package staff

import (
	"encoding/json"
	"net/http"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/services/media"
	"lendral3n/ordering-system/internal/utils"
	"strconv"

	"github.com/gorilla/mux"
)

type MenuHandler struct {
	menuRepo     repository.MenuRepository
	mediaRepo    repository.MediaRepository
	mediaService *media.Service
}

func NewMenuHandler(
	menuRepo repository.MenuRepository,
	mediaRepo repository.MediaRepository,
	mediaService *media.Service,
) *MenuHandler {
	return &MenuHandler{
		menuRepo:     menuRepo,
		mediaRepo:    mediaRepo,
		mediaService: mediaService,
	}
}

// Category handlers
func (h *MenuHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") != "false"
	
	categories, err := h.menuRepo.GetCategories(activeOnly)
	if err != nil {
		utils.ErrorResponse(w, "Failed to get categories", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Categories retrieved", categories)
}

func (h *MenuHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category models.MenuCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate
	if category.Name == "" {
		utils.ErrorResponse(w, "Category name is required", http.StatusBadRequest)
		return
	}
	
	if err := h.menuRepo.CreateCategory(&category); err != nil {
		utils.ErrorResponse(w, "Failed to create category", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Category created", category)
}

func (h *MenuHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	
	var category models.MenuCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	category.ID = categoryID
	
	if err := h.menuRepo.UpdateCategory(&category); err != nil {
		utils.ErrorResponse(w, "Failed to update category", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Category updated", category)
}

// Menu item handlers
func (h *MenuHandler) GetMenuItems(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := r.URL.Query().Get("category_id")
	availableOnly := r.URL.Query().Get("available_only") != "false"
	
	var categoryID *int
	if categoryIDStr != "" {
		id, err := strconv.Atoi(categoryIDStr)
		if err == nil {
			categoryID = &id
		}
	}
	
	items, err := h.menuRepo.GetMenuItems(categoryID, availableOnly)
	if err != nil {
		utils.ErrorResponse(w, "Failed to get menu items", http.StatusInternalServerError)
		return
	}
	
	// Get media files for each item
	for i := range items {
		mediaFiles, _ := h.mediaRepo.GetByMenuItemID(items[i].ID)
		items[i].MediaFiles = mediaFiles
	}
	
	utils.SuccessResponse(w, "Menu items retrieved", items)
}

func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	var item models.MenuItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate
	if item.Name == "" || item.CategoryID == 0 || item.Price <= 0 {
		utils.ErrorResponse(w, "Invalid menu item data", http.StatusBadRequest)
		return
	}
	
	if err := h.menuRepo.CreateMenuItem(&item); err != nil {
		utils.ErrorResponse(w, "Failed to create menu item", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Menu item created", item)
}

func (h *MenuHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid item ID", http.StatusBadRequest)
		return
	}
	
	var item models.MenuItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	item.ID = itemID
	
	if err := h.menuRepo.UpdateMenuItem(&item); err != nil {
		utils.ErrorResponse(w, "Failed to update menu item", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Menu item updated", item)
}

func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid item ID", http.StatusBadRequest)
		return
	}
	
	// Get media files to delete
	mediaFiles, _ := h.mediaRepo.GetByMenuItemID(itemID)
	
	// Delete menu item
	if err := h.menuRepo.DeleteMenuItem(itemID); err != nil {
		utils.ErrorResponse(w, "Failed to delete menu item", http.StatusInternalServerError)
		return
	}
	
	// Delete associated media files
	for _, media := range mediaFiles {
		h.mediaService.DeleteFile(media.FileURL)
		h.mediaRepo.Delete(media.ID)
	}
	
	utils.SuccessResponse(w, "Menu item deleted", nil)
}

func (h *MenuHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid item ID", http.StatusBadRequest)
		return
	}
	
	// Parse multipart form (max 50MB)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		utils.ErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	
	// Get file type
	fileType := r.FormValue("type")
	if fileType == "" {
		fileType = "image"
	}
	
	// Get uploaded files
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		utils.ErrorResponse(w, "No files uploaded", http.StatusBadRequest)
		return
	}
	
	// Get user ID from context
	userID := r.Context().Value("user_id").(int)
	
	uploadedFiles := make([]models.MediaFile, 0)
	
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			utils.ErrorResponse(w, "Failed to open file", http.StatusBadRequest)
			return
		}
		defer file.Close()
		
		// Upload file
		result, err := h.mediaService.UploadFile(file, fileHeader)
		if err != nil {
			utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		// Save media record
		mediaFile := models.MediaFile{
			FileType:     models.MediaFileType(fileType),
			FileURL:      result.FileURL,
			ThumbnailURL: &result.ThumbnailURL,
			FileSize:     result.FileSize,
			MimeType:     result.MimeType,
			MenuItemID:   &itemID,
			UploadedBy:   userID,
		}
		
		if err := h.mediaRepo.Create(&mediaFile); err != nil {
			// Rollback - delete uploaded file
			h.mediaService.DeleteFile(result.FileURL)
			utils.ErrorResponse(w, "Failed to save media record", http.StatusInternalServerError)
			return
		}
		
		uploadedFiles = append(uploadedFiles, mediaFile)
		
		// Update menu item URLs based on file type
		if fileType == "image" && len(uploadedFiles) == 1 {
			// First image becomes the main image
			item, _ := h.menuRepo.GetMenuItemByID(itemID)
			if item != nil {
				item.ImageURL = &result.FileURL
				h.menuRepo.UpdateMenuItem(item)
			}
		} else if fileType == "image_360" {
			item, _ := h.menuRepo.GetMenuItemByID(itemID)
			if item != nil {
				item.Image360URL = &result.FileURL
				h.menuRepo.UpdateMenuItem(item)
			}
		} else if fileType == "video" {
			item, _ := h.menuRepo.GetMenuItemByID(itemID)
			if item != nil {
				item.VideoURL = &result.FileURL
				h.menuRepo.UpdateMenuItem(item)
			}
		}
	}
	
	utils.SuccessResponse(w, "Files uploaded successfully", uploadedFiles)
}

func (h *MenuHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid item ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Quantity int    `json:"quantity"`
		Reason   string `json:"reason"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.menuRepo.UpdateStock(itemID, req.Quantity); err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Log inventory change
	userID := r.Context().Value("user_id").(int)
	// Create inventory log (implementation depends on inventory repo)
	
	utils.SuccessResponse(w, "Stock updated", nil)
}
