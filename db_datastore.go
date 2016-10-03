package fiandsri

import (
	// "appengine"
	// "appengine/datastore"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	_ "log"
	"net/http"
	"reflect"
	"time"
)

type DS struct {
	//ctx appengine.Context
	ctx context.Context
}

type DSErr struct {
	When time.Time
	What string
}

var (
	_        GuestDatabase = &DS{}
	_        error         = &DSErr{}
	entities               = map[string]string{"Guest": "guest"}
	total                  = int64(0)
)

func NewDS(r *http.Request) *DS {

	return &DS{ctx: appengine.NewContext(r)}

}

//Crude ad-hoc key gen
func (ds *DS) datastoreKey(id int64) *datastore.Key {

	c := ds.ctx
	return datastore.NewKey(c, "guest", "", id, nil)

}

func (ds *DS) datastoreKeyah(entity string, id ...int64) *datastore.Key {

	c := ds.ctx
	if len(id) != 0 {
		return datastore.NewKey(c, entity, "", id[0], nil)
	} else {
		return datastore.NewIncompleteKey(c, entity, nil)
	}
}

//Less crude interface key gen
func (ds *DS) dsKey(t reflect.Type, id ...interface{}) *datastore.Key {

	c := ds.ctx
	if entity, ok := entities[t.Name()]; ok {
		switch t.Name() {
		case "Counter":
			return datastore.NewKey(c, entity, "thekey", 0, nil)
		default:
			if len(id) > 0 {
				return datastore.NewKey(c, entity, "", id[0].(int64), nil)
			} else {
				return datastore.NewIncompleteKey(c, entity, nil) // shouldn't get here
			}
		}
	}
	return nil

}

func (ds *DS) dsChildKey(t reflect.Type, id int64, pk *datastore.Key) *datastore.Key {

	c := ds.ctx
	if entity, ok := entities[t.Name()]; ok {
		return datastore.NewKey(c, entity, "", id, pk)
	}
	return nil

}

/*func (ds *DS) Add(v interface{}) (int64, error) {

        c := ds.ctx
        k := ds.dsKey(reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem().Field(0).Interface())
        if k == nil {
                return 0, fmt.Errorf("Add guesty error - key create error")
        }
        _, err := datastore.Put(c, k, v)
        if err != nil {
                return 0, fmt.Errorf("Add guesty error - datastore put error")
        }
        return k.IntID(), nil

}*/

// Add creates an appropriate ds key for the entity passed via interface{} with the optional int64 used as Id of key
func (ds *DS) Add(v interface{}, n ...int64) (int64, error) {

	c := ds.ctx
	var k *datastore.Key
	if len(n) == 0 { // for Counter
		//k = ds.dsKey(reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem().Field(0).Interface())
		k = ds.dsKey(reflect.TypeOf(v).Elem())
	} else { // if extra args are provided use as key ID
		k = ds.dsKey(reflect.TypeOf(v).Elem(), n[0])
	}
	if k == nil {
		return 0, fmt.Errorf("Add error - key create error - unknown entity")
	}
	_, err := datastore.Put(c, k, v)
	if err != nil {
		return 0, fmt.Errorf("Add error - datastore put error: %v", err)
	}
	return k.IntID(), nil

}

// AddwParent creates keys and adds an interface along with it's parent. Parent must have Id field
func (ds *DS) AddwParent(parent interface{}, child interface{}, offset int64) error {

	if reflect.TypeOf(parent).Kind() != reflect.Ptr || reflect.TypeOf(child).Kind() != reflect.Ptr {
		return DSErr{When: time.Now(), What: "Get error: pointers reqd"}
	}

	c := ds.ctx

	pt := reflect.TypeOf(parent).Elem()
	if _, ok := pt.FieldByName("Id"); !ok {
		return DSErr{When: time.Now(), What: "Add w parent error: parent lacks Id"}
	}

	pv := reflect.ValueOf(parent).Elem().Field(0).Interface().(int64)
	pk := ds.dsKey(pt, pv)
	ck := ds.dsChildKey(reflect.TypeOf(child).Elem(), pv+offset, pk)

	if pk == nil || ck == nil {
		return DSErr{When: time.Now(), What: "Add w parent error: during key creation"}
	}
	pk, err := datastore.Put(c, pk, parent)
	if err != nil {
		return DSErr{When: time.Now(), What: "Add w parent error: during parent put" + err.Error()}
	}
	if pk != nil {
		_, err := datastore.Put(c, ck, child)
		if err != nil {
			return DSErr{When: time.Now(), What: "AddwParent error: during child put"}
		}
	}
	return nil

}

func (ds *DS) Get(v interface{}) error {

	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return DSErr{When: time.Now(), What: "Get error: pointer reqd"}
	}

	var id int64
	var k *datastore.Key
	// check whether Id field is available in struct - if it is, it shouldn't be 0
	if _, ok := reflect.TypeOf(v).Elem().FieldByName("Id"); ok {
		id = reflect.ValueOf(v).Elem().Field(0).Interface().(int64) // could also use FieldByName("Id") instead of Field(0)
		if id == 0 {                                                // shouldn't be zero
			return DSErr{When: time.Now(), What: "Get error: id not set"}
		}
		k = ds.dsKey(reflect.TypeOf(v).Elem(), id) //complete key
	} else {
		k = ds.dsKey(reflect.TypeOf(v).Elem())
	}

	c := ds.ctx
	if k == nil {
		return fmt.Errorf("Get error - key create error")
	}

	//if err := datastore.Get(c, k, reflect.ValueOf(v).Interface()); err != nil {
	if err := datastore.Get(c, k, v); err != nil {
		return fmt.Errorf("Get error - datastore get error: %v, key kind: %v", err, k.Kind())
	}
	return nil

}

