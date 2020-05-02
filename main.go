package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/namsral/flag"
	diff "github.com/olegfedoseev/image-diff"
	"github.com/tebeka/selenium"
)

var (
	captureWait  = flag.Int("capture-wait", 1000, "number of milliseconds to wait after a page is loaded to take the screenshot")
	alphaBaseURL = flag.String("alpha-base-url", "", "base url for alpha site")
	betaBaseURL  = flag.String("beta-base-url", "", "base url for beta site")
	testPath     = flag.String("test-path", "/", "path to test on both the sites")
	seleniumURL  = flag.String("webdriver-url", "", "full url (including port) to webdriver")
	outputDir    = flag.String("output-dir", "output", "directory to output screenshots, metadata, and web UI")
	runTitle     = flag.String("run-title", "", "name to use for the current test run")
	links        = make(map[string]Link)
	result       = Result{}
)

// Link describes the output of a particular page on both sites.
type Link struct {
	Path                string  `json:"path"`
	Captured            bool    `json:"captured"`
	AlphaScreenshotFile string  `json:"alphaScreenshotFile"`
	BetaScreenshotFile  string  `json:"betaScreenshotFile"`
	Similarity          float64 `json:"similarity"`
	DiffFile            string  `json:"diffFile"`
	AlphaBaseURL        string  `json:"alphaBaseURL"`
	BetaBaseURL         string  `json:"betaBaseURL"`
}

// Result holds the result of the visual regression test.
type Result struct {
	Links                     map[string]Link `json:"links"`
	OverallSimilarity         float64         `json:"overallSimilarity"`
	CompareStartTime          time.Time       `json:"compareStartTime"`
	CompareEndTime            time.Time       `json:"compareEndTime"`
	CompareStartTimeFormatted string          `json:"compareStartTimeFormatted"`
	CompareEndTimeFormatted   string          `json:"compareEndTimeFormatted"`
	TestPath                  string          `json:"testPath"`
	AlphaBaseURL              string          `json:"alphaBaseURL"`
	BetaBaseURL               string          `json:"betaBaseURL"`
	CaptureWait               int             `json:"captureWait"`
	Title                     string          `json:"title"`
}

func main() {
	// Parse the command line flags.
	flag.Parse()

	// Set up some of the result metadata.
	result.CompareStartTime = time.Now()
	result.TestPath = *testPath
	result.AlphaBaseURL = *alphaBaseURL
	result.BetaBaseURL = *betaBaseURL
	result.Title = *runTitle
	result.CaptureWait = *captureWait

	// Connect to the WebDriver instance.
	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, *seleniumURL)
	if err != nil {
		log.Println("Could not create new webdriver session.")
		panic(err)
	}
	defer wd.Quit()

	// Make sure the data dir and needed children exist.
	if _, err := os.Stat(*outputDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(*outputDir, os.ModePerm)
		} else {
			panic(err)
		}
	}
	if _, err := os.Stat(filepath.Join(*outputDir, "screenshots")); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(filepath.Join(*outputDir, "screenshots"), os.ModePerm)
		} else {
			panic(err)
		}
	}
	if _, err := os.Stat(filepath.Join(*outputDir, "diffs")); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(filepath.Join(*outputDir, "diffs"), os.ModePerm)
		} else {
			panic(err)
		}
	}

	// Capture the first link once and populate the list of links.
	log.Println("Scanning for links on alpha site")

	// Add the base page to the link list.
	links[*testPath] = Link{
		Path:                *testPath,
		Captured:            false,
		AlphaScreenshotFile: "",
		BetaScreenshotFile:  "",
		Similarity:          0,
		DiffFile:            "",
		AlphaBaseURL:        *alphaBaseURL,
		BetaBaseURL:         *betaBaseURL,
	}

	// Load the page
	if err := wd.Get(strings.Join([]string{*alphaBaseURL, *testPath}, "")); err != nil {
		log.Printf("Error loading %s%s: %s\n", *alphaBaseURL, *testPath, err.Error())
	}
	elements, err := wd.FindElements(selenium.ByCSSSelector, "a")
	if err != nil {
		panic(err)
	}

	for _, e := range elements {
		href, err := e.GetAttribute("href")
		if err != nil {
			panic(err)
		}

		if strings.Contains(href, *alphaBaseURL) && !strings.Contains(href, "#") {
			// Just get the path.
			newPath := strings.ReplaceAll(href, *alphaBaseURL, "")

			// Don't add the same link twice.
			if _, ok := links[newPath]; ok {
				continue
			}

			// If we got here, then we'll need to capture the page later.
			log.Printf("Adding new link for capture: %s\n", href)
			links[newPath] = Link{
				Path:                newPath,
				Captured:            false,
				AlphaScreenshotFile: "",
				BetaScreenshotFile:  "",
				Similarity:          0,
				DiffFile:            "",
				AlphaBaseURL:        *alphaBaseURL,
				BetaBaseURL:         *betaBaseURL,
			}
		}
	}

	// Now that we have a list of paths off of the alpha site, we'll loop through
	// everything, make the captures, and do the comparison.
	for path, link := range links {
		log.Printf("Capturing %s from both sites and comparing\n", path)
		newlink, err := captureAndCompare(path, link, wd)
		if err != nil {
			log.Printf("Error capturing and comparing %s: %s", path, err.Error())
		} else {
			links[path] = newlink
		}
	}

	// Add the full list of links to the result object.
	result.Links = links

	// Set the result end time.
	result.CompareEndTime = time.Now()

	// Calculate the overall similarity (average of all similarities).
	average := 0.0
	for _, link := range links {
		average = average + link.Similarity
	}
	average = math.Round((average/float64(len(links)))*100) / 100
	result.OverallSimilarity = average

	// Clean up file paths so that they're useful to the HTML and JSON consumers.
	for path, link := range links {
		// Strip the output dir and a path seperator from the beginning of all paths.
		pathToRemove := *outputDir + string(filepath.Separator)
		link.AlphaScreenshotFile = strings.ReplaceAll(link.AlphaScreenshotFile, pathToRemove, "")
		link.BetaScreenshotFile = strings.ReplaceAll(link.BetaScreenshotFile, pathToRemove, "")
		link.DiffFile = strings.ReplaceAll(link.DiffFile, pathToRemove, "")

		// Round similarity to nearest hundreth.
		link.Similarity = math.Round((link.Similarity)*100) / 100

		links[path] = link
	}

	// Format start and end times
	result.CompareStartTimeFormatted = result.CompareStartTime.Format(time.RFC850)
	result.CompareEndTimeFormatted = result.CompareEndTime.Format(time.RFC850)
	if result.Title == "" {
		result.Title = result.CompareStartTimeFormatted
	}

	// Write results.json
	resultdata, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Printf("Could not convert result data to JSON: %s\n", err.Error())
	} else {
		err = ioutil.WriteFile(filepath.Join(*outputDir, "results.json"), resultdata, 0644)
		if err != nil {
			log.Printf("Could not write results.json: %s\n", err.Error())
		}
	}

	// Write results.html
	resultHTMLdata, err := getResultsHTML(result)
	if err != nil {
		log.Printf("Could not get results HTML: %s\n", err.Error())
	} else {
		err = ioutil.WriteFile(filepath.Join(*outputDir, "results.html"), resultHTMLdata, 0644)
		if err != nil {
			log.Printf("Could not write results.html: %s\n", err.Error())
		}
	}

	log.Println("Done!")
}

