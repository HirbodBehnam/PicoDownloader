package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/dustin/go-humanize"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

type links struct {
	link, password string
}

const VERSION = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Please pass the file name as argument")
	}
	fmt.Println("Pico File Downloader by Hirbod Behnam v" + VERSION)
	fmt.Println("https://github.com/HirbodBehnam/PicoDownloader")
	log.Println("Reading the file...")
	var urls []links
	//Read file
	{
		file, err := os.Open(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.Split(scanner.Text()," ")
			l := links{link:line[0],password:""}
			if len(line) > 1{
				l.password = line[1]
			}
			urls = append(urls,l)
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	//Download each file
	for _,k := range urls{
		log.Println("Starting to download",k.link,"with password",k.password)
		log.Println("Getting link...")
		link,err := ExtractLink(k.link,k.password)
		if err != nil{
			log.Println("Error on getting link:",err.Error())
			log.Println()
			continue
		}
		log.Println("Extracted link:",link)
		_, file := path.Split(k.link)
		file = strings.TrimSuffix(file, path.Ext(file)) //Remove extension
		err = DownloadFile(file,link)
		if err != nil{
			log.Println("Error on downloading file:",err.Error())
			log.Println()
			continue
		}
		newFile, err := url.QueryUnescape(file)
		if err != nil{
			log.Println("Cannot unescape url to decode filename. The file name will be unescaped.")
			newFile = file
		}
		err = os.Rename(file+".tmp", newFile)
		if err != nil {
			log.Println("Cannot rename file:",err.Error())
		}
		log.Println("Download Finished")
		log.Println()
	}
}

func ExtractLink(u,password string) (string,error) {
	urlT,err := url.Parse(u)
	if err != nil{
		return "", err
	}
	query := "http://" + urlT.Host + "/file/generateDownloadLink?fileId=" + strings.Split(u,"/")[4]

	data := url.Values{}
	data.Set("password", password)
	req, err := http.NewRequest("POST", query, bytes.NewBuffer([]byte(data.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200{
		return "", fmt.Errorf("return code is not 200, it's %v; The body is %s",resp.StatusCode,body)
	}
	return string(body),nil
}

//Download stuff
//From https://golangcode.com/download-a-file-with-progress/
type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}
func DownloadFile(filepath string, url string) error {

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	// The progress use the same line so print a new line once it's finished downloading
	fmt.Print("\n")

	return nil
}