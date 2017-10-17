package amper

import (
	"encoding/json"
)

const IAMPolicyVersion = "2012-10-17"

type IAMPolicyDocRaw struct {
	Version string `json:",omitempty"`
	Id      string `json:",omitempty"`

	Statements []*IAMPolicyStatement `json:"Statement"`
}

type IAMPolicyDoc IAMPolicyDocRaw

func (d IAMPolicyDoc) MarshalJSON() ([]byte, error) {
	if d.Statements == nil {
		d.Statements = make([]*IAMPolicyStatement, 0)
	}

	if d.Version == "" {
		d.Version = IAMPolicyVersion
	}

	return json.Marshal(IAMPolicyDocRaw(d))
}

type IAMPolicyStatement struct {
	Sid          string
	Effect       string     `json:",omitempty"`
	Actions      StringList `json:"Action,omitempty"`
	NotActions   StringList `json:"NotAction,omitempty"`
	Resources    StringList `json:"Resource,omitempty"`
	NotResources StringList `json:"NotResource,omitempty"`

	Principals    map[string]StringList `json:"Principal,omitempty"`
	NotPrincipals map[string]StringList `json:"NotPrincipal,omitempty"`

	Conditions map[string]map[string]StringList `json:"Condition,omitempty"`

	size int `json:"-"`
}

func jsonSize(a interface{}) int {
	data, err := json.Marshal(a)

	if err != nil {
		return -1
	}

	return len(data)
}

func (s *IAMPolicyStatement) Size() int {
	return jsonSize(s)
}

func (s *IAMPolicyDoc) Size() int {
	return jsonSize(s)
}

type StringList []string

func (l StringList) MarshalJSON() ([]byte, error) {
	if len(l) == 1 {
		return json.Marshal(l[0])
	}

	return json.Marshal([]string(l))
}

func (p *StringList) UnmarshalJSON(data []byte) (err error) {
	*p = make([]string, 1)

	if err = json.Unmarshal(data, &(*p)[0]); err == nil {
		return
	}

	if _, ok := err.(*json.UnmarshalTypeError); !ok {
		return err
	}

	return json.Unmarshal(data, (*[]string)(p))
}

type ByPolicySize []*IAMPolicyStatement

func (a ByPolicySize) Len() int           { return len(a) }
func (a ByPolicySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPolicySize) Less(i, j int) bool { return a[i].size > a[j].size }
