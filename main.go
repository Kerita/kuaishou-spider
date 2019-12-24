package main

//author:keguoyu
//仅用于技术学习 不用于商业用途

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	"math/rand"
)

const (
	TextBlack = iota + 30
	TextRed
	TextGreen
	TextYellow
	TextBlue
	TextMagenta
	TextCyan
	TextWhite
)

const defaultID string = "3xra9abv3xq8i34"

const videoListQueryPayLoad string = `{"operationName":"publicFeedsQuery","variables":{"principalId":"%s","pcursor":"0","count":1000},"query":"query publicFeedsQuery($principalId: String, $pcursor: String, $count: Int) {\n  publicFeeds(principalId: $principalId, pcursor: $pcursor, count: $count) {\n    pcursor\n    live {\n      user {\n        id\n        avatar\n        name\n        __typename\n      }\n      watchingCount\n      poster\n      coverUrl\n      caption\n      id\n      playUrls {\n        quality\n        url\n        __typename\n      }\n      quality\n      gameInfo {\n        category\n        name\n        pubgSurvival\n        type\n        kingHero\n        __typename\n      }\n      hasRedPack\n      liveGuess\n      expTag\n      __typename\n    }\n    list {\n      id\n      thumbnailUrl\n      poster\n      workType\n      type\n      useVideoPlayer\n      imgUrls\n      imgSizes\n      magicFace\n      musicName\n      caption\n      location\n      liked\n      onlyFollowerCanComment\n      relativeHeight\n      timestamp\n      width\n      height\n      counts {\n        displayView\n        displayLike\n        displayComment\n        __typename\n      }\n      user {\n        id\n        eid\n        name\n        avatar\n        __typename\n      }\n      expTag\n      __typename\n    }\n    __typename\n  }\n}\n"}`

const videoDetailQueryPayLoad string = `{"operationName":"SharePageQuery","variables":{"photoId":"%s","principalId":"%s"},"query":"query SharePageQuery($principalId: String, $photoId: String) {\n feedById(principalId: $principalId, photoId: $photoId) {\n    currentWork {\n      playUrl\n      __typename\n    }\n    __typename\n  }\n}\n"}`

const hqlURL string = `https://live.kuaishou.com/m_graphql`

var userAgentList []string = []string{"Mozilla/5.0 (compatible, MSIE 10.0, Windows NT, DigExt)",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, 360SE)",
	"Mozilla/4.0 (compatible, MSIE 8.0, Windows NT 6.0, Trident/4.0)",
	"Mozilla/5.0 (compatible, MSIE 9.0, Windows NT 6.1, Trident/5.0,",
	"Opera/9.80 (Windows NT 6.1, U, en) Presto/2.8.131 Version/11.11",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, TencentTraveler 4.0)",
	"Mozilla/5.0 (Windows, U, Windows NT 6.1, en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (Macintosh, Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
	"Mozilla/5.0 (Macintosh, U, Intel Mac OS X 10_6_8, en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (Linux, U, Android 3.0, en-us, Xoom Build/HRI39) AppleWebKit/534.13 (KHTML, like Gecko) Version/4.0 Safari/534.13",
	"Mozilla/5.0 (iPad, U, CPU OS 4_3_3 like Mac OS X, en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, Trident/4.0, SE 2.X MetaSr 1.0, SE 2.X MetaSr 1.0, .NET CLR 2.0.50727, SE 2.X MetaSr 1.0)",
	"Mozilla/5.0 (iPhone, U, CPU iPhone OS 4_3_3 like Mac OS X, en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"MQQBrowser/26 Mozilla/5.0 (Linux, U, Android 2.3.7, zh-cn, MB200 Build/GRJ22, CyanogenMod-7) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1"}

type SingleVideoInfo struct {
	ID                	   string        `json:"id"`
	Caption                string        `json:"caption"`
	ThumbnailURL           string        `json:"thumbnailUrl"`
	Poster                 string        `json:"poster"`
	ViewCount              string        `json:"viewCount"`
	LikeCount              string        `json:"likeCount"`
	CommentCount           string        `json:"commentCount"`
	Timestamp              int64         `json:"timestamp"`
	WorkType               string        `json:"workType"`
	Type                   string        `json:"type"`
	UseVideoPlayer         bool          `json:"useVideoPlayer"`
	ImgUrls                []interface{} `json:"imgUrls"`
	ImgSizes               []interface{} `json:"imgSizes"`
	MagicFace              interface{}   `json:"magicFace"`
	MusicName              string        `json:"musicName"`
	Location               interface{}   `json:"location"`
	Liked                  bool          `json:"liked"`
	OnlyFollowerCanComment bool          `json:"onlyFollowerCanComment"`
	RelativeHeight         interface{}   `json:"relativeHeight"`
	Width                  int           `json:"width"`
	Height                 int           `json:"height"`
	User                   struct {
		ID       string `json:"id"`
		Eid      string `json:"eid"`
		Name     string `json:"name"`
		Profile  string `json:"profile"`
		Typename string `json:"__typename"`
	} `json:"user"`
	ExpTag   string `json:"expTag"`
	Typename string `json:"__typename"`
}

