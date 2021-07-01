package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Properties struct {
	SupportInternal   bool `json:"support_internal"` // for aliyun
	SupportIpv6       bool `json:"support_ipv6"`
	SupportAccelerate bool `json:"support_accelerate"` // 加速, support: aliyun, tencentcloud
	NeedAppID         bool `json:"need_appid"`         // 主账号 only for 腾讯cos
}

type Region struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
	Region   string `json:"region"`
	Endpoint string `json:"endpoint"`
	//IEndpoint  string
	Properties Properties `json:"properties"`
}

func main() {
	ls := make([]*Region, 0)

	l1 := AliyunOSS()
	ls = append(ls, l1...)

	l2 := TencentCOS()
	ls = append(ls, l2...)

	l3 := HuaweiOBS()
	ls = append(ls, l3...)

	for _, v := range ls {
		fmt.Printf("%+v\n", v)
	}

	fmt.Println("regino info by json:")
	data, _ := json.Marshal(ls)
	fmt.Println(string(data))
}

func AliyunOSS() []*Region {
	res, err := http.Get("https://help.aliyun.com/document_detail/31837.htm")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	ls := make([]*Region, 0)
	doc.Find("#tbody-pmx-2hz-rzy tr").Each(func(row int, tr *goquery.Selection) {
		tmp := &Region{}

		tr.Find("td").Each(func(col int, td *goquery.Selection) {
			tmp.Provider = "aliyun"
			//fmt.Println(row, col, td.Text())
			switch col {
			case 0:
				tmp.Name = strings.TrimSpace(td.Text())
			case 1:
				tmp.Region = strings.TrimSpace(td.Text())
			case 2:
				tmp.Properties.SupportIpv6 = strings.TrimSpace(td.Text()) == "是"
			case 3:
				tmp.Endpoint = strings.TrimSpace(td.Text())
			case 4:
				tmp.Properties.SupportInternal = strings.TrimSpace(td.Text()) != ""
			}
		})

		if tmp.Name == "" {
			return
		}

		ls = append(ls, tmp)
	})

	ls = append(ls,
		&Region{ // 不支持内网Endpoint
			Provider: "aliyun",
			Name:     "全球加速Endpoint",
			Region:   "oss-accelerate",
			Endpoint: "oss-accelerate.aliyuncs.com",
			Properties: Properties{
				SupportAccelerate: true,
			},
		},
		&Region{ // 不支持内网Endpoint
			Provider: "aliyun",
			Name:     "非中国内地加速Endpoint",
			Region:   "oss-accelerate-overseas",
			Endpoint: "oss-accelerate-overseas.aliyuncs.com",
			Properties: Properties{
				SupportAccelerate: true,
			},
		})

	return ls
}

func AliyunOSSInternalUrl(regionId string) string {
	return fmt.Sprintf("%s-internal.aliyuncs.com", regionId)
}

func TencentCOS() []*Region {
	res, err := http.Get("https://cloud.tencent.com/document/product/436/6224")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	ls := make([]*Region, 0)
	doc.Find("table tr").Each(func(row int, tr *goquery.Selection) {
		tmp := &Region{}

		count := 0
		tr.Find("td").Each(func(col int, td *goquery.Selection) {
			_, hasRowspan := td.Attr("rowspan")
			if hasRowspan {
				count++
				return
			}

			tmp.Provider = "tencentcloud"
			//fmt.Println(row, col-count, td.Text())
			switch col - count {
			case 0:
				tmp.Name = strings.TrimSpace(td.Text())
			case 1:
				tmp.Region = strings.TrimSpace(td.Text())
			case 2:
				tmp.Endpoint = strings.TrimSpace(td.Text())
			}
		})

		if tmp.Name == "" {
			return
		}

		if strings.HasSuffix(tmp.Region, "-fsi") { // 排除金融云
			return
		}

		tmp.Endpoint = strings.TrimPrefix(tmp.Endpoint, "<BucketName-APPID>.")
		tmp.Properties.NeedAppID = true

		ls = append(ls, tmp)
	})

	ls = append(ls,
		&Region{ // 不支持内网Endpoint
			Provider: "tencentcloud",
			Name:     "全球加速域名",
			Region:   "accelerate",
			Endpoint: "cos.accelerate.myqcloud.com",
			Properties: Properties{
				SupportAccelerate: true,
				NeedAppID:         true,
			},
		})

	return ls
}

func TencentCOSUrl(regionId, bucketName, appID string) string {
	return fmt.Sprintf("%s-%s.cos.%s.myqcloud.com", bucketName, appID, regionId)
}

func HuaweiOBS() []*Region {
	res, err := http.Get("https://developer.huaweicloud.com/intl/zh-cn/endpoint?OBS")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	ls := make([]*Region, 0)
	doc.Find("div.name-team").Each(func(row int, s *goquery.Selection) {
		sn := s.Find("div.service-name").First()
		if strings.TrimSpace(sn.Text()) != "对象存储服务 OBS" {
			return
		}

		s.Find("table tr").Each(func(row int, tr *goquery.Selection) {
			tmp := &Region{}
			tr.Find("td").Each(func(col int, td *goquery.Selection) {
				tmp.Provider = "huaweicloud"
				//fmt.Println(row, col, td.Text())
				switch col {
				case 0:
					tmp.Name = strings.TrimSpace(td.Text())
				case 1:
					tmp.Region = strings.TrimSpace(td.Text())
				case 2:
					tmp.Endpoint = strings.TrimSpace(td.Text())
				}
			})

			if tmp.Name == "" { // exclude table header
				return
			}

			ls = append(ls, tmp)
		})
	})

	return ls
}
