package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
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

const (
	PSEDBSystemNameCSVIndex             = 1
	PSEDBSystemURLCSVIndex              = 2
	PSEDBSystemOperatorSectorCSVIndex   = 3
	PSEDBSystemOperatorNameCSVIndex     = 4
	PSEDBSystemRegistrationDateCSVIndex = 5
	PSEDBSystemRegistrationIDCSVIndex   = 6
)

const (
	PSEDBOK                   = 0b00000000
	PSEDBCritical             = 0b00000001
	PSEDBInputError           = 0b00000010
	PSEDBOutputError          = 0b00000100
	PSEDBHTTPPreRequestError  = 0b00001000
	PSEDBHTTPPostRequestError = 0b00010000
	PSEDBParserError          = 0b00100000
)

var (
	PortRemovalRegex = regexp.MustCompile(`:\d+`)
	StrictUrlRegex   = regexp.MustCompile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
	UrlRegex         = regexp.MustCompile(`^((http|ftp|https):\/\/){0,1}([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?$`)
)

var SiteLists = [6][2]string{
	{"domestik-terdaftar", "https://pse.komdigi.go.id/api-public/tdpse?index=LOKAL_TERDAFTAR&page=@&hit_per_page=50"},
	{"domestik-dihentikan-sementara", "https://pse.komdigi.go.id/api-public/tdpse?index=LOKAL_DIHENTIKAN_SEMENTARA&page=@&hit_per_page=50"},
	{"domestik-dicabut", "https://pse.komdigi.go.id/api-public/tdpse?index=LOKAL_DICABUT&page=@&hit_per_page=50"},
	{"asing-terdaftar", "https://pse.komdigi.go.id/api-public/tdpse?index=ASING_TERDAFTAR&page=@&hit_per_page=50"},
	{"asing-dihentikan-sementara", "https://pse.komdigi.go.id/api-public/tdpse?index=ASING_DIHENTIKAN_SEMENTARA&page=@&hit_per_page=50"},
	{"asing-dicabut", "https://pse.komdigi.go.id/api-public/tdpse?index=ASING_DICABUT&page=@&hit_per_page=50"},
}

func main() {
	if (stage1() & PSEDBCritical) != PSEDBCritical {
		stage2()
	}
}

func stage1() int {
	fmt.Println("(#_ ): Starting Scraper...")

	var wg sync.WaitGroup
	for _, list := range SiteLists {
		page := 1
		outOf := 20 // Temporarily fetch the first 20 resources

		for page <= outOf {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fmt.Printf("| Downloading %s (page %4d/%4d)...\n", list[0], page, outOf)
			
				// Disable TLS verification
				http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			
				client := &http.Client{}
				req, err := http.NewRequest("GET", strings.Replace(list[1], "@", strconv.itoa(page)), nil)
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
			
				out, err := os.Create(fmt.Sprintf("raw/%s.%d.json", list[0], page))
				defer out.Close()
			
				_, err = io.Copy(out, resp.Body)
				if err != nil {
					fmt.Printf("| (#_ ): Failed to save file: %v\n", err)
					continue
				}
			
				fmt.Printf("| (#_ ): Saved file: raw/%s.%d.json\n", list[0], page)
				
				// Try to dynamically update the total pages
				var data struct {
					TotalPages int `json:"totalPages"`
				}

				err = json.Unmarshal(resp.Body, &data)
				if err != nil && data.TotalPages != outOf {
					outOf = data.TotalPages
				}
			}
		}
	}
	return PSEDBOK
}

