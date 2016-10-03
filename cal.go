package fiandsri

import (
	"errors"
	"flag"
	"fmt"
	"google.golang.org/api/calendar/v3"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	dateFormat        = "2006-01-02"
	day               = time.Second * 60 * 60 * 24
	defaultCalSummary = "Streaks"
	defaultEvtSummary = "Streak"
)

var (
	defaultCacheFile = filepath.Join(os.Getenv("HOME"), ".streak-request-token")
	cacheFile        = flag.String("cachefile", defaultCacheFile, "Authentication token cache file")
	offset           = flag.Int("offset", 0, "Day offset")
	remove           = flag.Bool("remove", false, "Remove day from streak")
	calendarName     = flag.String("cal", defaultCalSummary, "Streak calendar name")
	eventName        = flag.String("event", defaultEvtSummary, "Streak event name")
	createCalendar   = flag.Bool("create", false, "Create calendar if missing")
)

func GetMoonPhases() {

	tok, err := readToken(*cacheFile)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("reading token cache: %v", err)
	}
	client, tok := oauthClientToken(tok)
	if err := writeToken(*cacheFile, tok); err != nil {
		log.Fatalf("writing token cache: %v", err)
	}

	service, err := calendar.New(client)
	if err != nil {
		log.Fatal(err)
	}

	calId, err := streakCalendarId(service)
	if err != nil {
		log.Fatal(err)
	}

	cal := &Calendar{
		Id:      calId,
		Service: service,
	}

	today := time.Now().Add(time.Duration(*offset) * day)
	today = parseDate(today.Format(dateFormat)) // normalize

	log.Println("Remove: ", *remove)
	/*if *remove {
		err = cal.removeFromStreak(today)
	} else {
		err = cal.addToStreak(today)
	}
	if err != nil {
		log.Fatal(err)
	}*/

	var longest time.Duration
	cal.iterateEvents(func(e *calendar.Event, start, end time.Time) error {
		if d := end.Sub(start); d > longest {
			longest = d
		}
		return Continue
	})
	log.Println("Longest streak:", int(longest/day), "days")

}

type Calendar struct {
	Id                string
	*calendar.Service //embedded field - files and methods of calendar.Service promoted to Calendar struct
	//*CalImage
}

func (c *Calendar) addToStreak(today time.Time) (err error) {
	var (
		create = true
		prev   *calendar.Event
	)
	err = c.iterateEvents(func(e *calendar.Event, start, end time.Time) error {
		if prev != nil {
			log.Println("Prev != nil")
			// We extended the previous event; merge it with this one?
			if prev.End.Date == e.Start.Date {
				// Merge events.
				// Extend this event to begin where the previous one did.
				e.Start = prev.Start
				_, err := c.Events.Update(c.Id, e.Id, e).Do()
				if err != nil {
					return err
				}
				// Delete the previous event.
				return c.Events.Delete(c.Id, prev.Id).Do()
			}
			// We needn't look at any more events.
			return nil
		}
		if start.After(today) {
			if start.Add(-day).Equal(today) {
				// This event starts tomorrow, update it to start today.
				log.Println("Start Tomorrow")
				create = false
				e.Start.Date = today.Format(dateFormat)
				_, err = c.Events.Update(c.Id, e.Id, e).Do()
				return err
			}
			// This event is too far in the future.
			return Continue
		}
		if end.After(today) {
			log.Println("End after today")
			// Today fits inside this event, nothing to do.
			create = false
			return nil
		}
		if end.Equal(today) {
			log.Println("End today")
			// This event ends today, update it to end tomorrow.
			create = false
			e.End.Date = today.Add(day).Format(dateFormat)
			_, err := c.Events.Update(c.Id, e.Id, e).Do()
			if err != nil {
				return err
			}
			prev = e
			// Continue to the next event to see if merge is necessary.
		}
		return Continue
	})
	if err == nil && create {
		// No existing events cover or are adjacent to today, so create one.
		err = c.createEvent(today, today.Add(day))
	}
	return
}

func (c *Calendar) removeFromStreak(today time.Time) (err error) {
	err = c.iterateEvents(func(e *calendar.Event, start, end time.Time) error {
		log.Println("Remove Start")
		if start.After(today) || end.Before(today) || end.Equal(today) {
			// This event is too far in the future or past.
			log.Println("Too far past or future")
			return Continue
		}
		if start.Equal(today) {
			log.Println("Starts today")
			if end.Equal(today.Add(day)) {
				// Single day event; remove it.
				return c.Events.Delete(c.Id, e.Id).Do()
			}
			// Starts today; shorten to begin tomorrow.
			e.Start.Date = start.Add(day).Format(dateFormat)
			_, err := c.Events.Update(c.Id, e.Id, e).Do()
			return err
		}
		if end.Equal(today.Add(day)) {
			log.Println("Ends today")
			// Ends tomorrow; shorten to end today.
			e.End.Date = today.Format(dateFormat)
			_, err := c.Events.Update(c.Id, e.Id, e).Do()
			return err
		}
		log.Println("Remove End")

		// Split into two events.
		// Shorten first event to end today.
		e.End.Date = today.Format(dateFormat)
		_, err = c.Events.Update(c.Id, e.Id, e).Do()
		if err != nil {
			return err
		}
		// Create second event that starts tomorrow.
		return c.createEvent(today.Add(day), end)
	})
	return
}

var Continue = errors.New("continue")

type iteratorFunc func(e *calendar.Event, start, end time.Time) error

func (c *Calendar) iterateEvents(fn iteratorFunc) error {
	var pageToken string
	for {
		call := c.Events.List(c.Id).SingleEvents(true).OrderBy("startTime")
		if pageToken != "" {
			call.PageToken(pageToken)
		}
		events, err := call.Do()
		if err != nil {
			return err
		}
		for _, e := range events.Items {
			if e.Start.Date == "" || e.End.Date == "" || e.Summary != *eventName {
				// Skip non-all-day event or non-streak events.
				log.Println("Not all day: ", e.Summary, e.Start.Date, e.End.Date)
				continue
			}
			start, end := parseDate(e.Start.Date), parseDate(e.End.Date)
			log.Println("All day: ", e.Summary, e.Start.Date, e.End.Date)
			if err := fn(e, start, end); err != Continue {
				return err
			}
		}
		pageToken = events.NextPageToken
		if pageToken == "" {
			return nil
		}
	}
}

func (c *Calendar) createEvent(start, end time.Time) error {
	e := &calendar.Event{
		Summary:   *eventName,
		Start:     &calendar.EventDateTime{Date: start.Format(dateFormat)},
		End:       &calendar.EventDateTime{Date: end.Format(dateFormat)},
		Reminders: &calendar.EventReminders{UseDefault: false},
	}
	_, err := c.Events.Insert(c.Id, e).Do()
	return err
}

func parseDate(s string) time.Time {
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		panic(err)
	}
	return t
}

func streakCalendarId(service *calendar.Service) (string, error) {
	list, err := service.CalendarList.List().Do()
	if err != nil {
		return "", err
	}
	for _, entry := range list.Items {
		if entry.Summary == *calendarName {
			return entry.Id, nil
		}
	}

	if *createCalendar {
		cal, err := service.Calendars.Insert(&calendar.Calendar{Summary: *calendarName}).Do()
		if err != nil {
			return "", err
		}

		return cal.Id, nil
	}

	return "", errors.New(fmt.Sprintf("couldn't find calendar named '%s'", *calendarName))
}
