package libansible

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

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

type State bool

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

type Response struct {
	Msg        string     `json:"msg"`
	Changed    bool       `json:"changed"`
	Failed     bool       `json:"failed"`
	Stdout     string     `json:"stdout"`
	Invocation Invocation `json:"invocation"`
	Diff       Diff       `json:"diff"`
}

type Invocation struct {
	ModuleArgs interface{} `json:"module_args,omitempty"`
}

type Diff struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

func ExitJson(responseBody Response) {
	returnResponse(responseBody)
}

func FailJson(responseBody Response, err error) {
	responseBody.Msg = err.Error()
	responseBody.Failed = true
	returnResponse(responseBody)
}

func returnResponse(responseBody Response) {
	response, err := json.Marshal(responseBody)
	if err != nil {
		response, _ = json.Marshal(Response{Msg: "Invalid response object"})
	}
	fmt.Println(string(response))
	if responseBody.Failed {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func ReadInput() []byte {
	var response Response
	if len(os.Args) != 2 {
		FailJson(response, errors.New("No argument file provided"))
	}
	input, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		FailJson(response, fmt.Errorf("Could not read configuration file: %s", err))
	}
	return input
}
