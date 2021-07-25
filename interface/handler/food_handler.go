package handler

import (
	"fmt"
	"learning-golang-ddd/application"
	"learning-golang-ddd/domain/entity"
	"learning-golang-ddd/infrastructure/auth"
	"learning-golang-ddd/interface/fileupload"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type FoodHandler struct {
	fAi application.FoodAppInterface
	uAi application.UserAppInterface
	fui fileupload.UploadFileInterface
	ai  auth.AuthInterface
	ti  auth.TokenInterface
}

// FoodHandler constructor
func NewFoodHandler(
	fAi application.FoodAppInterface,
	uAi application.UserAppInterface,
	fui fileupload.UploadFileInterface,
	ai auth.AuthInterface,
	ti auth.TokenInterface,
) *FoodHandler {
	return &FoodHandler{
		fAi: fAi,
		uAi: uAi,
		fui: fui,
		ai:  ai,
		ti:  ti,
	}
}

func (h *FoodHandler) SaveFood(c *gin.Context) {
	// check is the user is authenticated first
	metadata, err := h.ti.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unautorized")
		return
	}

	// lookup the metadata in redis:
	uId, err := h.ai.FetchAuth(metadata.TokenUuid)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}

	// we are using FE vuejs, our errors to have keys for easy checking,
	// so we use a map to hold our errors
	var saveFoodError = make(map[string]string)

	title := c.PostForm("title")
	description := c.PostForm("description")
	if fmt.Sprintf("%T", title) != "string" || fmt.Sprintf("%T", description) != "string" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"invalid_json": "invalid_json",
		})
		return
	}

	// we initialize a new food for the purpose of validating:
	// in case the payload is empty or an invalid data type is used
	emptyFood := entity.Food{}
	emptyFood.Title = title
	emptyFood.Description = description
	saveFoodError = emptyFood.Validate("")
	if len(saveFoodError) > 0 {
		c.JSON(http.StatusUnprocessableEntity, saveFoodError)
		return
	}
	file, err := c.FormFile("food_image")
	if err != nil {
		saveFoodError["invalid_file"] = "a valid file is required"
		c.JSON(http.StatusUnprocessableEntity, saveFoodError)
		return
	}
	// check if the user exist
	_, err = h.uAi.GetUser(uId)
	if err != nil {
		c.JSON(http.StatusBadRequest, "user not found, unauthorized")
		return
	}
	uploadedFile, err := h.fui.UploadFile(file)
	if err != nil {
		saveFoodError["upload_error"] = err.Error() // this error can be any we defined the UploadFile method
		c.JSON(http.StatusUnprocessableEntity, saveFoodError)
		return
	}

	var food = entity.Food{}
	food.UserID = uId
	food.Title = title
	food.Description = description
	food.FoodImage = uploadedFile
	savedFood, saveErr := h.fAi.SaveFood(&food)
	if saveErr != nil {
		c.JSON(http.StatusInternalServerError, saveErr)
		return
	}
	c.JSON(http.StatusCreated, savedFood)
}

func (h *FoodHandler) UpdateFood(c *gin.Context) {
	// check if the user is authenticated first
	metadata, err := h.ti.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}

	// lookup the metadata in redis:
	uId, err := h.ai.FetchAuth(metadata.TokenUuid)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}

	// we are using FE vuejs, our errors need to have keys for easy checking,
	// so we use a map to hold our errors
	var updateFoodError = make(map[string]string)

	foodId, err := strconv.ParseUint(c.Param("food_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, "invalid request")
		return
	}

	// since it as multipart form data we sent, we will do a manual check
	// on each item
	title := c.PostForm("title")
	description := c.PostForm("description")
	if fmt.Sprintf("%T", title) != "string" && fmt.Sprintf("%T", description) != "string" {
		c.JSON(http.StatusUnprocessableEntity, "invalid json")
		return
	}

	// we initialize a new food for the purpose of validating:
	// in case the payload is empty or an invalid data type is used
	emptyFood := entity.Food{}
	emptyFood.Title = title
	emptyFood.Description = description
	updateFoodError = emptyFood.Validate("update")
	if len(updateFoodError) > 0 {
		c.JSON(http.StatusUnprocessableEntity, updateFoodError)
		return
	}
	user, err := h.uAi.GetUser(uId)
	if err != nil {
		c.JSON(http.StatusBadRequest, "user not found, unauthorized")
		return
	}

	// check if the food exist
	food, err := h.fAi.GetFood(foodId)
	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	// if the user id doesnt math with the on we have, dont update.
	// this is the case when an authenticated user tries to update
	// someone else post using postman, curl, etc
	if user.ID != uint16(food.UserID) {
		c.JSON(http.StatusUnauthorized, "you are not the owner of this food")
		return
	}

	// since this is an update request, a new image may or may not be given.
	// - If not image is given, an error occurs. We know this that is why we ignored
	//   the error and instead check if the file is nil.
	// - if not nil, we process the file by calling the "UploadFile" method.
	// - if nil, we used the old one whose path is saved in the database
	file, _ := c.FormFile("food_image")
	if file != nil {
		food.FoodImage, err = h.fui.UploadFile(file)
		// since i am using Digital Ocean (DO) spaces to save image,
		// i am appending my DO url here. You can comment this line
		// since you may be using Digital Ocean Spaces.
		food.FoodImage = os.Getenv("DO_SPACES_URL") + food.FoodImage
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"upload_error": err.Error(),
			})
			return
		}
	}

	// we dont need to update user's id
	food.Title = title
	food.Description = description
	food.UpdatedAt = time.Now()
	updatedFood, updateFoodErr := h.fAi.UpdateFood(food)
	if updateFoodErr != nil {
		c.JSON(http.StatusInternalServerError, updateFoodErr)
		return
	}

	c.JSON(http.StatusOK, updatedFood)
}

func (h *FoodHandler) GetAllFood(c *gin.Context) {
	allfood, err := h.fAi.GetAllFood()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, allfood)
}

func (h *FoodHandler) GetFoodAndCreator(c *gin.Context) {
	foodId, err := strconv.ParseUint(c.Param("food_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, "invalid request")
		return
	}
	food, err := h.fAi.GetFood(foodId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	user, err := h.uAi.GetUser(food.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	foodAndUser := map[string]interface{}{
		"food":    food,
		"creator": user.PublicUser(),
	}
	c.JSON(http.StatusOK, foodAndUser)
}

func (h *FoodHandler) DeleteFood(c *gin.Context) {
	metadata, err := h.ti.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	foodId, err := strconv.ParseUint(c.Param("food_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, "invalid request")
		return
	}
	_, err = h.uAi.GetUser(metadata.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	err = h.fAi.DeleteFood(foodId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, "food deleted")
}