func parseSiteUrl(siteUrl string, site [7]any, listType string) int {
	// 2. Parse URL from list
	// Eliminate all invalid URLs like "-"
	// Handle cases like "http(s)://reinhart1010.id" and "reinhart1010.id" (without protocol string)
	// String replacemenent due to cases like "co,id" instead of "co.id"
	filtered := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(siteUrl, ",", "."), "(", ""), ")", ""), ";", "")
	parsedUrl, err := url.Parse(fmt.Sprintf("http://%s", filtered))
	if err != nil {
		parsedUrl, err = url.Parse(siteUrl)
	}
	if err != nil {
		fmt.Printf("| (>_ ): Failed to parse url: %s (%s)\n", site[PSEDBSystemNameCSVIndex].(string), site[PSEDBSystemURLCSVIndex].(string))
		return PSEDBParserError
	}

	// 3. Remove unnecessary information
	urlDomain := strings.ToLower(strings.ReplaceAll(PortRemovalRegex.ReplaceAllString(parsedUrl.Host, ""), ";", ""))
	if len(urlDomain) == 0 || urlDomain == "http:" || urlDomain == "https:" {
		return PSEDBInputError
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
		fmt.Printf("| (>_ ): Failed to open site info file: %s (%s): %v\n", site[PSEDBSystemNameCSVIndex].(string), site[PSEDBSystemURLCSVIndex].(string), err)
		return PSEDBInputError
	}

	// 6. Check data
	tdpseRegex := regexp.MustCompile(`https:\/\/pse.komdigi.go.id\/tdpse-detail\/(\d+?)($|\D)`)
	tdpseId := tdpseRegex.FindStringSubmatch(site[PSEDBSystemRegistrationIDCSVIndex].(string))

	var siteInfoData map[string]PseEntry
	siteInfoRaw, err := ioutil.ReadAll(siteInfoFile)
	err = json.Unmarshal(siteInfoRaw, &siteInfoData)
	if err != nil {
		siteInfoData = make(map[string]PseEntry)
	}

	// 7. Assign data to PseEntry
	siteInfoData[tdpseId[1]] = PseEntry{
		RegistrationStatus: strings.ToUpper(listType),
		SystemName:         site[PSEDBSystemNameCSVIndex].(string),
		SystemURL:          site[PSEDBSystemURLCSVIndex].(string),
		OperatorSector:     site[PSEDBSystemOperatorSectorCSVIndex].(string),
		OperatorName:       site[PSEDBSystemOperatorNameCSVIndex].(string),
		RegistrationDate:   site[PSEDBSystemRegistrationDateCSVIndex].(string),
		RegistrationID:     site[PSEDBSystemRegistrationIDCSVIndex].(string),
	}

	// 8. Update JSON
	final, err := json.Marshal(siteInfoData)
	if err != nil {
		fmt.Printf("| (>_ ): Failed to create JSON: %s (%s)\n", site[PSEDBSystemNameCSVIndex], site[PSEDBSystemURLCSVIndex])
		return PSEDBOutputError
	}

	// 9. Rewrite JSON to file
	siteInfoFile.Truncate(0)
	siteInfoFile.Seek(0, 0)
	_, err = siteInfoFile.WriteString(string(final))
	if err != nil {
		fmt.Printf("| (>_ ): Failed to write JSON string: %s (%s)\n", site[PSEDBSystemNameCSVIndex], site[PSEDBSystemURLCSVIndex])
		return PSEDBOutputError
	}

	// 10. Close JSON
	err = siteInfoFile.Close()
	if err != nil {
		fmt.Printf("| (>_ ): Failed to save file to data/%s.json: %s (%s)\n", reversedUrl, site[PSEDBSystemNameCSVIndex], site[PSEDBSystemURLCSVIndex])
		return PSEDBOutputError
	}
	return PSEDBOK
}

func stage2() int {
	fmt.Println("(>_ ): Starting PSEDB...")
	for _, list := range SiteLists {
		fmt.Printf("| Opening raw/%s.json...\n", list[0])
		in, err := os.OpenFile(fmt.Sprintf("raw/%s.json", list[0]), os.O_RDONLY, 0644)
		if err != nil {
			fmt.Printf("| (>_ ): Failed to open file: %v\n", err)
			continue
		}

		var data struct {
			Data [][7]any `json:"hits"`
		}

		listRaw, err := ioutil.ReadAll(in)
		err = json.Unmarshal(listRaw, &data)
		if err != nil {
			fmt.Printf("| (>_ ): Failed to parse file: %v\n", err)
			continue
		}

		for _, site := range data.Data {
			// 1. Detect and iterate multiple URLs
			for _, siteUrl := range strings.Split(site[PSEDBSystemNameCSVIndex].(string), " ") {
				if !StrictUrlRegex.MatchString(siteUrl) {
					continue
				}
				parseSiteUrl(siteUrl, site, list[0])
			}
			for _, siteUrl := range strings.Split(site[PSEDBSystemURLCSVIndex].(string), " ") {
				if !UrlRegex.MatchString(siteUrl) {
					continue
				}
				parseSiteUrl(siteUrl, site, list[0])
			}
			for _, siteUrl := range append(StrictUrlRegex.FindAllString(site[PSEDBSystemNameCSVIndex].(string), -1), UrlRegex.FindAllString(site[PSEDBSystemURLCSVIndex].(string), -1)...) {
				parseSiteUrl(siteUrl, site, list[0])
			}
		}
	}
	return PSEDBOK
}
