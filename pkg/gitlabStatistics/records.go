package gitlabstatistics

func FilterWithProject(r []DatabaseBFileRecord, projectID int) (res []DatabaseBFileRecord) {
	for idx := range r {
		if r[idx].ProjectID == projectID {
			res = append(res, r[idx])
		}
	}
	return
}

func FilterWithGroup(r []DatabaseBFileRecord, groupID int) (res []DatabaseBFileRecord) {
	for idx := range r {
		if r[idx].GroupID == groupID {
			res = append(res, r[idx])
		}
	}
	return
}
