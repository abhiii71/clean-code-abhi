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

	if err := c.ShouldBindJSON(request); err != nil {
		h.respondWithError(c, http.StatusBadRequest, map[string]string{"request-parse": err.Error()})
		return
	}

	errVal := request.Valid()

	if len(errVal) > 0 {
		h.respondWithError(c, http.StatusBadRequest, map[string]interface{}{"invalid-request": errVal})
		return
	}
	err := h.service.Register(c, request)

	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, map[string]string{"err": err.Error()})
		return
	}
	h.respondWithData(c, http.StatusOK, "success", nil)
}
