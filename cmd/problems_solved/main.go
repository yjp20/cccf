package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/peterbourgon/ff"
	"github.com/yjp20/cccf/pkg"
	"google.golang.org/api/sheets/v4"
)

func main() {
	fs := flag.NewFlagSet("leaderboard", flag.ExitOnError)
	var (
		_           = fs.String("c", "config", "config location")
		sheetID     = fs.String("sheetid", "", "Spreadsheet ID")
		sheetRange  = fs.String("psrange", "", "ex) Sheet!A1:A4")
		memberRange = fs.String("memberrange", "", "ex) Sheet!A1:A4")
	)
	ff.Parse(fs, os.Args[1:],
		ff.WithIgnoreUndefined(true),
		ff.WithConfigFileFlag("c"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix("CCCF"),
	)

	println("Get members from google sheets")
	ss := pkg.MustService(pkg.GetSheetsService())
	md, err := pkg.GetMemberData(ss, *sheetID, *memberRange)
	if err != nil {
		log.Fatal(err)
	}
	problems, err := pkg.GetProblems(ss, *sheetID, *sheetRange)
	problems = problems[2:]
	if err != nil {
		log.Fatal(err)
	}
	idxMap := make(map[string]int)
	memberSolved := make([]map[string]string, len(md))

	for idx, problem := range problems {
		idxMap[problem] = idx
	}

	for idx := range md {
		memberSolved[idx] = make(map[string]string)
	}

	println("Get Codeforces member submissions")
	for idx, member := range md {
		params := url.Values{}
		params.Add("handle", member.Handle)
		submissionList := []pkg.CFSubmission{}
		err = pkg.GetCF("https://codeforces.com/api/user.status", &submissionList, params)
		if err != nil {
			log.Fatal(err)
		}
		for _, submission := range submissionList {
			if submission.Verdict == "OK" {
				problemString := strconv.Itoa(submission.Problem.ContestId) + submission.Problem.Index
				memberSolved[idx][problemString] = "1"
			}
		}
	}

	println("Write to sheets: " + *sheetRange)
	output := make([][]interface{}, len(md)+2)
	output[0] = make([]interface{}, 0)
	output[1] = make([]interface{}, 0)
	for idx := range md {
		output[idx+2] = make([]interface{}, len(problems)+2)
		for jdx, problem := range problems {
			output[idx+2][jdx+2] = memberSolved[idx][problem]
		}
	}
	vr := sheets.ValueRange{Values: output}
	err = pkg.SetRange(ss, *sheetID, *sheetRange, &vr)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
