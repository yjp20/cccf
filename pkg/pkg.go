package pkg

import (
	"context"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

type MemberData struct {
	Index    int
	Name     string
	Handle   string
	Rating   int
	Rank     string
	Contests map[int]float64
}

func GetHandleConcat(md []MemberData) string {
	handleConcat := strings.Builder{}
	for _, member := range md {
		handleConcat.WriteString(member.Handle)
		handleConcat.WriteString(";")
	}
	return handleConcat.String()
}

func GetSheetsService() (*sheets.Service, error) {
	c, err := google.DefaultClient(context.Background(), "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return nil, err
	}
	sheetsService, err := sheets.New(c)
	return sheetsService, nil
}

func GetMemberData(ss *sheets.Service, sheetId string, memberRange string) ([]MemberData, error) {
	resp, err := ss.Spreadsheets.Values.Get(sheetId, memberRange).Do()
	if err != nil {
		return nil, err
	}
	md := make([]MemberData, len(resp.Values))
	for idx, row := range resp.Values {
		member := MemberData{}
		member.Index = idx
		member.Name = row[0].(string)
		if len(row) > 2 {
			member.Handle = row[2].(string)
		}
		md[idx] = member
	}
	return md, nil
}

func GetProblems(ss *sheets.Service, sheetId, problemRange string) ([]string, error) {
	resp, err := ss.Spreadsheets.Values.Get(sheetId, problemRange).Do()
	if err != nil {
		return nil, err
	}
	problems := make([]string, len(resp.Values[0]))
	for idx, col := range resp.Values[0] {
		problems[idx] = col.(string)
	}
	return problems, nil
}

func SetRange(ss *sheets.Service, sheetId, r string, data *sheets.ValueRange) error {
	_, err := ss.Spreadsheets.Values.Update(sheetId, r, data).ValueInputOption("USER_ENTERED").Do()
	return err
}

func SliceToMap(md []MemberData) map[string]MemberData {
	m := make(map[string]MemberData)
	for _, member := range md {
		m[member.Handle] = member
	}
	return m
}

func MapToSlice(mm map[string]MemberData) []MemberData {
	mslice := make([]MemberData, len(mm))
	i := 0
	for _, m := range mm {
		mslice[i] = m
		i++
	}
	return mslice
}
