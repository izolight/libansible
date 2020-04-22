package libansible

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

// AnsibleArgs contains the builtin arguments which Ansible passes when invoking a module
type AnsibleArgs struct {
	AnsibleCheckMode              bool        `json:"_ansible_check_mode,omitempty"`
	AnsibleNoLog                  bool        `json:"_ansible_no_log,omitempty"`
	AnsibleDebug                  bool        `json:"_ansible_debug,omitempty"`
	AnsibleDiff                   bool        `json:"_ansible_diff,omitempty"`
	AnsibleVerbosity              uint        `json:"_ansible_verbosity,omitempty"`
	AnsibleVersion                string      `json:"_ansible_version,omitempty"`
	AnsibleModuleName             string      `json:"_ansible_module_name,omitempty"`
	AnsibleSyslogFacility         string      `json:"_ansible_syslog_facility,omitempty"`
	AnsibleSELinuxSpecialFS       []string    `json:"_ansible_se_linux_special_fs,omitempty"`
	AnsibleStringConversionAction string      `json:"_ansible_string_conversion_action,omitempty"`
	AnsibleSocket                 interface{} `json:"_ansible_socket,omitempty"`
	AnsibleShellExecutable        string      `json:"_ansible_shell_executable,omitempty"`
	AnsibleKeepRemoteFiles        bool        `json:"_ansible_keep_remote_files,omitempty"`
	AnsibleTmpDir                 string      `json:"_ansible_tmp_dir,omitempty"`
	AnsibleRemoteTmp              string      `json:"_ansible_remote_tmp,omitempty"`
}

// State maps true/false to absent/present
type State bool

// MarshalJSON converts a json bool value to absent/present
func (s State) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	if bool(s) {
		buffer.WriteString("present")
	} else {
		buffer.WriteString("absent")
	}
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON converts absent/present to a json bool value
func (s *State) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	switch j {
	case "present":
		*s = true
		return nil
	case "absent":
		*s = false
		return nil
	}
	return fmt.Errorf("State should be absent or present, was %s", j)
}

// String is a slice of strings that can also be a single string when converting from json
type String []string

// MarshalJSON converts a String type either to a list or single string, depending on length
func (s String) MarshalJSON() ([]byte, error) {
	switch len(s) {
	case 0:
		return []byte(`""`), nil
	case 1:
		return []byte(`"` + s[0] + `"`), nil
	default:
		return json.Marshal([]string(s))
	}
}

// UnmarshalJSON converts a single string or a list of strings to a slice of strings
func (s *String) UnmarshalJSON(b []byte) error {
	var j interface{}
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	switch v := j.(type) {
	case nil:
		return nil
	case string:
		*s = []string{v}
		return nil
	case []interface{}:
		for _, i := range v {
			k, ok := i.(string)
			if !ok {
				return fmt.Errorf("List element was not a string, was %t", i)
			}
			*s = append(*s, k)
		}
		return nil
	}
	return fmt.Errorf("Input should be string or list of strings, was %t", j)
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

type Bool struct {
	Value bool
	IsSet bool
}

func (b *Bool) UnmarshalJSON(in []byte) error {
	var j interface{}
	err := json.Unmarshal(in, &j)
	if err != nil {
		return err
	}
	switch v := j.(type) {
	case nil:
		return nil
	case string:
		yes := []string{"Yes", "yes", "True", "true"}
		no := []string{"No", "no", "False", "false"}
		if contains(yes, v) {
			b.Value = true
			b.IsSet = true
		} else if contains(no, v) {
			b.Value = false
			b.IsSet = true
		} else if v == "" {
			b.IsSet = false
		} else {
			return fmt.Errorf("Invalid value, needs to be truthy or falsy: %s", v)
		}
	case bool:
		b.Value = v
		b.IsSet = true
	}
	return nil
}

func (b Bool) MarshalJSON() ([]byte, error) {
	return json.Marshal(bool(b.Value))
}

// Response contains the fields which Ansible expects as module output
type Response struct {
	Changed    bool       `json:"changed"`
	Failed     bool       `json:"failed"`
	Stdout     string     `json:"stdout,omitempty"`
	Stderr     string     `json:"stderr,omitempty"`
	Invocation Invocation `json:"invocation"`
	Diff       Diff       `json:"diff,omitempty"`
}

// Invocation contains the Module Arguments (might be more in the future)
type Invocation struct {
	ModuleArgs interface{} `json:"module_args,omitempty"`
}

// Diff is a simple structure that contains a string for the state before and after module execution
type Diff struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

// ExitJson returns the Response Object as Json
func ExitJson(responseBody Response, w io.Writer) error {
	return returnResponse(responseBody, w)
}

// FailJson sets the stderr field to the provided error and sets the failed state to true and returns the response as json
func FailJson(responseBody Response, err error, w io.Writer) error {
	responseBody.Stderr = err.Error()
	responseBody.Failed = true
	return returnResponse(responseBody, w)
}

// returnResponse takes a Response struct and outputs json to stdout and sets the exit code according to failed state
func returnResponse(responseBody Response, w io.Writer) error {
	response, err := json.Marshal(responseBody)
	if err != nil {
		response, _ = json.Marshal(Response{Stderr: "Invalid response object"})
	}
	fmt.Fprintln(w, string(response))
	if responseBody.Failed {
		return fmt.Errorf("module failed: %s", responseBody.Stderr)
	}
	return nil
}

// ReadInput checks if the only argument is the provided json file and reads the file
func ReadInput(args []string) []byte {
	var response Response
	if len(args) != 2 {
		FailJson(response, errors.New("No argument file provided"))
	}
	input, err := ioutil.ReadFile(args[1])
	if err != nil {
		FailJson(response, fmt.Errorf("Could not read configuration file: %s", err))
	}
	return input
}

type Playbook struct {
	Name        string                 `json:"name"`
	Connection  string                 `json:"connection"`
	Hosts       string                 `json:"hosts"`
	Tasks       json.RawMessage        `json:"tasks"`
	Environment map[string]interface{} `json:"environment"`
	Become      bool                   `json:"become"`
}
