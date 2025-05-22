package userauth

import (
	"log"
	"net/http"
	"time"

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
	applicantApi.POST("/info", h.InsertUserInformation)

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
	userID, err := h.service.UserRegister(c, request)
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, map[string]string{"err": err.Error()})
		return
	}
	h.respondWithData(c, http.StatusOK, "success", map[string]string{"user_id": userID})
}

// userimformation
func (h *Handler) respondWithSuccess(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{
		"msg": msg,
	})
}

func (h *Handler) InsertUserInformation(c *gin.Context) {
	var req UserInformationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("[InsertUserInformation] Bind Error:", err)
		h.respondWithError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ID == 0 {
		h.respondWithError(c, http.StatusBadRequest, "missing required fields")
		return
	}

	err := h.service.UpdateUserInformation(c.Request.Context(), req)
	if err != nil {
		log.Println("[InsertUserInformation] Service Error:", err)
		h.respondWithError(c, http.StatusInternalServerError, "could not upsert user info")
		return
	}

	h.respondWithSuccess(c, http.StatusOK, "user info upserted successfully")
}

// // update user information
// func (s *service) UserInformation(c *gin.Context) {
// 	var req UserInformationRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request"})
// 		return
// 	}

// 	userUUID, exists := c.Get("user_uuid")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"msg": "user_uuid not found in token"})
// 		return
// 	}

// 	// Assign the user_uuid from token to request struct or pass separately
// 	req.UUID = userUUID.(string)

// 	err := s.repo.UpdateUserInformation(c.Request.Context(), req)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"msg": "could not upsert user info"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"msg": "user info updated"})
// }

// Get the user Profile
func calculateAge(dob time.Time) int {
	now := time.Now()
	years := now.Year() - dob.Year()
	if now.YearDay() < dob.YearDay() {
		years--
	}
	return years
}

func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		log.Println("[GetProfile] Error fetching user: ", err)
		h.respondWithError(c, http.StatusInternalServerError, "failed to fetch profile")
		return
	}

	age := calculateAge(user.DOB)

	h.respondWithData(c, http.StatusOK, "profile fetched successfully", gin.H{
		"user_id":    user.ID,
		"email":      user.Email,
		"dob":        user.DOB.Format("2006-01-02"),
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"gender":     user.Gender,
		"age":        age,
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

	// log.Println("[Login Handler] Request Body Parsed:", request)

	token, err := h.service.GetUserProfile(c.Request.Context(), request)
	if err != nil {
		log.Println("[Token error] ", token)
		h.respondWithError(c, http.StatusUnauthorized, err.Error())
		return
	}
	h.respondWithData(c, http.StatusOK, "login success", gin.H{"token": token})
}
