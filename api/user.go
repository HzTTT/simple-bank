package api

import (
	"net/http"

	db "github.com/HzTTT/simple_bank/db/sqlc"
	"github.com/HzTTT/simple_bank/util"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createUserResponse struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email string `json:"email"`
	PasswordChangedAt int64 `json:"password_changed_at"`
	CreatedAt int64 `json:"created_at"`
}

type createUserRequest struct {
	Username    string `json:"username" binding:"required,alphanum"`
	Password 	string `json:"password" binding:"required,min=6"`
	FullName 	string `json:"full_name" binding:"required"`
	Email 		string `json:"email" binding:"required,email"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}	
	arg := db.CreateUserParams{
		Username: req.Username,
		HashedPassword: hashedPassword,
		FullName: req.FullName,
		Email: req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
    if err != nil {
        if pqErr, ok := err.(*pq.Error); ok {
            switch pqErr.Code.Name() {
            case "unique_violation":
                ctx.JSON(http.StatusForbidden, errorResponse(err))
                return
            }
        }
        ctx.JSON(http.StatusInternalServerError, errorResponse(err))
        return
    }

	rsq := createUserResponse{
		Username: user.Username,
		FullName: user.FullName,
		Email: user.Email,
		PasswordChangedAt: user.PasswordChangedAt.Unix(),
		CreatedAt: user.CreatedAt.Unix(),
	}

	ctx.JSON(http.StatusOK,rsq)
}