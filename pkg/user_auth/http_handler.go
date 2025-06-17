package userauth

import (
	"log"
	"net/http"
	"path/filepath"
	"strconv"
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
	applicantApi.POST("/file", h.UploadPDF)
	applicantApi.Use(AuthMiddleware())
	applicantApi.GET("/profile", h.GetProfile)
	applicantApi.PATCH("/profile", h.UpdateUserInformation)

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
func (h *Handler) UpdateUserInformation(c *gin.Context) {
	var req UserInformationRequest

	// Bind JSON after resetting body
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("[UpdateUserInformation] Invalid JSON:", err)
		h.respondWithError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user ID from JWT context (set by middleware)
	userID := c.GetString("user_id")
	if userID == "" {
		h.respondWithError(c, http.StatusUnauthorized, "user ID missing from token")
		return
	}

	// Convert userID string to int (or whatever type UserInformationRequest.ID expects)
	idInt, err := strconv.Atoi(userID)
	if err != nil {
		log.Println("[InsertUserInformation] Invalid user_id in context:", err)
		h.respondWithError(c, http.StatusInternalServerError, "internal error")
		return
	}
	req.ID = idInt

	// Call service
	err = h.service.UpdateUserInfo(c.Request.Context(), req)
	if err != nil {
		log.Println("[InsertUserInformation] Service Error:", err)
		h.respondWithError(c, http.StatusInternalServerError, "could not upsert user info")
		return
	}

	log.Println("user_id from token:", req.ID)
	log.Printf("request body: %+v\n", req)

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
		"address":    user.Address,
		"vehicle":    user.Vehicle,
	})
}

// login handler
func (h *Handler) Login(c *gin.Context) {
	var request UserLoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "invalid request")
		return
	}

	token, err := h.service.GetUserProfile(c.Request.Context(), request)
	if err != nil {
		h.respondWithError(c, http.StatusUnauthorized, err.Error())
		return
	}
	h.respondWithData(c, http.StatusOK, "login success", gin.H{"token": token})
}

func (h *Handler) UploadPDF(c *gin.Context) {
	email := c.PostForm("email")
	file, err := c.FormFile("pdf")

	if email == "" || err != nil {
		log.Printf("Email: %s, FormFile error: %v\n", email, err)
		h.respondWithError(c, http.StatusBadRequest, map[string]string{"request-parse": "email and pdf are required"})
		return
	}

	if filepath.Ext(file.Filename) != ".pdf" {
		h.respondWithError(c, http.StatusBadRequest, map[string]string{"invalid-file": "only PDF files are allowed"})
		return
	}

	err = h.service.SaveUserPDF(c, email, file)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, map[string]string{"upload-failed": err.Error()})
		return
	}

	h.respondWithData(c, http.StatusOK, "upload successful", map[string]string{"email": email})
}
