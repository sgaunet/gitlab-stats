package jsonfile

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

type FileRepository struct {
	filepath string
}

func NewDBStats(dbFile string) *FileRepository {
	if dbFile == "" {
		dbFile = fmt.Sprintf("%s/.gitlab-stats/db.json", os.Getenv("HOME"))

		dirDB := filepath.Dir(dbFile)
		if _, err := os.Stat(dirDB); os.IsNotExist(err) {
			os.Mkdir(dirDB, 0750)
		}
	}
	return &FileRepository{
		filepath: dbFile,
	}
}

func (s *FileRepository) AddStats(projectID, groupID, all, closed, opened int, dateExec time.Time) error {
	var record databaseBFileRecord
	record.GroupID = groupID
	record.ProjectID = projectID
	record.DateExec = dateExec
	record.Counts.All = all
	record.Counts.Closed = closed
	record.Counts.Opened = opened

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

	records.Records = append(records.Records, record)
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

func (s *FileRepository) GetAllRecordsFromDB() (records DatabaseBFile, err error) {
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

func (s *FileRepository) getLastRecordOfProjectBetween(projectID int, begin, end time.Time) (r databaseBFileRecord, err error) {
	var zeroValue databaseBFileRecord
	db, err := s.GetAllRecordsFromDB()
	if err != nil {
		return
	}

	if len(db.Records) == 0 {
		return r, errors.New("no value in DB")
	}
	// records are sorted in GetAllRecordsFromDB
	for i := range db.Records {
		if db.Records[i].ProjectID == projectID {
			if db.Records[i].DateExec.After(begin) && db.Records[i].DateExec.Before(end) {
				r = db.Records[i]
			}
			if db.Records[i].DateExec == begin || db.Records[i].DateExec == end {
				r = db.Records[i]
			}
		}
	}
	if zeroValue == r {
		return r, errors.New("no value")
	}
	return r, nil
}

func (s *FileRepository) getLastRecordOfGroupBetween(groupID int, begin, end time.Time) (r databaseBFileRecord, err error) {
	var zeroValue databaseBFileRecord
	db, err := s.GetAllRecordsFromDB()
	if err != nil {
		return
	}

	if len(db.Records) == 0 {
		return r, errors.New("no value in DB")
	}
	// records are sorted in GetAllRecordsFromDB
	for i := range db.Records {
		if db.Records[i].GroupID == groupID {
			if db.Records[i].DateExec.After(begin) && db.Records[i].DateExec.Before(end) {
				r = db.Records[i]
			}
			if db.Records[i].DateExec == begin || db.Records[i].DateExec == end {
				r = db.Records[i]
			}
		}
	}
	if zeroValue == r {
		return r, errors.New("no value")
	}
	return r, nil
}

func (s *FileRepository) GetStatsByProjectId(projectID int, since int) (openedSerie []float64, closedSerie []float64, dateExecSerie []time.Time, err error) {
	tz := ""
	startOfNowMonth, _ := calcdatelib.NewDate(time.Now().Format("2006/01/02 15:04:05"), "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
	startOfNowMonth.SetBeginMonth()
	begin, _ := calcdatelib.NewDate(time.Now().Format("2006/01/02 15:04:05"), "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
	begin.SetBeginMonth()
	begin.AddMonth(-since)
	dateLoop := begin
	for i := 0; i < since; i++ {
		dBegin, _ := calcdatelib.NewDate(dateLoop.String(), "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
		dBegin.SetBeginDate()
		dEnd, _ := calcdatelib.NewDate(dateLoop.String(), "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
		dEnd.SetEndMonth()

		lastRecordInTheMont, err := s.getLastRecordOfProjectBetween(projectID, dBegin.Time(), dEnd.Time())
		if err != nil {
			fmt.Println("no value")
		} else {
			// fmt.Println(lastRecordInTheMont)
			openedSerie = append(openedSerie, float64(lastRecordInTheMont.Counts.Opened))
			closedSerie = append(closedSerie, float64(lastRecordInTheMont.Counts.Closed))
			dateExecSerie = append(dateExecSerie, lastRecordInTheMont.DateExec)
		}
		dateLoop.AddMonth(1)
	}
	return
}

func (s *FileRepository) GetStatsByGroupId(groupID int, since int) (openedSerie []float64, closedSerie []float64, dateExecSerie []time.Time, err error) {
	tz := ""
	startOfNowMonth, _ := calcdatelib.NewDate(time.Now().Format("2006/01/02 15:04:05"), "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
	startOfNowMonth.SetBeginMonth()
	begin, _ := calcdatelib.NewDate(time.Now().Format("2006/01/02 15:04:05"), "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
	begin.SetBeginMonth()
	begin.AddMonth(-since)
	dateLoop := begin
	for i := 0; i < since; i++ {
		dBegin, _ := calcdatelib.NewDate(dateLoop.String(), "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
		dBegin.SetBeginDate()
		dEnd, _ := calcdatelib.NewDate(dateLoop.String(), "%YYYY/%MM/%DD %hh:%mm:%ss", tz)
		dEnd.SetEndMonth()
		lastRecordInTheMont, err := s.getLastRecordOfGroupBetween(groupID, dBegin.Time(), dEnd.Time())
		if err != nil {
			fmt.Println("no value")
		} else {
			openedSerie = append(openedSerie, float64(lastRecordInTheMont.Counts.Opened))
			closedSerie = append(closedSerie, float64(lastRecordInTheMont.Counts.Closed))
			dateExecSerie = append(dateExecSerie, lastRecordInTheMont.DateExec)
		}
		dateLoop.AddMonth(1)
	}
	return
}
