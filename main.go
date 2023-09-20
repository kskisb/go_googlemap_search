package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"unicode/utf8"

	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func main() {
	// ハンドラの登録
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/callback", lineHandler)

	fmt.Println("http://localhost:8080 で起動中...")
	// HTTPサーバを起動
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	msg := "Hello World!!!!"
	fmt.Fprintf(w, msg)
}

func lineHandler(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// BOTを初期化
	secret := os.Getenv("LINE_CHANNEL_SECRET")
	token := os.Getenv("LINE_CHANNEL_TOKEN")
	bot, err := linebot.New(secret, token)
	if err != nil {
		log.Fatal(err)
	}

	// リクエストからBOTのイベントを取得
	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	for _, event := range events {
		// イベントがメッセージの受信だった場合
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			// メッセージがテキスト形式の場合
			case *linebot.TextMessage:
				replyMessage := message.Text
				_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do()
				if err != nil {
					log.Print(err)
				}
			// メッセージが位置情報の場合
			case *linebot.LocationMessage:
				sendRestoInfo(bot, event)
			}
		}
	}
}

// レストランの情報を送信する
func sendRestoInfo(bot *linebot.Client, e *linebot.Event) {
	msg := e.Message.(*linebot.LocationMessage)

	lat := strconv.FormatFloat(msg.Latitude, 'f', 2, 64)
	lng := strconv.FormatFloat(msg.Longitude, 'f', 2, 64)

	replyMsg := getRestoInfo(lat, lng)

	res := linebot.NewTemplateMessage(
		"レストラン一覧",
		linebot.NewCarouselTemplate(replyMsg...).WithImageOptions("rectangle", "cover"),
	)

	if _, err := bot.ReplyMessage(e.ReplyToken, res).Do(); err != nil {
		log.Print(err)
	}
}

// response APIレスポンス
type response struct {
	Results []shop `json:"results"`
}

// shop レストラン一覧
type shop struct {
	Name             string  `json:"name"`
	Address          string  `json:"vicinity"`
	URLS             string  `json:"place_id"`
	Rating           float64 `json:"rating"`
	UserRatingsTotal float64 `json:"user_ratings_total"`
}

// Places API で Nearby Search から取得した place_id を Place Details で URL に変換
func getURL(place_id string) string {
	apikey := os.Getenv("GOOGLE_MAPS_API_KEY")
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/place/details/json?place_id=%s&fields=url&key=%s", place_id, apikey)

	// HTTP GETリクエストを送信
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("リクエストエラー:", err)
		return ""
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取り
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("レスポンス読み取りエラー:", err)
		return ""
	}

	// JSONデコード
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("JSONデコードエラー:", err)
		return ""
	}

	// URLを抽出
	if result["status"].(string) != "OK" {
		return ""
	}
	url = result["result"].(map[string]interface{})["url"].(string)
	return url
}

// レストランの情報をGoogle Maps APIから取得する
func getRestoInfo(lat string, lng string) []*linebot.CarouselColumn {
	apikey := os.Getenv("GOOGLE_MAPS_API_KEY")
	url := fmt.Sprintf(
		"https://maps.googleapis.com/maps/api/place/nearbysearch/json?location=%s,%s&radius=1500&type=restaurant&key=%s&language=ja",
		lat, lng, apikey)

	// リクエストしてボディを取得
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data response
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal(err)
	}

	for i := range data.Results {
		// 店のurlを取得
		data.Results[i].URLS = getURL(data.Results[i].URLS)

		// 評価順に並べ替えるための重み付け
		data.Results[i].Rating = math.Pow(data.Results[i].Rating, 2.5) * math.Pow(data.Results[i].UserRatingsTotal, 0.25)
	}

	// 評価順にソート
	sort.Slice(data.Results, func(i, j int) bool {
		return data.Results[i].Rating > data.Results[j].Rating
	})

	var ccs []*linebot.CarouselColumn
	for _, shop := range data.Results {
		addr := shop.Address
		if utf8.RuneCountInString(addr) > 40 {
			addr = string([]rune(addr)[:40])
		}

		cc := linebot.NewCarouselColumn(
			"",
			shop.Name,
			addr,
			linebot.NewURIAction("Google Mapで開く", shop.URLS),
		).WithImageOptions("#FFFFFF")
		ccs = append(ccs, cc)
		if len(ccs) == 10 {
			break
		}
	}
	return ccs
}
