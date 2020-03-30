package member

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful-member/config"
	"github.com/readr-media/readr-restful-member/internal/args"
	"github.com/readr-media/readr-restful-member/internal/rrsql"
)

// var MemberStatus map[string]interface{}

type Member struct {
	ID       int64            `json:"id" db:"id"`
	MemberID string           `json:"member_id" db:"member_id"`
	UUID     string           `json:"uuid" db:"uuid"`
	Points   rrsql.NullInt    `json:"points" db:"points"`
	Name     rrsql.NullString `json:"name" db:"name"`
	Nickname rrsql.NullString `json:"nickname" db:"nickname"`
	// Cannot parse Date format
	Birthday rrsql.NullTime   `json:"birthday" db:"birthday"`
	Gender   rrsql.NullString `json:"gender" db:"gender"`
	Work     rrsql.NullString `json:"work" db:"work"`
	Mail     rrsql.NullString `json:"mail" db:"mail"`
	Phone    rrsql.NullString `json:"phone" db:"phone"`

	RegisterMode rrsql.NullString `json:"register_mode" db:"register_mode"`
	SocialID     rrsql.NullString `json:"social_id,omitempty" db:"social_id"`
	TalkID       rrsql.NullString `json:"talk_id" db:"talk_id"`

	CreatedAt     rrsql.NullTime   `json:"created_at" db:"created_at"`
	UpdatedAt     rrsql.NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy     rrsql.NullInt    `json:"updated_by" db:"updated_by"`
	Password      rrsql.NullString `json:"-" db:"password"`
	Salt          rrsql.NullString `json:"-" db:"salt"`
	PremiumBefore rrsql.NullTime   `json:"premium_before" db:"premium_before"`
	// Ignore password JSON marshall for now

	Description  rrsql.NullString `json:"description" db:"description"`
	ProfileImage rrsql.NullString `json:"profile_image" db:"profile_image"`
	Identity     rrsql.NullString `json:"identity" db:"identity"`

	Role   rrsql.NullInt `json:"role" db:"role"`
	Active rrsql.NullInt `json:"active" db:"active"`

	CustomEditor rrsql.NullBool `json:"custom_editor" db:"custom_editor"`
	HideProfile  rrsql.NullBool `json:"hide_profile" db:"hide_profile"`
	ProfilePush  rrsql.NullBool `json:"profile_push" db:"profile_push"`
	PostPush     rrsql.NullBool `json:"post_push" db:"post_push"`
	DailyPush    rrsql.NullBool `json:"daily_push" db:"daily_push"`
	CommentPush  rrsql.NullBool `json:"comment_push" db:"comment_push"`
}

