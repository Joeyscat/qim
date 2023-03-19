package database

import "time"

// create database qim_base default character set utf8mb4 collate utf8mb4_unicode_ci;
// create database qim_message default character set utf8mb4 collate utf8mb4_unicode_ci;

type Model struct {
	ID        int64     `gorm:"primary_key;auto_increment;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MessageIndex struct {
	ID        int64  `gorm:"primary_key;auto_increment;not null"`
	AccountA  string `gorm:"size:60;not null;index"`
	AccountB  string `gorm:"size:60;not null;comment:'The other party account'"`
	Direction byte   `gorm:"not null;default:0;comment:'0:A is the sender,1:A is the receiver'"`
	MessageID int64  `gorm:"not null"`
	Group     string `gorm:"size:30;not null;comment:'Group ID'"`
	SendTime  int64  `gorm:"not null;index"`
}

type MessageContent struct {
	ID       int64  `gorm:"primary_key;auto_increment;not null"`
	Type     byte   `gorm:"not null;default:0"`
	Body     string `gorm:"size:5000;not null"`
	Extra    string `gorm:"size:500;not null"`
	SendTime int64  `gorm:"not null;index"`
}

type User struct {
	Model
	Account  string `gorm:"size:60;not null"`
	App      string `gorm:"size:30;not null"`
	Password string `gorm:"size:30;not null"`
	Avatar   string `gorm:"size:200;not null"`
	Nickname string `gorm:"size:20;not null"`
}

type Group struct {
	Model
	Group       string `gorm:"size:30;not null;unique_index"`
	App         string `gorm:"size:30;not null"`
	Name        string `gorm:"size:50;not null"`
	Owner       string `gorm:"size:60;not null"`
	Avatar      string `gorm:"size:200;not null"`
	Introdution string `gorm:"size:300;not null"`
}

type GroupMember struct {
	Model
	Account string `gorm:"size:60;not null;unique_index:uni_gp_acc"`
	Group   string `gorm:"size:30;not null;unique_index:uni_gp_acc;index"`
	Alias   string `gorm:"size:30;not null"`
}
