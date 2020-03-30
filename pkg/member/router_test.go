package member

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful-member/config"
	"github.com/readr-media/readr-restful-member/internal/args"
	"github.com/readr-media/readr-restful-member/internal/rrsql"
	tc "github.com/readr-media/readr-restful-member/internal/test"
	"github.com/readr-media/readr-restful-member/internal/utils"
)

type mockMemberAPI struct{}

// Declare a backup struct for member test data
var mockMembers = []Member{
	Member{
		ID:           1,
		MemberID:     "superman@mirrormedia.mg",
		UUID:         "3d64e480-3e30-11e8-b94b-cfe922eb374f",
		Nickname:     rrsql.NullString{String: "readr", Valid: true},
		Active:       rrsql.NullInt{Int: 1, Valid: true},
		UpdatedAt:    rrsql.NullTime{Time: time.Date(2017, 6, 8, 16, 27, 52, 0, time.UTC), Valid: true},
		Mail:         rrsql.NullString{String: "superman@mirrormedia.mg", Valid: true},
		CustomEditor: rrsql.NullBool{Bool: true, Valid: true},
		Role:         rrsql.NullInt{Int: 9, Valid: true},
		Points:       rrsql.NullInt{Int: 0, Valid: true},
	},
	Member{
		ID:        2,
		MemberID:  "test6743@test.test",
		UUID:      "3d651126-3e30-11e8-b94b-cfe922eb374f",
		Nickname:  rrsql.NullString{String: "yeahyeahyeah", Valid: true},
		Active:    rrsql.NullInt{Int: 0, Valid: true},
		Birthday:  rrsql.NullTime{Time: time.Date(2001, 1, 3, 0, 0, 0, 0, time.UTC), Valid: true},
		UpdatedAt: rrsql.NullTime{Time: time.Date(2017, 11, 11, 23, 11, 37, 0, time.UTC), Valid: true},
		Mail:      rrsql.NullString{String: "Lulu_Brakus@yahoo.com", Valid: true},
		Role:      rrsql.NullInt{Int: 3, Valid: true},
		Points:    rrsql.NullInt{Int: 0, Valid: true},
	},
	Member{
		ID:        3,
		MemberID:  "Barney.Corwin@hotmail.com",
		UUID:      "3d6512e8-3e30-11e8-b94b-cfe922eb374f",
		Nickname:  rrsql.NullString{String: "reader", Valid: true},
		Active:    rrsql.NullInt{Int: 0, Valid: true},
		Gender:    rrsql.NullString{String: "M", Valid: true},
		UpdatedAt: rrsql.NullTime{Time: time.Date(2017, 1, 3, 19, 32, 37, 0, time.UTC), Valid: true},
		Birthday:  rrsql.NullTime{Time: time.Date(1939, 11, 9, 0, 0, 0, 0, time.UTC), Valid: true},
		Mail:      rrsql.NullString{String: "Barney.Corwin@hotmail.com", Valid: true},
		Role:      rrsql.NullInt{Int: 1, Valid: true},
		Points:    rrsql.NullInt{Int: 0, Valid: true},
	},
}

var mockMemberDS = []Member{}

func (a *mockMemberAPI) GetMembers(req *GetMembersArgs) (result []Member, err error) {

	if req.CustomEditor == true {
		result = []Member{mockMemberDS[0]}
		err = nil
		return result, err
	}

	if req.Role != nil {
		result = []Member{mockMemberDS[2]}
		err = nil
		return result, err
	}
	if len(req.Active) > 1 {
		return []Member{}, errors.New("Too many active lists")
	}
	for k, v := range req.Active {
		if k == "$nin" && reflect.DeepEqual(v, []int{0, -1}) {
			return []Member{mockMemberDS[0]}, nil
		} else if k == "$nin" && reflect.DeepEqual(v, []int{-1, 0, 1}) {
			return []Member{}, nil
		} else if reflect.DeepEqual(v, []int{-3, 0, 1}) {
			return []Member{}, errors.New("Not all active elements are valid")
		} else if reflect.DeepEqual(v, []int{3, 4}) {
			return []Member{}, errors.New("No valid active request")
		}
	}

	result = make([]Member, len(mockMemberDS))
	copy(result, mockMemberDS)
	switch req.Sorting {
	case "updated_at":
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].UpdatedAt.Before(result[j].UpdatedAt)
		})
		err = nil
	case "-updated_at":
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].UpdatedAt.After(result[j].UpdatedAt)
		})
		err = nil
	}
	if req.MaxResult == 2 {
		result = result[0:2]
	}
	return result, err
}

