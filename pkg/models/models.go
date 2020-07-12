package models

import (
	"time"

	"github.com/lib/pq"
)

type Registration struct {
	Id        int `gorm:"primary_key"`
	AccountId []byte
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Hostname struct {
	Id int `gorm:"primary_key"`

	Registration   *Registration
	RegistrationId int

	Hostname string
	Labels   pq.StringArray

	CreatedAt time.Time
	UpdatedAt time.Time
}
