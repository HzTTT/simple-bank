package gapi

import (
	"context"
	"database/sql"

	db "github.com/HzTTT/simple_bank/db/sqlc"
	"github.com/HzTTT/simple_bank/pb"
	"github.com/HzTTT/simple_bank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server)LoginUser(ctx context.Context,req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {

	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "username not fund")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user:%s", err)
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "password not right:")
	}

	assessToken, accseePayload, err := server.tokenMaker.CreateToken(req.GetUsername(), server.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal,err.Error())
	}

	refreshToken, refreshpayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal,err.Error())
	}

	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshpayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    "",
		ClientIp:     "",
		IsBlocked:    false,
		ExpiresAt:    refreshpayload.ExpiredAt,
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal,err.Error())
	}

	rsp := pb.LoginUserResponse{
		SessionId:             session.ID.String(),
		AccessToken:           assessToken,
		AccessTokenExpiresAt:  timestamppb.New(accseePayload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(accseePayload.ExpiredAt),
		User:                  convertUser(user),
	}

	return &rsp, nil
}