// Stunt could be regarded as an experimental, pre-transitional wrap of Member, which provide omitempty tag for json
// and Use *Null type instead of Null type to made omitempty work
// In this way we could control the fields returned by update SQL select fields
type Stunt struct {
	// Make ID, MemberID, UUID pointer to avoid situation we have to use IFNULL
	ID       *int64            `json:"id,omitempty" db:"id"`
	MemberID *string           `json:"member_id,omitempty" db:"member_id"`
	UUID     *string           `json:"uuid,omitempty" db:"uuid"`
	Points   *rrsql.NullInt    `json:"points,omitempty" db:"points"`
	Name     *rrsql.NullString `json:"name,omitempty" db:"name"`
	Nickname *rrsql.NullString `json:"nickname,omitempty" db:"nickname"`

	Birthday *rrsql.NullTime   `json:"birthday,omitempty" db:"birthday"`
	Gender   *rrsql.NullString `json:"gender,omitempty" db:"gender"`
	Work     *rrsql.NullString `json:"work,omitempty" db:"work"`
	Mail     *rrsql.NullString `json:"mail,omitempty" db:"mail"`
	Phone    *rrsql.NullString `json:"phone,omitempty" db:"phone"`

	RegisterMode *rrsql.NullString `json:"register_mode,omitempty" db:"register_mode"`
	SocialID     *rrsql.NullString `json:"social_id,omitempty,omitempty" db:"social_id"`
	TalkID       *rrsql.NullString `json:"talk_id,omitempty" db:"talk_id"`

	CreatedAt *rrsql.NullTime  `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt *rrsql.NullTime  `json:"updated_at,omitempty" db:"updated_at"`
	UpdatedBy *rrsql.NullInt   `json:"updated_by,omitempty" db:"updated_by"`
	Password  rrsql.NullString `json:"-" db:"password"`
	Salt      rrsql.NullString `json:"-" db:"salt"`

	Description  *rrsql.NullString `json:"description,omitempty" db:"description"`
	ProfileImage *rrsql.NullString `json:"profile_image,omitempty" db:"profile_image"`
	Identity     *rrsql.NullString `json:"identity,omitempty" db:"identity"`

	Role   *rrsql.NullInt `json:"role,omitempty" db:"role"`
	Active *rrsql.NullInt `json:"active,omitempty" db:"active"`

	CustomEditor *rrsql.NullBool `json:"custom_editor,omitempty" db:"custom_editor"`
	HideProfile  *rrsql.NullBool `json:"hide_profile,omitempty" db:"hide_profile"`
	ProfilePush  *rrsql.NullBool `json:"profile_push,omitempty" db:"profile_push"`
	PostPush     *rrsql.NullBool `json:"post_push,omitempty" db:"post_push"`
	DailyPush    *rrsql.NullBool `json:"daily_push,omitempty" db:"daily_push"`
	CommentPush  *rrsql.NullBool `json:"comment_push,omitempty" db:"comment_push"`
}

// Separate API and Member struct
type memberAPI struct{}

var MemberAPI MemberInterface = new(memberAPI)

type MemberInterface interface {
	DeleteMember(idType string, id string) error
	GetMember(req GetMemberArgs) (Member, error)
	GetMembers(req *GetMembersArgs) ([]Member, error)
	FilterMembers(args *FilterMemberArgs) ([]Stunt, error)
	InsertMember(m Member) (id int, err error)
	UpdateAll(ids []int64, active int) error
	UpdateMember(m Member) error
	Count(req args.ArgsParser) (result int, err error)
	GetIDsByNickname(params GetMembersKeywordsArgs) (result []Stunt, err error)
}

type GetMemberArgs struct {
	IDType string
	ID     string
	Mode   string
}

func (m *GetMemberArgs) parseRestricts() (restricts string, values []interface{}) {
	where := make([]string, 0)

	if m.ID != "" && m.IDType != "" {
		where = append(where, fmt.Sprintf("%s = ?", m.IDType))
		values = append(values, m.ID)
	}

	if m.Mode != "" {
		where = append(where, "register_mode = ?")
		values = append(values, m.Mode)
	}

	if len(where) > 1 {
		restricts = strings.Join(where, " AND ")
	} else if len(where) == 1 {
		restricts = where[0]
	}
	return restricts, values
}

type GetMembersArgs struct {
	MaxResult    uint8            `form:"max_result"`
	Page         uint16           `form:"page"`
	Sorting      string           `form:"sort"`
	CustomEditor bool             `form:"custom_editor"`
	Role         *int64           `form:"role"`
	Active       map[string][]int `form:"active"`
	IDs          []string         `form:"ids"`
	UUIDs        []string         `form:"uuids"`
	Total        bool             `form:"total"`
}

func (m *GetMembersArgs) SetDefault() {
	m.MaxResult = 20
	m.Page = 1
	m.Sorting = "-updated_at"
}

func (m *GetMembersArgs) DefaultActive() {
	// m.Active = map[string][]int{"$nin": []int{int(MemberStatus["delete"].(float64))}}
	m.Active = map[string][]int{"$nin": []int{config.Config.Models.Members["delete"]}}
}

func (m *GetMembersArgs) ParseCountQuery() (query string, values []interface{}) {

	if !m.anyFilter() {
		return `SELECT COUNT(*) FROM members`, values
	} else {
		restricts, values := m.parseRestricts()
		return fmt.Sprintf(`SELECT COUNT(*) FROM (SELECT id FROM members WHERE %s) AS subquery`, restricts), values
	}
}

func (m *GetMembersArgs) anyFilter() bool {
	return m.Active != nil || m.CustomEditor == true
}

func (m *GetMembersArgs) parseRestricts() (restricts string, values []interface{}) {
	where := make([]string, 0)

	if m.CustomEditor {
		where = append(where, "custom_editor = ?")
		values = append(values, m.CustomEditor)
	}
	if m.Active != nil {
		for k, v := range m.Active {
			where = append(where, fmt.Sprintf("%s %s (?)", "members.active", rrsql.OperatorHelper(k)))
			values = append(values, v)
		}
	}
	if m.Role != nil {
		where = append(where, "role = ?")
		values = append(values, *m.Role)
	}
	if len(m.IDs) > 0 {
		a := make([]string, len(m.IDs))
		for i := range a {
			a[i] = "?"
		}
		where = append(where, fmt.Sprintf("members.id IN (%s)", strings.Join(a, ", ")))
		for i := range m.IDs {
			values = append(values, m.IDs[i])
		}
	}
	if len(m.UUIDs) > 0 {
		a := make([]string, len(m.UUIDs))
		for i := range a {
			a[i] = "?"
		}
		where = append(where, fmt.Sprintf("members.uuid IN (%s)", strings.Join(a, ", ")))
		for i := range m.UUIDs {
			values = append(values, m.UUIDs[i])
		}
	}
	if len(where) > 1 {
		restricts = strings.Join(where, " AND ")
	} else if len(where) == 1 {
		restricts = where[0]
	}
	return restricts, values
}

func (m *GetMembersArgs) parseLimit() (restricts string, values []interface{}) {

	if m.Sorting != "" {
		restricts = fmt.Sprintf("%s ORDER BY %s", restricts, rrsql.OrderByHelper(m.Sorting))
	}

	if m.MaxResult > 0 {
		restricts = fmt.Sprintf("%s LIMIT ?", restricts)
		values = append(values, m.MaxResult)
		if m.Page > 0 {
			restricts = fmt.Sprintf("%s OFFSET ?", restricts)
			values = append(values, (m.Page-1)*uint16(m.MaxResult))
		}
	}
	return restricts, values
}

type FilterMemberArgs struct {
	args.FilterArgs
	Fields rrsql.Sqlfields
}

func (p FilterMemberArgs) ParseQuery() (query string, values []interface{}) {
	return p.parse(false)
}
func (p FilterMemberArgs) ParseCountQuery() (query string, values []interface{}) {
	return p.parse(true)
}

func (m *FilterMemberArgs) parse(doCount bool) (query string, values []interface{}) {
	selectedFields := m.Fields.GetFields(`%s "%s"`)

	restricts, restrictVals := m.parseFilterRestricts()
	limit, limitVals := m.parseLimit()
	values = append(values, restrictVals...)
	values = append(values, limitVals...)

	if doCount {
		query = fmt.Sprintf(`
		SELECT %s FROM members %s `,
			"COUNT(member_id)",
			restricts,
		)
		values = restrictVals
	} else {
		query = fmt.Sprintf(`
		SELECT %s FROM members %s `,
			selectedFields,
			restricts+limit,
		)
	}

	return query, values
}
func (m *FilterMemberArgs) parseFilterRestricts() (restrictString string, values []interface{}) {
	restricts := make([]string, 0)

	if m.ID != 0 {
		restricts = append(restricts, `CAST(members.id as CHAR) LIKE ?`)
		values = append(values, fmt.Sprintf("%s%d%s", "%", m.ID, "%"))
	}
	if m.Mail != "" {
		restricts = append(restricts, `CAST(members.mail as CHAR) LIKE ?`)
		values = append(values, fmt.Sprintf("%s%s%s", "%", m.Mail, "%"))
	}
	if m.Nickname != "" {
		restricts = append(restricts, `CAST(members.nickname as CHAR) LIKE ?`)
		values = append(values, fmt.Sprintf("%s%s%s", "%", m.Nickname, "%"))
	}
	if len(m.CreatedAt) != 0 {
		if v, ok := m.CreatedAt["$gt"]; ok {
			restricts = append(restricts, "members.created_at >= ?")
			values = append(values, v)
		}
		if v, ok := m.CreatedAt["$lt"]; ok {
			restricts = append(restricts, "members.created_at <= ?")
			values = append(values, v)
		}
	}
	if len(m.UpdatedAt) != 0 {
		if v, ok := m.UpdatedAt["$gt"]; ok {
			restricts = append(restricts, "members.updated_at >= ?")
			values = append(values, v)
		}
		if v, ok := m.UpdatedAt["$lt"]; ok {
			restricts = append(restricts, "members.updated_at <= ?")
			values = append(values, v)
		}
	}
	if len(restricts) > 1 {
		restrictString = fmt.Sprintf("WHERE %s", strings.Join(restricts, " AND "))
	} else if len(restricts) == 1 {
		restrictString = fmt.Sprintf("WHERE %s", restricts[0])
	}
	return restrictString, values
}

func (m *FilterMemberArgs) parseLimit() (restricts string, values []interface{}) {

	if m.Sorting != "" {
		restricts = fmt.Sprintf("%s ORDER BY %s", restricts, rrsql.OrderByHelper(m.Sorting))
	}

	if m.MaxResult > 0 {
		restricts = fmt.Sprintf("%s LIMIT ?", restricts)
		values = append(values, m.MaxResult)
		if m.Page > 0 {
			restricts = fmt.Sprintf("%s OFFSET ?", restricts)
			values = append(values, (m.Page-1)*m.MaxResult)
		}
	}
	return restricts, values
}

type GetMembersKeywordsArgs struct {
	Keywords string `form:"keyword"`
	Roles    map[string][]int
	Fields   rrsql.Sqlfields
}

func (a *GetMembersKeywordsArgs) Validate() (err error) {
	// Validate keyword
	if a.Keywords == "" {
		return errors.New("Invalid keyword")
	}
	// Validate field
	validFields := rrsql.GetStructDBTags("full", Stunt{})

CheckEachFieldLoop:
	for _, f := range a.Fields {
		for _, F := range validFields {
			if f == F {
				continue CheckEachFieldLoop
			}
		}
		return fmt.Errorf("Invalid fields: %s", f)
	}
	var containfield = func(field string) bool {
		for _, f := range a.Fields {
			if f == field {
				return true
			}
		}
		return false
	}
	// Set default fields id & nickname
	for _, fs := range []string{"id", "nickname"} {
		if !containfield(fs) {
			a.Fields = append(a.Fields, fs)
		}
	}
	return err
}

func (a *memberAPI) GetMembers(req *GetMembersArgs) (result []Member, err error) {

	restricts, values := req.parseRestricts()
	query := fmt.Sprintf(`SELECT * FROM members where %s `, restricts)

	query, args, err := sqlx.In(query, values...)
	if err != nil {
		return []Member{}, err
	}
	query = rrsql.DB.Rebind(query)
	query = query + fmt.Sprintf(`ORDER BY %s LIMIT ? OFFSET ?`, rrsql.OrderByHelper(req.Sorting))
	args = append(args, req.MaxResult, (req.Page-1)*uint16(req.MaxResult))
	err = rrsql.DB.Select(&result, query, args...)
	if err != nil {
		return []Member{}, err
	}
	if len(result) == 0 {
		return []Member{}, nil
	}
	return result, err
}

func (a *memberAPI) GetMember(req GetMemberArgs) (Member, error) {
	member := Member{}
	restricts, values := req.parseRestricts()
	query := fmt.Sprintf("SELECT * FROM members where %s", restricts)

	err := rrsql.DB.QueryRowx(query, values...).StructScan(&member)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("User Not Found")
		member = Member{}
	case err != nil:
		log.Fatal(err)
		member = Member{}
	default:
		err = nil
	}
	return member, err
}

func (a *memberAPI) FilterMembers(args *FilterMemberArgs) (result []Stunt, err error) {
	query, values := args.ParseQuery()

	rows, err := rrsql.DB.Queryx(query, values...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var asset Stunt
		if err = rows.StructScan(&asset); err != nil {
			return result, err
		}
		result = append(result, asset)
	}
	return result, nil
}

func (a *memberAPI) InsertMember(m Member) (id int, err error) {
	existedID := 0
	err = rrsql.DB.Get(&existedID, `SELECT id FROM members WHERE id=? OR member_id=? LIMIT 1;`, m.ID, m.MemberID)
	if err != nil {
		if err != sql.ErrNoRows {
			return 0, err
		}
	}
	if existedID != 0 {
		return 0, errors.New("Duplicate entry")
	}

	tags := rrsql.GetStructDBTags("partial", m)
	query := fmt.Sprintf(`INSERT INTO members (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))
	result, err := rrsql.DB.NamedExec(query, m)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return 0, errors.New("Duplicate entry")
		}
		return 0, err
	}
	rowCnt, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rowCnt > 1 {
		return 0, errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return 0, errors.New("No Row Inserted")
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last inserted ID when insert a member: %v", err)
		return 0, err
	}
	return int(lastID), nil
}

