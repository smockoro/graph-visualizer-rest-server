package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"

	"gopkg.in/resty.v1"

	"github.com/ant0ine/go-json-rest/rest"
)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			return origin == "http://localhost:3001"
		},
		AllowedMethods: []string{"GET", "POST", "PUT"},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "X-Custom-Header", "Origin"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})
	router, err := rest.MakeRouter(
		rest.Get("/graph", GetGraphMis),
		rest.Get("/graph/belong", GetGraphBelong),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type Node struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Group int    `json:"group"`
}

type Edge struct {
	From int `json:"from"`
	To   int `json:"to"`
}

var lock = sync.RWMutex{}

func GetGraphMis(w rest.ResponseWriter, r *rest.Request) {
	lock.RLock()
	bytes, err := ioutil.ReadFile("./graph_mis.json")
	if err != nil {
		log.Fatal(err)
	}

	var graph Graph
	if err := json.Unmarshal(bytes, &graph); err != nil {
		log.Fatal(err)
	}
	lock.RUnlock()

	w.WriteJson(graph)
}

func GetGraphBelong(w rest.ResponseWriter, r *rest.Request) {
	lock.RLock()

	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		Get("http://lookup-service-prod.mlb.com/json/named.team_all_season.bam?sport_code=%27mlb%27&all_star_sw=%27N%27&sort_order=%27name_asc%27&season=%272017%27")
	if err != nil {
		log.Fatal(err)
	}

	var seasonTeam SeasonTeam
	if err := json.Unmarshal(resp.Body(), &seasonTeam); err != nil {
		log.Fatal(err)
	}

	nodeList := make([]Node, 0)
	edgeList := make([]Edge, 0)
	playerList := make([]Players, 0)
	playerMap := map[int]string{}
	for _, row := range seasonTeam.TeamAllSeason.QueryResults.Row {
		var players Players
		resp := mlbResponceCreator(
			"http://lookup-service-prod.mlb.com/json/named.roster_team_alltime.bam?start_season=%272017%27&end_season=%272017%27&team_id=%27" + row.TeamID + "%27")
		if err := json.Unmarshal(resp, &players); err != nil {
			log.Fatal(err)
		}
		playerList = append(playerList, players)

		// generate team node
		team_id, _ := strconv.Atoi(row.TeamID)
		leage_id, _ := strconv.Atoi(row.LeagueID)
		team_node := Node{
			ID:    team_id,
			Label: row.Name,
			Group: leage_id,
		}
		nodeList = append(nodeList, team_node)

		// generate team and player edges and player node
		for _, row := range players.RosterTeamAlltime.QueryResults.Row {
			player_id, _ := strconv.Atoi(row.PlayerID)
			if _, ok := playerMap[player_id]; !ok {
				playerMap[player_id] = "exist"
				player_node := Node{
					ID:    player_id,
					Label: row.NameLastFirst,
					Group: team_id,
				}
				nodeList = append(nodeList, player_node)
			}
			edge := Edge{
				From: team_id,
				To:   player_id,
			}
			edgeList = append(edgeList, edge)
		}

	}
	graph := Graph{
		Nodes: nodeList,
		Edges: edgeList,
	}

	lock.RUnlock()

	w.WriteJson(graph)
}

func mlbResponceCreator(url string) []byte {
	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		Get(url)
	if err != nil {
		log.Fatal(err)
	}
	return resp.Body()
}

