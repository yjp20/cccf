package pkg

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

type CFReturner struct {
	Contest CFContest
	Rows    []CFRanklistRow
}

type CFRanklistRow struct {
	Party struct {
		Members []struct {
			Handle string
		}

		ContestId        int
		ParticipantType  string
		Ghost            bool
		Room             int
		StartTimeSeconds int64
	}

	ProblemResults []struct {
		Points                    float64
		RejectedAttemptCount      int
		Type                      string
		BestSubmissionTimeSeconds int
	}

	Rank                  int
	Points                float64
	Penalty               int
	SuccessfulHackCount   int
	UnsuccessfulHackCount int
}

type CFContest struct {
	Id                  int
	Name                string
	Type                string
	Phase               string
	Frozen              bool
	DurationSeconds     int64
	StartTimeSeconds    int64
	RelativeTimeSeconds int64
}

type CFUser struct {
	Handle string
	Rating int
	Rank   string
}

type CFSubmission struct {
	Id        int
	ContestId int
	Problem   CFProblem
	Verdict   string
}

type CFProblem struct {
	ContestId   int
	ProblemName string
	Index       string
	Name        string
	Rating      int
}

type CFVerdict int

const (
	CF_FAILED  = 0
	CF_OK      = 1
	CF_PARTIAL = 2
)

type wrapper struct {
	Status string
	Result interface{}
}

func GetCF(apiurl string, v interface{}, params url.Values) error {
	w := wrapper{Result: v}
	res, err := http.Get(apiurl + "?" + params.Encode())
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &w)
	if err != nil {
		return err
	}
	return nil
}
