package gitlabstatistics

func FilterWithProject(r []Record, projectID int) (res []Record) {
	for idx := range r {
		if r[idx].ProjectID == projectID {
			res = append(res, r[idx])
		}
	}
	return
}

func FilterWithGroup(r []Record, groupID int) (res []Record) {
	for idx := range r {
		if r[idx].GroupID == groupID {
			res = append(res, r[idx])
		}
	}
	return
}
