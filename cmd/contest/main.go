package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/peterbourgon/ff"
	"github.com/yjp20/cccf/pkg"
	"google.golang.org/api/sheets/v4"
)

type Cache struct {
	MemberList  []pkg.MemberData
	ContestList map[int]bool
}

func main() {
	fs := flag.NewFlagSet("leaderboard", flag.ExitOnError)
	fs.String("c", "config", "config location")
	sheetID := fs.String("sheetid", "", "")
	sheetrange := fs.String("ctrange", "", "ex) Sheet!A1:Z10")
	_ = fs.String("lbrange", "", "")
	err := ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("c"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix("CCCF"),
	)
	if err != nil {
		log.Fatal(err)
	}

	println("Get members from google sheets")
	startTime := time.Date(2019, time.September, 0, 0, 0, 0, 0, time.UTC)
	ss := pkg.MustService(pkg.GetSheetsService())
	cache := Cache{}
	err = cache.readCache(ss, *sheetID)
	if err != nil {
		log.Fatal(err)
	}
	mm := pkg.SliceToMap(cache.MemberList)
	hc := pkg.GetHandleConcat(cache.MemberList)

	println("Get contest data from Codeforces")
	contestlist := []pkg.CFContest{}
	contestmap := make(map[int]int)
	err = pkg.GetCF("https://codeforces.com/api/contest.list", &contestlist, url.Values{})
	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(contestlist, func(i, j int) bool {
		return contestlist[i].StartTimeSeconds < contestlist[j].StartTimeSeconds
	})

	println("Get contest participation from Codeforces")
	i := 3
	for _, contest := range contestlist {
		contestId := contest.Id
		contestTime := contest.StartTimeSeconds
		beforeStart := contestTime < startTime.Unix()
		afterNow := contestTime > time.Now().Unix()
		_, alreadyCached := cache.ContestList[contestId]
		if beforeStart || alreadyCached || afterNow {
			continue
		}
		appendContestToMap(contest, cache, hc, mm)
		contestmap[contestId] = i
		i++
	}

	println("Write to google sheets")
	i = 1
	m := make([][]interface{}, len(mm)+1)
	for i := 0; i < len(mm)+1; i++ {
		m[i] = make([]interface{}, len(contestmap)+3)
		for j := 0; j < len(contestmap)+3; j++ {
			m[i][j] = " "
		}
	}
	m[0][0] = "Name"
	m[0][1] = "Handle"
	m[0][2] = "Counter"
	for key, index := range contestmap {
		m[0][index] = strconv.Itoa(key)
	}
	mslice := pkg.MapToSlice(mm)
	sort.Slice(mslice, func(i, j int) bool {
		return mslice[i].Index < mslice[j].Index
	})
	for _, member := range mslice {
		idx := member.Index + 1 + 1
		m[i][0] = member.Name
		m[i][1] = member.Handle
		m[i][2] = fmt.Sprintf("=COUNT(D%d:ZZ%d)", idx, idx)
		for id, score := range member.Contests {
			m[i][contestmap[id]] = score
		}
		i++
	}

	vr := sheets.ValueRange{}
	vr.Values = m
	err = pkg.SetRange(ss, *sheetID, *sheetrange, &vr)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Cache) readCache(ss *sheets.Service, sheetID string) error {
	file, err := ioutil.ReadFile("members.json")
	if err == nil {
		err = json.Unmarshal(file, &c)
		if err != nil {
			return err
		}
	} else {
		md, err := pkg.GetMemberData(ss, sheetID)
		if err != nil {
			return err
		}
		c.MemberList = md
	}
	return nil
}

func appendContestToMap(
	contest pkg.CFContest,
	cache Cache,
	hc string,
	mm map[string]pkg.MemberData,
) {
	contestId := contest.Id
	params := url.Values{}
	params.Add("contestId", strconv.Itoa(contestId))
	params.Add("handles", hc)
	ret := pkg.CFReturner{}
	err := pkg.GetCF("https://codeforces.com/api/contest.standings", &ret, params)
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range ret.Rows {
		handle := row.Party.Members[0].Handle
		member := mm[handle]
		if member.Contests == nil {
			member.Contests = make(map[int]float64)
		}
		member.Contests[contestId] = row.Points
		mm[handle] = member
	}
}
