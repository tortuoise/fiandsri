<A name="toc0_1" title="Fiandsri"/>
# An rsvp app for our wedding using calendar & maps API on GAE

##Contents     
**<a href="toc1_1">Calendar</a>**  
**<a href="toc1_2">Maps</a>**  
**<a href="toc1_3">Plan</a>**  
**<a href="toc1_4">References</a>**  


<A name="toc1_1" title="Calendar" />
## Calendar ##
The calendar is a generated image with day colours corresponding to the phase of the moon with black being new(no) moon and white being full moon. 
<A name="toc1_2" title="Maps" />
## Maps ##
The maps lists the guests location who have rsvp'd. The location is found from the ip address of the connection when guest rsvp's.

<A name="toc1_3" title="Plan" />
## Plan ##
1. Get calendar api working on gae


Check if dday is in future. Do nothing if dday is in past. 
Check number of months

<A name="toc1_4" title="References" />
## References ##
[Writing images to templates](http://www.sanarias.com/blog/1214PlayingwithimagesinHTTPresponseingolang)
[Appengine datastore api](https://godoc.org/google.golang.org/appengine/datastore)
[GCP Appengine Console](https://console.cloud.google.com/appengine?project=fiandsri)
[Method: apps.repair](https://cloud.google.com/appengine/docs/admin-api/reference/rest/v1/apps/repair)


