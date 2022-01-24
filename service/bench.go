package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hmatsu47/kusocode-bench/apimodel"
	"github.com/hmatsu47/kusocode-bench/dbaccess"
	"github.com/hmatsu47/kusocode-bench/dbmodel"
)

func Bench(ipAddress string) apimodel.Result {
	var result apimodel.Result
	// IP アドレスチェック
	team, err := dbaccess.FetchTeam(ipAddress)
	if err != nil {
		// チームテーブルにリクエスト元アドレスが見つからなかった or エラーが発生
		result.Id = 0
		result.Name = ""
		result.Score = 0
		result.Message = "不正なリクエスト元アドレスです"
		return result
	}
	result.Id = team.Id
	result.Name = team.Name
	result.Score = 0
	// 実行中？
	if team.ExecFlag != 0 {
		// 実行中
		result.Message = "前のベンチマークが実行中です"
		return result
	}
	// レスポンスチェック用のデータを初期化
	pictures, err := dbaccess.FetchPicturesAll()
	if err != nil || len(pictures) != 25 {
		// 写真テーブルからのデータ取得でエラーが発生
		result.Message = "内部エラーが発生しました。運営に連絡してください : ベンチマーク用 picture テーブルからのデータ取得でエラーが発生しました"
		return result
	}
	// 実行中にする
	team.Score = 0
	team.ExecFlag = 1
	err = dbaccess.UpdateTeam(team)
	if err != nil {
		result.Message = "内部エラーが発生しました。運営に連絡してください : ベンチマーク用 team テーブルの更新でエラーが発生しました"
		return result
	}
	// 以降、エラー発生時はベンチマーク実行フラグを 0 で更新する
	team.ExecFlag = 0
	// ベンチマーク先のデータを初期化
	message, err := initData(ipAddress)
	if err != nil {
		// 初期化処理でエラーが発生
		err = dbaccess.UpdateTeam(team)
		if err != nil {
			result.Message = "内部エラーが発生しました。運営に連絡してください : ベンチマーク用 team テーブルの更新でエラーが発生しました"
			return result
		}
		result.Message = "初期化リクエストが失敗しました : " + message
		return result
	}
	// 計測時間の起点を取得
	now := time.Now()
	// 最初の一覧データを取得
	items, err := listItem(ipAddress, 10)
	if err != nil {
		// 最初の一覧データ取得でエラーが発生
		err = dbaccess.UpdateTeam(team)
		if err != nil {
			result.Message = "内部エラーが発生しました。運営に連絡してください : ベンチマーク用 team テーブルの更新でエラーが発生しました"
			return result
		}
		result.Message = "一覧データ取得リクエストが失敗しました : 応答がないか不正です"
		return result
	}
	// 最初の一覧データをチェック
	message, _ = checkItem(items, true, -1, pictures)
	if message != "" {
		// 最初の一覧データ取得のチェックが失敗
		err = dbaccess.UpdateTeam(team)
		if err != nil {
			result.Message = "内部エラーが発生しました。運営に連絡してください : ベンチマーク用 team テーブルの更新でエラーが発生しました"
			return result
		}
		result.Message = "一覧データ取得リクエストが失敗しました : " + message
		return result
	}
	// チェック用のパラメータを設定
	threads := 10
	latency := time.Since(now).Milliseconds()
	if latency >= 5000 {
		threads = 2
	} else if latency >= 2500 {
		threads = 4
	} else if latency >= 2000 {
		threads = 5
	}
	offset := latency / int64(threads)
	result.Score = 10
	team.Score = result.Score
	// 設定スレッド数でリクエストを流す
	var wg sync.WaitGroup
	wg.Add(threads)
	var mu sync.Mutex
	i := 0
	for i < threads {
		go func() {
			// スレッド毎の初期値を設定
			thscore := 0
			thlastcount := -1
			thmessage := ""
			if i > 0 {
				time.Sleep((time.Duration(offset) * time.Millisecond))
			}
			for thmessage == "" && time.Since(now).Seconds() < 60 {
				// 一覧データを取得
				thitems, therr := listItem(ipAddress, 30)
				if therr != nil {
					// 一覧データ取得でエラーが発生
					thmessage = "一覧データ取得リクエストが失敗しました : 応答がないか不正です"
				} else if time.Since(now).Seconds() < 60 {
					// 時間内なら一覧データをチェック
					tmpmessage, tmpcount := checkItem(thitems, false, thlastcount, pictures)
					if tmpmessage != "" {
						// 一覧データ取得のチェックが失敗
						thmessage = "一覧データ取得リクエストが失敗しました : " + tmpmessage
					} else {
						// 一覧データ取得のスコアを加算
						thscore += 10
						thlastcount = tmpcount
						if time.Since(now).Seconds() < 60 {
							// 時間内なら写真画像データを取得
							tmpmessage = getImage(ipAddress, thitems[24].PictureId, pictures)
							if tmpmessage != "" {
								// 写真画像データのチェックが失敗
								thmessage = "写真画像データ取得リクエストが失敗しました : " + tmpmessage
							} else if time.Since(now).Seconds() < 60 {
								// 時間内なら写真画像データのスコアを加算
								thscore += 2
							}
						}
					}
				}
			}
			// 結果を返す（スコア加算・メッセージ返却）
			mu.Lock()
			defer mu.Unlock()
			defer wg.Done()
			result.Score += thscore
			if thmessage != "" {
				result.Message = thmessage
			}
		}()
		i++
	}
	wg.Wait()
	// DB に結果を書き込む
	team.Score = result.Score
	err = dbaccess.UpdateTeam(team)
	if err != nil {
		result.Message = "内部エラーが発生しました。運営に連絡してください : ベンチマーク用 team テーブルの更新でエラーが発生しました"
		return result
	}
	result.Message = "ベンチマークが完走しました!!"
	return result
}