type SeasonTeam struct {
	TeamAllSeason struct {
		CopyRight    string `json:"copyRight"`
		QueryResults struct {
			Created   string `json:"created"`
			TotalSize string `json:"totalSize"`
			Row       []struct {
				PhoneNumber        string `json:"phone_number"`
				VenueName          string `json:"venue_name"`
				FranchiseCode      string `json:"franchise_code"`
				AllStarSw          string `json:"all_star_sw"`
				SportCode          string `json:"sport_code"`
				AddressCity        string `json:"address_city"`
				City               string `json:"city"`
				NameDisplayFull    string `json:"name_display_full"`
				SpringLeagueAbbrev string `json:"spring_league_abbrev"`
				TimeZoneAlt        string `json:"time_zone_alt"`
				SportID            string `json:"sport_id"`
				VenueID            string `json:"venue_id"`
				MlbOrgID           string `json:"mlb_org_id"`
				TimeZoneGeneric    string `json:"time_zone_generic"`
				MlbOrg             string `json:"mlb_org"`
				LastYearOfPlay     string `json:"last_year_of_play"`
				LeagueFull         string `json:"league_full"`
				HomeOpenerTime     string `json:"home_opener_time"`
				AddressProvince    string `json:"address_province"`
				LeagueID           string `json:"league_id"`
				NameAbbrev         string `json:"name_abbrev"`
				BisTeamCode        string `json:"bis_team_code"`
				League             string `json:"league"`
				SpringLeague       string `json:"spring_league"`
				BaseURL            string `json:"base_url"`
				AddressZip         string `json:"address_zip"`
				SportCodeDisplay   string `json:"sport_code_display"`
				MlbOrgShort        string `json:"mlb_org_short"`
				TimeZone           string `json:"time_zone"`
				AddressLine1       string `json:"address_line1"`
				MlbOrgBrief        string `json:"mlb_org_brief"`
				AddressLine2       string `json:"address_line2"`
				Season             string `json:"season"`
				AddressLine3       string `json:"address_line3"`
				DivisionAbbrev     string `json:"division_abbrev"`
				NameDisplayShort   string `json:"name_display_short"`
				TeamID             string `json:"team_id"`
				ActiveSw           string `json:"active_sw"`
				AddressIntl        string `json:"address_intl"`
				State              string `json:"state"`
				AddressCountry     string `json:"address_country"`
				MlbOrgAbbrev       string `json:"mlb_org_abbrev"`
				Division           string `json:"division"`
				TeamCode           string `json:"team_code"`
				Name               string `json:"name"`
				WebsiteURL         string `json:"website_url"`
				SportCodeName      string `json:"sport_code_name"`
				FirstYearOfPlay    string `json:"first_year_of_play"`
				LeagueAbbrev       string `json:"league_abbrev"`
				NameDisplayLong    string `json:"name_display_long"`
				StoreURL           string `json:"store_url"`
				TimeZoneText       string `json:"time_zone_text"`
				NameShort          string `json:"name_short"`
				HomeOpener         string `json:"home_opener"`
				AddressState       string `json:"address_state"`
				DivisionFull       string `json:"division_full"`
				TimeZoneNum        string `json:"time_zone_num"`
				SpringLeagueFull   string `json:"spring_league_full"`
				Address            string `json:"address"`
				NameDisplayBrief   string `json:"name_display_brief"`
				FileCode           string `json:"file_code"`
				DivisionID         string `json:"division_id"`
				SpringLeagueID     string `json:"spring_league_id"`
				VenueShort         string `json:"venue_short"`
			} `json:"row"`
		} `json:"queryResults"`
	} `json:"team_all_season"`
}

type Players struct {
	RosterTeamAlltime struct {
		CopyRight    string `json:"copyRight"`
		QueryResults struct {
			Created   string `json:"created"`
			TotalSize string `json:"totalSize"`
			Row       []struct {
				NameFirstLast       string `json:"name_first_last"`
				Weight              string `json:"weight"`
				PrimaryPosition     string `json:"primary_position"`
				BirthDate           string `json:"birth_date"`
				Throws              string `json:"throws"`
				StatYears           string `json:"stat_years"`
				HeightInches        string `json:"height_inches"`
				NameSort            string `json:"name_sort"`
				StatusShort         string `json:"status_short"`
				JerseyNumber        string `json:"jersey_number"`
				PlayerFirstLastHTML string `json:"player_first_last_html"`
				Bats                string `json:"bats"`
				PositionDesig       string `json:"position_desig"`
				FortyManSw          string `json:"forty_man_sw"`
				PlayerHTML          string `json:"player_html"`
				HeightFeet          string `json:"height_feet"`
				PlayerID            string `json:"player_id"`
				NameLastFirst       string `json:"name_last_first"`
				CurrentSw           string `json:"current_sw"`
				TeamID              string `json:"team_id"`
				RosterYears         string `json:"roster_years"`
				ActiveSw            string `json:"active_sw"`
			} `json:"row"`
		} `json:"queryResults"`
	} `json:"roster_team_alltime"`
}
