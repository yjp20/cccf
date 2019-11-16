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
	var (
		_           = fs.String("c", "config", "config location")
		sheetID     = fs.String("sheetid", "", "spreadsheet ID")
		sheetRange  = fs.String("ctrange", "", "ex) Sheet!A1:Z10")
		memberRange = fs.String("memberrange", "", "ex) Sheet!A1:Z10")
	)
	err := ff.Parse(fs, os.Args[1:],
		ff.WithIgnoreUndefined(true),
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
	err = cache.readCache(ss, *sheetID, *memberRange)
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
		cache.ContestList[contestId] = true
		i++
	}

	println("Write to google sheets")
	i = 1
	// Init all cells, required to overwrite cells that are outdated
	output := make([][]interface{}, len(mm)+1)
	for i := 0; i < len(mm)+1; i++ {
		output[i] = make([]interface{}, len(contestmap)+3)
		for j := 0; j < len(contestmap)+3; j++ {
			output[i][j] = " "
		}
	}

	// Write the first row, which include column headers such as the
	// Name, Handle, Counter (total num of contests participated in)
	// and the index for each contest.
	output[0][0] = "Name"
	output[0][1] = "Handle"
	output[0][2] = "Counter"
	for key, index := range contestmap {
		output[0][index] = strconv.Itoa(key)
	}
	mslice := pkg.MapToSlice(mm)
	sort.Slice(mslice, func(i, j int) bool {
		return mslice[i].Index < mslice[j].Index
	})

	// Fill out information for each member, where it matches up the contest
	// participated by the member to the right column.
	for _, member := range mslice {
		idx := member.Index + 1 + 1
		output[i][0] = member.Name
		output[i][1] = member.Handle
		output[i][2] = fmt.Sprintf("=COUNT(D%d:ZZ%d)", idx, idx)
		for id, score := range member.Contests {
			output[i][contestmap[id]] = score
		}
		i++
	}

	vr := sheets.ValueRange{Values: output}
	err = pkg.SetRange(ss, *sheetID, *sheetRange, &vr)
	if err != nil {
		log.Fatal(err)
	}
	err = cache.writeCache()
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Cache) readCache(ss *sheets.Service, sheetID, memberRange string) error {
	file, err := ioutil.ReadFile("contest_cache.json")
	if err == nil {
		err = json.Unmarshal(file, &c)
		if err != nil {
			return err
		}
	} else {
		md, err := pkg.GetMemberData(ss, sheetID, memberRange)
		if err != nil {
			return err
		}
		c.ContestList = make(map[int]bool)
		c.MemberList = md
	}
	return nil
}

func (c *Cache) writeCache() error {
	file, _ := json.MarshalIndent(c, "", " ")
	return ioutil.WriteFile("contest_cache.json", file, 0644)
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
