package userauth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) MountRoutes(engine *gin.Engine) {
	applicantApi := engine.Group(basePath)
	applicantApi.POST("/register", h.Register)

	protected := applicantApi.Group("/")
	protected.Use(AuthMiddleware())
	protected.GET("/profile", h.GetProfile)
}

func (h *Handler) respondWithError(c *gin.Context, code int, msg interface{}) {
	resp := gin.H{
		"msg": msg,
	}

	c.JSON(code, resp)
	c.Abort()
}

func (h *Handler) respondWithData(c *gin.Context, code int, message interface{}, data interface{}) {
	resp := gin.H{
		"msg":  message,
		"data": data,
	}
	c.JSON(code, resp)
}

func (h *Handler) Register(c *gin.Context) {
	var request UserRegisterRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.respondWithError(c, http.StatusBadRequest, map[string]string{"request-parse": err.Error()})
		return
	}

	errVal := request.Valid()

	if len(errVal) > 0 {
		h.respondWithError(c, http.StatusBadRequest, map[string]interface{}{"invalid-request": errVal})
		return
	}
	token, err := h.service.Register(c, request)
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, map[string]string{"err": err.Error()})
		return
	}
	h.respondWithData(c, http.StatusOK, "success", map[string]string{"token: ": token})
}

// Get the user Profile
func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetString("userID")
	email := c.GetString("email")

	h.respondWithData(c, http.StatusOK, "profile fetched successfully", gin.H{
		"user_id": userID,
		"email":   email,
	})
}
