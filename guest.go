package fiandsri

import (
	//"appengine/datastore"
	"google.golang.org/appengine/datastore"
	"time"
)

type Counter struct {
	Rsvps    int64 `json:"rsvps"`
	Confirms int64 `json:"confirms"`
	Visitors int64 `json:"visitors"`
}

type Guest struct {
	Id        int64     `json:"id" schema:"-"`
	Name      string    `json:"name" schema:"fullname"`
	Email     string    `json:"email" schema:"email"`
	Addr      string    `json:"addr"`
	Note      string    `json:"note" schema:"special"`
	JDte      time.Time `json:"jdte"`
	Group     int       `json:"group" schema:"group"`
	Party     bool      `json:"conf" schema:"party"`
	Confirmed bool      `json:"conf"`
}

type GuestDatabase interface {
	ListGuests() ([]*Guest, error)

	AddGuest(guesty *Guest) (int64, error) //create

	GetGuest(id int64) (*Guest, error) //retrieve by id

	GetGuestwEmail(email string) (*Guest, error) //retrieve by email

	GetGuestKey(email string) (*Guest, *datastore.Key, error)

	UpdateGuest(guesty *Guest) error //update

	DeleteGuest(id int64) error //delete

	Close() error
}

func NewGuest(id int64) *Guest {

	return &Guest{Id: id, JDte: time.Now(), Confirmed: false}

}
