package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultUrl string = "http://live.kuaishou.com/profile/3xra9abv3xq8i34"

const videoListQueryFormat string = `{publicFeeds(principalId: "%s", pcursor: "0", count: %d) {
   pcursor
   live {
     user {
       id
       kwaiId
       eid
       profile
       name
       living
       __typename
     }
     watchingCount
     src
     title
     gameId
     gameName
     categoryId
     liveStreamId
     playUrls {
       quality
       url
       __typename
     }
     followed
     type
     living
     redPack
     liveGuess
     anchorPointed
     latestViewed
     expTag
     __typename
   }
   list {
     photoId
     caption
     thumbnailUrl
     poster
     viewCount
     likeCount
     commentCount
     timestamp
     workType
     type
     useVideoPlayer
     imgUrls
     imgSizes
     magicFace
     musicName
     location
     liked
     onlyFollowerCanComment
     relativeHeight
     width
     height
     user {
       id
       eid
       name
       profile
       __typename
     }
     expTag
     __typename
   }
   __typename
  }}`

const videoDetailQueryFormat string = `{feedById(photoId:"%s",principalId:"%s"){currentWork{playUrl,__typename},__typename}}`

const hqlUrl string = `https://live.kuaishou.com/graphql`

type SingleVideoInfo struct {
	PhotoID                string        `json:"photoId"`
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
	fmt.Print("请输入待爬取的个人首页url(不输入爬取默认用户视频 例如：http://live.kuaishou.com/profile/3xra9abv3xq8i34)：")
	var homeUrl string = defaultUrl
	_, err := fmt.Scanln(&homeUrl)
	if err != nil {
		fmt.Println("你貌似没有输入任何数据，将默认为你爬取默认用户的视频...")
		homeUrl = defaultUrl
	}
	fmt.Printf("爬取的主页是：%v\n", homeUrl)

	slice := strings.Split(homeUrl, "/")
	var last string = slice[len(slice)-1]
	index := strings.Index(last, "?")
	var id string
	if index <= 0 {
		id = last
	} else {
		id = substring(last, 0, index)
	}
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
			go getVideoDetail(index, video.PhotoID, video.User.ID, video.ThumbnailURL, id, ch)
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

func substring(source string, start int, end int) string {
	var r = []rune(source)
	length := len(r)

	if start < 0 || end > length || start > end {
		return ""
	}

	if start == 0 && end == length {
		return source
	}

	return string(r[start:end])
}

//通过接口拿到剩下的视频
func getVideoListByInterface(id string, ch chan []SingleVideoInfo) {
	query := fmt.Sprintf(videoListQueryFormat, id, 1000)
	body, err := getHttpResponse(query)

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
func getVideoDetail(index int, photoId, id, imageUrl, uid string, ch chan bool) {
	query := fmt.Sprintf(videoDetailQueryFormat, photoId, id)
	body, err := getHttpResponse(query)

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

	playUrl := videoDetail.Data.FeedByID.CurrentWork.PlayURL

	downloadImageToDisk(imageUrl, uid)
	downloadVideoToDisk(playUrl, uid)
	ch <- true
}

func getHttpResponse(query string) (buf []byte, err error) {
	values := url.Values{}
	values.Set("query", query)
	req, err := http.NewRequest(http.MethodPost,
		hqlUrl,
		strings.NewReader(values.Encode()))

	if err != nil {
		return
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Add("Content-Type",
		"application/x-www-form-urlencoded; param=value")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.132 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
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
