package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultID string = "3xra9abv3xq8i34"

const videoListQueryPayLoad string = `{"operationName":"publicFeedsQuery","variables":{"principalId":"%s","pcursor":"0","count":1000},"query":"query publicFeedsQuery($principalId: String, $pcursor: String, $count: Int) {\n  publicFeeds(principalId: $principalId, pcursor: $pcursor, count: $count) {\n    pcursor\n    live {\n      user {\n        id\n        avatar\n        name\n        __typename\n      }\n      watchingCount\n      poster\n      coverUrl\n      caption\n      id\n      playUrls {\n        quality\n        url\n        __typename\n      }\n      quality\n      gameInfo {\n        category\n        name\n        pubgSurvival\n        type\n        kingHero\n        __typename\n      }\n      hasRedPack\n      liveGuess\n      expTag\n      __typename\n    }\n    list {\n      id\n      thumbnailUrl\n      poster\n      workType\n      type\n      useVideoPlayer\n      imgUrls\n      imgSizes\n      magicFace\n      musicName\n      caption\n      location\n      liked\n      onlyFollowerCanComment\n      relativeHeight\n      timestamp\n      width\n      height\n      counts {\n        displayView\n        displayLike\n        displayComment\n        __typename\n      }\n      user {\n        id\n        eid\n        name\n        avatar\n        __typename\n      }\n      expTag\n      __typename\n    }\n    __typename\n  }\n}\n"}`

const videoDetailQueryPayLoad string = `{"operationName":"SharePageQuery","variables":{"photoId":"%s","principalId":"%s"},"query":"query SharePageQuery($principalId: String, $photoId: String) {\n feedById(principalId: $principalId, photoId: $photoId) {\n    currentWork {\n      playUrl\n      __typename\n    }\n    __typename\n  }\n}\n"}`

const hqlURL string = `https://live.kuaishou.com/m_graphql`

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
	fmt.Print("请输入待爬取的用户快手ID(在用户的快手主页可查看 不输入爬取默认的用户(ID=3xra9abv3xq8i34)):")
	var id string = defaultID
	_, err := fmt.Scanln(&id)
	if err != nil {
		fmt.Println("你貌似没有输入任何数据，将默认为你爬取默认用户的视频...")
		id = defaultID
	}
	fmt.Printf("爬取的用户ID是：%v\n", id)
	fmt.Printf("即将开始爬取(爬取默认超时10s)...id=%v\n", id)
	c := make(chan []SingleVideoInfo)
	go getVideoListByInterface(id, c)
	var videos []SingleVideoInfo
	start := time.Now().Second()
	select {
	case videos = <-c:
		fmt.Printf("数据爬取完成，耗时%v秒\n", time.Now().Second()-start)
		close(c)
	case <-time.After(10 * time.Second):
		close(c)
		fmt.Println("数据爬取超时...")
		selectMenu()
		return
	}

	var count = len(videos)
	if count == 0 {
		fmt.Println("爬取到网页中的有效视频数量是0...这有可能是因为页面请求失败,可以选择爬取别的主页")
		selectMenu()
		return
	} else {
		_ = os.Mkdir("kuaishou_spider_images", os.ModePerm)
		_ = os.Mkdir("kuaishou_spider_videos", os.ModePerm)
		_ = os.Mkdir("kuaishou_spider_images"+"/"+id, os.ModePerm)
		_ = os.Mkdir("kuaishou_spider_videos"+"/"+id, os.ModePerm)
		fmt.Printf("爬取到网页中的有效视频数量是%v\n", len(videos))
		time.Sleep(time.Second)
		fmt.Println("即将开始下载封面图和视频...")
		wd, _ := os.Getwd()
		fmt.Printf("所有封面图将保存到：%v\n", wd+"images/"+id+"/")
		fmt.Printf("所有视频将保存到：%v\n", wd+"videos/"+id+"/")
		time.Sleep(time.Second)
		ch := make(chan bool, count)
		for index, video := range videos {
			go getVideoDetail(index, video.ID, video.User.ID, video.ThumbnailURL, id, ch)
		}
		var succ, failed int = 0, 0
		for b := range ch {
			// 每次从ch中接收数据，表明一个活动的协程结束
			count--
			// 当所有活动的协程都结束时，关闭管道
			if count == 0 {
				close(ch)
			}
			if b {
				succ++
			} else {
				failed++
			}
		}
		fmt.Printf("%v个任务成功,%v个任务失败\n", succ, failed)
		selectMenu()
	}
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
	req, err := http.NewRequest(http.MethodPost, hqlURL, strings.NewReader(query))

	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36")
	
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		return
	}

	defer resp.Body.Close()

	buf, err = ioutil.ReadAll(resp.Body)

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
	err := ioutil.WriteFile(fileName, buf, 0644)
	if err != nil {
		fmt.Printf("文件 %v ----> 写入失败，原因%v\n", fileName, err)
	} else {
		fmt.Printf("文件 %v ----> 写入成功\n", fileName)
	}
}

//获取写入的文件名
func getFileName(root, id, url string) (fileName string) {
	splits := strings.Split(url, "/")
	fileName = root + id + "/" + splits[len(splits)-1]
	return
}

func selectMenu() {
	fmt.Println("输入1继续爬取别的用户视频,输入其它退出")
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