func (a *mockMemberAPI) GetMember(req GetMemberArgs) (result Member, err error) {
	intID, _ := strconv.Atoi(req.ID)
	for _, value := range mockMemberDS {
		if req.IDType == "id" && value.ID == int64(intID) {
			return value, nil
		} else if req.IDType == "member_id" && value.MemberID == req.ID {
			return value, nil
		} else if req.IDType == "mail" && req.ID == "registerdupeuser@mirrormedia.mg" {
			return Member{RegisterMode: rrsql.NullString{"ordinary", true}}, nil
		}
	}
	err = errors.New("User Not Found")
	return result, err
}

func (a *mockMemberAPI) FilterMembers(args *FilterMemberArgs) (result []Stunt, err error) {
	return result, nil
}

func (a *mockMemberAPI) InsertMember(m Member) (id int, err error) {
	for _, member := range mockMemberDS {
		if member.MemberID == m.MemberID {
			return 0, errors.New("Duplicate entry")
		}
	}
	m.ID = int64(len(mockMemberDS) + 1)
	mockMemberDS = append(mockMemberDS, m)
	err = nil
	return int(m.ID), err
}
func (a *mockMemberAPI) UpdateMember(m Member) error {

	err := errors.New("User Not Found")
	for index, member := range mockMemberDS {
		if member.ID == m.ID {
			mockMemberDS[index] = m
			err = nil
		}
	}
	return err
}

func (a *mockMemberAPI) DeleteMember(idType string, id string) error {

	err := errors.New("User Not Found")
	intID, _ := strconv.Atoi(id)
	for index, value := range mockMemberDS {
		if int64(intID) == value.ID {
			// mockMemberDS[index].Active = rrsql.NullInt{Int: int64(MemberStatus["delete"].(float64)), Valid: true}
			mockMemberDS[index].Active = rrsql.NullInt{Int: int64(config.Config.Models.Members["delete"]), Valid: true}
			return nil
		}
	}
	return err
}

func (a *mockMemberAPI) UpdateAll(ids []int64, active int) (err error) {

	result := make([]int, 0)
	for _, value := range ids {
		for i, v := range mockMemberDS {
			if v.ID == value {
				mockMemberDS[i].Active = rrsql.NullInt{Int: int64(active), Valid: true}
				result = append(result, i)
			}
		}
	}
	if len(result) == 0 {
		err = errors.New("Members Not Found")
		return err
	}
	return err
}

func (a *mockMemberAPI) Count(req args.ArgsParser) (result int, err error) {
	query, _ := req.ParseCountQuery()
	result = 0
	err = errors.New("Members Not Found")
	if strings.Contains(query, "custom_editor") {
		return 1, nil
	}
	if strings.Contains(query, "role") {
		return 1, nil
	}
	if strings.Contains(query, "members.active NOT IN (?))") {
		return 4, nil
	}
	if strings.Contains(query, "members.active IN (?))") {
		return 2, nil
	}
	if strings.Contains(query, "CAST(members.mail as CHAR) LIKE") {
		return 1, nil
	}
	return result, err
}

func (a *mockMemberAPI) GetIDsByNickname(params GetMembersKeywordsArgs) (result []Stunt, err error) {
	if params.Keywords == "readr" {
		if params.Roles != nil {
			result = append(result, Stunt{ID: &(mockMemberDS[0].ID), Nickname: &(mockMemberDS[0].Nickname)})
			return result, err
		}
		result = append(result, Stunt{ID: &(mockMemberDS[0].ID), Nickname: &(mockMemberDS[0].Nickname)})
		return result, err

	}
	return result, err
}

