package handler

import (
	"fmt"
	"learning-golang-ddd/application"
	"learning-golang-ddd/domain/entity"
	"learning-golang-ddd/infrastructure/auth"
	"net/http"
	"os"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type AuthenticateHandler struct {
	uAi application.UserAppInterface
	ai  auth.AuthInterface
	ti  auth.TokenInterface
}

func NewAuthenticateHandler(uAi application.UserAppInterface, ai auth.AuthInterface, ti auth.TokenInterface) *AuthenticateHandler {
	return &AuthenticateHandler{
		uAi: uAi,
		ai:  ai,
		ti:  ti,
	}
}

func (h *AuthenticateHandler) Login(c *gin.Context) {
	var user *entity.User
	var tokenErr = map[string]string{}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "invalid json provided")
		return
	}

	// validate request
	validateUser := user.Validate("login")
	if len(validateUser) > 0 {
		c.JSON(http.StatusUnprocessableEntity, validateUser)
		return
	}
	u, userErr := h.uAi.GetUserByFilter(user)
	if userErr != nil {
		c.JSON(http.StatusInternalServerError, userErr)
		return
	}

	// token details
	td, tErr := h.ti.CreateToken(uint64(u.ID))
	if tErr != nil {
		tokenErr["token_error"] = tErr.Error()
		c.JSON(http.StatusUnprocessableEntity, tokenErr)
		return
	}

	// save
	saveErr := h.ai.CreateAuth(uint64(u.ID), td)
	if saveErr != nil {
		c.JSON(http.StatusInternalServerError, saveErr.Error())
		return
	}

	uData := make(map[string]interface{})
	uData["access_token"] = td.AccessToken
	uData["refresh_token"] = td.RefreshToken
	uData["id"] = u.ID
	uData["first_name"] = u.FirstName
	uData["last_name"] = u.LastName

	c.JSON(http.StatusOK, uData)
}

func (h *AuthenticateHandler) Logout(c *gin.Context) {
	// check is the user is authenticated first
	metadata, err := h.ti.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unathorized")
		return
	}
	// if the access token exist and it is still valid, then delete both the access token the refresh token
	deletedErr := h.ai.DeleteTokens(metadata)
	if deletedErr != nil {
		c.JSON(http.StatusUnauthorized, deletedErr.Error())
		return
	}
	c.JSON(http.StatusOK, "Successfully logged out")
}

func (h *AuthenticateHandler) Refresh(c *gin.Context) {
	mapToken := map[string]string{}
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	refreshToken := mapToken["refresh_token"]

	// verify the token
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		// make sure that token method conform to "SigningMethodHMAC"
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})

	// any error may be due to token expiration
	if err != nil {
		c.JSON(http.StatusUnauthorized, err.Error())
		return
	}

	// is token valid?
	if _, ok := token.Claims.(jwt.MapClaims); !ok && !token.Valid {
		c.JSON(http.StatusUnauthorized, err.Error())
		return
	}

	// since token is valid, get the uuid
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		refUuid, ok := claims["refresh_uuid"].(string) // convert interface{} to string
		if !ok {
			c.JSON(http.StatusUnprocessableEntity, "cannot get uuid")
			return
		}
		uId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, "error occured")
			return
		}
		// delete the previous refresh token
		delErr := h.ai.DeleteRefresh(refUuid)
		if delErr != nil { // if any goes wrong
			c.JSON(http.StatusUnauthorized, "unauthorized")
			return
		}
		// create new pairs of refresh and access tokens
		td, createErr := h.ti.CreateToken(uId)
		if createErr != nil {
			c.JSON(http.StatusForbidden, createErr.Error())
			return
		}
		// save the tokens metadata to redis
		saveErr := h.ai.CreateAuth(uId, td)
		if saveErr != nil {
			c.JSON(http.StatusForbidden, saveErr.Error())
			return
		}
		tokens := map[string]string{
			"access_token":  td.AccessToken,
			"refresh_token": td.RefreshToken,
		}
		c.JSON(http.StatusCreated, tokens)
	} else {
		c.JSON(http.StatusUnauthorized, "refresh token expired")
	}
}
