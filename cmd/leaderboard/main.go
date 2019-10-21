package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"sort"
	"strconv"

	"github.com/peterbourgon/ff"
	"github.com/yjp20/cccf/pkg"
	"google.golang.org/api/sheets/v4"
)

func main() {
	fs := flag.NewFlagSet("leaderboard", flag.ExitOnError)
	fs.String("c", "config", "config location")
	sheetID := fs.String("sheetid", "", "Spreadsheet ID")
	sheetrange := fs.String("lbrange", "", "ex) Sheet!A1:A4")
	ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("c"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix("CCCF"),
	)

	println("Get members from google sheets")
	ss := pkg.MustService(pkg.GetSheetsService())
	md, err := pkg.GetMemberData(ss, *sheetID)
	if err != nil {
		log.Fatal(err)
	}
	mm := pkg.SliceToMap(md)
	hc := pkg.GetHandleConcat(md)

	println("Get Codeforces member data")
	params := url.Values{}
	params.Add("handles", hc)
	userList := []pkg.CFUser{}
	err = pkg.GetCF("https://codeforces.com/api/user.info", &userList, params)
	if err != nil {
		log.Fatal(err)
	}

	for _, user := range userList {
		member := mm[user.Handle]
		member.Rating = user.Rating
		if len(user.Rank) == 0 {
			member.Rank = "Trash"
		} else {
			member.Rank = user.Rank
		}
		mm[user.Handle] = member
	}

	println("Write to sheets: " + *sheetrange)
	mslice := pkg.MapToSlice(mm)
	minterface := make([][]interface{}, len(mm))
	sort.Slice(mslice, func(i, j int) bool {
		return mslice[i].Rating > mslice[j].Rating
	})
	for index, m := range mslice {
		minterface[index] = []interface{}{m.Name, m.Handle, strconv.Itoa(m.Rating), m.Rank}
	}
	vr := sheets.ValueRange{Values: minterface}
	err = pkg.SetRange(ss, *sheetID, *sheetrange, &vr)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
