package httpapp

import (
	"errors"
	"net/http"

	models "github.com/Nikolay-Yakushev/mango/internal/domain"
	"github.com/gin-gonic/gin"
)

func(a *Adapter) BindError(ctx *gin.Context, err error) {

	a.log.Sugar().Errorf("request failed: %s", err.Error())

	switch {
		case errors.Is(err, models.ForbiddenErr), errors.Is(err, models.TokenInvalidErr):
			ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied",})
		case errors.Is(err, models.TokenExpiredErr):
			ctx.JSON(http.StatusForbidden, gin.H{"error": "token expired",})
		case errors.Is(err, models.NotFoundErr):
			ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied",
			})
		case errors.Is(err, models.ConflictErr):
			ctx.JSON(http.StatusConflict, gin.H{"error": "user already exists",})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(),})
	}
}