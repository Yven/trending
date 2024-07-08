package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

type ProjectList struct {
	Since    Since
	Language Language
	List     []GithubProject
}

type TechStackInfoList struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Icon     string `json:"icon"`
	Homepage string `json:"homepage"`
}

type TechStackList map[string][]TechStackInfoList

type StdTechStackInfo struct {
	OriginUrl string
	IconUrl   string
	Name      string
}

type StdTechStackList struct {
	Key  string
	Info []StdTechStackInfo
}

type GithubProject struct {
	Link          string
	Name          string
	StarNums      string
	Desc          string
	StarTrending  string
	ForkNums      string
	TechStackList []StdTechStackList
}

type Since struct {
	Name string
	Tag  string
}

type Language struct {
	Language string
}

type Template struct {
	SinceList    []Since
	LanguageList []Language
	ProjectList  []ProjectList
}

func main() {
	args := os.Args
	var saveFile, tempFile string
	if len(args) > 1 {
		for i := 1; i < len(args)-1; i += 2 {
			if args[i] == "-o" {
				saveFile = args[i+1]
			} else if args[i] == "-i" {
				tempFile = args[i+1]
			}
		}
	} else {
		if len(saveFile) == 0 {
			saveFile = "today.html"
		}
		if len(tempFile) == 0 {
			tempFile = "template.html"
		}
	}

	sinceList := []Since{
		{"今日", "today"},
		// {"本周", "weekly"},
	}

	languageList := []Language{
		{"Go"},
		// {"PHP"},
	}

	projectList := make([]ProjectList, 0, len(sinceList)*len(languageList))
	for _, since := range sinceList {
		for _, lang := range languageList {
			url := "https://github.com/trending/%s?since=%s"
			list := scrape(fmt.Sprintf(url, lang.Language, since.Tag))
			fmt.Println("GET", since.Name, lang.Language)
			if len(list) == 0 {
				log.Fatalln("爬取到的页面数据结构为空")
			}

			projectList = append(projectList, ProjectList{since, lang, list})
		}
	}

	t, err := template.ParseFiles(tempFile)
	if err != nil {
		log.Fatal("Parse error:", err)
	}

	f, err := os.OpenFile(saveFile, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = f.Truncate(0)
	if err != nil {
		log.Fatal("清空文件内容失败:", err)
	}

	err = t.Execute(f, Template{sinceList, languageList, projectList})
	if err != nil {
		log.Fatal("Execute error:", err)
	}
}

func scrape(url string) []GithubProject {
	list := make([]GithubProject, 0, 18)
	c := colly.NewCollector(colly.MaxDepth(1))
	c.OnHTML(".Box .Box-row", func(e *colly.HTMLElement) {
		project := GithubProject{}
		// author & repository name
		authorRepoName := e.ChildText("h2.h3 > a")
		parts := strings.Split(authorRepoName, "/")
		project.Name = strings.TrimSpace(parts[0]) + "/" + strings.TrimSpace(strings.Trim(parts[1], "\n"))
		// link
		project.Link = "https://github.com" + e.ChildAttr("h2.h3 > a", "href")

		resp, err := http.Get("https://techstack.zeabur.app/repo?url=" + project.Link)
		if err == nil {
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			techList := make(TechStackList)
			json.Unmarshal(body, &techList)
			stdList := make([]StdTechStackList, 0, len(techList))
			for k, infoList := range techList {
				tmp := StdTechStackList{
					Key: k,
				}
				for _, v := range infoList {
					tmp.Info = append(tmp.Info, StdTechStackInfo{
						IconUrl:   v.Icon,
						Name:      v.Name,
						OriginUrl: v.Homepage,
					})
				}
				stdList = append(stdList, tmp)
			}
			project.TechStackList = stdList
		}

		// description
		project.Desc = e.ChildText("p.pr-4")

		// language
		// repo.Lang = strings.TrimSpace(e.ChildText("div.mt-2 > span.mr-3 > span[itemprop]"))

		// star & fork
		starForkStr := e.ChildText("div.mt-2 > a.mr-3")
		starForkStr = strings.Replace(strings.TrimSpace(starForkStr), ",", "", -1)
		parts = strings.Split(starForkStr, "\n")
		project.StarNums = strings.TrimSpace(parts[0])
		project.ForkNums = strings.TrimSpace(parts[len(parts)-1])

		// stars trending
		starTrendingStr := e.ChildText("div.mt-2 > span.float-sm-right")
		project.StarTrending = starTrendingStr

		list = append(list, project)
	})

	c.Visit(url)
	return list
}
