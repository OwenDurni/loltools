package ddragon

import (
  "bytes"
  "encoding/json"
  "fmt"
  "net/http"
  "image"
  "io"
  "io/ioutil"
  "os"
  "strconv"

  _ "image/png"
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

type DDragon struct {
  Debug io.Writer
  Region string
  Version string
  Language string
  CdnRoot string
  
  Items map[int]DDItem
  Champions map[int]DDChampion
  Summoners map[int]DDSummoner
  SpriteSheets map[string]DDSpriteSheet
}
func NewDDragon(region string, debug io.Writer) *DDragon {
  dd := new(DDragon)
  dd.Debug = debug
  dd.Region = region
  dd.Items = make(map[int]DDItem)
  dd.Champions = make(map[int]DDChampion)
  dd.Summoners = make(map[int]DDSummoner)
  dd.SpriteSheets = make(map[string]DDSpriteSheet)
  return dd
}
type DDSpriteSheet struct {
  Url string
  H int
  W int
}
type DDSprite struct {
  Url string
  X int
  Y int
  W int
  H int
}
type DDItem struct {
  Id        int
  Name      string
  ImageUrl  string
  Sprite DDSprite
}
type DDChampion struct {
  Id int
  Name string
  ImageUrl string
  SplashUrl string
  Sprite DDSprite
}
type DDSummoner struct {
  Id int
  Name string
  ImageUrl string
  Sprite DDSprite
}
func (dd *DDragon) Debugf(format string, args ...interface{}) {
  if dd.Debug == nil { return }
  fmt.Fprintf(dd.Debug, fmt.Sprintf("%s\n", format), args...)
}
func (dd *DDragon) UrlBase() string {
  return "http://ddragon.leagueoflegends.com"
}
func (dd *DDragon) UrlVersionJson() string {
  return fmt.Sprintf("%s/realms/%s.json", dd.UrlBase(), dd.Region)
}
func (dd *DDragon) UrlItemJson() string {
  return fmt.Sprintf("%s/%s/data/%s/item.json", dd.CdnRoot, dd.Version, dd.Language)
}
func (dd *DDragon) UrlChampionJson() string {
  return fmt.Sprintf("%s/%s/data/%s/champion.json", dd.CdnRoot, dd.Version, dd.Language)
}
func (dd *DDragon) UrlSummonerJson() string {
  return fmt.Sprintf("%s/%s/data/%s/summoner.json", dd.CdnRoot, dd.Version, dd.Language)
}
func (dd *DDragon) UrlItemImage(filename string) string {
  return fmt.Sprintf("%s/%s/img/item/%s", dd.CdnRoot, dd.Version, filename)
}
func (dd *DDragon) UrlChampionImage(filename string) string {
  return fmt.Sprintf("%s/%s/img/champion/%s", dd.CdnRoot, dd.Version, filename)
}
func (dd *DDragon) UrlSummonerImage(filename string) string {
  return fmt.Sprintf("%s/%s/img/summoner/%s", dd.CdnRoot, dd.Version, filename)
}
func (dd *DDragon) UrlSprite(filename string) string {
  url := fmt.Sprintf("%s/%s/img/sprite/%s", dd.CdnRoot, dd.Version, filename)
  dd.AddSpriteSheet(url)
  return url
}

func (dd *DDragon) AddSpriteSheet(url string) {
  if _, exists := dd.SpriteSheets[url]; exists {
    return
  }
  var ss DDSpriteSheet
  ss.Url = url

  imageBytes := FetchUrl(url)
  image, _, err := image.Decode(bytes.NewReader(imageBytes))
  if err != nil { panic(err) }
  ss.W = image.Bounds().Dx()
  ss.H = image.Bounds().Dy()

  dd.SpriteSheets[url] = ss
}

// "v": "{{Version}}"
// "l": "{{Language}}"
// "cdn": "{{CdnRoot}}"
type versionRootJson struct {
  Version string `json:"v"`
  Language string `json:"l"`
  CdnRoot string `json:"cdn"`
}
func (dd *DDragon) ParseVersionJson(jsonData []byte) {
  versionRootJ := new(versionRootJson)
  if err := json.Unmarshal(jsonData, versionRootJ); err != nil {
    panic(err)
  }
  dd.Version = versionRootJ.Version
  dd.Language = versionRootJ.Language
  dd.CdnRoot = versionRootJ.CdnRoot
  
  dd.Debugf("DD.Version: %s", dd.Version)
  dd.Debugf("DD.Language: %s", dd.Language)
  dd.Debugf("DD.CdnRoot: %s", dd.CdnRoot)
}

// "data":
//   "{{ItemNumber}}":
//     "name": "{{Name}}"
//     "image":
//       "full": "{{ItemImageFile}}"
//       "sprite":
//         "group": "{{ItemSpriteGroup}}"
//         "x": {{ItemSpriteX}}
//         "y": {{ItemSpriteY}}
//         "w": {{ItemSpriteW}}
//         "h": {{ItemSpriteH}}
type itemRootJson struct {
  Data map[string]itemJson `json:"data"`
}
type itemJson struct {
  Name  string    `json:"name"`
  Image imageJson `json:"image"`
}
type imageJson struct {
  Full string `json:"full"`
  Sprite string `json:"sprite"`
  SpriteX int `json:"x"`
  SpriteY int `json:"y"`
  SpriteW int `json:"w"`
  SpriteH int `json:"h"`
}
func (dd *DDragon) ParseItemJson(jsonData []byte) {
  itemRootJ := new(itemRootJson)
  if err := json.Unmarshal(jsonData, itemRootJ); err != nil {
    panic(err)
  }
  dd.Debugf("DD.Items: ")
  for itemIdStr, itemJ := range itemRootJ.Data {
    itemId, err := strconv.Atoi(itemIdStr)
    if err != nil { panic(err) }
    var item DDItem
    item.Id = itemId
    item.Name = itemJ.Name
    
    imageJ := itemJ.Image
    item.ImageUrl = dd.UrlItemImage(imageJ.Full)
    item.Sprite.Url = dd.UrlSprite(imageJ.Sprite)
    item.Sprite.X = imageJ.SpriteX
    item.Sprite.Y = imageJ.SpriteY
    item.Sprite.W = imageJ.SpriteW
    item.Sprite.H = imageJ.SpriteH
    
    dd.Items[itemId] = item
    dd.Debugf("DD.Items[%d]: %v\n", item.Id, item)
  }
}

// "data":
//   "{{ChampoinName}}":
//     "key": "{{RiotChampionId}}"
//     "image":
//       "full": "{{ItemImageFile}}"
//       "sprite":
//         "group": "{{ItemSpriteGroup}}"
//         "x": {{ItemSpriteX}}
//         "y": {{ItemSpriteY}}
//         "w": {{ItemSpriteW}}
//         "h": {{ItemSpriteH}}
type championRootJson struct {
  Data map[string]championJson `json:"data"`
}
type championJson struct {
  Key   string    `json:"key"`
  Name  string    `json:"name"`
  Image imageJson `json:"image"`
}
func (dd *DDragon) ParseChampionJson(jsonData []byte) {
  championRootJ := new(championRootJson)
  if err := json.Unmarshal(jsonData, championRootJ); err != nil {
    panic(err)
  }
  dd.Debugf("DD.Champions: ")
  for _, championJ := range championRootJ.Data {
    championId, err := strconv.Atoi(championJ.Key)
    if err != nil { panic(err) }
    var champion DDChampion
    champion.Id = championId
    champion.Name = championJ.Name
    
    imageJ := championJ.Image
    champion.ImageUrl = dd.UrlChampionImage(imageJ.Full)
    
    champion.Sprite.Url = dd.UrlSprite(imageJ.Sprite)
    champion.Sprite.X = imageJ.SpriteX
    champion.Sprite.Y = imageJ.SpriteY
    champion.Sprite.W = imageJ.SpriteW
    champion.Sprite.H = imageJ.SpriteH
    
    dd.Champions[championId] = champion
    dd.Debugf("DD.Champions[%d]: %v\n", champion.Id, champion)
  }
}

// "data":
//   "{{SummonerName}}":
//     "key": "{{RiotSummonerId}}"
//     "image":
//       "full": "{{ItemImageFile}}"
//       "sprite":
//         "group": "{{ItemSpriteGroup}}"
//         "x": {{ItemSpriteX}}
//         "y": {{ItemSpriteY}}
//         "w": {{ItemSpriteW}}
//         "h": {{ItemSpriteH}}
type summonerRootJson struct {
  Data map[string]summonerJson `json:"data"`
}
type summonerJson struct {
  Key string `json:"key"`
  Name string `json:"name"`
  Image imageJson `json:"image"`
}
func (dd *DDragon) ParseSummonerJson(jsonData []byte) {
  summonerRootJ := new(summonerRootJson)
  if err := json.Unmarshal(jsonData, summonerRootJ); err != nil {
    panic(err)
  }
  dd.Debugf("DD.Summoners: ")
  for _, summonerJ := range summonerRootJ.Data {
    summonerId, err := strconv.Atoi(summonerJ.Key)
    if err != nil { panic(err) }
    var summoner DDSummoner
    summoner.Id = summonerId
    summoner.Name = summonerJ.Name
    
    imageJ := summonerJ.Image
    summoner.ImageUrl = dd.UrlSummonerImage(imageJ.Full)
    
    summoner.Sprite.Url = dd.UrlSprite(imageJ.Sprite)
    summoner.Sprite.X = imageJ.SpriteX
    summoner.Sprite.Y = imageJ.SpriteY
    summoner.Sprite.W = imageJ.SpriteW
    summoner.Sprite.H = imageJ.SpriteH
    
    dd.Summoners[summonerId] = summoner
    dd.Debugf("DD.Summoners[%d]: %v\n", summoner.Id, summoner)
  }
}
