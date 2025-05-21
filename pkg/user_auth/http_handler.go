package userauth

import (
	"log"
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
	applicantApi.POST("/login", h.Login)

	applicantApi.Use(AuthMiddleware())
	applicantApi.GET("/profile", h.GetProfile)
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
	userID, err := h.service.Register(c, request)
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, map[string]string{"err": err.Error()})
		return
	}
	h.respondWithData(c, http.StatusOK, "success", map[string]string{"user_id": userID})
}

// Get the user Profile
func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		log.Println("[GetProfile] Error fetching user: ", err)
		h.respondWithError(c, http.StatusInternalServerError, "failed to fetched profile")
		return
	}

	log.Printf("[GetProfile] User fetched: %+v\n", user)
	// email := c.GetString("email")
	// dob := c.GetString("age")
	// firstName := c.GetString("first_name")
	// lastName := c.GetString("last_name")
	// gender := c.GetString("gender")
	// age := c.GetString("age")
	// //debug
	log.Printf("[GetProfile] User fetched: %+v\n", user)

	h.respondWithData(c, http.StatusOK, "profile fetched successfully", gin.H{
		"user_id":    user.ID,
		"email":      user.Email,
		"dob":        user.DOB.Format("2006-01-02"),
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"gender":     user.Gender,
		"age":        user.Age,
	})
}

// login handler
func (h *Handler) Login(c *gin.Context) {
	log.Println("[Login Handler] called")
	var request UserLoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Println("[Login Handler] JSON Bind Error:", err)
		h.respondWithError(c, http.StatusBadRequest, "invalid request")
		return
	}

	log.Println("[Login Handler] Request Body Parsed:", request)

	token, err := h.service.Login(c.Request.Context(), request)
	if err != nil {
		log.Println("[Token error] ", token)
		h.respondWithError(c, http.StatusUnauthorized, err.Error())
		return
	}
	h.respondWithData(c, http.StatusOK, "login success", gin.H{"token": token})

	// // validate email
	// if !emailregex.MatchString(req.Email)
	// if err != nil {
	// 	c.JSON(400, gin.H.StatusBadRequest)
	// }
}
