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

func GetMemberData(ss *sheets.Service, sid string) ([]MemberData, error) {
	sheetRange := "C2:E100" // Assume there's never going to be more than 100 members (there better not be)
	resp, err := ss.Spreadsheets.Values.Get(sid, sheetRange).Do()
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

func SetRange(ss *sheets.Service, sid, r string, data *sheets.ValueRange) error {
	_, err := ss.Spreadsheets.Values.Update(sid, r, data).ValueInputOption("USER_ENTERED").Do()
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