type VideoList struct {
	Data struct {
		PublicFeeds struct {
			Pcursor  string            `json:"pcursor"`
			Live     interface{}       `json:"live"`
			List     []SingleVideoInfo `json:"list"`
			Typename string            `json:"__typename"`
		} `json:"publicFeeds"`
	} `json:"data"`
}

type VideoDetail struct {
	Data struct {
		FeedByID struct {
			CurrentWork struct {
				PlayURL  string `json:"playUrl"`
				Typename string `json:"__typename"`
			} `json:"currentWork"`
			Typename string `json:"__typename"`
		} `json:"feedById"`
	} `json:"data"`
}

func main() {
	fmt.Println("\t\t*****Welcome to kuaishou-spider by keguoyu*****")
	mainLoop()
}

func mainLoop() {
	fmt.Print("Please input the id that will be spided: ")
	var id string = defaultID
	_, err := fmt.Scanln(&id)
	if err != nil {
		fmt.Println("Your input is invalid, will spide default user who's id is 3xm9nrxtd9qbb3w")
		id = defaultID
	}
	fmt.Printf("Will spide the data of %v, timeout is 10s \n", id)
	c := make(chan []SingleVideoInfo)
	go getVideoListByInterface(id, c)
	var videos []SingleVideoInfo
	start := time.Now().Second()
	select {
	case videos = <-c:
		fmt.Printf("Get data complete, it costs %vs\n", time.Now().Second()-start)
		close(c)
	case <-time.After(10 * time.Second):
		close(c)
		fmt.Println("Get data timeout")
		selectMenu()
		return
	}

	var count = len(videos)

	if count == 0 {
		fmt.Println("The valid data is 0, you may try other ids")
		selectMenu()
		return
	} else {
		_ = os.Mkdir("kuaishou_spider_images", os.ModePerm)
		_ = os.Mkdir("kuaishou_spider_videos", os.ModePerm)
		_ = os.Mkdir("kuaishou_spider_images"+"/"+id, os.ModePerm)
		_ = os.Mkdir("kuaishou_spider_videos"+"/"+id, os.ModePerm)
		fmt.Printf("The number of videos is %v\n", len(videos))
		time.Sleep(time.Second)
		fmt.Println("Will download images and videos...")
		wd, _ := os.Getwd()
		fmt.Printf("All images will be saved at %v\n", wd+"images/"+id+"/")
		fmt.Printf("All videos will be saved at %v\n", wd+"videos/"+id+"/")
		time.Sleep(time.Second)
		ch := make(chan bool, count)
		for index, video := range videos {
			// fmt.Println(video.ID,",",video.Caption)
			go getVideoDetail(index, video.ID, video.User.ID, video.ThumbnailURL, id, ch)
		}
		var succ, failed int = 0, 0
		var done int = 0
		for b := range ch {
			// 每次从ch中接收数据，表明一个活动的协程结束
			done++
			// 当所有活动的协程都结束时，关闭管道
			if done == count {
				close(ch)
			}
			// percent := int(float32(done) * 100.0 / float32(count))
			// show(percent, count)
			updateProgress(done,count)
			if b {
				succ++
			} else {
				failed++
			}
		}

		fmt.Printf("\n%v succed, %v failed\n", succ, failed)
		selectMenu()
	}
}

//用于更新进度 我们最多展示50个# 如果不足50那么就展示那么多
func updateProgress(done, total int) {
	var showCount int
	if total < 50 {
		showCount = total
	} else {
		showCount = 50
	}
	arr := make([]string, showCount)

	percent := float32(done) / float32(total)

	show := int(percent * float32(showCount))

	for j := 0;j < showCount; j++ {
		if j <= show {
			arr[j] = "#"
		} else {
			arr[j] = " "
		}
	}
	bar := fmt.Sprintf("[%s]", strings.Join(arr, ""))
	dis := int(show * 100 / showCount)
    fmt.Printf("\r%s %%%d", bar, dis)
}

