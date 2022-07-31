package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	//"sync"
)

var wg sync.WaitGroup

type IPTV struct {
	extinf string
	url    string
	isok   bool
}

func main() {
	var url = "https://raw.githubusercontent.com/Art2Cat/iptv-source/main/iptv-source"
	downloadFile("test.txt", url)

	m3us := readFileLines("test.txt")

	sl := make([]string, 0)
	for i, i2 := range m3us {
		tmp := fmt.Sprintf("tmp%d", i)
		downloadFile(tmp, i2)

		ls := readFileLines(tmp)
		sl = append(sl, ls...)
		err := os.Remove(tmp)
		if err != nil {
			log.Panic(err)
		}

	}

	iptvs := make([]*IPTV, 0)
	for i := 1; i < len(sl); i++ {
		l := sl[i]
		if strings.HasPrefix(l, "http") {
			iptvs[len(iptvs)-1].url = l
		} else if strings.HasPrefix(l, "#EXTINF") {
			iptvs = append(iptvs, &IPTV{extinf: l})
		}
	}
	iptvs = unique(iptvs)
	for _, iptv := range iptvs {
		wg.Add(1)
		go verifyM3u(iptv)
	}
	wg.Wait()
	result := make([]IPTV, 0)
	for _, iptv := range iptvs {
		// 移除省市地方台 及MIGU
		if iptv.isok && !strings.Contains(iptv.extinf, "省市地方") && !strings.Contains(iptv.extinf, "MIGU") {
			fmt.Printf("%+v\n", iptv)
			result = append(result, *iptv)
		}
	}

	println(len(result))

	saveM3u(result)

}

func saveM3u(result []IPTV) {
	f, err := os.Create("iptv.m3u")
	if err != nil {
		log.Panic(err)
	}
	w := bufio.NewWriter(f)
	_, err = w.WriteString("#EXTM3U\n")
	if err != nil {
		log.Panic(err)
	}
	for _, iptv := range result {
		_, err = w.WriteString(iptv.extinf + "\n" + iptv.url + "\n")
		if err != nil {
			log.Panic(err)
		}
	}
}

func downloadFile(filepath string, url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		e := resp.Body.Close()
		if e != nil {
			err = e
		}
	}()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		e := out.Close()
		if e != nil {
			log.Panic(e)
		}
	}()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Panic(err)
	}
}
func mergeSliceWithOutDuplicate(a []*IPTV, b []*IPTV) []*IPTV {
	d := append(a, b...)
	return unique(d)
}

func unique(a []*IPTV) []*IPTV {

	type key struct{ url string }
	res := make([]*IPTV, 0)
	check := make(map[key]int)

	for i, val := range a {
		k := key{
			url: val.url,
		}
		check[k] = i
	}

	for _, v := range check {
		res = append(res, a[v])
	}

	return res
}

func verifyM3u(i *IPTV) {
	defer wg.Done()
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(i.url)
	if err != nil {
		// log.Fatal(err)
		i.isok = false
		return
	}

	// Print the HTTP Status Code and Status Name
	fmt.Println("HTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))

	if resp.StatusCode == 200 {
		fmt.Println(i.url + ": ok")
		i.isok = true
	} else {
		fmt.Println(i.url + ": Argh! Broken")
		i.isok = false
	}
}

func readFileLines(path string) []string {

	lines := make([]string, 0)
	f, err := os.Open(path)
	if err != nil {
		return lines
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Panic(err)
		}
	}(f)
	r4 := bufio.NewReader(f)

	for {
		line, _, err := r4.ReadLine()

		if err == io.EOF {
			break
		}
		s := string(line)
		lines = append(lines, s)
	}
	return lines
}