func captureAndCompare(path string, link Link, wd selenium.WebDriver) (Link, error) {
	// Force the window size to be 1920x1280
	err := wd.ResizeWindow("", 1920, 1280)
	if err != nil {
		return Link{}, err
	}

	// Load the page from the alpha site.
	url := strings.Join([]string{*alphaBaseURL, path}, "")
	filename, err := capturePage(url, wd)
	if err != nil {
		return Link{}, err
	}
	link.Captured = true
	link.AlphaScreenshotFile = filename

	// Load the page from the beta site.
	url = strings.Join([]string{*betaBaseURL, path}, "")
	filename, err = capturePage(url, wd)
	if err != nil {
		return Link{}, err
	}
	link.BetaScreenshotFile = filename

	diff, percent, err := diff.CompareFiles(link.AlphaScreenshotFile, link.BetaScreenshotFile)
	if err != nil {
		return Link{}, err
	}

	link.Similarity = 100 - percent

	buf := new(bytes.Buffer)
	err = png.Encode(buf, diff)
	if err != nil {
		return Link{}, err
	}
	diffBytes := buf.Bytes()

	algorithm := sha1.New()
	algorithm.Write(diffBytes)
	filename = hex.EncodeToString(algorithm.Sum(nil))
	filename = filepath.Join(*outputDir, "diffs", filename+".png")
	err = ioutil.WriteFile(filename, diffBytes, 0644)
	if err != nil {
		return Link{}, err
	}

	link.DiffFile = filename

	return link, nil
}

func capturePage(url string, wd selenium.WebDriver) (string, error) {
	// Load the page.
	if err := wd.Get(url); err != nil {
		return "", err
	}

	// Wait a bit and then capture the screenshot.
	time.Sleep(time.Duration(*captureWait) * time.Millisecond)
	screendata, err := wd.Screenshot()
	if err != nil {
		return "", err
	}

	// Generate the filename and save the screenshot.
	algorithm := sha1.New()
	algorithm.Write(screendata)
	filename := hex.EncodeToString(algorithm.Sum(nil))
	filename = filepath.Join(*outputDir, "screenshots", filename+".png")
	err = ioutil.WriteFile(filename, screendata, 0644)
	if err != nil {
		return "", err
	}

	return filename, nil
}