func initData(ipAddress string) (string, error) {
	var message string
	// URL を生成
	u := &url.URL{}
	u.Scheme = "http"
	u.Host = ipAddress + ":8080"
	u.Path = "/kusocode3/InitData"
	uStr := u.String()
	// ポストデータ（ダミー）
	values := url.Values{}
	values.Add("dummy", "dummy")
	// タイムアウトを 10 秒に指定
	client := &http.Client{Timeout: time.Duration(10) * time.Second}
	// POST リクエスト発行
	resp, err := client.Post(uStr, "application/json; charset=utf-8", strings.NewReader(values.Encode()))
	if err != nil {
		fmt.Println(err)
		message = "応答がないか不正です"
		return message, err
	}
	// 関数を抜ける際に response を close
	defer resp.Body.Close()
	// レスポンスを取得
	body, _ := ioutil.ReadAll(resp.Body)
	var status apimodel.InitData
	err = json.Unmarshal(body, &status)
	if err != nil || status.Status != "OK" {
		fmt.Println(err)
		message = "応答が不正です"
		return message, err
	}
	message = ""
	return message, err
}

func listItem(ipAddress string, timeout int) ([]apimodel.Item, error) {
	var items []apimodel.Item
	// URL を生成
	u := &url.URL{}
	u.Scheme = "http"
	u.Host = ipAddress + ":8080"
	u.Path = "/kusocode3/ListItem"
	uStr := u.String()
	// タイムアウトを 10 秒に指定
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	// GET リクエスト発行
	resp, err := client.Get(uStr)
	if err != nil {
		fmt.Println(err)
		return items, err
	}
	// 関数を抜ける際に response を close
	defer resp.Body.Close()
	// レスポンスを取得
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &items)
	if err != nil {
		fmt.Println(err)
		return items, err
	}
	return items, err
}

func checkItem(items []apimodel.Item, zerocheck bool, lasttotal int, pictures [25]dbmodel.Picture) (string, int) {
	// 注 : 余分な項目があるのは許容・count が 0 の場合、count 項目自体が存在しないのは許容
	var message string
	totalcount := 0
	// 25 行あるか？
	if len(items) != 25 {
		message = "一覧の行数が違います"
		return message, totalcount
	}
	var flag [25]bool
	lastcount := 0
	// 全行チェック
	for i := 0; i < 25; i++ {
		// pictureId の範囲は 1 〜 25 か？
		pid := items[i].PictureId
		if pid < 1 || 25 < pid {
			message = "pictureId の値が不正です"
			return message, totalcount
		}
		// pictureId の重複はないか？
		if flag[pid-1] {
			message = "pictureId の値が重複しています"
			return message, totalcount
		}
		flag[pid-1] = true
		// タイトルと説明は正しいか？
		if items[i].Title != pictures[pid-1].Title {
			message = "title が違います"
			return message, totalcount
		}
		if items[i].Description != pictures[pid-1].Description {
			message = "description が違います"
			return message, totalcount
		}
		// zerocheck が true の場合、count は 0 か？
		count := items[i].Count
		if zerocheck && count != 0 {
			message = "アクセス数の初期化が行われていません"
			return message, totalcount
		}
		// 2 行目以降の場合、前行の count 以下か？
		if i > 0 && lastcount < count {
			message = "アクセス数が降順ではありません"
			return message, totalcount
		}
		lastcount = count
		totalcount += count
	}
	// count の合計は前回の合計値より大きいか？
	if totalcount <= lasttotal {
		message = "アクセス数が加算されていません"
		return message, totalcount
	}
	message = ""
	return message, totalcount
}

func getImage(ipAddress string, pictureId int, pictures [25]dbmodel.Picture) string {
	var message string
	// URL を生成
	u := &url.URL{}
	u.Scheme = "http"
	u.Host = ipAddress + ":8080"
	u.Path = "/kusocode3/GetImage"
	u.RawQuery = "pictureId=" + strconv.Itoa(pictureId)
	uStr := u.String()
	// タイムアウトを 10 秒に指定
	client := &http.Client{Timeout: time.Duration(30) * time.Second}
	// GET リクエスト発行
	resp, err := client.Get(uStr)
	if err != nil {
		fmt.Println(err)
		message = "応答がないか不正です"
		return message
	}
	// 関数を抜ける際に response を close
	defer resp.Body.Close()
	// レスポンスを取得
	body, _ := ioutil.ReadAll(resp.Body)
	var image apimodel.Image
	err = json.Unmarshal(body, &image)
	if err != nil {
		fmt.Println(err)
		message = "応答が不正です"
		return message
	}
	// pictureId をチェック
	pid := image.PictureId
	if pid != pictureId {
		message = "pictureId の値が不正です"
		return message
	}
	// image をチェック
	tmpimage := image.Image
	if tmpimage != base64.StdEncoding.EncodeToString(pictures[pid-1].Image) {
		message = "写真画像の内容が違います"
		return message
	}
	return message
}
