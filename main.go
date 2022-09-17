package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

const (
	ActivePatron = "active_patron"
	Campaign     = "campaign"
	Tier         = "tier"
)

var (
	version string
)

type campaign struct {
	id     string
	userID string
}

type member struct {
	id                           string
	email                        string
	fullName                     string
	tierID                       []string
	currentlyEntitledAmountCents int
}

type Data struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type V2Campaign struct {
	Data []struct {
		ID            string `json:"id"`
		Type          string `json:"type"`
		Relationships struct {
			Creator struct {
				Data Data `json:"data"`
			} `json:"creator"`
		} `json:"relationships"`
	} `json:"data"`
}

type V2Members struct {
	Data []struct {
		ID            string `json:"id"`
		Type          string `json:"type"`
		Relationships struct {
			CurrentlyEntitledTiers struct {
				Data []Data `json:"data"`
			} `json:"currently_entitled_tiers"`
		} `json:"relationships"`
		Attributes struct {
			CurrentlyEntitledAmountCents int    `json:"currently_entitled_amount_cents"`
			Email                        string `json:"email"`
			FullName                     string `json:"full_name"`
			PatronStatus                 string `json:"patron_status"`
		} `json:"attributes"`
	} `json:"data"`
	Included []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Title string `json:"title"` // a number of different include types are returned but we only care about tier types
		} `json:"attributes"`
	} `json:"included"`
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
}

func main() {
	accessToken := flag.String("access-token", "", "Patreon access-token")
	getPledges := flag.Bool("get-pledges", false, "Get pledge info")
	debug := flag.Bool("debug", false, "Add debugging info")
	showVersion := flag.Bool("version", false, "Show version and exit.")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}

	if *accessToken == "" {
		fmt.Println("Please provide your Patreon access-token")
		flag.Usage()
		os.Exit(0)
	}

	if *getPledges {
		client := &http.Client{}

		// Get campaign
		campaign := campaign{}
		{
			fmt.Fprintln(os.Stderr, "Fetching campaign details")
			req, err := http.NewRequest("GET", "https://www.patreon.com/api/oauth2/v2/campaigns?include=creator", nil)
			if err != nil {
				log.Fatalf("HTTP error %s", err.Error())
			}
			req.Header.Set("Authorization", "Bearer "+*accessToken)

			res, err := client.Do(req)
			if err != nil {
				log.Fatalf("Error making request for campaigns: %s\n", err)
			}

			if res.StatusCode != http.StatusOK {
				log.Fatalf("get %s returned HTTP %d (%s)\n", "my-info", res.StatusCode, http.StatusText(res.StatusCode))
			}

			if res.Body != nil {
				defer res.Body.Close()
			}

			body, readErr := ioutil.ReadAll(res.Body)
			if readErr != nil {
				log.Fatal(readErr)
			}

			campaigns := V2Campaign{}
			jsonErr := json.Unmarshal(body, &campaigns)
			if jsonErr != nil {
				log.Fatal(jsonErr)
			}

			if len(campaigns.Data) != 1 {
				fmt.Printf("The number of campaigns is %d, which I don't understand.  I was expecting exactly 1.  Call or email Mark.  Exiting.", len(campaigns.Data))
				os.Exit(0)
			}

			if campaigns.Data[0].Type != Campaign {
				fmt.Printf("The response is not a campaign: type==%s\n", campaigns.Data[0].Type)
				os.Exit(0)
			}

			campaign.id = campaigns.Data[0].ID
			campaign.userID = campaigns.Data[0].Relationships.Creator.Data.ID
		}

		// get members

		{
			fmt.Fprintln(os.Stderr, "Fetching members details")

			tiers := make(map[string]string, 0)
			members := make(map[string]member, 0)

			pages := 1
			nextPage := fmt.Sprintf("https://www.patreon.com/api/oauth2/v2/campaigns/%s/members?fields%%5Bmember%%5D=email%%2Cfull_name%%2Cpatron_status%%2Ccurrently_entitled_amount_cents&fields%%5Btier%%5D=title&fields%%5Buser%%5D=email&include=currently_entitled_tiers%%2Cpledge_history%%2Cuser", campaign.id)
			for nextPage != "" {
				r, err := http.NewRequest("GET", nextPage, nil)
				if err != nil {
					log.Fatalf("Cannot parse initial URL: %s\n", err)
				}
				r.Header.Set("Authorization", "Bearer "+*accessToken)

				res, err := client.Do(r)
				if err != nil {
					log.Fatalf("Error making request for my-info: %s\n", err)
				}

				if res.StatusCode != http.StatusOK {
					log.Fatalf("get %s returned HTTP %d (%s)\n", "get-members", res.StatusCode, http.StatusText(res.StatusCode))

				}
				if res.Body != nil {
					defer res.Body.Close()
				}

				body, readErr := ioutil.ReadAll(res.Body)
				if readErr != nil {
					log.Fatal(readErr)
				}

				// dump json for debugging
				if *debug {
					dumpFile := fmt.Sprintf("page-%d.json", pages)
					fmt.Fprintf(os.Stderr, "writing page of pledge raw JSON to %s\n", dumpFile)
					if os.WriteFile(dumpFile, body, 0644) != nil {
						fmt.Fprintf(os.Stderr, "error writing %s\n", dumpFile)
					}
				}

				patrons := V2Members{}
				if err := json.Unmarshal(body, &patrons); err != nil {
					log.Fatalf("Error unmarshaling V2Members: %s\n", err)
				}

				for _, s := range patrons.Included {
					if s.Type != Tier {
						continue
					}
					tiers[s.ID] = s.Attributes.Title
				}

				for _, m := range patrons.Data {
					if m.Attributes.PatronStatus != ActivePatron {
						continue
					}

					mem := member{}
					mem.id = m.ID
					mem.fullName = strings.Join(strings.Fields(m.Attributes.FullName), "")
					mem.email = m.Attributes.Email
					mem.currentlyEntitledAmountCents = m.Attributes.CurrentlyEntitledAmountCents
					for _, t := range m.Relationships.CurrentlyEntitledTiers.Data {
						mem.tierID = append(mem.tierID, t.ID)
					}

					if mem.fullName == "" {
						mem.fullName = "_none_"
					}
					members[m.ID] = mem
				}

				nextPage = patrons.Links.Next

				if pages%10 == 0 {
					fmt.Fprintf(os.Stderr, "Fetched %d total member pages\n", pages)
				}
				pages = pages + 1
			}

			for _, u := range members {
				// build entitled tiers
				var s []string
				for _, v := range u.tierID {
					s = append(s, strings.Join(strings.Fields(tiers[v]), ""))
				}
				sort.Strings(s)

				fmt.Printf("fullName=%s email=%s pledgeAmount:%d, tiers: %s\n", u.fullName, u.email, u.currentlyEntitledAmountCents, strings.Join(s, ","))
			}

		}

	}

}