func show(percent, total int) {
    middle := int(percent * total / 100.0)
    arr := make([]string, total)
    for j := 0; j < total; j++ {
        if j <= middle-1 {
            arr[j] = "#"
        }  else {
            arr[j] = " "
        }
    }
    bar := fmt.Sprintf("[%s]", strings.Join(arr, ""))
    fmt.Printf("\r%s %%%d", bar, percent)
}

//通过接口拿到剩下的视频
func getVideoListByInterface(id string, ch chan []SingleVideoInfo) {
	query := fmt.Sprintf(videoListQueryPayLoad, id)
	body, err := getHTTPResponse(query)

	if err != nil {
		return
	}

	var videoList VideoList
	err = json.Unmarshal(body, &videoList)

	if err != nil {
		return
	}

	videos := videoList.Data.PublicFeeds.List

	ch <- videos
}

//爬取单个视频详情
func getVideoDetail(index int, photoID, id, imageURL, uid string, ch chan bool) {
	query := fmt.Sprintf(videoDetailQueryPayLoad, photoID, id)
	body, err := getHTTPResponse(query)

	if err != nil {
		ch <- false
		return
	}

	var videoDetail VideoDetail
	err = json.Unmarshal(body, &videoDetail)
	if err != nil {
		ch <- false
		return
	}

	playURL := videoDetail.Data.FeedByID.CurrentWork.PlayURL

	downloadImageToDisk(imageURL, uid)
	downloadVideoToDisk(playURL, uid)
	ch <- true
}

func getHTTPResponse(query string) (buf []byte, err error) {
	// fmt.Println("query=",query)
	req, err := http.NewRequest(http.MethodPost, hqlURL, strings.NewReader(query))

	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", getRandomUserAgent())
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cookie" ,"clientid=3; did=web_8cc7fb39cbb74656ab5616ca05577583; client_key=65890b29; didv=1577092573000; userId=224866748; userId=224866748; kuaishou.live.bfb1s=3e261140b0cf7444a0ba411c6f227d88; kuaishou.live.web_st=ChRrdWFpc2hvdS5saXZlLndlYi5zdBKgAS_4HNcabTk1EWoNnD3-2JYWyyFejF50w0OfgFprUgEj99ffcHabzKHrIejI3uhLBhqh7osAPHyLdC9jmNADP9RXY78XPhiSN0Ab1bhKZwZYcf5naeGTPpHk2OCP8TCE9KzoYlV0y6TLtXD-o_wAsWa5lMJv8cdOhmOMGjF4EcYTRZWBMejAE5APDAcM5qgMN-AT1vqXGbFZwlUlOs9akokaEqVj6czba05AmU7FMveU3Qsu3iIgQ8GsDp6asscOGTAy0RQHzGEX3kIggSEupjC11-HN9h0oBTAB; kuaishou.live.web_ph=c6f5d0ff8e5b68734e8fa9d8962381538acd")
	
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		return
	}

	defer resp.Body.Close()

	buf, err = ioutil.ReadAll(resp.Body)

	// fmt.Println(string(buf))

	return buf, err
}

func downloadVideoToDisk(url, uid string) {
	buf, err := getBytesResp(url)
	if err != nil {
		return
	}

	fileName := getFileName("kuaishou_spider_videos/", uid, url)
	saveFile(fileName, buf)

}

//下载图片
func downloadImageToDisk(url, uid string) {
	buf, err := getBytesResp(url)
	if err != nil {
		return
	}
	fileName := getFileName("kuaishou_spider_images/", uid, url)
	saveFile(fileName, buf)
}

//获取响应字节流
func getBytesResp(url string) (buf []byte, err error) {
	res, err := http.Get(url)
	if err != nil {
		return
	}

	defer res.Body.Close()

	buf, err = ioutil.ReadAll(res.Body)

	if err != nil {
		return
	}
	return
}

//写入文件
func saveFile(fileName string, buf []byte) {
	_  = ioutil.WriteFile(fileName, buf, 0644)
}

//获取写入的文件名
func getFileName(root, id, url string) (fileName string) {
	splits := strings.Split(url, "/")
	fileName = root + id + "/" + splits[len(splits)-1]
	return
}

func SetColor(msg string, conf, bg, text int) string {
    return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, conf, bg, text, msg, 0x1B)
}

func getRandomUserAgent() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return userAgentList[r.Intn(len(userAgentList))]
}

func selectMenu() {
	fmt.Println("Press 1 to continue, press others to exit")
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		os.Exit(0)
	} else {
		if input == "1" {
			mainLoop()
		} else {
			os.Exit(0)
		}
	}
}
