package gitlabstatistics

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/sgaunet/calcdate/calcdatelib"
	"github.com/sgaunet/gitlab-stats/gitlabRequest"
)

func init() {
	DBFile := DBFileName()
	dirDB := filepath.Dir(DBFile)
	if _, err := os.Stat(dirDB); os.IsNotExist(err) {
		os.Mkdir(dirDB, 0750)
	}
}

type RequestStatistics struct {
	uri               string
	fieldFilterAfter  string
	valueFilterAfter  time.Time
	fieldFilterBefore string
	valueFilterBefore time.Time
}

func NewRequestStatistics() *RequestStatistics {
	// list issues of a project : /projects/:id/issues
	// list issues of a group   : /groups/:id/issues
	r := RequestStatistics{
		uri: "",
	}
	return &r
}

func (r *RequestStatistics) SetProjectId(projectId int) {
	r.uri = fmt.Sprintf("projects/%d/issues_statistics?", projectId)
}

func (r *RequestStatistics) SetGroupId(groupId int) {
	r.uri = fmt.Sprintf("groups/%d/issues_statistics?", groupId)
}

func (r *RequestStatistics) SetFilterAfter(field string, d time.Time) {
	r.fieldFilterAfter = field
	r.valueFilterAfter = d
}

func (r *RequestStatistics) SetFilterBefore(field string, d time.Time) {
	r.fieldFilterBefore = field
	r.valueFilterBefore = d
}

func (r *RequestStatistics) Url() (url string) {
	url = r.uri
	if r.fieldFilterAfter != "" {
		url = fmt.Sprintf("%s&%s=%s", url, r.fieldFilterAfter, r.valueFilterAfter.Format(time.RFC3339))
	}
	if r.fieldFilterBefore != "" {
		url = fmt.Sprintf("%s&%s=%s", url, r.fieldFilterBefore, r.valueFilterBefore.Format(time.RFC3339))
	}
	return url
	//rqt = fmt.Sprintf("issues?state=%s&%s=%s&%s=%s&page=1", state, fieldFilterAfter, dBegin.Format(time.RFC3339), fieldFilterBefore, dEnd.Format(time.RFC3339))
}

func (r *RequestStatistics) GetStatistics() (result Statistics, err error) {
	if r.uri == "" {
		return result, errors.New("no project or group specified")
	}
	_, body, _ := gitlabRequest.Request(r.Url())
	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}
	return result, nil
}

func AppendStatsToFile(r Record) error {
	db, err := os.ReadFile(DBFileName())
	if err != nil {
		return err
	}

	f, err := os.OpenFile(DBFileName(), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	var records structDBFile
	if len(db) != 0 {
		err = json.Unmarshal(db, &records)
		if err != nil {
			return err
		}
	}

	records.Records = append(records.Records, r)

	newDB, err := json.Marshal(records)
	if err != nil {
		return err
	}
	_, err = f.Write(newDB)
	return err
}

func DBFileName() string {
	return fmt.Sprintf("%s/.gitlab-stats/db.json", os.Getenv("HOME"))
}

func GetAllRecordsFromDB() (records structDBFile, err error) {
	db, err := os.ReadFile(DBFileName())
	if err != nil {
		return
	}

	if len(db) != 0 {
		err = json.Unmarshal(db, &records)
		if err != nil {
			return
		}
	}

	// sort result by date
	sort.Slice(records.Records, func(i, j int) bool {
		return records.Records[i].DateExec.After(records.Records[i].DateExec)
	})
	return
}

func GetLastRecordBetween(begin, end time.Time) (r Record, err error) {
	var zeroValue Record
	db, err := GetAllRecordsFromDB()
	if err != nil {
		return
	}

	if len(db.Records) == 0 {
		return r, errors.New("no value in DB")
	}
	// records are sorted in GetAllRecordsFromDB
	for i := range db.Records {
		if db.Records[i].DateExec.After(begin) && db.Records[i].DateExec.Before(end) {
			r = db.Records[i]
		}
	}
	if zeroValue == r {
		return r, errors.New("no value")
	}
	return r, nil
}

func GetLastMonthStatsFromFile(since int) (res []Record, err error) {
	db, err := os.ReadFile(DBFileName())
	if err != nil {
		return
	}

	var records structDBFile
	if len(db) != 0 {
		err = json.Unmarshal(db, &records)
		if err != nil {
			return
		}
	} else {
		return res, errors.New("no stats")
	}

	tz := ""
	startOfNowMonth, _ := calcdatelib.NewDate(time.Now().Format("2006/01/02 15:04:05"), "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
	startOfNowMonth.SetBeginMonth()

	// fmt.Println("startOfNowMonth=", startOfNowMonth)

	begin, _ := calcdatelib.NewDate(time.Now().Format("2006/01/02 15:04:05"), "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
	begin.SetBeginMonth()
	begin.AddMonth(-since)
	// fmt.Println("begin=", begin)

	dateLoop := begin
	for i := 0; i < since; i++ {
		dBegin, err := calcdatelib.NewDate(dateLoop.String(), "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
		if err != nil {
			return res, err
		}
		dBegin.SetBeginDate()
		dEnd, err := calcdatelib.NewDate(dateLoop.String(), "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
		if err != nil {
			return res, err
		}
		dEnd.SetEndMonth()
		// fmt.Println(dBegin)
		// fmt.Println(dEnd)

		lastRecordInTheMont, err := GetLastRecordBetween(dBegin.Time(), dEnd.Time())
		if err != nil {
			fmt.Println("no value")
		} else {
			// fmt.Println(lastRecordInTheMont)
			res = append(res, lastRecordInTheMont)
		}
		dateLoop.AddMonth(1)
	}
	return
}
