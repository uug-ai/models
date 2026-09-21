package models

import (
	"bytes"
	"encoding/json"
	"reflect"
)

// These defined types retain the model fields and tags without their JSON methods.
type monitorStageJSON MonitorStage
type pipelineUserJSON User
type pipelineUser User

// MarshalJSON excludes unused user audit history from pipeline transport only.
func (m MonitorStage) MarshalJSON() ([]byte, error) {
	// Check the master chain before invoking nested custom marshalers, which
	// would otherwise restart encoding/json's cycle detection at each level.
	seen := make(map[*User]bool)
	for user := &m.User; user != nil; user = user.Master {
		if seen[user] {
			return nil, &json.UnsupportedValueError{
				Value: reflect.ValueOf(user),
				Str:   "encountered a cycle via User.Master",
			}
		}
		seen[user] = true
	}
	return json.Marshal(struct {
		*monitorStageJSON
		User *pipelineUser `json:"user,omitempty"`
	}{
		monitorStageJSON: (*monitorStageJSON)(&m),
		User:             (*pipelineUser)(&m.User),
	})
}

// UnmarshalJSON accepts both historical audit arrays and current audit objects
// before queue handlers run. Other fields retain encoding/json's merge behavior.
func (m *MonitorStage) UnmarshalJSON(data []byte) error {
	wire := struct {
		*monitorStageJSON
		User pipelineUser `json:"user,omitempty"`
	}{
		monitorStageJSON: (*monitorStageJSON)(m),
		User:             pipelineUser(m.User),
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	m.User = User(wire.User)
	return nil
}

func (u pipelineUser) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		*pipelineUserJSON
		Audit  json.RawMessage `json:"audit,omitempty"`
		Master *pipelineUser   `json:"master"`
	}{
		pipelineUserJSON: (*pipelineUserJSON)(&u),
		Master:           (*pipelineUser)(u.Master),
	})
}

func (u *pipelineUser) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil
	}
	wire := struct {
		*pipelineUserJSON
		Audit  json.RawMessage `json:"audit,omitempty"`
		Master *pipelineUser   `json:"master"`
	}{
		pipelineUserJSON: (*pipelineUserJSON)(u),
		Master:           (*pipelineUser)(u.Master),
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	u.Audit = Audit{}
	u.Master = (*User)(wire.Master)
	return nil
}