func (a *memberAPI) UpdateMember(m Member) error {
	// query, _ := rrsql.GenerateSQLStmt("partial_update", "members", m)
	tags := rrsql.GetStructDBTags("partial", m)
	fields := rrsql.MakeFieldString("update", `%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE members SET %s WHERE id = :id`, strings.Join(fields, ", "))
	result, err := rrsql.DB.NamedExec(query, m)

	if err != nil {
		log.Fatal(err)
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("User Not Found")
	}
	return nil
}

func (a *memberAPI) DeleteMember(idType string, id string) error {

	// result, err := rrsql.DB.Exec(fmt.Sprintf("UPDATE members SET active = %d WHERE %s = ?", int(MemberStatus["delete"].(float64)), idType), id)
	result, err := rrsql.DB.Exec(fmt.Sprintf("UPDATE members SET active = %d WHERE %s = ?", config.Config.Models.Members["delete"], idType), id)
	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("User Not Found")
	}
	return err
}

func (a *memberAPI) UpdateAll(ids []int64, active int) (err error) {
	prep := fmt.Sprintf("UPDATE members SET active = %d WHERE id IN (?);", active)
	query, args, err := sqlx.In(prep, ids)
	if err != nil {
		return err
	}
	query = rrsql.DB.Rebind(query)
	result, err := rrsql.DB.Exec(query, args...)
	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > int64(len(ids)) {
		return errors.New("More Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Members Not Found")
	}
	return err
}

