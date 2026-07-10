package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_IsBanned_Active(t *testing.T) {
	u := &User{Status: UserStatusActive}
	assert.False(t, u.IsBanned())
}

func TestUser_IsBanned_Banned(t *testing.T) {
	u := &User{Status: UserStatusBanned}
	assert.True(t, u.IsBanned())
}

func TestUser_IsBanned_DefaultStatus(t *testing.T) {
	u := &User{} // 默认 Status = 0
	assert.False(t, u.IsBanned())
}

func TestUser_CanCheckin_Active(t *testing.T) {
	u := &User{Status: UserStatusActive}
	assert.NoError(t, u.CanCheckin())
}

func TestUser_CanCheckin_Banned(t *testing.T) {
	u := &User{Status: UserStatusBanned}
	err := u.CanCheckin()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suspended")
}

func TestUser_DisplayName_WithNickname(t *testing.T) {
	u := &User{Nickname: "武术达人"}
	assert.Equal(t, "武术达人", u.DisplayName())
}

func TestUser_DisplayName_EmptyNickname(t *testing.T) {
	u := &User{Nickname: ""}
	assert.Equal(t, "匿名用户", u.DisplayName())
}

func TestUser_DisplayName_WhitespaceNickname(t *testing.T) {
	u := &User{Nickname: "  "}
	assert.Equal(t, "  ", u.DisplayName()) // 空格不算空
}
