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
)

type fileRepository struct {
	filepath string
}

func NewDBStats(dbFile string) *fileRepository {
	if dbFile == "" {
		dbFile = fmt.Sprintf("%s/.gitlab-stats/db.json", os.Getenv("HOME"))

		dirDB := filepath.Dir(dbFile)
		if _, err := os.Stat(dirDB); os.IsNotExist(err) {
			os.Mkdir(dirDB, 0750)
		}
	}
	return &fileRepository{
		filepath: dbFile,
	}
}

func (s *fileRepository) AddStats(r DatabaseBFileRecord) error {
	db, err := os.ReadFile(s.filepath)
	if err != nil {
		return err
	}

	var records DatabaseBFile
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

	os.Remove(s.filepath)

	f, err := os.OpenFile(s.filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(newDB)
	return err
}

func (s *fileRepository) GetAllRecordsFromDB() (records DatabaseBFile, err error) {
	db, err := os.ReadFile(s.filepath)
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

func (s *fileRepository) GetLastRecordBetween(begin, end time.Time) (r DatabaseBFileRecord, err error) {
	var zeroValue DatabaseBFileRecord
	db, err := s.GetAllRecordsFromDB()
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
		if db.Records[i].DateExec == begin || db.Records[i].DateExec == end {
			r = db.Records[i]
		}
		// fmt.Println(db.Records[i].Counts.Opened)
	}
	if zeroValue == r {
		return r, errors.New("no value")
	}
	return r, nil
}

func (s *fileRepository) GetLastMonthStatsFromFile(since int) (res []DatabaseBFileRecord, err error) {
	db, err := os.ReadFile(s.filepath)
	if err != nil {
		return
	}

	var records DatabaseBFile
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

		lastRecordInTheMont, err := s.GetLastRecordBetween(dBegin.Time(), dEnd.Time())
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
