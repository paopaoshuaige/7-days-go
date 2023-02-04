package session

import (
	"testing"
)

type User struct {
	Name string `peeorm:"PRIMARY KEY"`
	Age  int
}

func TestSession_CreateTable(t *testing.T) {
	s := NewSession().Moedl(&User{})
	_ = s.DropTable()
	_ = s.CreateTable()
	if !s.HasTable() {
		t.Fatal("Failed to create table User")
	}
}

func TestSession_Model(t *testing.T) {
	s := NewSession().Moedl(&User{})
	table := s.RefTable()
	// Session默认名字是Session
	s.Moedl(&Session{})
	if table.Name != "User" || s.RefTable().Name != "Session" {
		t.Fatal("Failed to change model")
	}
}
