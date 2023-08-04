package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type PseEntry struct {
	RegistrationStatus string `json:"registrationStatus"`
	SystemName         string `json:"systemName"`
	SystemURL          string `json:"systemUrl"`
	OperatorSector     string `json:"operatorSector"`
	OperatorName       string `json:"operatorName"`
	RegistrationDate   string `json:"registrationDate"`
	RegistrationID     string `json:"registrationID"`
}

const CSV_SYSTEM_NAME = 1
const CSV_SYSTEM_URL = 2
const CSV_OPERATOR_SECTOR = 3
const CSV_OPERATOR_NAME = 4
const CSV_REGISTRATION_DATE = 5
const CSV_REGISTRATION_ID = 6

var SiteLists = [6][2]string{
	{"domestik-terdaftar", "https://pse.kominfo.go.id/api/v1/retrieve-json-all?status=LOKAL_TERDAFTAR"},
	{"domestik-dihentikan-sementara", "https://pse.kominfo.go.id/api/v1/retrieve-json-all?status=LOKAL_DIHENTIKAN_SEMENTARA"},
	{"domestik-dicabut", "https://pse.kominfo.go.id/api/v1/retrieve-json-all?status=LOKAL_DICABUT"},
	{"asing-terdaftar", "https://pse.kominfo.go.id/api/v1/retrieve-json-all?status=ASING_TERDAFTAR"},
	{"asing-dihentikan-sementara", "https://pse.kominfo.go.id/api/v1/retrieve-json-all?status=ASING_DIHENTIKAN_SEMENTARA"},
	{"asing-dicabut", "https://pse.kominfo.go.id/api/v1/retrieve-json-all?status=ASING_DICABUT"},
}

func main() {
	stage2()
}

func stage1() {
	fmt.Println("(#_ ): Starting Scraper...")
	for _, list := range SiteLists {
		fmt.Printf("| Downloading %s...\n", list[0])

		client := &http.Client{}
		req, err := http.NewRequest("GET", list[1], nil)
		if err != nil {
			fmt.Printf("| (#_ ): Failed to initiate request: %v\n", err)
			continue
		}

		req.Header.Add("User-Agent", "psedb-bot/1.0 (+https://psedb.reinhart1010.id)")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("| (#_ ): Failed to download: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("| (#_ ): Failed to download: %v\n", err)
			continue
		}

		fmt.Println("| (#_ ): Saving raw data...")

		out, err := os.Create(fmt.Sprintf("raw/%s.json", list[0]))
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			fmt.Printf("| (#_ ): Failed to save file: %v\n", err)
			continue
		}

		fmt.Printf("| (#_ ): Saved file: raw/%s.json\n", list[0])
	}
}

