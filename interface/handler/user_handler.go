package handler

import (
	"learning-golang-ddd/application"
	"learning-golang-ddd/domain/entity"
	"learning-golang-ddd/infrastructure/auth"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UsersHandler struct defines the dependencies that will be used
type UsersHandler struct {
	uAi application.UserAppInterface
	ai  auth.AuthInterface
	ti  auth.TokenInterface
}

// UsersHandler constructor
func NewUsersHandler(uAi application.UserAppInterface, ai auth.AuthInterface, ti auth.TokenInterface) *UsersHandler {
	return &UsersHandler{
		uAi: uAi,
		ai:  ai,
		ti:  ti,
	}
}

func (h *UsersHandler) SaveUser(c *gin.Context) {
	var user entity.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"invalid_json": "invalid_json",
		})
		return
	}

	// validate the request
	validateErr := user.Validate("")
	if len(validateErr) > 0 {
		c.JSON(http.StatusUnprocessableEntity, validateErr)
		return
	}

	err := user.BeforeSave()
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, "unable to save credentials")
		return
	}

	newUser, saveErr := h.uAi.SaveUser(&user)
	if saveErr != nil {
		c.JSON(http.StatusInternalServerError, saveErr)
		return
	}

	c.JSON(http.StatusCreated, newUser.PublicUser())
}

func (h *UsersHandler) GetUsers(c *gin.Context) {
	var users entity.Users // customize user
	var err error
	users, err = h.uAi.GetUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, users.PublicUsers())
}

func (h *UsersHandler) GetUser(c *gin.Context) {
	uId, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.uAi.GetUser(uId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, user.PublicUser())
}
