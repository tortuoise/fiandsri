package fiandsri

import (
	//local "appengine"
	"encoding/json"
	"golang.org/x/net/context"
	_ "google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"net"
	"net/http"
)

type Loc struct {
	Ip           string  `json:"ip"`
	Country_code string  `json:"country_code"`
	Country_name string  `json:"country_name"`
	Region_code  string  `json:"region_code"`
	Region_name  string  `json:"region_name"`
	City         string  `json:"city"`
	Zip_code     string  `json:"zip_code"`
	Time_zone    string  `json:"time_zone"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Metro_code   int     `json:"region_name"`
}

func GetLoc(ctx context.Context, host string) (*Loc, error) {

	client := urlfetch.Client(ctx)
	resp, err := client.Get("https://freegeoip.net/json/" + host)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var loc *Loc
	dcdr := json.NewDecoder(resp.Body)
	err = dcdr.Decode(loc)
	if err != nil {
		return nil, err
	}
	return loc, nil
}

func GetIp(r *http.Request) string {

	if ipprxy := r.Header.Get("X-FORWARDED-FOR"); len(ipprxy) > 0 {
		return ipprxy
	} else {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return ""
		}
		return host
	}

}