func stage2() {
	fmt.Println("(>_ ): Starting Parser...")
	for _, list := range SiteLists {
		fmt.Printf("| Opening raw/%s.json...\n", list[0])
		in, err := os.OpenFile(fmt.Sprintf("raw/%s.json", list[0]), os.O_RDONLY, 0644)
		if err != nil {
			fmt.Printf("| (>_ ): Failed to open file: %v\n", err)
			continue
		}

		var data struct {
			Data [][7]any `json:"data"`
		}

		listRaw, err := ioutil.ReadAll(in)
		err = json.Unmarshal(listRaw, &data)
		if err != nil {
			fmt.Printf("| (>_ ): Failed to parse file: %v\n", err)
			continue
		}

		urlRegex := regexp.MustCompile(`((http|ftp|https):\/\/){0,1}([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
		strictUrlRegex := regexp.MustCompile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
		portRemovalRegex := regexp.MustCompile(`:\d+`)

		for _, site := range data.Data {
			// 1. Detect and iterate multiple URLs
			for _, siteUrl := range append(strictUrlRegex.FindAllString(site[CSV_SYSTEM_NAME].(string), -1), urlRegex.FindAllString(site[CSV_SYSTEM_URL].(string), -1)...) {
				// 2. Parse URL from list
				// Eliminate all invalid URLs like "-"
				// Handle cases like "http(s)://reinhart1010.id" and "reinhart1010.id" (without protocol string)
				// String replacemenent due to cases like "co,id" instead of "co.id"
				parsedUrl, err := url.Parse(fmt.Sprintf("http://%s", strings.ReplaceAll(siteUrl, ",", ".")))
				if err != nil {
					parsedUrl, err = url.Parse(siteUrl)
				}
				if err != nil {
					fmt.Printf("| (>_ ): Failed to parse url: %s (%s)\n", site[CSV_SYSTEM_NAME].(string), site[CSV_SYSTEM_URL].(string))
					continue
				}

				// 3. Remove unnecessary information
				urlDomain := strings.ToLower(strings.ReplaceAll(portRemovalRegex.ReplaceAllString(parsedUrl.Host, ""), ";", ""))
				if len(urlDomain) == 0 || urlDomain == "http:" || urlDomain == "https:" {
					continue
				}

				// 4. Split and reverse array
				// e.g.: google.com -> ["google", "com"] -> ["com", "google"] -> "com/google"
				urlParts := strings.Split(urlDomain, ".")
				for i := 0; i < len(urlParts)/2; i++ {
					j := len(urlParts) - i - 1
					urlParts[i], urlParts[j] = urlParts[j], urlParts[i]
				}
				reversedUrl := strings.Join(urlParts, "/")
				reversedUrlPath := strings.Join(urlParts[:len(urlParts)-1], "/")

				// 5. Open or create site info file
				err = os.MkdirAll(fmt.Sprintf("data/%s", reversedUrlPath), os.ModePerm)
				siteInfoFile, err := os.OpenFile(fmt.Sprintf("data/%s.json", reversedUrl), os.O_CREATE|os.O_RDWR, 0644)
				if err != nil {
					fmt.Printf("| (>_ ): Failed to open site info file: %s (%s): %v\n", site[CSV_SYSTEM_NAME].(string), site[CSV_SYSTEM_URL].(string), err)
					continue
				}

				// 6. Check data
				tdpseRegex := regexp.MustCompile(`https:\/\/pse.kominfo.go.id\/tdpse-detail\/(\d+?)($|\D)`)
				tdpseId := tdpseRegex.FindStringSubmatch(site[CSV_REGISTRATION_ID].(string))

				var siteInfoData map[string]PseEntry
				siteInfoRaw, err := ioutil.ReadAll(siteInfoFile)
				err = json.Unmarshal(siteInfoRaw, &siteInfoData)
				if err != nil {
					siteInfoData = make(map[string]PseEntry)
				}

				// 7. Assign data to PseEntry
				siteInfoData[tdpseId[1]] = PseEntry{
					RegistrationStatus: strings.ToUpper(list[0]),
					SystemName:         site[CSV_SYSTEM_NAME].(string),
					SystemURL:          site[CSV_SYSTEM_URL].(string),
					OperatorSector:     site[CSV_OPERATOR_SECTOR].(string),
					OperatorName:       site[CSV_OPERATOR_NAME].(string),
					RegistrationDate:   site[CSV_REGISTRATION_DATE].(string),
					RegistrationID:     site[CSV_REGISTRATION_ID].(string),
				}

				// 8. Update JSON
				final, err := json.Marshal(siteInfoData)
				if err != nil {
					fmt.Printf("| (>_ ): Failed to create JSON: %s (%s)\n", site[CSV_SYSTEM_NAME], site[CSV_SYSTEM_URL])
					continue
				}

				// 9. Rewrite JSON to file
				siteInfoFile.Truncate(0)
				siteInfoFile.Seek(0, 0)
				_, err = siteInfoFile.WriteString(string(final))
				if err != nil {
					fmt.Printf("| (>_ ): Failed to write JSON string: %s (%s)\n", site[CSV_SYSTEM_NAME], site[CSV_SYSTEM_URL])
					continue
				}

				// 10. Close JSON
				err = siteInfoFile.Close()
				if err != nil {
					fmt.Printf("| (>_ ): Failed to save file to data/%s.json: %s (%s)\n", reversedUrl, site[CSV_SYSTEM_NAME], site[CSV_SYSTEM_URL])
					continue
				}
			}
		}
	}
}