func TestMain(m *testing.M) {

	_, err := config.LoadConfig("../../config/main.json")
	if err != nil {
		panic(fmt.Errorf("Invalid application configuration: %s", err))
	}

	tc.SetRoutes(&Router)
	MemberAPI = new(mockMemberAPI)
	os.Exit(m.Run())
}

func TestRouteMembers(t *testing.T) {

	if os.Getenv("db_driver") == "mysql" {
		_, _ = rrsql.DB.Exec("truncate table members;")
	} else {
		mockMemberDS = []Member{}
	}

	for _, m := range mockMembers {
		_, err := MemberAPI.InsertMember(m)
		if err != nil {
			log.Printf("Init member test fail %s", err.Error())
		}
	}

	asserter := func(resp string, tc tc.GenericTestcase, t *testing.T) {
		type response struct {
			Items []Member `json:"_items"`
		}

		var Response response
		var expected []Member = tc.Resp.([]Member)

		err := json.Unmarshal([]byte(resp), &Response)
		if err != nil {
			t.Errorf("%s, Unexpected result body: %v", resp, err.Error())
		}

		if len(Response.Items) != len(expected) {
			t.Errorf("%s expect member length to be %v but get %v", tc.Name, len(expected), len(Response.Items))
		}

		for i, resp := range Response.Items {
			exp := expected[i]
			if resp.ID == exp.ID &&
				resp.Active == exp.Active &&
				resp.UpdatedAt == exp.UpdatedAt &&
				resp.Role == exp.Role {
				continue
			}
			t.Errorf("%s, expect to get %v, but %v ", tc.Name, exp, resp)
		}
	}

	t.Run("GetMembers", func(t *testing.T) {
		for _, testcase := range []tc.GenericTestcase{
			tc.GenericTestcase{"UpdatedAtDescending", "GET", "/members", ``, http.StatusOK, []Member{mockMembers[1], mockMembers[0], mockMembers[2]}},
			tc.GenericTestcase{"UpdatedAtAscending", "GET", "/members?sort=updated_at", ``, http.StatusOK, []Member{mockMembers[2], mockMembers[0], mockMembers[1]}},
			tc.GenericTestcase{"max_result", "GET", "/members?max_result=2", ``, http.StatusOK, []Member{mockMembers[1], mockMembers[0]}},
			tc.GenericTestcase{"ActiveFilter", "GET", `/members?active={"$nin":[0,-1]}`, ``, http.StatusOK, []Member{mockMembers[0]}},
			tc.GenericTestcase{"CustomEditorFilter", "GET", `/members?custom_editor=true`, ``, http.StatusOK, []Member{mockMembers[0]}},
			tc.GenericTestcase{"NoMatchMembers", "GET", `/members?active={"$nin":[-1,0,1]}`, ``, http.StatusOK, `{"_items":[]}`},
			tc.GenericTestcase{"MoreThanOneActive", "GET", `/members?active={"$nin":[1,0], "$in":[-1,3]}`, ``, http.StatusBadRequest, `{"Error":"Too many active lists"}`},
			tc.GenericTestcase{"NotEntirelyValidActive", "GET", `/members?active={"$in":[-3,0,1]}`, ``, http.StatusBadRequest, `{"Error":"Not all active elements are valid"}`},
			tc.GenericTestcase{"NoValidActive", "GET", `/members?active={"$nin":[3,4]}`, ``, http.StatusBadRequest, `{"Error":"No valid active request"}`},
			tc.GenericTestcase{"Role", "GET", `/members?role=1`, ``, http.StatusOK, []Member{mockMembers[2]}},
		} {
			tc.GenericDoTest(testcase, t, asserter)
		}
	})
	t.Run("GetMember", func(t *testing.T) {
		for _, testcase := range []tc.GenericTestcase{
			tc.GenericTestcase{"Current", "GET", "/member/1", ``, http.StatusOK, []Member{mockMembers[0]}},
			tc.GenericTestcase{"NotExisted", "GET", "/member/24601", ``, http.StatusNotFound, `{"Error":"User Not Found"}`},
			tc.GenericTestcase{"NotExisted", "GET", "/member/superman@mirrormedia.mg", ``, http.StatusOK, []Member{mockMembers[0]}},
		} {
			tc.GenericDoTest(testcase, t, asserter)
		}
	})
	t.Run("PostMember", func(t *testing.T) {
		for _, testcase := range []tc.GenericTestcase{
			tc.GenericTestcase{"New", "POST", "/member", `{"member_id":"spaceoddity", "name":"Major Tom", "mail":"spaceoddity"}`, http.StatusOK, `{"_items":{"last_id":4}}`},
			//tc.GenericTestcase{"EmptyPayload", "POST", "/member", `{}`, http.StatusBadRequest, `{"Error":"Invalid User"}`},
			//tc.GenericTestcase{"Existed", "POST", "/member", `{"id": 1, "member_id":"superman@mirrormedia.mg"}`, http.StatusBadRequest, `{"Error":"User Already Existed"}`},
		} {
			tc.GenericDoTest(testcase, t, asserter)
		}
	})
	t.Run("CountMembers", func(t *testing.T) {
		for _, testcase := range []tc.GenericTestcase{
			tc.GenericTestcase{"SimpleCount", "GET", "/members/count", ``, http.StatusOK, `{"_meta":{"total":4}}`},
			tc.GenericTestcase{"CountActive", "GET", `/members/count?active={"$in":[1,-1]}`, ``, http.StatusOK, `{"_meta":{"total":2}}`},
			tc.GenericTestcase{"CountCustomEditor", "GET", `/members/count?custom_editor=true`, ``, http.StatusOK, `{"_meta":{"total":1}}`},
			tc.GenericTestcase{"MoreThanOneActive", "GET", `/members/count?active={"$nin":[1,0], "$in":[-1,3]}`, ``, http.StatusBadRequest, `{"Error":"Too many active lists"}`},
			tc.GenericTestcase{"NotEntirelyValidActive", "GET", `/members/count?active={"$in":[-3,0,1]}`, ``, http.StatusBadRequest, `{"Error":"Not all active elements are valid"}`},
			tc.GenericTestcase{"NoValidActive", "GET", `/members/count?active={"$nin":[3,4]}`, ``, http.StatusBadRequest, `{"Error":"No valid active request"}`},
			tc.GenericTestcase{"Role", "GET", "/members/count?role=9", ``, http.StatusOK, `{"_meta":{"total":1}}`},
		} {
			tc.GenericDoTest(testcase, t, asserter)
		}
	})
	t.Run("KeyNickname", func(t *testing.T) {
		for _, testcase := range []tc.GenericTestcase{
			tc.GenericTestcase{"Keyword", "GET", `/members/nickname?keyword=readr`, ``, http.StatusOK, `{"_items":[{"id":1,"nickname":"readr"}]}`},
			tc.GenericTestcase{"KeywordAndRoles", "GET", `/members/nickname?keyword=readr&roles={"$in":[3,9]}`, ``, http.StatusOK, `{"_items":[{"id":1,"nickname":"readr"}]}`},
			tc.GenericTestcase{"InvalidKeyword", "GET", `/members/nickname`, ``, http.StatusBadRequest, `{"Error":"Invalid keyword"}`},
			tc.GenericTestcase{"InvalidFields", "GET", `/members/nickname?keyword=readr&fields=["line"]`, ``, http.StatusBadRequest, `{"Error":"Invalid fields: line"}`},
		} {
			tc.GenericDoTest(testcase, t, asserter)
		}
	})
	t.Run("PutMember", func(t *testing.T) {
		for _, testcase := range []tc.GenericTestcase{
			tc.GenericTestcase{"New", "PUT", "/member", `{"id":1, "name":"Clark Kent"}`, http.StatusOK, ``},
			tc.GenericTestcase{"UpdateDailyPush", "PUT", "/member", `{"id":1, "daily_push":true}`, http.StatusOK, ``},
			tc.GenericTestcase{"NotExisted", "PUT", "/member", `{"id":24601, "name":"spaceoddity"}`, http.StatusBadRequest, `{"Error":"User Not Found"}`},
		} {
			tc.GenericDoTest(testcase, t, asserter)
		}
	})

	t.Run("DeleteMembers", func(t *testing.T) {
		for _, testcase := range []tc.GenericTestcase{
			tc.GenericTestcase{"Delete", "DELETE", `/members?ids=[2]`, ``, http.StatusOK, ``},
			tc.GenericTestcase{"Empty", "DELETE", `/members?ids=[]`, ``, http.StatusBadRequest, `{"Error":"ID List Empty"}`},
			//tc.GenericTestcase{"InvalidQueryArray", "DELETE", `/members?ids=["superman@mirrormedia.mg,"test6743"]`, ``, http.StatusBadRequest, `{"Error":"invalid character 't' after array element"}`},
			tc.GenericTestcase{"NotFound", "DELETE", `/members?ids=[24601, 24602]`, ``, http.StatusBadRequest, `{"Error":"Members Not Found"}`},
		} {
			tc.GenericDoTest(testcase, t, asserter)
		}
	})
	t.Run("DeleteMember", func(t *testing.T) {
		for _, testcase := range []tc.GenericTestcase{
			tc.GenericTestcase{"Current", "DELETE", `/member/3`, ``, http.StatusOK, ``},
			tc.GenericTestcase{"NonExisted", "DELETE", `/member/24601`, ``, http.StatusNotFound, `{"Error":"User Not Found"}`},
		} {
			tc.GenericDoTest(testcase, t, asserter)
		}
	})
	t.Run("ActivateMultipleMembers", func(t *testing.T) {
		for _, testcase := range []tc.GenericTestcase{
			tc.GenericTestcase{"CurrentMembers", "PUT", `/members`, `{"ids": [1,2]}`, http.StatusOK, ``},
			tc.GenericTestcase{"NotFound", "PUT", `/members`, `{"ids": [24601, 24602]}`, http.StatusNotFound, `{"Error":"Members Not Found"}`},
			tc.GenericTestcase{"InvalidPayload", "PUT", `/members`, `{}`, http.StatusBadRequest, `{"Error":"Invalid Request Body"}`},
		} {
			tc.GenericDoTest(testcase, t, asserter)
		}
	})

}

