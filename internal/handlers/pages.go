package handlers

import (
	"html/template"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/context"

	"github.com/APIENG/apieng/internal/db"
	"github.com/APIENG/apieng/internal/models"
)

// EndpointStat holds aggregated metrics for a single unique endpoint.
type EndpointStat struct {
	Endpoint      string
	Count         int
	AvgResponseMs float64
	AvgEnergy     float64
	LastStatus    int
	LastSeen      time.Time
}

// EndpointsHandler displays the distinct endpoints a user has measured,
// aggregated with summary stats.
func EndpointsHandler(w http.ResponseWriter, r *http.Request) {
	dr, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer dr.Close()

	token := context.Get(r, "user")
	strToken, _ := token.(string)

	metrics, err := FetchMetrics(dr, strToken)
	if err != nil {
		http.Error(w, "Unable to fetch metrics", http.StatusInternalServerError)
		return
	}

	// Aggregate by endpoint
	grouped := make(map[string]*EndpointStat)
	for _, m := range metrics {
		stat, ok := grouped[m.APIEndpoint]
		if !ok {
			stat = &EndpointStat{Endpoint: m.APIEndpoint}
			grouped[m.APIEndpoint] = stat
		}
		stat.Count++
		stat.AvgResponseMs += float64(m.ResponseTime.Milliseconds())
		stat.AvgEnergy += m.EnergyConsumption
		if m.Timestamp.After(stat.LastSeen) {
			stat.LastSeen = m.Timestamp
			stat.LastStatus = m.Status
		}
	}

	stats := make([]EndpointStat, 0, len(grouped))
	for _, s := range grouped {
		if s.Count > 0 {
			s.AvgResponseMs = s.AvgResponseMs / float64(s.Count)
			s.AvgEnergy = s.AvgEnergy / float64(s.Count)
		}
		stats = append(stats, *s)
	}
	// Most recently seen first
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].LastSeen.After(stats[j].LastSeen)
	})

	tmpl, err := template.ParseFiles("templates/endpoints.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, struct{ Endpoints []EndpointStat }{Endpoints: stats}); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// KeysHandler displays the user's API key and lets them (re)generate one.
func KeysHandler(w http.ResponseWriter, r *http.Request) {
	dr, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer dr.Close()

	token := context.Get(r, "user")
	strToken, _ := token.(string)

	user, err := FetchEachUser(dr, strToken)
	if err != nil {
		// Not fatal — user may simply have no key yet.
		user = models.Users{}
	}

	apiKey := ""
	if user.Apikey.Valid {
		apiKey = user.Apikey.String
	}

	tmpl, err := template.ParseFiles("templates/keys.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	data := struct {
		APIKey string
		HasKey bool
	}{APIKey: apiKey, HasKey: apiKey != ""}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// SettingsHandler displays the current user's account information.
func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	dr, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer dr.Close()

	token := context.Get(r, "user")
	strToken, _ := token.(string)

	user, err := FetchEachUser(dr, strToken)
	if err != nil {
		user = models.Users{Iid: strToken}
	}

	hasKey := user.Apikey.Valid && user.Apikey.String != ""

	tmpl, err := template.ParseFiles("templates/settings.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	data := struct {
		User   models.Users
		HasKey bool
	}{User: user, HasKey: hasKey}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// DocsHandler renders the public documentation page.
func DocsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/docs.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// ContactHandler renders the public contact page.
func ContactHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/contact.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
