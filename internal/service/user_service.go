package service

import (
	"datagateway/internal/model"
	"datagateway/proto/userpb"
	"encoding/json"
	"errors"
	"strings"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"
)

type UserServiceServer struct {
	userpb.UnimplementedUserServiceServer
	DB *gorm.DB
}

func NewUserServiceServer(db *gorm.DB) *UserServiceServer {
	return &UserServiceServer{DB: db}
}

func toProtoUser(u *model.User) *userpb.User {
	if u == nil {
		return nil
	}
	return &userpb.User{
		Id:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (s *UserServiceServer) CreateUser(req *userpb.CreateUserRequest) (*userpb.Response, error) {
	name := strings.TrimSpace(req.GetName())
	email := strings.TrimSpace(req.GetEmail())

	// 檢查必填欄位
	if name == "" {
		return makeResponse(400, "Name is required", nil)
	}
	if email == "" {
		return makeResponse(400, "Email is required", nil)
	}

	u := &model.User{
		Name:  name,
		Email: email,
	}

	// 嘗試寫入 DB
	if err := s.DB.Create(u).Error; err != nil {

		if strings.Contains(err.Error(), "idx_users_email") {
			return makeResponse(500, "Email existed", nil)
		}

		return makeResponse(500, err.Error(), nil)
	}

	return makeResponse(0, "OK", nil)
}

func (s *UserServiceServer) GetUser(req *userpb.GetUserRequest) (*userpb.Response, error) {
	if req.GetId() == 0 {
		return makeResponse(400, "ID is required", nil)
	}

	var u model.User
	err := s.DB.First(&u, req.GetId()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return makeResponse(400, "User not found", nil)
		}
		return makeResponse(500, "Database error", nil)
	}
	return makeResponse(0, "OK", toProtoUser(&u))
}

func (s *UserServiceServer) UpdateUser(req *userpb.UpdateUserRequest) (*userpb.Response, error) {
	if req.GetId() == 0 {
		return makeResponse(400, "ID is required", nil)
	}

	var u model.User

	err := s.DB.First(&u, req.GetId()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return makeResponse(400, "User not found", nil)
		}
		return makeResponse(500, err.Error(), nil)
	}

	if req.GetName() != "" {
		u.Name = req.GetName()
	}
	if req.GetEmail() != "" {
		u.Email = req.GetEmail()
	}

	if err = s.DB.Save(&u).Error; err != nil {
		if strings.Contains(err.Error(), "idx_users_email") {
			return makeResponse(400, "Email address existed", nil)
		}

		return makeResponse(500, err.Error(), nil)
	}

	// 成功回傳統一格式
	return makeResponse(0, "OK", nil)
}

func (s *UserServiceServer) DeleteUser(req *userpb.DeleteUserRequest) (*userpb.Response, error) {
	if req.GetId() == 0 {
		return makeResponse(401, "ID is required", nil)
	}

	var u model.User

	err := s.DB.First(&u, req.GetId()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return makeResponse(402, "User not found", nil)
		}
		return makeResponse(500, err.Error(), nil)
	}

	// 軟刪除：只會填入 DeletedAt，不會真的刪掉資料
	if err = s.DB.Delete(&u).Error; err != nil {
		return makeResponse(500, err.Error(), nil)
	}

	return makeResponse(0, "OK", nil)
}

func (s *UserServiceServer) ListUsers(_ *emptypb.Empty) (*userpb.Response, error) {
	var users []model.User
	if err := s.DB.Find(&users).Error; err != nil {
		return makeResponse(500, "Database error", nil)
	}
	return makeResponse(0, "OK", users)
}

func makeResponse(code int32, message string, data any) (*userpb.Response, error) {
	var pbData *structpb.Value

	if data != nil {
		bytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		var jsonData any
		if err = json.Unmarshal(bytes, &jsonData); err != nil {
			return nil, err
		}

		pbData, err = structpb.NewValue(jsonData)
		if err != nil {
			return nil, err
		}
	}

	return &userpb.Response{
		Code:    code,
		Message: message,
		Data:    pbData,
	}, nil
}
