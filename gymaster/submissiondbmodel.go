package main

type SubmissionDbModel struct {
	db DbContext
}

func NewSubmissionDbModel() (SubmissionDbModel, error) {
	sdm := SubmissionDbModel{}
	db, err := OpenDatabase()
	if err != nil {
		return sdm, err
	}
	sdm.db = db
	return sdm, err
}

func (sdm *SubmissionDbModel) Close() error {
	return sdm.db.Close()
}

func (sdm *SubmissionDbModel) GetSubmissionList(userId int, problemId int) ([]SubmissionInfo, error) {
	db := sdm.db
	query := `SELECT s.id, s.id_problem, s.id_user, s.id_cache, s.lang, s.code, s.status, s.details, s.score, s.submit_time, s.compile_time, 
        s.compile_stdout, s.compile_stderr, u.display_name, p.problem_name, c.title 
        FROM ((({{.TablePrefix}}submissions AS s INNER JOIN {{.TablePrefix}}users AS u ON s.id_user = u.id)
        INNER JOIN {{.TablePrefix}}problems AS p ON s.id_problem = p.id) 
        INNER JOIN {{.TablePrefix}}contests AS c ON p.contest_id = c.id)
        WHERE ((id_user = ?) OR (? = 0)) AND ((id_problem = ?) OR (? = 0)) ORDER BY id DESC`
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(userId, userId, problemId, problemId)
	if err != nil {
		return nil, err
	}
	var subs []SubmissionInfo
	for rows.Next() {
		sb := SubmissionInfo{}
		err = rows.Scan(
			&sb.Id,
			&sb.ProblemId,
			&sb.UserId,
			&sb.CacheId,
			&sb.Language,
			&sb.Code,
			&sb.Status,
			&sb.Details,
			&sb.Score,
			&sb.SubmitTime,
			&sb.CompileTime,
			&sb.CompileStdout,
			&sb.CompileStderr,
			&sb.UserDisplayName,
			&sb.ProblemName,
			&sb.ContestName,
		)
		subs = append(subs, sb)
	}
	return subs, nil
}
