package main

import (
  "fmt"
  "github.com/OwenDurni/loltools/ddragon"
  "net/http"
  "io"
  "io/ioutil"
  "os"
)

const (
  Region = "na"
)

func FetchUrl(loc string) []byte {
  fmt.Fprintf(os.Stderr, "Fetch: %s\n", loc)
  res, err := http.Get(loc)
  if err != nil { panic(err) }
  defer res.Body.Close()
  data, err := ioutil.ReadAll(res.Body)
  if err != nil { panic(err) }
  return data
}

func main() {
  var debug io.Writer = nil // io.Writer(os.Stderr)
  dd := ddragon.NewDDragon(Region, debug)

  dd.ParseVersionJson(FetchUrl(dd.UrlVersionJson()))
  dd.ParseItemJson(FetchUrl(dd.UrlItemJson()))
  dd.ParseChampionJson(FetchUrl(dd.UrlChampionJson()))
  dd.ParseSummonerJson(FetchUrl(dd.UrlSummonerJson()))

  fmt.Fprintf(os.Stderr, "Most recent ddragon version is %s\n", dd.Version)
  fmt.Fprintf(os.Stderr, "Found %d Items.\n", len(dd.Items))
  fmt.Fprintf(os.Stderr, "Found %d Champions.\n", len(dd.Champions))
  fmt.Fprintf(os.Stderr, "Found %d Summoners.\n", len(dd.Summoners))
  
  dd.Debug = nil
  f := os.Stdout
  fmt.Fprintf(f, "package riot\n")
  fmt.Fprintf(f, "\n")
  fmt.Fprintf(f, "import (\n")
  fmt.Fprintf(f, "  \"github.com/OwenDurni/loltools/ddragon\"\n")
  fmt.Fprintf(f, "  \"io\"\n")
  fmt.Fprintf(f, ")\n")
  fmt.Fprintf(f, "\n")
  fmt.Fprintf(f, "//\n")
  fmt.Fprintf(f, "// This file is autogenerated\n")
  fmt.Fprintf(f, "//\n")
  fmt.Fprintf(f, "\n")
  fmt.Fprintf(f, "var Lookup ddragon.DDragon = %#v\n", *dd)
}
