package fiandsri

import (
	"encoding/base64"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"googlemaps.github.io/maps"
	"html/template"
	"io/ioutil"
	"net/http"
)

var (
	tmpl_logs = template.Must(template.ParseFiles("templates/logs"))
)

const recordsPerPage = 10

func Logs(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)
	var data struct {
		Records []*log.Record
		Offset  string
	}

	query := &log.Query{AppLogs: true}

	if offset := r.FormValue("offset"); offset != "" {
		query.Offset, _ = base64.URLEncoding.DecodeString(offset)
	}

	res := query.Run(ctx)

	for i := 0; i < recordsPerPage; i++ {
		rec, err := res.Next()
		if err == log.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "Reading log records: %v", err)
			break
		}

		data.Records = append(data.Records, rec)
		if i == recordsPerPage-1 {
			data.Offset = base64.URLEncoding.EncodeToString(rec.Offset)
		}
	}

	if err := tmpl_logs.Execute(w, data); err != nil {
		log.Errorf(ctx, "Rendering template: %v", err)
	}

}

func Map(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)
	key, err := ioutil.ReadFile("mapsapi_sc.key")
	if err != nil {
		log.Errorf(ctx, "Maps API key read: %v", err)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": ""})
		return
	}
	client, err := maps.NewClient(maps.WithAPIKey(string(key)), maps.WithRateLimit(2))
	if err != nil {
		log.Errorf(ctx, "Maps new client: %v", err)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": ""})
		return
	}
	dr := &maps.DirectionsRequest{
		Origin:      "London SW83TP",
		Destination: "London E143AG",
	}
	dr.Mode = maps.TravelModeDriving
	dr.Units = maps.UnitsMetric
	//dr.TransitRoutingPreference = maps.TransitRoutingPreferenceFewerTransfers
	dr.TrafficModel = maps.TrafficModelOptimistic

	routes, _, err := client.Directions(ctx, dr)
	if err != nil {
		log.Errorf(ctx, "Maps client directions: %v", err)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": ""})
		return
	}
	var data struct {
		Routes []maps.Route
	}
	data.Routes = routes
	if err := tmpl_maps.Execute(w, data); err != nil {
		log.Errorf(ctx, "Rendering template: %v", err)
		tmpl_err.ExecuteTemplate(w, "base", map[string]interface{}{"Message": err, "Filename": ""})
		return
	}

}