//Guest specific
func (ds *DS) ListGuests() ([]*Guest, error) {

	c := ds.ctx
	cs := make([]*Guest, 0)
	q := datastore.NewQuery("guest").Order("Id")
	ks, err := q.GetAll(c, &cs)
	if err != nil {
		return nil, fmt.Errorf("Get guesty list error")
	}
	for i, k := range ks {
		cs[i].Id = k.IntID()
	}
	return cs, nil

}

//AddGuest adds guest which must already have Id (from CreateGuest) to be used as datastore key id. Doesn't touch counter. Returns either (id, nil) / (0, error)
func (ds *DS) AddGuest(guesty *Guest) (int64, error) {

	c := ds.ctx
	k := ds.datastoreKey(guesty.Id)
	_, err := datastore.Put(c, k, guesty)
	if err != nil {
		return 0, err
	}
	return k.IntID(), nil

}

func (ds *DS) GetGuest(id int64) (*Guest, error) {

	c := ds.ctx
	k := ds.datastoreKey(id)
	cst := &Guest{}
	err := datastore.Get(c, k, cst)
	if err != nil {
		return nil, fmt.Errorf("Get by id error")
	}
	return cst, nil

}

func (ds *DS) GetGuestwEmail(email string) (*Guest, error) {

	c := ds.ctx
	q := datastore.NewQuery("guest").Filter("Email=", email)
	cst := make([]*Guest, 0)
	spk, err := q.GetAll(c, &cst) // *[]*Guest
	if err != nil {
		return nil, fmt.Errorf("Get by email error %v", err)
	}
	if len(spk) > 0 {
		return cst[0], nil
	}
	return nil, nil

}

func (ds *DS) GetGuestKey(email string) (*Guest, *datastore.Key, error) {

	c := ds.ctx
	q := datastore.NewQuery("guest").Filter("Email=", email)
	cst := &Guest{}
	k, err := q.GetAll(c, cst)
	if err != nil {
		return nil, nil, fmt.Errorf("Get by email error %v", err)
	}
	return cst, k[0], nil

}

func (ds *DS) UpdateGuest(guesty *Guest) error {

	c := ds.ctx
	k := ds.datastoreKey(guesty.Id)
	_, err := datastore.Put(c, k, guesty)
	if err != nil {
		return fmt.Errorf("Updating error %v", err)
	}
	return nil

}

func (ds *DS) DeleteGuest(id int64) error {

	c := ds.ctx
	k := ds.datastoreKey(id)
	if err := datastore.Delete(c, k); err != nil {
		return fmt.Errorf("Deleting error")
	}
	return nil

}

//Creates Guest and updates counter. Returns nil if there was an error retrieving/updating counter or creating Guest
func (ds *DS) CreateGuest() *Guest {

	counter := ds.GetCounter()
	if counter != nil {
		counter.Rsvps++
		err := ds.PutCounter(counter)
		if err != nil {
			log.Errorf(ds.ctx, "Create guest Put counter: %v", err)
			return nil
		}
		return NewGuest(counter.Rsvps)
	} else {
		return nil
	}
	//total++
	//return NewGuest(total)

}

func (ds *DS) Close() error {

	return nil

}

func (dse DSErr) Error() string {

	return fmt.Sprintf("%v: %v", dse.When, dse.What)

}

//Location specific
func (ds *DS) AddLoc(loc *Loc) (int64, error) {

	c := ds.ctx
	k := ds.datastoreKeyah("loc")
	_, err := datastore.Put(c, k, loc)
	if err != nil {
		return 0, fmt.Errorf("Add loc error")
	}
	return k.IntID(), nil

}

func (ds *DS) ListLocs() ([]*Loc, error) {

	c := ds.ctx
	cs := make([]*Loc, 0)
	q := datastore.NewQuery("loc").Order("Ip")
	_, err := q.GetAll(c, &cs)
	if err != nil {
		return nil, fmt.Errorf("Get locations list error")
	}
	//for i, k := range ks {
	//	cs[i].Id = k.IntID()
	//}
	return cs, nil

}

//Counter specific
//Get Counter gets the singleton counter from datastore. Doesn't try to create. Returns the counter or nil
func (ds *DS) GetCounter() *Counter {

	c := ds.ctx
	k := ds.datastoreKeyah("counter", 1234567890)
	counter := &Counter{}
	err := datastore.Get(c, k, counter)
	if err != nil {
		log.Errorf(c, "Couldn't get counter: %v", err)
		return nil
	}
	return counter

}

//CreateCounter creates a counter and returns nil (success) or error
func (ds *DS) CreateCounter() error {

	c := ds.ctx
	counter := &Counter{Rsvps: int64(0), Visitors: int64(0), Confirms: int64(0)}
	k := ds.datastoreKeyah("counter", 1234567890)
	_, err := datastore.Put(c, k, counter)
	if err != nil {
		log.Errorf(c, "Couldn't create counter: %v", err)
		return err
	}
	return nil

}

func (ds *DS) PutCounter(counter *Counter) error {

	c := ds.ctx
	k := ds.datastoreKeyah("counter", 1234567890)
	_, err := datastore.Put(c, k, counter)
	if err != nil {
		log.Errorf(c, "Couldn't put counter: %v", err)
		return err
	}
	return nil

}
