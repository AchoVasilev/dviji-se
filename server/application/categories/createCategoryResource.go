package categories

type CreateCategoryResource struct {
	Name     string `json:"name" binding:"required"`
	ImageUrl string `json:"image_url" binding:"required"`
}
