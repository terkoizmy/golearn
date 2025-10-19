package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/terkoizmy/golearn/services/user/domain"
	"github.com/terkoizmy/golearn/services/user/repository"
	"github.com/terkoizmy/golearn/services/user/service"
)

type HTTPHandler struct {
	service service.UserService
}

func NewHTTPHandler(service service.UserService) *HTTPHandler {
	return &HTTPHandler{service: service}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with email, name, and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body domain.RegisterRequest true "User registration data"
// @Success 201 {object} domain.User "User created successfully"
// @Failure 400 {object} domain.ErrorResponse "Invalid input"
// @Failure 409 {object} domain.ErrorResponse "Email already exists"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/register [post]
func (h *HTTPHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			c.JSON(http.StatusConflict, domain.ErrorResponse{Error: "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password, returns JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body domain.LoginRequest true "Login credentials"
// @Success 200 {object} domain.LoginResponse "Login successful"
// @Failure 400 {object} domain.ErrorResponse "Invalid input"
// @Failure 401 {object} domain.ErrorResponse "Invalid credentials"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/login [post]
func (h *HTTPHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	response, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to login"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get user information by user ID
// @Tags Users
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} domain.User "User found"
// @Failure 404 {object} domain.ErrorResponse "User not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/users/{id} [get]
func (h *HTTPHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user)
}
