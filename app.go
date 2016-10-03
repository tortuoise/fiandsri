package fiandsri

import (
	//local "appengine"
	//"appengine/mail"
	"bytes"
	"encoding/base64"
	_ "encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"golang.org/x/net/context"
	_ "google.golang.org/api/calendar/v3"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/mail"
	"html/template"
	"image/jpeg"
	_ "io/ioutil" //for dev_appserver testing
	_ "log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	//tmpl_cmn = template.Must(template.ParseGlob("templates/*"))
	tmpl_cmn     = template.Must(template.ParseFiles("templates/base", "templates/head", "templates/menu", "templates/body", "templates/footer"))
	tmpl_page    = template.Must(template.ParseFiles("templates/base", "templates/head", "templates/menu", "templates/page_body", "templates/footer"))
	tmpl_rsvp    = template.Must(template.ParseFiles("templates/base", "templates/head", "templates/menu", "templates/body", "templates/rsvp_footer"))
	tmpl_err     = template.Must(template.ParseFiles("templates/base", "templates/head", "templates/menu", "templates/err_body", "templates/err_footer"))
	tmpl_success = template.Must(template.ParseFiles("templates/base", "templates/head", "templates/menu", "templates/success_body", "templates/success_footer"))
	tmpl_list    = template.Must(template.ParseFiles("templates/base", "templates/head", "templates/menu", "templates/list_body", "templates/success_footer"))
	tmpl_locs    = template.Must(template.ParseFiles("templates/base", "templates/head", "templates/menu", "templates/locs_body", "templates/success_footer"))
	tmpl_maps    = template.Must(template.ParseFiles("templates/base", "templates/head", "templates/menu", "templates/maps_body", "templates/success_footer"))
	filenames    = []string{"f0.jpg", "f1.jpg", "f2.jpg"}
	months       = map[time.Month]string{9: "sep.jpg", 10: "oct.jpg", 11: "nov.jpg"}
	invimages    = map[string]string{"venue": "images/Fi_and_Sri_Wedding_Invite_2.png", "directions": "images/Fi_and_Sri_Wedding_Invite_4.png", "schedule": "images/Fi_and_Sri_Wedding_Invite_3.png", "background": "images/Fi_and_Sri_Wedding_Invite_1.png", "accommodation": "images/Fi_and_Sri_Wedding_Invite_1.png"}
	validEmail   = regexp.MustCompile("^.*@.*\\.(com|org|in|mail|io)$")
	validPath    = regexp.MustCompile(`^/(confirm|venue|rsvp|list|accommodation|schedule)?/?(.*)$`)
	bucket       = "fiandsri.appspot.com"
)

const confirmMessage = `Thanks for RSVPing. Please confirm your attendance by clicking on the link: %s`
const thanksMessage = `Thanks for RSVPing. Please check your email.`
const sorryMessage = `Shame you can't make it. If you change your mind, please confirm your attendance by clicking on the link: %s`
const thanksAgainMessage = `Thanks for confirming. Looking forward to seeing you!`

type Render struct { //for most purposes
	Message string   `json:"message"`
	Images  []string `json:"images"`
}

func handleNotFound(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	err := tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": "404 Nothing to see here " + r.URL.Path, "Filename": ""})
	if err != nil {
		log.Errorf(ctx, "Couldn't execute common NotFound template: %v\n", err)
	}
	return
}

func handleFound(ctx context.Context, page string, w http.ResponseWriter, r *http.Request) {
	img := invimages[page]
	data := Render{Message: "", Images: []string{img}}
	err := tmpl_page.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Errorf(ctx, "Couldn't excute page template: %v", err)
	}
	return
	return
}

func handleRoot(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	x := daystogo()
	images := make([]string, 0)
	for _, f := range filenames {
		buffer := new(bytes.Buffer)
		//b, err := ioutil.ReadFile(f) // for dev_appserver testing only
		img, err := ReadCloudImage(c, f) //ReadCloudImage (*image.Image, error)
		if err != nil {
			log.Errorf(c, "error reading from gcs %v \n", err)
			tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": f})
			return
		}
		//img, err := jpeg.Decode(bytes.NewReader(b)) //for dev_appserver testing only
		//if err != nil { //testing only
		//        log.Printf("error reading from gcs %v \n", err)
		//        tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message":err, "Filename":f})
		//        return
		//}//for dev_appserver testing only
		if err := jpeg.Encode(buffer, *img, nil); err != nil { //change *img to img for dev_appserver testing
			log.Errorf(c, "error reading image from gcs %v \n", err)
			tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": f})
			return
		}
		str := base64.StdEncoding.EncodeToString(buffer.Bytes())
		images = append(images, str) //buffer.Bytes())
	}
	data := Render{Message: strconv.Itoa(x) + " days to go", Images: images}
	err := tmpl_cmn.ExecuteTemplate(w, "base", data)
	//w.Header().Set("Content-Type", "image/jpeg")
	//w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	//_, err := w.Write(buffer.Bytes())
	if err != nil {
		log.Errorf(c, "Couldn't execute common template: %v\n", err)
	}
	return

}

func handleCalUpdate(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	//x := daystogo()
	//data := Render{strconv.Itoa(x)+" days to go",}
	err := CreateImages(c)
	if err != nil {
		log.Errorf(c, "error creating image: %v\n", err)
	}
	return

}

func handleMenu(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	vars := mux.Vars(r)
	switch vars["page"] {

	case "venue":
		handleFound(c, "venue", w, r)
		return
	case "schedule":
		handleFound(c, "schedule", w, r)
		return
	case "rsvp":
		handleRSVP(c, w, r)
		return
	case "accommodation":
		handleFound(c, "accommodation", w, r)
		return
	case "directions":
		handleFound(c, "directions", w, r)
		return
	case "list":
		handleList(c, w, r)
		return
	case "locs":
		handleLocList(c, w, r)
		return
	case "dirs":
		Map(w, r)
		return
	case "logs":
		Logs(w, r)
		return
	default:
		handleNotFound(c, w, r)
		return

	}

}

func handleRSVP(c context.Context, w http.ResponseWriter, r *http.Request) {

	images := make([]string, 0)
	for _, f := range filenames {
		buffer := new(bytes.Buffer)
		//b, err := ioutil.ReadFile(f) // for dev_appserver testing only
		img, err := ReadCloudImage(c, f) //ReadCloudImage (*image.Image, error)
		if err != nil {
			log.Errorf(c, "error reading from gcs %v \n", err)
			tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": f})
			return
		}
		//img, err := jpeg.Decode(bytes.NewReader(b)) //for dev_appserver testing only
		//if err != nil { //testing only
		//        log.Printf("error reading from gcs %v \n", err)
		//        tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message":err, "Filename":f})
		//        return
		//}//for dev_appserver testing only
		if err := jpeg.Encode(buffer, *img, nil); err != nil { //change *img to img for dev_appserver testing
			log.Errorf(c, "error reading image from gcs %v \n", err)
			tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": f})
			return
		}
		str := base64.StdEncoding.EncodeToString(buffer.Bytes())
		images = append(images, str) //buffer.Bytes())
	}
	data := Render{Images: images}
	err := tmpl_rsvp.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Errorf(c, "Couldn't execute common template: %v\n", err)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": ""})
	}
	return

}

func handleConfirm(w http.ResponseWriter, r *http.Request) {

	//c := local.NewContext(r)
	c := appengine.NewContext(r)
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		log.Errorf(c, "Bad email or path: %v\n", r.URL.Path)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": r.URL.Path, "Filename": ""})
		return
	}
	adsc := &DS{ctx: c}
	counter := adsc.GetCounter()
	if counter != nil {
		counter.Confirms++
		err := adsc.PutCounter(counter)
		if err != nil {
			log.Errorf(c, "handleConfirm Put counter: %v", err)
		}
	}
	guest, err := adsc.GetGuestwEmail(m[2])
	if err != nil {
		log.Errorf(c, "Couldn't find that email: %v\n", err)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": ""})
		return
	}
	if guest == nil {
		log.Errorf(c, "Couldn't find that email: %v\n", m[2])
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": "Couldn't find that email", "Filename": m[2]})
		return
	}
	guest.Confirmed = true
	err = adsc.UpdateGuest(guest)
	if err != nil {
		log.Errorf(c, "Couldn't confirm attendance: %v\n", err)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": ""})
		return
	}
	data := Render{Message: thanksAgainMessage}
	tmpl_success.ExecuteTemplate(w, "base", data)
	return

}

func handleList(c context.Context, w http.ResponseWriter, r *http.Request) {

	adsc := NewDS(r) // &DS{ctx: c}
	list, err := adsc.ListGuests()
	if err != nil {
		log.Errorf(c, "Couldn't list guest: %v\n", err)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": ""})
	}
	//w.Header().Set("Content-Type", "text/json")
	//enc := json.NewEncoder(w)
	//if err := enc.Encode(list); err != nil {
	//        tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message":err, "Filename":""})
	//        return
	//}
	tmpl_list.ExecuteTemplate(w, "base", map[string]interface{}{"Guests": list})
	return

}

func handleLocList(c context.Context, w http.ResponseWriter, r *http.Request) {

	adsc := NewDS(r) // &DS{ctx: c}
	list, err := adsc.ListLocs()
	if err != nil {
		log.Errorf(c, "Couldn't list guest: %v\n", err)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": ""})
	}
	//w.Header().Set("Content-Type", "text/json")
	//enc := json.NewEncoder(w)
	//if err := enc.Encode(list); err != nil {
	//        tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message":err, "Filename":""})
	//        return
	//}
	tmpl_locs.ExecuteTemplate(w, "base", map[string]interface{}{"Locations": list})
	return

}

func handleMenuP(w http.ResponseWriter, r *http.Request) {

	//c := local.NewContext(r)
	c := appengine.NewContext(r)
	addr := GetIp(r) //returns host part only
	_ = r.ParseForm()
	adsc := &DS{ctx: c}
	//Need to make sure counter is alive before creating/adding guests
	counter := adsc.GetCounter()
	if counter == nil {
		err := adsc.CreateCounter()
		if err != nil {
			log.Errorf(c, "handleMenu Create counter: %v", err)
			tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": "Couldn't create counter", "Filename": ""})
			return
		}
	}
	g1 := adsc.CreateGuest()
	decoder := schema.NewDecoder()
	err := decoder.Decode(g1, r.PostForm)
	if err != nil {
		log.Errorf(c, "Couldn't decode posted form: %v\n", err)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": ""})
		return
	}
	m := validEmail.FindStringSubmatch(g1.Email)
	if m == nil {
		log.Errorf(c, "Invalid email entered: %v\n", g1.Email)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": "Invalid email", "Filename": ""})
		return
	}
	guest, err := adsc.GetGuestwEmail(g1.Email)
	if guest != nil {
		log.Errorf(c, "Email already in use: %v\n", g1.Email)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": "Email already in use", "Filename": g1.Email})
		return
	}
	confirm(c, g1.Email, g1.Party)
	if _, err = adsc.AddGuest(g1); err != nil {
		log.Errorf(c, "Couldn't add guest: %v\n", err)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": ""})
		return
	}
	tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": thanksMessage, "Filename": ""})
	if addr != "" {
		loc, err := GetLoc(c, addr)
		if err != nil {
			log.Errorf(c, "Couldn't get location: %v\n", err)
		}
		_, err = adsc.AddLoc(loc)
		if err != nil {
			log.Errorf(c, "Couldn't add location: %v\n", err)
		}
	}
	return

}

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			handleNotFound(appengine.NewContext(r), w, r)
			return
		}
		fn(w, r) //, m[2])
	}
}

func init() {

	r := mux.NewRouter()
	r.HandleFunc("/", makeHandler(handleRoot)).Methods("GET")
	r.HandleFunc("/calupdate", makeHandler(handleCalUpdate)).Methods("GET")
	r.HandleFunc("/confirm/{email}", makeHandler(handleConfirm)).Methods("GET")
	r.HandleFunc("/{page}", makeHandler(handleMenu)).Methods("GET")
	r.HandleFunc("/{page}", makeHandler(handleMenuP)).Methods("POST")
	http.Handle("/", r)

}

func daystogo() int {

	now := time.Now() // today's date
	ven, err := time.LoadLocation("IST")
	if err != nil {
		ven, _ = time.LoadLocation("UTC")
	}
	dday := time.Date(2016, time.November, 02, 8, 30, 0, 0, ven)

	ds := 0
	for ds = 0; now.Before(dday); ds++ {
		now = now.Add(time.Hour * 24)
	}
	return ds

}

//func confirm(c local.Context, email string, party bool) {
func confirm(c context.Context, email string, party bool) {
	url := fmt.Sprintf("%s%s", "https://fiandsri.appspot.com/confirm/", email)
	var body string
	if party {
		body = fmt.Sprintf(confirmMessage, url)
	} else {
		body = fmt.Sprintf(sorryMessage, url)
	}
	msg := &mail.Message{
		Sender:  "Fi & Sri <fiandsri@fiandsri.appspotmail.com>",
		To:      []string{email},
		Subject: "Confirm your attendance",
		Body:    body,
	}
	if err := mail.Send(c, msg); err != nil {
		//(*c).Errorf("Couldn't send email: %v", err)
	}
}