func (a *memberAPI) Count(req args.ArgsParser) (result int, err error) {

	query, values := req.ParseCountQuery()

	query, args, err := sqlx.In(query, values...)
	if err != nil {
		return 0, err
	}
	query = rrsql.DB.Rebind(query)
	count, err := rrsql.DB.Queryx(query, args...)
	if err != nil {
		return 0, err
	}
	for count.Next() {
		if err = count.Scan(&result); err != nil {
			return 0, err
		}
	}

	return result, err
}

// GetMembersByNickname select nickname and uuid from active members only
// when their nickname fits certain keyword
func (a *memberAPI) GetIDsByNickname(params GetMembersKeywordsArgs) (result []Stunt, err error) {

	query := fmt.Sprintf(`SELECT %s FROM members WHERE active = ? AND nickname LIKE ?`, strings.Join(params.Fields, ", "))
	values := []interface{}{}
	values = append(values, config.Config.Models.Members["active"], params.Keywords+"%")

	if len(params.Roles) != 0 {
		for k, v := range params.Roles {
			query = fmt.Sprintf("%s AND %s %s (?)", query, "members.role", rrsql.OperatorHelper(k))
			values = append(values, v)
		}
	}
	query, values, err = sqlx.In(query, values...)
	query = rrsql.DB.Rebind(query)
	if err = rrsql.DB.Select(&result, query, values...); err != nil {
		return []Stunt{}, err
	}
	return result, err
}