func TestRouteMemberUpdatePassword(t *testing.T) {

	var r *gin.Engine
	r = gin.New()
	Router.SetRoutes(r)

	type ChangePWCaseIn struct {
		ID       string `json:"id,omitempty"`
		Password string `json:"password,omitempty"`
	}

	var TestRouteChangePWCases = []struct {
		name     string
		in       ChangePWCaseIn
		httpcode int
	}{
		{"ChangePWOK", ChangePWCaseIn{ID: "1", Password: "angrypug"}, http.StatusOK},
		{"ChangePWFail", ChangePWCaseIn{ID: "1"}, http.StatusBadRequest},
		{"ChangePWNoID", ChangePWCaseIn{Password: "angrypug"}, http.StatusBadRequest},
		{"ChangePWMemberNotFound", ChangePWCaseIn{ID: "24601", Password: "angrypug"}, http.StatusNotFound},
	}

	for _, testcase := range TestRouteChangePWCases {
		jsonStr, err := json.Marshal(&testcase.in)
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}
		req, _ := http.NewRequest("PUT", "/member/password", bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != testcase.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.httpcode, w.Code, testcase.name)
			t.Fail()
		}

		if w.Code == http.StatusOK {
			member, err := MemberAPI.GetMember(GetMemberArgs{
				ID:     testcase.in.ID,
				IDType: "id",
			})
			if err != nil {
				t.Errorf("Cannot get user after update PW, testcase %s", testcase.name)
				t.Fail()
			}

			newPW, err := utils.CryptGenHash(testcase.in.Password, member.Salt.String)
			switch {
			case err != nil:
				t.Errorf("Error when hashing password, testcase %s", testcase.name)
				t.Fail()
			case newPW != member.Password.String:
				t.Errorf("%v", member.ID)
				t.Errorf("Password update fail Want %v but get %v, testcase %s", newPW, member.Password.String, testcase.name)
				t.Fail()
			}
		}
	}
}
